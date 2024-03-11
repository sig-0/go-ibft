package store

import (
	"github.com/madz-lab/go-ibft/message/types"
)

// MessageStore is a thread-safe storage for consensus messages with a built-in Feed mechanism
type MessageStore struct {
	ProposalMessages    MessageCollection[*types.MsgProposal]
	PrepareMessages     MessageCollection[*types.MsgPrepare]
	CommitMessages      MessageCollection[*types.MsgCommit]
	RoundChangeMessages MessageCollection[*types.MsgRoundChange]
}

// NewMessageStore returns a new MessageStore instance
func NewMessageStore() *MessageStore {
	return &MessageStore{
		ProposalMessages:    NewMessageCollection[*types.MsgProposal](),
		PrepareMessages:     NewMessageCollection[*types.MsgPrepare](),
		CommitMessages:      NewMessageCollection[*types.MsgCommit](),
		RoundChangeMessages: NewMessageCollection[*types.MsgRoundChange](),
	}
}

func (s *MessageStore) Feed() Feed {
	return Feed{s}
}

type Feed struct {
	*MessageStore
}

func (f Feed) ProposalMessages(
	view *types.View,
	futureRounds bool,
) (types.Subscription[*types.MsgProposal], func()) {
	return f.MessageStore.ProposalMessages.Subscribe(view, futureRounds)
}

func (f Feed) PrepareMessages(
	view *types.View,
	futureRounds bool,
) (types.Subscription[*types.MsgPrepare], func()) {
	return f.MessageStore.PrepareMessages.Subscribe(view, futureRounds)
}

func (f Feed) CommitMessages(
	view *types.View,
	futureRounds bool,
) (types.Subscription[*types.MsgCommit], func()) {
	return f.MessageStore.CommitMessages.Subscribe(view, futureRounds)
}

func (f Feed) RoundChangeMessages(
	view *types.View,
	futureRounds bool,
) (types.Subscription[*types.MsgRoundChange], func()) {
	return f.MessageStore.RoundChangeMessages.Subscribe(view, futureRounds)
}
