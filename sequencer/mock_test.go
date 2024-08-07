package sequencer

import (
	"github.com/sig-0/go-ibft/message"
)

type MockValidator struct {
	address           []byte
	signFn            func([]byte) []byte
	buildProposalFn   func(uint64) []byte
	isValidProposalFn func(uint64, []byte) bool
}

func (v MockValidator) Address() []byte {
	return v.address
}

func (v MockValidator) Sign(digest []byte) []byte {
	return v.signFn(digest)
}

func (v MockValidator) BuildProposal(sequence uint64) []byte {
	return v.buildProposalFn(sequence)
}

func (v MockValidator) IsValidProposal(proposal []byte, sequence uint64) bool {
	return v.isValidProposalFn(sequence, proposal)
}

type MockValidatorSet struct {
	isValidatorFn func([]byte, uint64) bool
	isProposerFn  func([]byte, uint64, uint64) bool
	hasQuorumFn   func([]message.Message) bool
}

func (vs MockValidatorSet) IsValidator(addr []byte, sequence uint64) bool {
	return vs.isValidatorFn(addr, sequence)
}

func (vs MockValidatorSet) IsProposer(addr []byte, sequence, round uint64) bool {
	return vs.isProposerFn(addr, sequence, round)
}

func (vs MockValidatorSet) HasQuorum(messages []message.Message) bool {
	return vs.hasQuorumFn(messages)
}

type MockKeccak func([]byte) []byte

func (k MockKeccak) Hash(digest []byte) []byte {
	return k(digest)
}

type MockSignatureVerifier func([]byte, []byte, []byte) error

func (s MockSignatureVerifier) Verify(signature, digest []byte, msg []byte) error {
	return s(signature, digest, msg)
}
