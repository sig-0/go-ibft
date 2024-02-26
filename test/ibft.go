package test

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/madz-lab/go-ibft"
)

var (
	Block = []byte("block")

	DefaultKeccak ibft.Keccak = KeccakFn(func(data []byte) []byte {
		return crypto.Keccak256(data)
	})
	//ECRecover ibft.SigRecover = SigRecoverFn(func(digest, sig []byte) []byte {
	//	pubKey, err := crypto.SigToPub(digest, sig)
	//	if err != nil {
	//		panic(fmt.Errorf("failed to extract pub key: %w", err).Error())
	//	}
	//
	//	return crypto.PubkeyToAddress(*pubKey).Bytes()
	//})
)

type ECDSAKey struct {
	*ecdsa.PrivateKey
}

func NewECDSAKey() ECDSAKey {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic(fmt.Errorf("failed to generate ecdsa key: %w", err).Error())
	}

	return ECDSAKey{key}
}

func (k ECDSAKey) Sign(digest []byte) []byte {
	sig, err := crypto.Sign(digest, k.PrivateKey)
	if err != nil {
		panic(fmt.Errorf("failed to sign: %w", err).Error())
	}

	return sig
}

func (k ECDSAKey) ID() []byte {
	pubKeyECDSA, ok := k.Public().(*ecdsa.PublicKey)
	if !ok {
		panic("failed to cast pub key to *ecdsa.PublicKey")
	}

	return crypto.PubkeyToAddress(*pubKeyECDSA).Bytes()
}

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

type IBFTVerifier struct {
	network IBFTNetwork
}

func NewIBFTVerifier(network IBFTNetwork) ibft.Verifier {
	return IBFTVerifier{network}
}

func (v IBFTVerifier) IsValidSignature(sender []byte, digest []byte, sig []byte) bool {
	pubKey, err := crypto.SigToPub(digest, sig)
	if err != nil {
		panic(fmt.Errorf("failed to extract pub key: %w", err).Error())
	}

	if !bytes.Equal(sender, crypto.PubkeyToAddress(*pubKey).Bytes()) {
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

type (
	TransportFn  func(ibft.Message)
	QuorumFn     func([]ibft.Message) bool
	SigRecoverFn func(digest, sig []byte) []byte
	KeccakFn     func([]byte) []byte
)

func (f TransportFn) Multicast(msg ibft.Message) {
	f(msg)
}

func (f QuorumFn) HasQuorum(messages []ibft.Message) bool {
	return f(messages)
}

func (f SigRecoverFn) From(digest, sig []byte) []byte {
	return f(digest, sig)
}

func (f KeccakFn) Hash(data []byte) []byte {
	return f(data)
}
