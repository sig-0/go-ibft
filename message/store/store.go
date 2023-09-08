package store

import (
	"bytes"
	"errors"

	"github.com/madz-lab/go-ibft/message/types"
)

var (
	ErrUnknownType      = errors.New("unknown message type")
	ErrInvalidSignature = errors.New("invalid signature")
)

type msgVerifier struct {
	types.SigRecover
}

// Store is a thread-safe storage for consensus messages with a built-in Feed mechanism
type Store struct {
	verifier msgVerifier

	proposal    *syncCollection[types.MsgProposal]
	prepare     *syncCollection[types.MsgPrepare]
	commit      *syncCollection[types.MsgCommit]
	roundChange *syncCollection[types.MsgRoundChange]
}

// New returns a new Store instance. Messages added to this store
// have their signatures verified before being included
func New(sigRecover types.SigRecover) *Store {
	s := &Store{
		verifier:    msgVerifier{sigRecover},
		proposal:    newSyncCollection[types.MsgProposal](),
		prepare:     newSyncCollection[types.MsgPrepare](),
		commit:      newSyncCollection[types.MsgCommit](),
		roundChange: newSyncCollection[types.MsgRoundChange](),
	}

	return s
}

// AddMessage stores the provided msg if its signature is valid
func (s *Store) AddMessage(msg types.Msg) error {
	if err := s.isValidSignature(msg); err != nil {
		return err
	}

	switch msg := msg.(type) {
	case *types.MsgProposal:
		s.proposal.AddMessage(msg, msg.View, msg.From)
	case *types.MsgPrepare:
		s.prepare.AddMessage(msg, msg.View, msg.From)
	case *types.MsgCommit:
		s.commit.AddMessage(msg, msg.View, msg.From)
	case *types.MsgRoundChange:
		s.roundChange.AddMessage(msg, msg.View, msg.From)
	default:
		return ErrUnknownType
	}

	return nil
}

func (s *Store) isValidSignature(msg types.Msg) error {
	if !bytes.Equal(msg.GetFrom(), s.verifier.From(msg.Payload(), msg.GetSignature())) {
		return ErrInvalidSignature
	}

	return nil
}

/*	MsgProposal	*/

func (s *Store) GetProposalMessages(view *types.View) []*types.MsgProposal {
	return s.proposal.GetMessages(view)
}

func (s *Store) RemoveProposalMessages(view *types.View) {
	s.proposal.Remove(view)
}

/*	MsgPrepare	*/

func (s *Store) GetPrepareMessages(view *types.View) []*types.MsgPrepare {
	return s.prepare.GetMessages(view)
}

func (s *Store) RemovePrepareMessages(view *types.View) {
	s.prepare.Remove(view)
}

/*	MsgCommit	*/

func (s *Store) GetCommitMessages(view *types.View) []*types.MsgCommit {
	return s.commit.GetMessages(view)
}

func (s *Store) RemoveCommitMessages(view *types.View) {
	s.commit.Remove(view)
}

/*	MsgRoundChange	*/

func (s *Store) GetRoundChangeMessages(view *types.View) []*types.MsgRoundChange {
	return s.roundChange.GetMessages(view)
}

func (s *Store) RemoveRoundChangeMessages(view *types.View) {
	s.roundChange.Remove(view)
}
