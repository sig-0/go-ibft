package mock

import (
	"bytes"
	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

var (
	DummyMsgProposalTransport    = ibft.TransportFn[*types.MsgProposal](func(_ *types.MsgProposal) {})
	DummyMsgPrepareTransport     = ibft.TransportFn[*types.MsgPrepare](func(_ *types.MsgPrepare) {})
	DummyMsgCommitTransport      = ibft.TransportFn[*types.MsgCommit](func(_ *types.MsgCommit) {})
	DummyMsgRoundChangeTransport = ibft.TransportFn[*types.MsgRoundChange](func(_ *types.MsgRoundChange) {})

	DummyKeccak = ibft.KeccakFn(func(_ []byte) []byte { return []byte("block hash") })

	DummyValidatorIDFn = func() []byte { return []byte("my validator") }
	DummySigner        = ibft.SignerFn(func(_ []byte) []byte { return []byte("signature") })

	NoQuorum      = ibft.QuorumFn(func(_ []ibft.Message) bool { return false })
	NonZeroQuorum = ibft.QuorumFn(func(messages []ibft.Message) bool { return len(messages) > 0 })

	AlwaysValidBlock     = func(_ []byte, _ uint64) bool { return true }
	AlwaysValidSignature = func(_, _, _ []byte) bool { return true }
	AlwaysValidator      = func(_ []byte, _ uint64) bool { return true }
)

type ValidatorSet map[string]struct{}

func NewValidatorSet(ids ...string) ValidatorSet {
	vs := make(ValidatorSet)

	for _, id := range ids {
		vs[id] = struct{}{}
	}

	return vs
}

func (vs ValidatorSet) IsValidator(id []byte, _ uint64) bool {
	_, ok := vs[string(id)]
	return ok
}

func NewDummyKeccak(digest string) ibft.KeccakFn {
	return func(_ []byte) []byte {
		return []byte(digest)
	}
}

type ValidatorID []byte

func NewValidatorID(id string) ValidatorID {
	return []byte(id)
}

func (id ValidatorID) ID() []byte {
	return id
}

func QuorumOf(n int) ibft.QuorumFn {
	return func(messages []ibft.Message) bool {
		return len(messages) == n
	}
}

type Proposer struct {
	ID    []byte
	Round uint64
}

func ProposersInRounds(proposers ...Proposer) func([]byte, uint64, uint64) bool {
	pp := make(map[uint64][]byte)
	for _, p := range proposers {
		pp[p.Round] = p.ID
	}

	return func(id []byte, _, round uint64) bool {
		return bytes.Equal(id, pp[round])
	}
}

type Validator struct {
	ibft.Signer
	ibft.Verifier

	IDFn            func() []byte
	BuildProposalFn func(uint64) []byte
}

func (v Validator) ID() []byte {
	return v.IDFn()
}

func (v Validator) BuildProposal(sequence uint64) []byte {
	return v.BuildProposalFn(sequence)
}

type Verifier struct {
	IsValidSignatureFn func([]byte, []byte, []byte) bool
	IsValidatorFn      func([]byte, uint64) bool
	IsValidProposalFn  func([]byte, uint64) bool
	IsProposerFn       func([]byte, uint64, uint64) bool
}

func (v Verifier) IsValidSignature(sender, digest, sig []byte) bool {
	return v.IsValidSignatureFn(sender, digest, sig)
}

func (v Verifier) IsValidator(id []byte, sequence uint64) bool {
	return v.IsValidatorFn(id, sequence)
}

func (v Verifier) IsValidProposal(proposal []byte, sequence uint64) bool {
	return v.IsValidProposalFn(proposal, sequence)
}

func (v Verifier) IsProposer(id []byte, sequence, round uint64) bool {
	return v.IsProposerFn(id, sequence, round)
}
