package jitter

import (
	"github.com/huandu/go-assert"
	"testing"
)

func assertGet(t *testing.T, b *Buffer, expectedValue []byte, expectedOk bool, expectedTS int64) {
	res, ok, ts := b.Get()
	assert.Equal(t, res, expectedValue)
	assert.Equal(t, ok, expectedOk)
	assert.Equal(t, ts, expectedTS)
}

func Test_basic(t *testing.T) {
	b := NewBuffer(20, 100, 400, 1000)

	b.Put(1000, []byte{1})
	b.Put(1020, []byte{2})
	b.Put(1040, []byte{3})

	assertGet(t, b, nil, false, 900)
	assertGet(t, b, nil, false, 920)
	assertGet(t, b, nil, false, 940)
	assertGet(t, b, nil, false, 960)
	assertGet(t, b, nil, false, 980)
	assertGet(t, b, []byte{1}, true, 1000)
	assertGet(t, b, []byte{2}, true, 1020)
	assertGet(t, b, []byte{3}, true, 1040)

	assert.Equal(t, b.targetTime(), int64(1060))
	assert.Equal(t, b.latency, int64(100))
}

func Test_basic2(t *testing.T) {
	b := NewBuffer(20, 100, 400, 1000)

	b.Put(1000, []byte{1})

	assertGet(t, b, nil, false, 900)
	assertGet(t, b, nil, false, 920)
	assertGet(t, b, nil, false, 940)
	assertGet(t, b, nil, false, 960)
	assertGet(t, b, nil, false, 980)
	assertGet(t, b, []byte{1}, true, 1000)
	assertGet(t, b, nil, false, 1020)
	assertGet(t, b, nil, false, 1040)
	assertGet(t, b, nil, false, 1060)
	assertGet(t, b, nil, false, 1080)
	assertGet(t, b, nil, false, 1100)
	assertGet(t, b, nil, false, 1120)
	assertGet(t, b, nil, false, 1140)
	assertGet(t, b, nil, false, 1160)
	assertGet(t, b, nil, false, 1180)

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
		b.Get()
	}

	b.adaptive()

	// assert.Equal(t, b.targetTime(), int64(1200))
	assert.Equal(t, b.latency, int64(100))
}

func Test_overflow(t *testing.T) {
	b := NewBuffer(20, 100, 400, 1000)

	b.Put(1<<32-20, []byte{1})
	b.Put(0, []byte{2})
	b.Put(20, []byte{3})

	assertGet(t, b, nil, false, 1<<32-120)
	assertGet(t, b, nil, false, 1<<32-100)
	assertGet(t, b, nil, false, 1<<32-80)
	assertGet(t, b, nil, false, 1<<32-60)
	assertGet(t, b, nil, false, 1<<32-40)
	assertGet(t, b, []byte{1}, true, 1<<32-20)
	assertGet(t, b, []byte{2}, true, 1<<32)
	assertGet(t, b, []byte{3}, true, 1<<32+20)
}
