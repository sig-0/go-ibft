package test

import (
	"encoding/binary"
)

var (
	Block          = []byte("block")
	ValidBlockByte = byte(1)
)

type IBFTValidator struct {
	key ECDSAKey
}

func NewIBFTValidator() IBFTValidator {
	return IBFTValidator{key: NewECDSAKey()}
}

func (v IBFTValidator) Sign(digest []byte) []byte {
	return v.key.Sign(digest)
}

func (v IBFTValidator) ID() []byte {
	return v.key.ID()
}

func (v IBFTValidator) BuildProposal(sequence uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, sequence)

	return append(buf, Block...)
}
