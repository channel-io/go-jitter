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
	b := NewBuffer(20, 100, 400, 1000)

	b.Put(1000, []byte{1})
	b.Put(1020, []byte{2})
	b.Put(1040, []byte{3})

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
	b := NewBuffer(20, 100, 400, 1000)

	b.Put(1000, []byte{1})

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

	b.Put(1020, []byte{2}) // late 180
	b.Put(1040, []byte{3}) // late 160

	assert.Equal(t, b.targetTime(), int64(1200))
	assert.Equal(t, b.latency, int64(100))

	b.adaptive()

	// 	assert.Equal(t, b.targetTime(), int64(1000))
	assert.Equal(t, b.latency, int64(280))

	//

	i := 0
	for ; i < 12; i++ {
		b.Put(int64(1060+i*20), []byte{4 + byte(i)})
	}

	for ; i < 100; i++ {
		b.Put(int64(1060+i*20), []byte{4 + byte(i)})
		b.getDry()
	}

	b.adaptive()

	// assert.Equal(t, b.targetTime(), int64(1200))
	assert.Equal(t, b.latency, int64(100))
}
