package store

import "github.com/sig-0/go-ibft/message"

func (s *MsgStore) Feed() Feed {
	return Feed{s}
}

type Feed struct {
	*MsgStore
}

func (f Feed) SubscribeProposal(
	view *message.View,
	futureRounds bool,
) (message.Subscription[*message.MsgProposal], func()) {
	return f.MsgStore.ProposalMessages.Subscribe(view, futureRounds)
}

func (f Feed) SubscribePrepare(
	view *message.View,
	futureRounds bool,
) (message.Subscription[*message.MsgPrepare], func()) {
	return f.MsgStore.PrepareMessages.Subscribe(view, futureRounds)
}

func (f Feed) SubscribeCommit(
	view *message.View,
	futureRounds bool,
) (message.Subscription[*message.MsgCommit], func()) {
	return f.MsgStore.CommitMessages.Subscribe(view, futureRounds)
}

func (f Feed) SubscribeRoundChange(
	view *message.View,
	futureRounds bool,
) (message.Subscription[*message.MsgRoundChange], func()) {
	return f.MsgStore.RoundChangeMessages.Subscribe(view, futureRounds)
}
