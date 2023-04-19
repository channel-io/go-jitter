package jitter

import (
	"github.com/huandu/go-assert"
	"testing"
)

func (b *Buffer) getDry() []byte {
	if res, ok := b.Get(); ok {
		return res
	} else {
		return nil
	}
}

func Test_basic(t *testing.T) {
	b := NewBuffer(100, 400, 20*50, 20)

	b.Put(&Packet{Timestamp: 1000, Data: []byte{1}, SampleCnt: 20})
	b.Put(&Packet{Timestamp: 1020, Data: []byte{2}, SampleCnt: 20})
	b.Put(&Packet{Timestamp: 1040, Data: []byte{3}, SampleCnt: 20})

	assert.Equal(t, b.getDry(), nil)       // 900
	assert.Equal(t, b.getDry(), nil)       // 920
	assert.Equal(t, b.getDry(), nil)       // 940
	assert.Equal(t, b.getDry(), nil)       // 960
	assert.Equal(t, b.getDry(), nil)       // 980
	assert.Equal(t, b.getDry(), []byte{1}) // 1000
	assert.Equal(t, b.getDry(), []byte{2}) // 1020
	assert.Equal(t, b.getDry(), []byte{3}) // 1040

	assert.Equal(t, b.targetTime(), int64(1060))
	assert.Equal(t, b.latency, int64(100))
}

func Test_basic2(t *testing.T) {
	b := NewBuffer(100, 400, 20*50, 20)

	b.Put(&Packet{Timestamp: 1000, Data: []byte{1}, SampleCnt: 20})

	assert.Equal(t, b.getDry(), nil)       // 900
	assert.Equal(t, b.getDry(), nil)       // 920
	assert.Equal(t, b.getDry(), nil)       // 940
	assert.Equal(t, b.getDry(), nil)       // 960
	assert.Equal(t, b.getDry(), nil)       // 980
	assert.Equal(t, b.getDry(), []byte{1}) // 1000
	assert.Equal(t, b.getDry(), nil)       // 1020
	assert.Equal(t, b.getDry(), nil)       // 1040
	assert.Equal(t, b.getDry(), nil)       // 1060
	assert.Equal(t, b.getDry(), nil)       // 1080
	assert.Equal(t, b.getDry(), nil)       // 1100
	assert.Equal(t, b.getDry(), nil)       // 1120
	assert.Equal(t, b.getDry(), nil)       // 1140
	assert.Equal(t, b.getDry(), nil)       // 1160
	assert.Equal(t, b.getDry(), nil)       // 1180

	b.Put(&Packet{Timestamp: 1020, Data: []byte{2}, SampleCnt: 20}) // late 180
	b.Put(&Packet{Timestamp: 1040, Data: []byte{3}, SampleCnt: 20}) // late 160

	assert.Equal(t, b.targetTime(), int64(1200))
	assert.Equal(t, b.latency, int64(100))

	b.adaptive()

	// 	assert.Equal(t, b.targetTime(), int64(1000))
	assert.Equal(t, b.latency, int64(280))

	//

	i := 0
	for ; i < 12; i++ {
		b.Put(&Packet{Timestamp: int64(1060 + i*20), Data: []byte{4 + byte(i)}, SampleCnt: 20})
	}

	for ; i < 100; i++ {
		b.Put(&Packet{Timestamp: int64(1060 + i*20), Data: []byte{4 + byte(i)}, SampleCnt: 20})
		b.getDry()
	}

	b.adaptive()

	// assert.Equal(t, b.targetTime(), int64(1200))
	assert.Equal(t, b.latency, int64(100))
}
