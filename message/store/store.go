package store

import (
	"bytes"
	"errors"
	"github.com/madz-lab/go-ibft/message/types"
)

var (
	ErrInvalidSignature = errors.New("invalid signature")
)

type Codec interface {
	RecoverFrom(data []byte, sig []byte) []byte
}

type Store struct {
	cdc Codec // todo: wrap

	proposal     collection[types.MsgProposal] // todo: replace with thread-safe version
	proposalSubs subscriptions[types.MsgProposal]
}

func New(cdc Codec) *Store {
	s := &Store{
		cdc:          cdc,
		proposal:     newCollection[types.MsgProposal](),
		proposalSubs: newSubscriptions[types.MsgProposal](),
	}

	return s
}

func (s *Store) AddMsgProposal(msg *types.MsgProposal) error {
	if !bytes.Equal(msg.From, s.cdc.RecoverFrom(msg.Payload(), msg.Signature)) {
		return ErrInvalidSignature
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
