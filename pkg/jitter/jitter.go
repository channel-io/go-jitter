package jitter

import (
	"github.com/huandu/skiplist"
	"github.com/samber/lo"
	"math"

	"sync"
)

type Factory struct {
	minLatency, maxLatency, window, defaultTickInterval int64
}

func NewFactory(minLatency, maxLatency, window, defaultTickInterval int64) *Factory {
	return &Factory{
		minLatency:          minLatency,
		maxLatency:          maxLatency,
		window:              window,
		defaultTickInterval: defaultTickInterval,
	}
}

func (f *Factory) CreateBuffer() Buffer {
	return NewJitter(f.minLatency, f.maxLatency, f.window, f.defaultTickInterval)
}

type deltaWithSampleCnt struct {
	delta     int64
	sampleCnt int64
}

type Jitter struct {
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

func NewJitter(minLatency, maxLatency, window, defaultTickInterval int64) *Jitter {
	b := &Jitter{
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

func (b *Jitter) init(ts int64) {
	b.firstTime = ts
	b.current = 0
	b.latency = b.minLatency
	b.marked = true
}

func (b *Jitter) Put(p *Packet) {
	b.Lock()
	defer b.Unlock()

	if !b.marked || math.Abs(float64(p.Timestamp-b.targetTime())) > 100_000 {
		b.init(p.Timestamp)
	}

	b.list.Set(p.Timestamp, p)

	delta := p.Timestamp - b.targetTime()
	if delta >= 0 {
		b.normal.Set(p.Timestamp, deltaWithSampleCnt{delta: delta, sampleCnt: p.SampleCnt})
	} else if delta < 0 && delta > -b.maxLatency { // 늦게 온 것이면, 단 너무 늦으면 버림
		b.late.Set(p.Timestamp, deltaWithSampleCnt{delta: -delta, sampleCnt: p.SampleCnt}) // 늦은 시간을 기록
	}
}

func (b *Jitter) Get() ([]*Packet, bool) {
	b.Lock()
	defer b.Unlock()

	if !b.marked {
		return nil, false
	}

	b.adaptive()

	removeLessThan(b.list, b.targetTime())
	removeLessThan(b.normal, b.targetTime()-b.window)
	removeLessThan(b.late, b.targetTime()-b.window)
	removeLessThan(b.loss, b.targetTime()-b.window)

	var ret []*Packet

	for {
		node := b.list.Front()
		if node == nil {
			break
		}

		pkt := node.Value.(*Packet)
		if pkt.Timestamp >= b.targetTime()+b.defaultTickInterval {
			break
		}

		b.list.RemoveFront()
		newTargetTime := pkt.Timestamp + pkt.SampleCnt
		delta := newTargetTime - b.targetTime()
		b.current += delta
		ret = append(ret, pkt)
	}

	if len(ret) == 0 {
		b.loss.Set(b.targetTime(), nil)
		b.current += b.defaultTickInterval
		return nil, false
	}

	return ret, true
}

func (b *Jitter) adaptive() {
	// late 가 너무 많다면 b.latency 를 늦춤
	if b.sumTsOfLatePackets() > b.window*2/100 { // late 패킷들의 ptime 합이 윈도우의 2% 를 초과시
		candidate := b.latency + maxInList(b.late)
		b.latency = lo.Min([]int64{candidate, b.maxLatency})
		b.late.Init()
	}

	if b.loss.Len() == 0 && b.late.Len() == 0 { // loss 와 late 가 모두 없으면
		candidate := b.latency - minInList(b.normal)
		b.latency = lo.Max([]int64{candidate, b.minLatency})
		b.late.Init()
	}
}

func (b *Jitter) sumTsOfLatePackets() int64 {
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

func (b *Jitter) targetTime() int64 {
	return b.firstTime + b.current - b.latency
}
