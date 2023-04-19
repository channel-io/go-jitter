package jitter

import (
	"github.com/huandu/skiplist"
	"github.com/samber/lo"
	"math"

	"sync"
)

type deltaWithSampleCnt struct {
	delta     int64
	sampleCnt int64
}

type Packet struct {
	Data      []byte
	SampleCnt int64
	Timestamp int64
	SSRC      uint32
}

type Buffer struct {
	sync.Mutex

	list *skiplist.SkipList

	normal *skiplist.SkipList
	late   *skiplist.SkipList
	loss   *skiplist.SkipList

	current int64
	latency int64

	marked    bool
	firstTime int64

	minLatency int64 // 200ms
	maxLatency int64 // 400ms
	window     int64 // 2000ms

	defaultTickInterval int64
}

func NewBuffer(minLatency, maxLatency, window, defaultTickInterval int64) *Buffer {
	b := &Buffer{
		normal:              skiplist.New(skiplist.Int64),
		list:                skiplist.New(skiplist.Int64),
		late:                skiplist.New(skiplist.Int64),
		loss:                skiplist.New(skiplist.Int64),
		current:             0,
		latency:             minLatency,
		minLatency:          minLatency,
		maxLatency:          maxLatency,
		window:              window,
		defaultTickInterval: defaultTickInterval,
	}
	return b
}

func (b *Buffer) Put(p *Packet) {
	b.Lock()
	defer b.Unlock()

	if !b.marked {
		b.firstTime = p.Timestamp
		b.marked = true
	}

	b.list.Set(p.Timestamp, p)

	delta := b.targetTime() - p.Timestamp
	if delta < 0 {
		b.normal.Set(p.Timestamp, deltaWithSampleCnt{delta: -delta, sampleCnt: p.SampleCnt})
	} else if delta > 0 && delta < b.maxLatency { // 늦게 온 것이면, 단 너무 늦으면 버림
		b.late.Set(p.Timestamp, deltaWithSampleCnt{delta: delta, sampleCnt: p.SampleCnt}) // 늦은 시간을 기록
	}
}

func (b *Buffer) Get() ([]byte, bool) {
	b.Lock()
	defer b.Unlock()

	b.adaptive()

	targetTime := b.targetTime()

	removeLessThan(b.list, targetTime)
	removeLessThan(b.normal, targetTime-b.window)
	removeLessThan(b.late, targetTime-b.window)
	removeLessThan(b.loss, targetTime-b.window)

	front := b.list.Front()

	if front != nil && front.Key() != nil && front.Key().(int64) == targetTime {
		b.list.RemoveFront()
		pkt := front.Value.(*Packet)
		b.current += pkt.SampleCnt
		return pkt.Data, true
	} else {
		b.loss.Set(targetTime, nil)
		b.current += b.defaultTickInterval
		return nil, false
	}
}

func (b *Buffer) adaptive() {
	// late 가 너무 많다면 b.latency 를 늦춤
	if b.sumTsOfLatePackets() > b.window*2/100 { // late 패킷들의 ptime 합이 윈도우의 2% 를 초과시
		candidate := b.latency + maxInList(b.late)
		b.latency = lo.Min([]int64{candidate, b.maxLatency})
		b.late.Init()
	}

	// loss 가 없이 안정적이라면
	if b.loss.Len() == 0 && b.late.Len() == 0 { // loss 와 late 가 모두 없으면
		candidate := b.latency - minInList(b.normal)
		b.latency = lo.Max([]int64{candidate, b.minLatency})
		b.late.Init()
	}
}

func (b *Buffer) sumTsOfLatePackets() int64 {
	var ret int64

	elem := b.late.Front()
	for elem != nil {
		ret += elem.Value.(deltaWithSampleCnt).sampleCnt
		elem = elem.Next()
	}

	return ret
}

func maxInList(list *skiplist.SkipList) int64 {
	var res int64 = math.MinInt64
	for el := list.Front(); el != nil; el = el.Next() {
		res = lo.Max([]int64{res, el.Value.(deltaWithSampleCnt).delta})
	}
	return res
}

func minInList(list *skiplist.SkipList) int64 {
	var res int64 = math.MaxInt64
	for el := list.Front(); el != nil; el = el.Next() {
		res = lo.Min([]int64{res, el.Value.(deltaWithSampleCnt).delta})
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
