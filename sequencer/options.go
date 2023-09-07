package sequencer

import (
	"time"

	"github.com/madz-lab/go-ibft/message/types"
)

var (
	DefaultRound0Duration = time.Second * 10
	DummyTransport        = TransportFn(func(_ types.Msg) {})
	//TrueQuorum            = QuorumFn(func(_ []types.Msg) bool { return true })
)
