package service

import (
	"go.uber.org/atomic"
)

// segment 表示一个ID缓冲区
type segment struct {
	value  *atomic.Int64  // 发放的id
	maxID  int64          // 当前号段最大id
	step   int64          // 步长 TODO: 暂不允许设定
	buffer *segmentBuffer // 当前号段所属segmentBuffer
}

// newSegment 创建一个新的Buffer实例
func newSegment(buffer *segmentBuffer) *segment {
	return &segment{
		value:  atomic.NewInt64(0),
		buffer: buffer,
	}
}

func (s *segment) GetAndInc() int64 {
	return s.value.Inc() - 1
}

func (s *segment) GetValue() int64 {
	return s.value.Load()
}

func (s *segment) SetValue(value int64) {
	s.value.Store(value)
}

func (s *segment) GetMax() int64 {
	return s.maxID
}

func (s *segment) SetMax(max int64) {
	s.maxID = max
}

func (s *segment) GetStep() int64 {
	return s.step
}

func (s *segment) SetStep(step int64) {
	s.step = step
}

func (s *segment) GetIdle() int64 {
	return s.GetMax() - s.value.Load()
}

// GetBuffer 获取当前号段所属的SegmentBuffer
func (s *segment) GetBuffer() *segmentBuffer {
	return s.buffer
}
