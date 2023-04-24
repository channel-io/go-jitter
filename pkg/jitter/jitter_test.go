package jitter

import (
	"github.com/huandu/go-assert"
	"testing"
)

const (
	firstTimestamp  = 1000
	samplePerPacket = 20
)

func packet(seq byte) *Packet {
	return &Packet{
		Timestamp: firstTimestamp + int64(seq-1)*samplePerPacket,
		Data:      []byte{seq},
		SampleCnt: samplePerPacket,
	}
}

func assertGet(t *testing.T, b *Jitter, expectedValue byte, expectedTs int64) {
	res, ok := b.Get()
	assert.Equal(t, ok, true)
	assert.Equal(t, len(res), 1)
	assert.Equal(t, res[0].Data, []byte{expectedValue})
	assert.Equal(t, b.targetTime(), expectedTs)
}

func assertLoss(t *testing.T, b *Jitter, expectedTs int64) {
	_, ok := b.Get()
	assert.Equal(t, ok, false)
	assert.Equal(t, b.targetTime(), expectedTs)
}

func Test_basic(t *testing.T) {
	b := NewJitter(100, 400, 20*50, 20, nil)

	b.Put(packet(1))
	b.Put(packet(2))
	b.Put(packet(3))

	assertLoss(t, b, 920)
	assertLoss(t, b, 940)
	assertLoss(t, b, 960)
	assertLoss(t, b, 980)
	assertLoss(t, b, 1000)

	assertGet(t, b, 1, 1020)
	assertGet(t, b, 2, 1040)
	assertGet(t, b, 3, 1060)

	assert.Equal(t, b.targetTime(), int64(1060))
	assert.Equal(t, b.latency, int64(100))
}

func Test_basic2(t *testing.T) {
	b := NewJitter(100, 400, 20*50, 20, nil)

	b.Put(packet(1))

	assertLoss(t, b, 920)
	assertLoss(t, b, 940)
	assertLoss(t, b, 960)
	assertLoss(t, b, 980)
	assertLoss(t, b, 1000)

	assertGet(t, b, 1, 1020)

	assertLoss(t, b, 1040)
	assertLoss(t, b, 1060)
	assertLoss(t, b, 1080)
	assertLoss(t, b, 1100)
	assertLoss(t, b, 1120)
	assertLoss(t, b, 1140)
	assertLoss(t, b, 1160)
	assertLoss(t, b, 1180)
	assertLoss(t, b, 1200)

	b.Put(packet(2)) // late 180
	b.Put(packet(3)) // late 160

	assert.Equal(t, b.latency, int64(100))

	b.adaptive()

	assert.Equal(t, b.latency, int64(280))

	i := 0
	for ; i < 12; i++ {
		b.Put(packet(byte(3 + i)))
	}

	for ; i < 100; i++ {
		b.Put(packet(byte(3 + i)))
		b.Get()
	}

	b.adaptive()

	assert.Equal(t, b.latency, int64(100))
}
