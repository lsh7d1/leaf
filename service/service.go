package service

import (
	"context"
	"errors"
	"fmt"
	"leaf/dal/model"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const (
	maxStep = 100_0000
)

var (
	ErrTagNotFound   = errors.New("TAG NOT FOUND")
	ErrServiceClosed = errors.New("SERVICE IS CLOSED")

	errTwoSegExhausted = errors.New("TWO SEGMENT EXHAUSTED")
)

type SegmentService struct {
	repo    SegmentRepo
	barrier singleflight.Group
	buffers sync.Map // k--BizTag, v--*segmentBuffer

	reqSegmentCh chan struct{}
}

func (svc *SegmentService) ShowBuffers() {
	svc.buffers.Range(func(key, value any) bool {
		fmt.Printf("BizTag: %s, === sb: %#v\n", key.(string), value.(*segmentBuffer))
		return true
	})
}

func NewSegmentService(repo SegmentRepo) *SegmentService {
	svc := &SegmentService{
		repo:         repo,
		barrier:      singleflight.Group{},
		buffers:      sync.Map{},
		reqSegmentCh: make(chan struct{}),
	}

	_ = svc.loadTags()
	go svc.loadProc()

	return svc
}

func (svc *SegmentService) GetID(ctx context.Context, tag string) (int64, error) {
	value, ok := svc.buffers.Load(tag)
	if !ok {
		return -1, fmt.Errorf("[service:GetID] svc.buffers.Load failed, err: %v", ErrTagNotFound)
	}
	sb := value.(*segmentBuffer)
	if !sb.IsInitOk() {
		_, err, _ := svc.barrier.Do(tag, func() (interface{}, error) {
			if !sb.IsInitOk() {
				if err := svc.updateSegmentFromDB(ctx, tag, sb.GetCurrentSegment()); err != nil {
					return nil, err
				}
				sb.SetInitOk(true)
			}
			return nil, nil
		})
		if err != nil {
			return -1, nil
		}
	}
	return svc.getIdFromSegmentBuffer(ctx, sb)
}

const (
	defaultInterval = int64(time.Minute * 15)
	defaultMaxStep  = 100_0000
)

func (svc *SegmentService) updateSegmentFromDB(ctx context.Context, tag string, seg *segment) (err error) {
	var leaf *model.LeafAlloc
	sb := seg.GetBuffer()

	if !sb.IsInitOk() { // 第一次进，没有在db做初始化
		leaf, err = svc.repo.UpdateAndGetLeaf(ctx, tag)
		if err != nil {
			return fmt.Errorf("[service:updateSegmentFromDB] svc.repo.GetLeafByTag failed, err: %v", err)
		}
		sb.SetStep(leaf.Step)
		sb.SetMinStep(leaf.Step)
	} else { // 第二次进，时间戳默认为0
		interval := time.Now().Unix() - sb.GetUpdateTimeStamp()
		nextStep := sb.GetStep()

		// 按时间间隔动态调整Step
		if interval < defaultInterval { // 15min以内，step*2，最大为 defaultMaxStep
			if nextStep*2 <= defaultMaxStep {
				nextStep *= 2
			}
		} else if interval > 2*defaultInterval { // 30min以上，step/2，最小为 sb.minStep
			if nextStep/2 >= sb.GetMinStep() {
				nextStep /= 2
			}
		}
		leaf, err = svc.repo.UpdateAndGetLeafWithStep(ctx, tag, nextStep)
		if err != nil {
			return fmt.Errorf("[service:updateSegmentFromDB] svc.repo.UpdateAndGetLeafWithStep failed, err: %v", err)
		}
		sb.SetStep(nextStep)
		seg.SetStep(nextStep)
	}

	sb.SetUpdateTimeStamp(time.Now().Unix())

	// fmt.Printf("leaf: %#v\n -> sb: %#v\n", leaf, sb)

	value := leaf.MaxID - sb.GetStep()
	seg.SetValue(value)
	seg.SetMax(leaf.MaxID)

	return
}

func (svc *SegmentService) getIdFromSegmentBuffer(ctx context.Context, sb *segmentBuffer) (value int64, err error) {
	var seg *segment

	for {
		if value = func() int64 {
			sb.RLock()
			defer sb.RUnLock()

			seg = sb.GetCurrentSegment()
			if !sb.IsNextReady() &&
				seg.GetIdle() < int64(0.9*float64(seg.GetStep())) &&
				sb.GetThreadRunning().CompareAndSwap(false, true) {
				// 传入空context，防止cancel
				go svc.loadNextSegmentFromDB(context.TODO(), sb)
			}

			v := seg.GetAndInc()
			if v < seg.GetMax() {
				return v
			}
			return -1
		}(); value != -1 {
			return
		}

		// 等待异步协程发号
		waitAndSleep(sb)

		if value = func() int64 {
			sb.WLock()
			defer sb.WUnLock()

			seg = sb.GetCurrentSegment()
			v := seg.GetAndInc()
			if v < seg.GetMax() {
				return v
			}

			// 执行到这里说明其他协程没有切换segment
			// 并且当前号段发放完毕
			// 如果下一号段存在，直接切换
			if sb.IsNextReady() {
				sb.SwitchPos()
				sb.SetNextReady(false)
			} else {
				return -1
			}
			return -1
		}(); value != -1 {
			return
		}
	}
}

// waitAndSleep 让等待nextSegment的segmentBuffer强行等待10000轮次并休眠10ms
func waitAndSleep(sb *segmentBuffer) {
	round := 0
	for sb.GetThreadRunning().Load() && round < 10000 {
		round++
	}
	time.Sleep(time.Millisecond * 10)
}

func (svc *SegmentService) loadNextSegmentFromDB(ctx context.Context, sb *segmentBuffer) {
	nextSeg := sb.GetSegments()[sb.NextPos()]
	err := svc.updateSegmentFromDB(ctx, sb.GetTag(), nextSeg)
	if err != nil {
		sb.GetThreadRunning().Store(false)
		return
	}

	sb.WLock()
	defer sb.WUnLock()
	sb.SetNextReady(true)
	sb.GetThreadRunning().Store(false)

	return
}

func (svc *SegmentService) loadTags() error {
	bizTags, err := svc.repo.ListAllTags(context.TODO())
	if err != nil {
		return fmt.Errorf("[service:loadTags()] svc.repo.ListAllTags failed, err: %v", err)
	}
	if len(bizTags) == 0 {
		return nil
	}

	addTags, delTags := make([]string, 0), make([]string, 0)
	bufferTags := make(map[string]bool)
	svc.buffers.Range(func(k, v any) bool {
		bufferTags[k.(string)] = true
		return true
	})

	// db新加的addTags加入buffers，并实例化对应的segmentBuffer
	for _, tag := range bizTags {
		if _, ok := bufferTags[tag]; !ok {
			addTags = append(addTags, tag)
		}
	}
	for _, tag := range addTags {
		sb := newSegmentBuffer(tag)
		svc.buffers.Store(tag, sb)
		bufferTags[tag] = true
	}

	// buffers失效的tag删除
	for _, tag := range bizTags {
		bufferTags[tag] = false
	}
	for tag, v := range bufferTags {
		if v {
			delTags = append(delTags, tag)
		}
	}
	for _, tag := range delTags {
		svc.buffers.Delete(tag)
	}

	return nil
}

func (svc *SegmentService) loadProc() {
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-ticker.C:
			_ = svc.loadTags()
		case <-svc.reqSegmentCh:
			_ = svc.loadTags()
			ticker.Reset(time.Minute)
		}
	}
}
