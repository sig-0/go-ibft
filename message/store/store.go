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

type Store struct {
	verifier msgVerifier

	proposal    *syncCollection[types.MsgProposal]
	prepare     *syncCollection[types.MsgPrepare]
	commit      *syncCollection[types.MsgCommit]
	roundChange *syncCollection[types.MsgRoundChange]
}

func New(recover types.SigRecover) *Store {
	s := &Store{
		verifier:    msgVerifier{recover},
		proposal:    newSyncCollection[types.MsgProposal](),
		prepare:     newSyncCollection[types.MsgPrepare](),
		commit:      newSyncCollection[types.MsgCommit](),
		roundChange: newSyncCollection[types.MsgRoundChange](),
	}

	return s
}

func (s *Store) AddMessage(msg types.Msg) error {
	if err := s.isValidSignature(msg); err != nil {
		return err
	}

	switch msg := msg.(type) {
	case *types.MsgProposal:
		s.proposal.addMessage(msg, msg.View, msg.From)
	case *types.MsgPrepare:
		s.prepare.addMessage(msg, msg.View, msg.From)
	case *types.MsgCommit:
		s.commit.addMessage(msg, msg.View, msg.From)
	case *types.MsgRoundChange:
		s.roundChange.addMessage(msg, msg.View, msg.From)
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
	return s.proposal.getMessages(view)
}

func (s *Store) RemoveProposalMessages(view *types.View) {
	s.proposal.remove(view)
}

/*	MsgPrepare	*/

func (s *Store) GetPrepareMessages(view *types.View) []*types.MsgPrepare {
	return s.prepare.getMessages(view)
}

func (s *Store) RemovePrepareMessages(view *types.View) {
	s.prepare.remove(view)
}

/*	MsgCommit	*/

func (s *Store) GetCommitMessages(view *types.View) []*types.MsgCommit {
	return s.commit.getMessages(view)
}

func (s *Store) RemoveCommitMessages(view *types.View) {
	s.commit.remove(view)
}

/*	MsgRoundChange	*/

func (s *Store) GetRoundChangeMessages(view *types.View) []*types.MsgRoundChange {
	return s.roundChange.getMessages(view)
}

func (s *Store) RemoveRoundChangeMessages(view *types.View) {
	s.roundChange.remove(view)
}
