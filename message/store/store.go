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

	proposal map[uint64]map[uint64][]*types.MsgProposal
}

func (s *Store) AddMsgProposal(msg *types.MsgProposal) error {
	if !bytes.Equal(msg.From, s.cdc.RecoverFrom(msg.Payload(), msg.Signature)) {
		return errors.New("invalid signature")
	}

	return nil
}

func New(cdc Codec) *Store {
	s := &Store{
		cdc:      cdc,
		proposal: map[uint64]map[uint64][]*types.MsgProposal{},
	}

	return s
}
