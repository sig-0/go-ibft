package store

import (
	"github.com/madz-lab/go-ibft/message/types"
)

type message interface {
	types.IBFTMessage

	GetView() *types.View
	GetSender() []byte
}

// MessageStore is a thread-safe storage for consensus messages with a built-in Feed mechanism
type MessageStore struct {
	ProposalMessages    Collection[*types.MsgProposal]
	PrepareMessages     Collection[*types.MsgPrepare]
	CommitMessages      Collection[*types.MsgCommit]
	RoundChangeMessages Collection[*types.MsgRoundChange]
}

// NewMsgStore returns a new MessageStore instance. MsgNotificationFn added to this store
// have their signatures verified before being included
func NewMsgStore() *MessageStore {
	return &MessageStore{
		ProposalMessages:    NewCollection[*types.MsgProposal](),
		PrepareMessages:     NewCollection[*types.MsgPrepare](),
		CommitMessages:      NewCollection[*types.MsgCommit](),
		RoundChangeMessages: NewCollection[*types.MsgRoundChange](),
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
