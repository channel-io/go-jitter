package jitter

import (
	"github.com/huandu/go-assert"
	"testing"
)

func TestOverflow(t *testing.T) {
	factory := NewFactory(100, 400, 20*50, 20)
	packetBuffer := NewPacketBuffer(factory)

	packetBuffer.Put(&Packet{Timestamp: 1<<32 - 20, Data: []byte{1}, SampleCnt: 20})

	packetBuffer.Put(&Packet{Timestamp: 0, Data: []byte{1}, SampleCnt: 20})
	assert.Equal(t, int64(1<<32), packetBuffer.lastTimestamp())

	packetBuffer.Put(&Packet{Timestamp: 20, Data: []byte{1}, SampleCnt: 20})
	assert.Equal(t, int64(1<<32+20), packetBuffer.lastTimestamp())
}
