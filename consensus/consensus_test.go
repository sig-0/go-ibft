package consensus

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_InitSequence_ExternalError(t *testing.T) {
	t.Parallel()

	err := errors.New("external error")
	vs := theRealMockVS{getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
		return nil, err
	}}

	cons := New(vs, nil, nil)
	assert.ErrorIs(t, err, cons.InitSequence(context.Background(), 101))
}

func Test_InitSequence_Ok(t *testing.T) {
	t.Parallel()

	vs := theRealMockVS{getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
		return [][]byte{[]byte("alice"), []byte("bob")}, nil
	}}

	cons := New(vs, nil, nil)
	require.NoError(t, cons.InitSequence(context.Background(), 101))

	require.Len(t, cons.currentValidators, 2)

	_, ok := cons.currentValidators["alice"]
	assert.True(t, ok)
	_, ok = cons.currentValidators["bob"]
	assert.True(t, ok)
}

type theRealMockVS struct {
	getProposersFn  func(ctx context.Context, sequence, round uint64) ([]byte, error)
	getValidatorsFn func(ctx context.Context, sequence uint64) ([][]byte, error)
	checkQuorumFn   func(ctx context.Context, sequence uint64, validators [][]byte) (bool, error)
}

func (b theRealMockVS) GetProposer(ctx context.Context, sequence, round uint64) ([]byte, error) {
	return b.getProposersFn(ctx, sequence, round)
}

func (b theRealMockVS) GetValidators(ctx context.Context, sequence uint64) ([][]byte, error) {
	return b.getValidatorsFn(ctx, sequence)
}

func (b theRealMockVS) CheckQuorum(ctx context.Context, sequence uint64, validators [][]byte) (bool, error) {
	return b.checkQuorumFn(ctx, sequence, validators)
}

type mockValidatorSet struct {
	proposer   []byte
	validators [][]byte
	minQuorum  int
}

func (vs mockValidatorSet) GetProposer(ctx context.Context, sequence, round uint64) ([]byte, error) {
	return vs.proposer, nil
}

func (vs mockValidatorSet) GetValidators(ctx context.Context, sequence uint64) ([][]byte, error) {
	return vs.validators, nil
}

func (vs mockValidatorSet) CheckQuorum(ctx context.Context, sequence uint64, validators [][]byte) (bool, error) {
	return len(validators) >= vs.minQuorum, nil
}

type mockProposalVerifier struct {
	valid bool
}

func (vrf mockProposalVerifier) VerifyProposal(ctx context.Context, sequence uint64, proposal []byte) error {
	if !vrf.valid {
		return errors.New("invalid proposal")
	}

	return nil
}

type mockDeriver struct {
	addr []byte
}

func (m mockDeriver) DeriveSender(digest, signature []byte) ([]byte, error) {
	return m.addr, nil
}

type mockSignatureVerifier struct {
	valid bool
}

func (vrf mockSignatureVerifier) Verify(sender, digest, signature []byte) error {
	if !vrf.valid {
		return errors.New("invalid signature")
	}

	return nil
}
