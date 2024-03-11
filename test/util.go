package test

import (
	"bytes"
	"math"

	"golang.org/x/crypto/sha3"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

func QuorumOf(n int) ibft.QuorumFn {
	var quorum int
	if (n-1)/3 == 0 {
		quorum = n
	} else {
		quorum = int(math.Ceil(2 * float64(n) / 3))
	}

	return func(messages []types.Message) bool {
		return len(messages) >= quorum
	}
}

func DefaultKeccak() ibft.KeccakFn {
	return func(data []byte) []byte {
		hash := sha3.New256()
		hash.Write(data)

		return hash.Sum(data)
	}
}

type Network chan types.Message

func NewNetwork() Network {
	return make(Network)
}

func GetTransport[M types.IBFTMessage](n Network) ibft.TransportFn[M] {
	return func(msg M) { n <- any(msg).(types.Message) } //nolint:forcetypeassert // guarded by types.IBFTMessage
}

func (n Network) Gossip(fn func(msg types.Message)) {
	for msg := range n {
		fn(msg)
	}
}

type ValidatorSet map[string]*IBFTValidator

func NewValidatorSet(names ...string) ValidatorSet {
	vs := make(ValidatorSet)
	for _, name := range names {
		vs[name] = &IBFTValidator{Name: name}
	}

	proposerAlgo := RoundRobinProposer(vs)
	for _, v := range vs {
		v.IsProposerFn = proposerAlgo
	}

	return vs
}

func RoundRobinProposer(vs ValidatorSet) func([]byte, uint64, uint64) bool {
	validators := make([]*IBFTValidator, 0, len(vs))

	for _, v := range vs {
		validators = append(validators, v)
	}

	return func(id []byte, _, r uint64) bool {
		num := len(validators)
		elected := int(r) % num

		return bytes.Equal(id, validators[elected].ID())
	}
}

type IBFTValidator struct {
	IsProposerFn func([]byte, uint64, uint64) bool
	Name         string
}

func (v IBFTValidator) Sign(digest []byte) []byte {
	return append(digest, v.ID()...)
}

func (v IBFTValidator) IsValidSignature(sender, digest, sig []byte) bool {
	if !bytes.Equal(digest, sig[:len(digest)]) {
		return false
	}

	if !bytes.Equal(sender, sig[len(digest):]) {
		return false
	}

	return true
}

func (v IBFTValidator) IsValidator(_ []byte, _ uint64) bool {
	return true
}

func (v IBFTValidator) IsValidProposal(_ []byte, _ uint64) bool {
	return true
}

func (v IBFTValidator) IsProposer(id []byte, sequence, round uint64) bool {
	return v.IsProposerFn(id, sequence, round)
}

func (v IBFTValidator) ID() []byte {
	return []byte(v.Name)
}

func (v IBFTValidator) BuildProposal(_ uint64) []byte {
	return []byte("valid block")
}
