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
	cdc Codec

	proposal collection[types.MsgProposal]
}

func New(cdc Codec) *Store {
	s := &Store{
		cdc:      cdc,
		proposal: newCollection[types.MsgProposal](),
	}

	return s
}

func (s *Store) AddMsgProposal(msg *types.MsgProposal) error {
	if !bytes.Equal(msg.From, s.cdc.RecoverFrom(msg.Payload(), msg.Signature)) {
		return ErrInvalidSignature
	}

	s.proposal.addMessage(msg, msg.View, msg.From)

	return nil
}

func (s *Store) GetProposalMessages(view *types.View) []*types.MsgProposal {
	return s.proposal.getMessages(view)
}
