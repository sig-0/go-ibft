package test

import (
	"encoding/binary"
)

var (
	Block            = []byte("block")
	ValidBlockByte   = byte(1)
	InvalidBlockByte = byte(0)
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

type SilentValidator IBFTValidator

func (v SilentValidator) Sign(keccak []byte) []byte {
	//TODO implement me
	panic("implement me")
}

func (v SilentValidator) ID() []byte {
	//TODO implement me
	panic("implement me")
}

func (v SilentValidator) BuildProposal(_ uint64) []byte {
	return nil
}

type HonestValidator IBFTValidator

func (v HonestValidator) Sign(keccak []byte) []byte {
	return IBFTValidator(v).Sign(keccak)
}

func (v HonestValidator) ID() []byte {
	return IBFTValidator(v).ID()
}

func (v HonestValidator) BuildProposal(sequence uint64) []byte {
	return append(IBFTValidator(v).BuildProposal(sequence), ValidBlockByte)
}

type MaliciousValidator IBFTValidator

func (v MaliciousValidator) BuildProposal(sequence uint64) []byte {
	return append(IBFTValidator(v).BuildProposal(sequence), InvalidBlockByte)
}
