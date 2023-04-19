package jitter

type Packet struct {
	Data      []byte
	SampleCnt int64
	Timestamp int64
	SSRC      uint32
}

type Buffer interface {
	Put(p *Packet)
	Get() ([]*Packet, bool)
}

type BufferFactory interface {
	CreateBuffer() Buffer
}
