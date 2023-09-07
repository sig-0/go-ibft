package store

import (
	"bytes"
	"errors"
	"github.com/madz-lab/go-ibft/message/types"
)

var (
	ErrInvalidSignature = errors.New("invalid signature")
)

type msgVerifier struct {
	types.SigRecover
}

type Store struct {
	verifier msgVerifier

	proposal     collection[types.MsgProposal] // todo: replace with thread-safe version
	proposalSubs subscriptions[types.MsgProposal]
}

func New(recover types.SigRecover) *Store {
	s := &Store{
		verifier:     msgVerifier{recover},
		proposal:     newCollection[types.MsgProposal](),
		proposalSubs: newSubscriptions[types.MsgProposal](),
	}

	return s
}

func (s *Store) isValidSignature(msg types.Msg) error {
	if !bytes.Equal(msg.GetFrom(), s.verifier.From(msg.Payload(), msg.GetSignature())) {
		return ErrInvalidSignature
	}

	return nil
}

func (s *Store) AddMsgProposal(msg *types.MsgProposal) error {
	if err := s.isValidSignature(msg); err != nil {
		return err
	}

	s.proposal.addMessage(msg, msg.View, msg.From)

	s.proposalSubs.notify(func(sub subscription[types.MsgProposal]) {
		if sub.View.Sequence != msg.View.Sequence {
			return
		}

		if sub.View.Round < msg.View.Round {
			return
		}

		sub.notify(s.proposal.unwrapFn(sub.View, sub.FutureRounds))
	})

	return nil
}

func (s *Store) GetProposalMessages(view *types.View) []*types.MsgProposal {
	return s.proposal.getMessages(view)
}
