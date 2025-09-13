package message

import (
	"errors"
	"fmt"
)

var ErrInvalidMessage = errors.New("invalid consensus message")

// MsgStore is a thread-safe storage for consensus messages with a built-in sequencer.Feed mechanism
type MsgStore struct {
	sigVerifier SignatureVerifier

	ProposalMessages    *MsgCollection[*MsgProposal]
	PrepareMessages     *MsgCollection[*MsgPrepare]
	CommitMessages      *MsgCollection[*MsgCommit]
	RoundChangeMessages *MsgCollection[*MsgRoundChange]
}

// NewMsgStore returns a new MsgStore instance
func NewMsgStore(vrf SignatureVerifier) *MsgStore {
	return &MsgStore{
		sigVerifier:         vrf,
		ProposalMessages:    NewMsgCollection[*MsgProposal](),
		PrepareMessages:     NewMsgCollection[*MsgPrepare](),
		CommitMessages:      NewMsgCollection[*MsgCommit](),
		RoundChangeMessages: NewMsgCollection[*MsgRoundChange](),
	}
}

// Add includes the message in the store
func (s *MsgStore) Add(msg Message) error {
	info := msg.GetInfo()
	if info == nil {
		return fmt.Errorf("%w: missing info field", ErrInvalidMessage)
	}

	if info.Sender == nil {
		return fmt.Errorf("%w: missing sender field", ErrInvalidMessage)
	}

	if info.Signature == nil {
		return fmt.Errorf("%w: missing signature field", ErrInvalidMessage)
	}

	switch msg := msg.(type) {
	case *MsgProposal:
		if msg.BlockHash == nil {
			return fmt.Errorf("%w: missing block_hash field", ErrInvalidMessage)
		}

		if msg.ProposedBlock == nil {
			return fmt.Errorf("%w: missing proposed_block field", ErrInvalidMessage)
		}

		if err := s.sigVerifier.Verify(msg.GetInfo().Sender, msg.Payload(), msg.GetInfo().Signature); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidMessage, err)
		}

		s.ProposalMessages.Add(msg)
	case *MsgPrepare:
		if msg.BlockHash == nil {
			return fmt.Errorf("%w: missing block_hash field", ErrInvalidMessage)
		}

		if err := s.sigVerifier.Verify(msg.GetInfo().Sender, msg.Payload(), msg.GetInfo().Signature); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidMessage, err)
		}

		s.PrepareMessages.Add(msg)
	case *MsgCommit:
		if msg.BlockHash == nil {
			return fmt.Errorf("%w: missing block_hash field", ErrInvalidMessage)
		}

		if msg.CommitSeal == nil {
			return fmt.Errorf("%w: missing commit_seal field", ErrInvalidMessage)
		}

		if err := s.sigVerifier.Verify(msg.GetInfo().Sender, msg.Payload(), msg.GetInfo().Signature); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidMessage, err)
		}

		s.CommitMessages.Add(msg)
	case *MsgRoundChange:
		if err := s.sigVerifier.Verify(msg.GetInfo().Sender, msg.Payload(), msg.GetInfo().Signature); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidMessage, err)
		}

		s.RoundChangeMessages.Add(msg)
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

func (s *MsgStore) Feed() Feed {
	return Feed{s}
}

type Feed struct {
	*MsgStore
}

func (f Feed) SubscribeProposal(
	sequence, round uint64,
	futureRounds bool,
) (chan func() []*MsgProposal, func()) {
	return f.MsgStore.ProposalMessages.Subscribe(sequence, round, futureRounds)
}

func (f Feed) SubscribePrepare(
	sequence, round uint64,
	futureRounds bool,
) (chan func() []*MsgPrepare, func()) {
	return f.MsgStore.PrepareMessages.Subscribe(sequence, round, futureRounds)
}

func (f Feed) SubscribeCommit(
	sequence, round uint64,
	futureRounds bool,
) (chan func() []*MsgCommit, func()) {
	return f.MsgStore.CommitMessages.Subscribe(sequence, round, futureRounds)
}

func (f Feed) SubscribeRoundChange(
	sequence, round uint64,
	futureRounds bool,
) (chan func() []*MsgRoundChange, func()) {
	return f.MsgStore.RoundChangeMessages.Subscribe(sequence, round, futureRounds)
}
