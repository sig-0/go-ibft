package store

import "github.com/sig-0/go-ibft/message"

func (s *MsgStore) Feed() Feed {
	return Feed{s}
}

type Feed struct {
	*MsgStore
}

func (f Feed) SubscribeProposal(
	sequence, round uint64,
	futureRounds bool,
) (message.Subscription[*message.MsgProposal], func()) {
	return f.MsgStore.ProposalMessages.Subscribe(sequence, round, futureRounds)
}

func (f Feed) SubscribePrepare(
	sequence, round uint64,
	futureRounds bool,
) (message.Subscription[*message.MsgPrepare], func()) {
	return f.MsgStore.PrepareMessages.Subscribe(sequence, round, futureRounds)
}

func (f Feed) SubscribeCommit(
	sequence, round uint64,
	futureRounds bool,
) (message.Subscription[*message.MsgCommit], func()) {
	return f.MsgStore.CommitMessages.Subscribe(sequence, round, futureRounds)
}

func (f Feed) SubscribeRoundChange(
	sequence, round uint64,
	futureRounds bool,
) (message.Subscription[*message.MsgRoundChange], func()) {
	return f.MsgStore.RoundChangeMessages.Subscribe(sequence, round, futureRounds)
}
