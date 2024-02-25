package test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/madz-lab/go-ibft"
)

type IBFTVerifier struct {
	network IBFTNetwork
}

func NewIBFTVerifier(network IBFTNetwork) ibft.Verifier {
	return IBFTVerifier{network}
}

func (v IBFTVerifier) HasValidSignature(msg ibft.Message) bool {
	from := msg.GetFrom()
	sig := msg.GetSignature()
	digest := crypto.Keccak256(msg.Payload())

	pubKey, err := crypto.SigToPub(digest, sig)
	if err != nil {
		panic(fmt.Errorf("failed to extract pub key: %w", err).Error())
	}

	if !bytes.Equal(from, crypto.PubkeyToAddress(*pubKey).Bytes()) {
		return false
	}

	return true
}

func (v IBFTVerifier) IsValidator(id []byte, sequence uint64) bool {
	return true
}

func (v IBFTVerifier) IsValidProposal(proposal []byte, sequence uint64) bool {
	if sequence != binary.BigEndian.Uint64(proposal[:8]) {
		return false
	}

	if !bytes.Equal(proposal[8:], Block) {
		return false
	}

	return true
}

func (v IBFTVerifier) IsProposer(id []byte, _, round uint64) bool {
	set := v.network.ValidatorSet()
	num := len(set)
	idx := int(round) % num

	if !bytes.Equal(id, set[idx].ID()) {
		return false
	}

	return true
}
