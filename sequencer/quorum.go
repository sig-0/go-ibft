package sequencer

import "github.com/madz-lab/go-ibft/message/types"

type Quorum interface {
	HasQuorum([]types.Msg) bool
}

type QuorumFn func([]types.Msg) bool

func (q QuorumFn) HasQuorum(messages []types.Msg) bool {
	return q(messages)
}
