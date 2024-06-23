package store

import (
	"github.com/madz-lab/go-ibft/message/types"
)

// MsgStore is a thread-safe storage for consensus messages with a built-in sequencer.MsgFeed mechanism
type MsgStore struct {
	ProposalMessages    MsgCollection[*types.MsgProposal]
	PrepareMessages     MsgCollection[*types.MsgPrepare]
	CommitMessages      MsgCollection[*types.MsgCommit]
	RoundChangeMessages MsgCollection[*types.MsgRoundChange]
}

// NewMsgStore returns a new MsgStore instance
func NewMsgStore() *MsgStore {
	return &MsgStore{
		ProposalMessages:    NewMsgCollection[*types.MsgProposal](),
		PrepareMessages:     NewMsgCollection[*types.MsgPrepare](),
		CommitMessages:      NewMsgCollection[*types.MsgCommit](),
		RoundChangeMessages: NewMsgCollection[*types.MsgRoundChange](),
	}
}

// Add includes the message in the store
func (s *MsgStore) Add(m types.Message) {
	switch m := m.(type) {
	case *types.MsgProposal:
		s.ProposalMessages.Add(m)
	case *types.MsgPrepare:
		s.PrepareMessages.Add(m)
	case *types.MsgCommit:
		s.CommitMessages.Add(m)
	case *types.MsgRoundChange:
		s.RoundChangeMessages.Add(m)
	}
}

// Clear removes all messages from store
func (s *MsgStore) Clear() {
	s.ProposalMessages.Clear()
	s.PrepareMessages.Clear()
	s.CommitMessages.Clear()
	s.RoundChangeMessages.Clear()
}

func (s *MsgStore) Feed() Feed {
	return Feed{s}
}

type Feed struct {
	*MsgStore
}

func (f Feed) ProposalMessages(
	view *types.View,
	futureRounds bool,
) (types.Subscription[*types.MsgProposal], func()) {
	return f.MsgStore.ProposalMessages.Subscribe(view, futureRounds)
}

func (f Feed) PrepareMessages(
	view *types.View,
	futureRounds bool,
) (types.Subscription[*types.MsgPrepare], func()) {
	return f.MsgStore.PrepareMessages.Subscribe(view, futureRounds)
}

func (f Feed) CommitMessages(
	view *types.View,
	futureRounds bool,
) (types.Subscription[*types.MsgCommit], func()) {
	return f.MsgStore.CommitMessages.Subscribe(view, futureRounds)
}

func (f Feed) RoundChangeMessages(
	view *types.View,
	futureRounds bool,
) (types.Subscription[*types.MsgRoundChange], func()) {
	return f.MsgStore.RoundChangeMessages.Subscribe(view, futureRounds)
}
