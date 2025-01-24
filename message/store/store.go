package store

import (
	"errors"
	"fmt"

	"github.com/sig-0/go-ibft/message"
)

var ErrInvalidMessage = errors.New("invalid consensus message")

// MsgStore is a thread-safe storage for consensus messages with a built-in sequencer.Feed mechanism
type MsgStore struct {
	sigVerifier message.SignatureVerifier

	ProposalMessages    *MsgCollection[*message.MsgProposal]
	PrepareMessages     *MsgCollection[*message.MsgPrepare]
	CommitMessages      *MsgCollection[*message.MsgCommit]
	RoundChangeMessages *MsgCollection[*message.MsgRoundChange]
}

// NewMsgStore returns a new MsgStore instance
func NewMsgStore(vrf message.SignatureVerifier) *MsgStore {
	return &MsgStore{
		sigVerifier:         vrf,
		ProposalMessages:    NewMsgCollection[*message.MsgProposal](),
		PrepareMessages:     NewMsgCollection[*message.MsgPrepare](),
		CommitMessages:      NewMsgCollection[*message.MsgCommit](),
		RoundChangeMessages: NewMsgCollection[*message.MsgRoundChange](),
	}
}

// Add includes the message in the store
func (s *MsgStore) Add(msg message.Message) error {
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
	case *message.MsgProposal:
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
	case *message.MsgPrepare:
		if msg.BlockHash == nil {
			return fmt.Errorf("%w: missing block_hash field", ErrInvalidMessage)
		}

		if err := s.sigVerifier.Verify(msg.GetInfo().Sender, msg.Payload(), msg.GetInfo().Signature); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidMessage, err)
		}

		s.PrepareMessages.Add(msg)
	case *message.MsgCommit:
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
	case *message.MsgRoundChange:
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
) (chan func() []*message.MsgProposal, func()) {
	return f.MsgStore.ProposalMessages.Subscribe(sequence, round, futureRounds)
}

func (f Feed) SubscribePrepare(
	sequence, round uint64,
	futureRounds bool,
) (chan func() []*message.MsgPrepare, func()) {
	return f.MsgStore.PrepareMessages.Subscribe(sequence, round, futureRounds)
}

func (f Feed) SubscribeCommit(
	sequence, round uint64,
	futureRounds bool,
) (chan func() []*message.MsgCommit, func()) {
	return f.MsgStore.CommitMessages.Subscribe(sequence, round, futureRounds)
}

func (f Feed) SubscribeRoundChange(
	sequence, round uint64,
	futureRounds bool,
) (chan func() []*message.MsgRoundChange, func()) {
	return f.MsgStore.RoundChangeMessages.Subscribe(sequence, round, futureRounds)
}
