package sequencer

import "github.com/madz-lab/go-ibft/message/types"

type Quorum interface {
	HasQuorum(sequence uint64, msgs []types.Msg) bool
}

type QuorumFn func(uint64, []types.Msg) bool

func (q QuorumFn) HasQuorum(sequence uint64, msgs []types.Msg) bool {
	return q(sequence, msgs)
}
