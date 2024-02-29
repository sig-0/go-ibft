package store

import (
	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

type message interface {
	types.IBFTMessage

	GetView() *types.View
	GetSender() []byte
}

// MsgStore is a thread-safe storage for consensus messages with a built-in Feed mechanism
type MsgStore struct {
	ProposalMessages    Collection[*types.MsgProposal]
	PrepareMessages     Collection[*types.MsgPrepare]
	CommitMessages      Collection[*types.MsgCommit]
	RoundChangeMessages Collection[*types.MsgRoundChange]
}

// NewMsgStore returns a new MsgStore instance. NotificationFn added to this store
// have their signatures verified before being included
func NewMsgStore() *MsgStore {
	s := &MsgStore{
		ProposalMessages:    NewCollection[*types.MsgProposal](),
		PrepareMessages:     NewCollection[*types.MsgPrepare](),
		CommitMessages:      NewCollection[*types.MsgCommit](),
		RoundChangeMessages: NewCollection[*types.MsgRoundChange](),
	}

	return s
}

func (s *MsgStore) Feed() ibft.MessageFeed {
	return Feed{s}
}

type Feed struct {
	*MsgStore
}

func (f Feed) ProposalMessages(
	view *types.View,
	futureRounds bool,
) (ibft.Subscription[*types.MsgProposal], func()) {
	return f.MsgStore.ProposalMessages.Subscribe(view, futureRounds)
}

func (f Feed) PrepareMessages(
	view *types.View,
	futureRounds bool,
) (ibft.Subscription[*types.MsgPrepare], func()) {
	return f.MsgStore.PrepareMessages.Subscribe(view, futureRounds)
}

func (f Feed) CommitMessages(
	view *types.View,
	futureRounds bool,
) (ibft.Subscription[*types.MsgCommit], func()) {
	return f.MsgStore.CommitMessages.Subscribe(view, futureRounds)
}

func (f Feed) RoundChangeMessages(
	view *types.View,
	futureRounds bool,
) (ibft.Subscription[*types.MsgRoundChange], func()) {
	return f.MsgStore.RoundChangeMessages.Subscribe(view, futureRounds)
}
