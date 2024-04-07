package service

import (
	"sync"

	"go.uber.org/atomic"
)

type segmentBuffer struct {
	bizTag  string
	rwMutex *sync.RWMutex

	step    int64 // 运行过程中的step
	minStep int64 // db配置的step

	currentIndex     int          // 当前使用的segment的index
	nextReady        bool         // 下一segment是否可切换
	initOk           bool         // DB的数据号段是否初始化
	threadGetRunning *atomic.Bool // 准备下一segment的线程是否在运行

	updateTime int64

	segments []*segment
}

func newSegmentBuffer(bizTag string) *segmentBuffer {
	sb := &segmentBuffer{
		bizTag:           bizTag,
		rwMutex:          &sync.RWMutex{},
		currentIndex:     0,
		nextReady:        false,
		initOk:           false,
		threadGetRunning: atomic.NewBool(false),
		segments:         make([]*segment, 0, 2),
	}
	sb.segments = append(sb.segments, newSegment(sb), newSegment(sb))
	return sb
}

func (sb *segmentBuffer) GetTag() string {
	return sb.bizTag
}

func (sb *segmentBuffer) SetTag(key string) {
	sb.bizTag = key
}

func (sb *segmentBuffer) GetSegments() []*segment {
	return sb.segments
}

func (sb *segmentBuffer) GetAnotherSegment() *segment {
	return sb.segments[sb.NextPos()]
}

func (sb *segmentBuffer) GetCurrentSegment() *segment {
	return sb.segments[sb.currentIndex]
}

func (sb *segmentBuffer) GetCurrentIndex() int {
	return sb.currentIndex
}

func (sb *segmentBuffer) NextPos() int {
	return (sb.currentIndex + 1) & 1 // X % 2
}

func (sb *segmentBuffer) SwitchPos() {
	sb.currentIndex = sb.NextPos()
}

func (sb *segmentBuffer) IsInitOk() bool {
	return sb.initOk
}

func (sb *segmentBuffer) SetInitOk(initOk bool) {
	sb.initOk = initOk
}

func (sb *segmentBuffer) IsNextReady() bool {
	return sb.nextReady
}

func (sb *segmentBuffer) SetNextReady(nextReady bool) {
	sb.nextReady = nextReady
}

func (sb *segmentBuffer) GetThreadRunning() *atomic.Bool {
	return sb.threadGetRunning
}

func (sb *segmentBuffer) RLock() {
	sb.rwMutex.RLock()
}

func (sb *segmentBuffer) RUnLock() {
	sb.rwMutex.RUnlock()
}

func (sb *segmentBuffer) WLock() {
	sb.rwMutex.Lock()
}

func (sb *segmentBuffer) WUnLock() {
	sb.rwMutex.Unlock()
}

func (sb *segmentBuffer) GetStep() int64 {
	return sb.step
}

func (sb *segmentBuffer) SetStep(step int64) {
	sb.step = step
}

func (sb *segmentBuffer) GetMinStep() int64 {
	return sb.minStep
}

func (sb *segmentBuffer) SetMinStep(minStep int64) {
	sb.minStep = minStep
}

func (sb *segmentBuffer) GetUpdateTimeStamp() int64 {
	return sb.updateTime
}

func (sb *segmentBuffer) SetUpdateTimeStamp(ts int64) {
	sb.updateTime = ts
}
