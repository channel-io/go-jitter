package jitter

import (
	"github.com/huandu/skiplist"
	"github.com/samber/lo"
	"math"

	"sync"
)

type Factory struct {
	minLatency, maxLatency, window, defaultTickInterval int64
	listener                                            Listener
}

func NewFactory(minLatency, maxLatency, window, defaultTickInterval int64, listener Listener) *Factory {
	return &Factory{
		minLatency:          minLatency,
		maxLatency:          maxLatency,
		window:              window,
		defaultTickInterval: defaultTickInterval,
		listener:            listener,
	}
}

func (f *Factory) CreateBuffer() Buffer {
	return NewJitter(f.minLatency, f.maxLatency, f.window, f.defaultTickInterval, f.listener)
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

	listener Listener
}

func NewJitter(minLatency, maxLatency, window, defaultTickInterval int64, listener Listener) *Jitter {
	if listener == nil {
		listener = &NullListener{}
	}
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
		listener:            listener,
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

	b.listener.OnPacketEnqueue(b.current, b.sumRemainingTs(), p)

	if !b.marked || math.Abs(float64(p.Timestamp-b.targetTime())) > float64(b.maxLatency) {
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

	targetTime := b.targetTime()

	removeLessThan(b.list, targetTime)
	removeLessThan(b.normal, targetTime-b.window)
	removeLessThan(b.late, targetTime-b.window)
	removeLessThan(b.loss, targetTime-b.window)

	ret := b.dequeuePackets()

	if len(ret) == 0 {
		b.loss.Set(targetTime, nil)
		b.current += b.defaultTickInterval
		b.listener.OnPacketLoss(b.current, b.sumRemainingTs())
		return nil, false
	}

	lastPkt := ret[len(ret)-1]
	newTargetTime := lastPkt.Timestamp + lastPkt.SampleCnt
	incr := newTargetTime - targetTime
	b.current += incr
	b.listener.OnPacketDequeue(b.current, b.sumRemainingTs(), ret)

	return ret, true
}

func (b *Jitter) dequeuePackets() []*Packet {
	var ret []*Packet

	threshold := b.targetTime() + b.defaultTickInterval

	for {
		node := b.list.Front()
		if node == nil {
			break
		}

		pkt := node.Value.(*Packet)
		if pkt.Timestamp >= threshold {
			break
		}

		b.list.RemoveFront()
		ret = append(ret, pkt)
	}

	return ret
}

func (b *Jitter) adaptive() {
	// late 가 너무 많다면 b.latency 를 늦춤
	if b.sumTsOfLatePackets() > b.window*2/100 { // late 패킷들의 ptime 합이 윈도우의 2% 를 초과시
		candidate := b.latency + maxInList(b.late)
		b.latency = lo.Min([]int64{candidate, b.maxLatency})
		b.listener.OnLatencyChanged(b.latency)
		b.late.Init()
	}

	if b.loss.Len() == 0 && b.late.Len() == 0 { // loss 와 late 가 모두 없으면
		candidate := b.latency - minInList(b.normal)
		b.latency = lo.Max([]int64{candidate, b.minLatency})
		b.listener.OnLatencyChanged(b.latency)
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

func (b *Jitter) sumRemainingTs() int64 {
	ret := int64(0)
	for el := b.list.Front(); el != nil; el = el.Next() {
		pkt := el.Value.(*Packet)
		if pkt.Timestamp >= b.current {
			ret += pkt.SampleCnt
		}
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
