package jitter

type Listener interface {
	OnPacketLoss(currentTs, remainingTs int64)
	OnPacketEnqueue(currentTs, remainingTs int64, pkt *Packet)
	OnPacketDequeue(currentTs, remainingTs int64, pkt []*Packet)
	OnLatencyChanged(new int64)
}

type NullListener struct {
}

func (n NullListener) OnPacketLoss(currentTs, remainingTs int64) {
}

func (n NullListener) OnLatencyChanged(new int64) {
}

func (n NullListener) OnPacketEnqueue(currentTs, remainingTs int64, pkt *Packet) {
}

func (n NullListener) OnPacketDequeue(currentTs, remainingTs int64, pkt []*Packet) {
}
