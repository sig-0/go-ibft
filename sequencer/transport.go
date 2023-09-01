package sequencer

import "github.com/madz-lab/go-ibft/message/types"

type Transport interface {
	Multicast(msg types.Msg)
}

type TransportFn func(types.Msg)

func (t TransportFn) Multicast(msg types.Msg) {
	t(msg)
}
