package jitter

type Listener interface {
	OnPacketLoss(currentTs, targetTs, remainingTs int64)
	OnPacketEnqueue(currentTs, targetTs, remainingTs int64, pkt *Packet)
	OnPacketDequeue(currentTs, targetTs, remainingTs int64, pkt []*Packet)
	OnReSyncTriggered(oldCurrent int64, newCurrent int64)
}

type NullListener struct {
}

func (n NullListener) OnPacketLoss(currentTs, targetTs, remainingTs int64) {
}

func (n NullListener) OnPacketEnqueue(currentTs, targetTs, remainingTs int64, pkt *Packet) {
}

func (n NullListener) OnPacketDequeue(currentTs, targetTs, remainingTs int64, pkt []*Packet) {
}

func (n NullListener) OnReSyncTriggered(oldCurrent int64, newCurrent int64) {
}
