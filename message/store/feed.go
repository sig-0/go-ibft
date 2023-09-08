package store

import (
	"github.com/madz-lab/go-ibft/message/types"
)

type Feed struct {
	*Store
}

func (f Feed) Proposal(view *types.View, futureRounds bool) (<-chan func() []*types.MsgProposal, func()) {
	sub, cancelSub := f.Store.proposal.subscribe(view, futureRounds)

	return sub.Channel, cancelSub
}

func (f Feed) Prepare(view *types.View, futureRounds bool) (<-chan func() []*types.MsgPrepare, func()) {
	sub, cancelSub := f.Store.prepare.subscribe(view, futureRounds)

	return sub.Channel, cancelSub
}

func (f Feed) Commit(view *types.View, futureRounds bool) (<-chan func() []*types.MsgCommit, func()) {
	sub, cancelSub := f.Store.commit.subscribe(view, futureRounds)

	return sub.Channel, cancelSub
}

func (f Feed) RoundChange(view *types.View, futureRounds bool) (<-chan func() []*types.MsgRoundChange, func()) {
	sub, cancelSub := f.Store.roundChange.subscribe(view, futureRounds)

	return sub.Channel, cancelSub
}
