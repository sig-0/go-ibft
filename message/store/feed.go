package store

import (
	"github.com/madz-lab/go-ibft/message/types"
)

type Feed struct {
	*Store
}

func (f Feed) Proposal(view *types.View, futureRounds bool) (<-chan func() []*types.MsgProposal, func()) {
	return f.Store.proposal.subscribe(view, futureRounds)
}

func (f Feed) Prepare(view *types.View, futureRounds bool) (<-chan func() []*types.MsgPrepare, func()) {
	return f.Store.prepare.subscribe(view, futureRounds)
}

func (f Feed) Commit(view *types.View, futureRounds bool) (<-chan func() []*types.MsgCommit, func()) {
	return f.Store.commit.subscribe(view, futureRounds)
}

func (f Feed) RoundChange(view *types.View, futureRounds bool) (<-chan func() []*types.MsgRoundChange, func()) {
	return f.Store.roundChange.subscribe(view, futureRounds)
}
