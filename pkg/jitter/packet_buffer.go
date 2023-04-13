package jitter

import (
	"github.com/pion/rtp"
	"sync"
)

type PacketBuffer struct {
	sync.Mutex

	buffer    *Buffer
	firstTime uint32
	ssrc      uint32

	sampleRate   int64 // 48 per milliseconds
	tickInterval int64 // 20ms
	minLatency   int64 // 200ms
	maxLatency   int64 // 400ms
	window       int64 // 2000ms
}

func NewPacketBuffer(sampleRate, tickInterval, minLatency, maxLatency, window int64) *PacketBuffer {
	return &PacketBuffer{
		sampleRate:   sampleRate,
		tickInterval: tickInterval,
		minLatency:   minLatency,
		maxLatency:   maxLatency,
		window:       window,
	}
}

func (p *PacketBuffer) init(packet *rtp.Packet) {
	p.buffer = NewBuffer(p.tickInterval, p.minLatency, p.maxLatency, p.window)
	p.firstTime = packet.Timestamp
	p.ssrc = packet.SSRC
}

func (p *PacketBuffer) Put(packet *rtp.Packet) {
	p.Lock()
	defer p.Unlock()

	if p.buffer == nil || p.ssrc != packet.SSRC {
		p.init(packet)
	}

	ts := int64(packet.Timestamp)
	if p.firstTime > packet.Timestamp {
		ts += 1 << 32 // overflow 처리
	}

	p.buffer.Put(ts/p.sampleRate, packet.Payload)
}

func (p *PacketBuffer) Get() ([]byte, bool, int64) {
	return p.buffer.Get()
}
