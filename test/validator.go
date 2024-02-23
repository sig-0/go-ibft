package test

import "encoding/binary"

var (
	Block            = []byte("block")
	ValidBlockByte   = byte(1)
	InvalidBlockByte = byte(0)
)

type IBFTValidator struct {
	id []byte
}

func (v IBFTValidator) Sign(data []byte) []byte {
	//TODO implement me
	panic("implement me")
}

func (v IBFTValidator) ID() []byte {
	return v.id
}

func (v IBFTValidator) BuildProposal(sequence uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, sequence)

	return append(buf, Block...)
}

type SilentValidator struct {
	IBFTValidator
}

func (v SilentValidator) BuildProposal(_ uint64) []byte {
	return nil
}

type HonestValidator struct {
	IBFTValidator
}

func (v HonestValidator) BuildProposal(sequence uint64) []byte {
	return append(v.IBFTValidator.BuildProposal(sequence), ValidBlockByte)
}

type MaliciousValidator struct {
	IBFTValidator
}

func (v MaliciousValidator) BuildProposal(sequence uint64) []byte {
	return append(v.IBFTValidator.BuildProposal(sequence), InvalidBlockByte)
}
