package test

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/madz-lab/go-ibft"
)

const Round0Timeout = 10 * time.Second

var (
	DefaultKeccak ibft.Keccak = KeccakFn(func(data []byte) []byte {
		return crypto.Keccak256(data)
	})
	ECRecover ibft.SigRecover = SigRecoverFn(func(digest, sig []byte) []byte {
		pubKey, err := crypto.SigToPub(digest, sig)
		if err != nil {
			panic(fmt.Errorf("failed to extract pub key: %w", err).Error())
		}

		return crypto.PubkeyToAddress(*pubKey).Bytes()
	})
)

type (
	TransportFn  func(ibft.Message)
	QuorumFn     func(uint64, []ibft.Message) bool
	SigRecoverFn func(digest, sig []byte) []byte
	KeccakFn     func([]byte) []byte
)

func (f TransportFn) Multicast(msg ibft.Message) {
	f(msg)
}

func (f QuorumFn) HasQuorum(sequence uint64, messages []ibft.Message) bool {
	return f(sequence, messages)
}

func (f SigRecoverFn) From(digest, sig []byte) []byte {
	return f(digest, sig)
}

func (f KeccakFn) Hash(data []byte) []byte {
	return f(data)
}

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
