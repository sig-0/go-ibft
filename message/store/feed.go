package store

import (
	"github.com/madz-lab/go-ibft/message/types"
)

type Feed struct {
	*Store
}

func (f Feed) SubscribeToProposalMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgProposal, func()) {
	sub := newSubscription[types.MsgProposal](view, futureRounds)
	sub.notify(f.Store.proposal.unwrapFn(view, futureRounds))

	id := f.Store.proposalSubs.add(sub)
	return sub.Channel, func() { f.Store.proposalSubs.remove(id) }
}

func (f Feed) SubscribeToPrepareMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgPrepare, func()) {
	//TODO implement me
	panic("implement me")
}

func (f Feed) SubscribeToCommitMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgCommit, func()) {
	//TODO implement me
	panic("implement me")
}

func (f Feed) SubscribeToRoundChangeMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgRoundChange, func()) {
	//TODO implement me
	panic("implement me")
}
