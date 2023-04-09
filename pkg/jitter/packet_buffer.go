package jitter

import "github.com/pion/rtp"

type PacketBuffer struct {
	buffer *Buffer
	ssrc   uint32

	tickInterval int64
	minLatency   int64 // 200ms
	maxLatency   int64 // 400ms
	window       int64 // 2000ms
}

func NewPacketBuffer(tickInterval, minLatency, maxLatency, window int64) *PacketBuffer {
	return &PacketBuffer{
		tickInterval: tickInterval,
		minLatency:   minLatency,
		maxLatency:   maxLatency,
		window:       window,
	}
}

func (p *PacketBuffer) init(ssrc uint32) {
	p.buffer = NewBuffer(p.tickInterval, p.minLatency, p.maxLatency, p.window)
	p.ssrc = ssrc
}

func (p *PacketBuffer) Put(packet *rtp.Packet) {
	if p.ssrc != packet.SSRC {
		p.init(p.ssrc)
	}

	p.buffer.Put(int64(packet.Timestamp), packet.Payload)
}

func (p *PacketBuffer) Get() ([]byte, bool) {
	return p.buffer.Get()
}
