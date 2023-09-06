package store

import (
	"bytes"
	"errors"
	"github.com/madz-lab/go-ibft/message/types"
)

type Codec interface {
	RecoverFrom(data []byte, sig []byte) []byte
}

type Store struct {
	cdc Codec

	proposal map[uint64]map[uint64]map[string]*types.MsgProposal
}

func New(cdc Codec) *Store {
	s := &Store{
		cdc:      cdc,
		proposal: map[uint64]map[uint64]map[string]*types.MsgProposal{},
	}

	return s
}

func (s *Store) AddMsgProposal(msg *types.MsgProposal) error {
	if !bytes.Equal(msg.From, s.cdc.RecoverFrom(msg.Payload(), msg.Signature)) {
		return errors.New("invalid signature")
	}

	// init once for every new sequence
	sequenceMessages, ok := s.proposal[msg.View.Sequence]
	if !ok {
		s.proposal[msg.View.Sequence] = map[uint64]map[string]*types.MsgProposal{}
		sequenceMessages = s.proposal[msg.View.Sequence]
	}

	// init once for every new round
	roundMessages, ok := sequenceMessages[msg.View.Round]
	if !ok {
		sequenceMessages[msg.View.Round] = map[string]*types.MsgProposal{}
		roundMessages = sequenceMessages[msg.View.Round]
	}

	roundMessages[string(msg.From)] = msg

	return nil
}

func (s *Store) GetProposalMessages(view *types.View) []*types.MsgProposal {
	// todo: nil cases

	messages := make([]*types.MsgProposal, 0)
	for _, msg := range s.proposal[view.Sequence][view.Round] {
		messages = append(messages, msg)

	}

	return messages
}
