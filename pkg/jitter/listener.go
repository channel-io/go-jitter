package jitter

type Listener interface {
	OnPacketLoss(targetTs, remainingTs int64)
	OnPacketEnqueue(targetTs, remainingTs int64, pkt *Packet)
	OnPacketDequeue(targetTs, remainingTs int64, pkt []*Packet)
	OnLatencyChanged(new int64)
	OnReSyncTriggered(oldCurrent int64, newCurrent int64)
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

func (n NullListener) OnReSyncTriggered(oldCurrent int64, newCurrent int64) {

}
