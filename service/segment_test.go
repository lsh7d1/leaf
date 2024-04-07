package service

import "testing"

func TestSegmentBuffer(t *testing.T) {
	sb := newSegmentBuffer("aaaa")
	t.Logf("sb: %#v", sb)
}
