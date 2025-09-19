//nolint:dupl //because
package sequencer

import (
	"context"
	"errors"

	"github.com/sig-0/go-ibft/message"
)

var (
	Alice = []byte("Alice")
	Bob   = []byte("Bob")
	Chris = []byte("Chris")
	Nina  = []byte("Nina")

	DummyKeccak      = KeccakFn(func(_ []byte) []byte { return DummyKeccakValue })
	DummyKeccakValue = []byte("keccak")
	DummySignFn      = func(_ []byte) []byte { return nil }
)

type dummyTransport struct{}

func (t dummyTransport) MulticastProposal(_ *message.Proposal) {}

func (t dummyTransport) MulticastPrepare(_ *message.Prepare) {}

func (t dummyTransport) MulticastCommit(_ *message.Commit) {}

func (t dummyTransport) MulticastRoundChange(_ *message.RoundChange) {}

type mockValidator struct {
	signFn            func([]byte) []byte
	buildProposalFn   func(uint64) []byte
	isValidProposalFn func(uint64, []byte) bool
	address           []byte
}

func (v mockValidator) Address() []byte {
	return v.address
}

func (v mockValidator) Sign(digest []byte) []byte {
	return v.signFn(digest)
}

func (v mockValidator) BuildProposal(sequence uint64) []byte {
	return v.buildProposalFn(sequence)
}

type mockVerifier struct {
	isValidatorFn      func([]byte, uint64) bool
	isProposerFn       func([]byte, uint64, uint64) bool
	hasQuorumFn        func([][]byte, uint64) bool
	isValidProposalFn  func([]byte, uint64) bool
	isValidSignatureFn func([]byte, []byte, []byte) error
}

func (vs mockVerifier) IsValidProposal(proposal []byte, sequence uint64) bool {
	return vs.isValidProposalFn(proposal, sequence)
}

func (vs mockVerifier) IsValidSignature(signer, digest, signature []byte) error {
	return vs.isValidSignatureFn(signer, digest, signature)
}

func (vs mockVerifier) IsValidator(addr []byte, sequence uint64) bool {
	return vs.isValidatorFn(addr, sequence)
}

func (vs mockVerifier) IsProposer(addr []byte, sequence, round uint64) bool {
	return vs.isProposerFn(addr, sequence, round)
}

func (vs mockVerifier) HasQuorum(addresses [][]byte, sequence uint64) bool {
	return vs.hasQuorumFn(addresses, sequence)
}

type mockSignatureVerifier func([]byte, []byte, []byte) error

func (s mockSignatureVerifier) Verify(signature, digest, msg []byte) error {
	return s(signature, digest, msg)
}

type mockProposerAlgo func(ctx context.Context, sequence, round uint64) ([]byte, error)

func (m mockProposerAlgo) GetProposer(ctx context.Context, sequence, round uint64) ([]byte, error) {
	return m(ctx, sequence, round)
}

type allGoodVrf struct {
}

func (m allGoodVrf) CheckProposal(ctx context.Context, sequence Sequence, messages []*message.Proposal) (*message.Proposal, error) {
	if len(messages) == 0 {
		return nil, errors.New("messages is empty")
	}

	return messages[0], nil
}

func (m allGoodVrf) CheckPrepare(ctx context.Context, sequence Sequence, messages []*message.Prepare) ([]*message.Prepare, error) {
	if len(messages) == 0 {
		return nil, errors.New("messages is empty")
	}

	return messages, nil
}

func (m allGoodVrf) CheckCommit(ctx context.Context, sequence Sequence, messages []*message.Commit) ([]*message.Commit, error) {
	if len(messages) == 0 {
		return nil, errors.New("messages is empty")
	}

	return messages, nil
}

func (m allGoodVrf) CheckRoundChange(ctx context.Context, sequence Sequence, messages []*message.RoundChange) ([]*message.RoundChange, error) {
	if len(messages) == 0 {
		return nil, errors.New("messages is empty")
	}

	return messages, nil
}

type mockValidatorSet struct {
	getValidatorsFn func(ctx context.Context, sequence uint64) ([][]byte, error)
	getProposerFn   func(ctx context.Context, sequence, round uint64) ([]byte, error)
}

func (m mockValidatorSet) GetValidators(ctx context.Context, sequence uint64) ([][]byte, error) {
	return m.getValidatorsFn(ctx, sequence)
}

func (m mockValidatorSet) GetProposer(ctx context.Context, sequence, round uint64) ([]byte, error) {
	return m.getProposerFn(ctx, sequence, round)
}
