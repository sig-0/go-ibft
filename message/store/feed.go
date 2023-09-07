package store

import (
	"github.com/madz-lab/go-ibft/message/types"
)

type Feed struct {
	*Store
}

func (f Feed) SubscribeToProposalMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgProposal, func()) {
	sub, cancelSub := f.Store.proposal.subscribe(view, futureRounds)
	return sub.Channel, cancelSub
}

//
//func (f Feed) SubscribeToPrepareMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgPrepare, func()) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (f Feed) SubscribeToCommitMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgCommit, func()) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (f Feed) SubscribeToRoundChangeMessages(view *types.View, futureRounds bool) (<-chan func() []*types.MsgRoundChange, func()) {
//	//TODO implement me
//	panic("implement me")
//}
