package jitter

import (
	"github.com/huandu/skiplist"
	"github.com/samber/lo"
	"math"

	"sync"
)

type Buffer struct {
	sync.Mutex

	list *skiplist.SkipList

	early *skiplist.SkipList
	late  *skiplist.SkipList
	loss  *skiplist.SkipList

	current int64
	latency int64

	marked    bool
	firstTime int64

	tickInterval int64
	minLatency   int64 // 200ms
	maxLatency   int64 // 400ms
	window       int64 // 2000ms
}

func NewBuffer(tickInterval, minLatency, maxLatency, window int64) *Buffer {
	b := &Buffer{
		list:         skiplist.New(skiplist.Int64),
		early:        skiplist.New(skiplist.Int64),
		late:         skiplist.New(skiplist.Int64),
		loss:         skiplist.New(skiplist.Int64),
		current:      0,
		latency:      minLatency,
		tickInterval: tickInterval,
		minLatency:   minLatency,
		maxLatency:   maxLatency,
		window:       window,
	}
	return b
}

func (b *Buffer) Put(ts int64, data []byte) {
	b.Lock()
	defer b.Unlock()

	if !b.marked {
		b.firstTime = ts
		b.marked = true
	}

	b.list.Set(ts, data)

	delta := ts - b.targetTime()
	if delta > 0 { // check
		b.early.Set(ts, delta) // 일찍 온 시간을 기록
	} else if delta > -b.maxLatency { // 늦게 온 것이면, 단 너무 늦으면 버림
		b.late.Set(ts, -delta) // 늦은 시간을 기록
	}

}

func (b *Buffer) Get() ([]byte, bool) {
	b.Lock()
	defer b.Unlock()

	defer func() {
		b.current += b.tickInterval
	}()

	b.adaptive()

	targetTime := b.targetTime()

	removeLessThan(b.list, targetTime)
	removeLessThan(b.early, targetTime-b.window)
	removeLessThan(b.late, targetTime-b.window)
	removeLessThan(b.loss, targetTime-b.window)

	front := b.list.Front()
	if front != nil && front.Key() != nil && front.Key().(int64) == targetTime {
		b.list.RemoveFront()
		return front.Value.([]byte), true
	} else {
		// loss
		b.loss.Set(targetTime, nil)
		return nil, false
	}
}

func (b *Buffer) adaptive() {
	// late 가 너무 많다면 b.latency 를 늦춤
	if b.late.Len() > int(b.window/b.tickInterval*2/100) { // 2% 보다 크면
		candidate := b.latency + maxInList(b.late)
		b.latency = lo.Min([]int64{candidate, b.maxLatency})
		b.late.Init()
		b.early.Init()
	}

	// loss 가 없이 안정적이라면
	if b.loss.Len() == 0 && b.early.Len() > int(b.window/b.tickInterval*2/100) { // 2% 보다 크면
		candidate := b.latency - minInList(b.early)
		b.latency = lo.Max([]int64{candidate, b.minLatency})
		b.late.Init()
		b.early.Init()
	}
}

func maxInList(list *skiplist.SkipList) int64 {
	var res int64 = math.MinInt64
	for el := list.Front(); el != nil; el = el.Next() {
		res = lo.Max([]int64{res, el.Value.(int64)})
	}
	return res
}

func minInList(list *skiplist.SkipList) int64 {
	var res int64 = math.MaxInt64
	for el := list.Front(); el != nil; el = el.Next() {
		res = lo.Min([]int64{res, el.Value.(int64)})
	}
	return res
}

func removeLessThan(list *skiplist.SkipList, ts int64) {
	for {
		front := list.Front()
		if front == nil || front.Key() == nil || front.Key().(int64) >= ts {
			break
		}
		list.RemoveFront()
	}
}

func (b *Buffer) targetTime() int64 {
	return b.firstTime + b.current - b.latency
}
