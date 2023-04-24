package jitter

import (
	"sync"
)

type PacketBuffer struct {
	sync.Mutex

	buffer  Buffer
	factory BufferFactory

	ssrc    uint32
	firstTs int64
	lastTs  int64
}

func NewPacketBuffer(factory BufferFactory) *PacketBuffer {
	return &PacketBuffer{
		factory: factory,
	}
}

func (p *PacketBuffer) init(packet *Packet) {
	p.buffer = p.factory.CreateBuffer()
	p.ssrc = packet.SSRC
	p.firstTs = packet.Timestamp
}

func (p *PacketBuffer) Put(packet *Packet) {
	p.Lock()
	defer p.Unlock()

	if p.buffer == nil || p.ssrc != packet.SSRC {
		p.init(packet)
	}

	if p.firstTs > packet.Timestamp {
		packet.Timestamp += 1 << 32 // overflow 처리
	}

	p.lastTs = packet.Timestamp
	p.buffer.Put(packet)
}

func (p *PacketBuffer) Get() ([]*Packet, bool) {
	p.Lock()
	defer p.Unlock()

	if p.buffer == nil {
		return nil, false
	}

	return p.buffer.Get()
}

func (p *PacketBuffer) lastTimestamp() int64 {
	return p.lastTs
}
