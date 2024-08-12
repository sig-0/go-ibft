package store

import (
	"errors"
	"fmt"

	"github.com/sig-0/go-ibft/message"
)

var ErrInvalidMessage = errors.New("invalid consensus message")

// MsgStore is a thread-safe storage for consensus messages with a built-in sequencer.Feed mechanism
type MsgStore struct {
	ProposalMessages    MsgCollection[*message.MsgProposal]
	PrepareMessages     MsgCollection[*message.MsgPrepare]
	CommitMessages      MsgCollection[*message.MsgCommit]
	RoundChangeMessages MsgCollection[*message.MsgRoundChange]
}

// NewMsgStore returns a new MsgStore instance
func NewMsgStore() *MsgStore {
	return &MsgStore{
		ProposalMessages:    NewMsgCollection[*message.MsgProposal](),
		PrepareMessages:     NewMsgCollection[*message.MsgPrepare](),
		CommitMessages:      NewMsgCollection[*message.MsgCommit](),
		RoundChangeMessages: NewMsgCollection[*message.MsgRoundChange](),
	}
}

// Add includes the message in the store
func (s *MsgStore) Add(m message.Message) error {
	info := m.GetInfo()
	if info == nil {
		return fmt.Errorf("%w: missing info field", ErrInvalidMessage)
	}

	if info.Sender == nil {
		return fmt.Errorf("%w: missing sender field", ErrInvalidMessage)
	}

	if info.Signature == nil {
		return fmt.Errorf("%w: missing signature field", ErrInvalidMessage)
	}

	switch m := m.(type) {
	case *message.MsgProposal:
		if m.BlockHash == nil {
			return fmt.Errorf("%w: missing block_hash field", ErrInvalidMessage)
		}

		if m.ProposedBlock == nil {
			return fmt.Errorf("%w: missing proposed_block field", ErrInvalidMessage)
		}

		s.ProposalMessages.Add(m)
	case *message.MsgPrepare:
		if m.BlockHash == nil {
			return fmt.Errorf("%w: missing block_hash field", ErrInvalidMessage)
		}

		s.PrepareMessages.Add(m)
	case *message.MsgCommit:
		if m.BlockHash == nil {
			return fmt.Errorf("%w: missing block_hash field", ErrInvalidMessage)
		}

		if m.CommitSeal == nil {
			return fmt.Errorf("%w: missing commit_seal field", ErrInvalidMessage)
		}

		s.CommitMessages.Add(m)
	case *message.MsgRoundChange:
		s.RoundChangeMessages.Add(m)
	}

	return nil
}

// Clear removes all messages from store
func (s *MsgStore) Clear() {
	s.ProposalMessages.Clear()
	s.PrepareMessages.Clear()
	s.CommitMessages.Clear()
	s.RoundChangeMessages.Clear()
}
