package jitter

type Listener interface {
	OnPacketLoss(currentTs, remainingTs int64)
	OnLatencyChanged(new int64)
	OnPacketReceived(currentTs, remainingTs int64, pkt *Packet)
	OnPacketDequeue(currentTs, remainingTs int64, pkt []*Packet)
}

type NullListener struct {
}

func (n NullListener) OnPacketLoss(currentTs, remainingTs int64) {
	//TODO implement me
	panic("implement me")
}

func (n NullListener) OnLatencyChanged(new int64) {
	//TODO implement me
	panic("implement me")
}

func (n NullListener) OnPacketReceived(currentTs, remainingTs int64, pkt *Packet) {
	//TODO implement me
	panic("implement me")
}

func (n NullListener) OnPacketDequeue(currentTs, remainingTs int64, pkt []*Packet) {
	//TODO implement me
	panic("implement me")
}
