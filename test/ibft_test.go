package test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AllValidProposals(network IBFTNetwork, proposals FinalizedProposals) bool {
	if len(network.Validators) != len(proposals) {
		return false // all validators must have finalized proposals
	}

	allProposals := make([][]byte, 0, len(proposals))
	for _, p := range proposals {
		allProposals = append(allProposals, p.Proposal)
	}

	p := allProposals[0]
	for _, pp := range allProposals[1:] {
		if !bytes.Equal(p, pp) {
			return false
		}
	}

	allRounds := make([]uint64, 0, len(proposals))
	for _, p := range proposals {
		allRounds = append(allRounds, p.Round)
	}

	r := allRounds[0]
	for _, round := range allRounds {
		if r != round {
			return false
		}
	}

	// unique number of seals
	seals := make(map[string][]byte)
	for _, p := range proposals {
		for _, seal := range p.Seals {
			seals[string(seal.From)] = seal.CommitSeal
		}
	}

	if len(proposals) != len(seals) {
		return false
	}

	return true
}

func Test_All_Honest_Validators(t *testing.T) {
	network := NewIBFTNetwork(
		NewIBFTValidator(),
		NewIBFTValidator(),
		NewIBFTValidator(),
		NewIBFTValidator(),
	)

	validatorProposals, err := network.FinalizeSequence(101, Round0Timeout)
	require.NoError(t, err)
	assert.True(t, AllValidProposals(network, validatorProposals))
}

func Test_Sequence_Finalized_In_Round_1(t *testing.T) {
	network := NewIBFTNetwork(
		NewIBFTValidator(),
		NewIBFTValidator(),
		NewIBFTValidator(),
		NewIBFTValidator(),
	).WithTransport(
		ExcludeMsgIf(IsMsgProposal(), HasRound(0)),
	)

	validatorProposals, err := network.FinalizeSequence(101, 100*time.Millisecond)
	require.NoError(t, err)
	assert.True(t, AllValidProposals(network, validatorProposals))
}

func Test_Sequence_Finalized_In_Round_1_2(t *testing.T) {
	network := NewIBFTNetwork(
		NewIBFTValidator(),
		NewIBFTValidator(),
		NewIBFTValidator(),
		NewIBFTValidator(),
	).WithTransport(
		ExcludeMsgIf(HasRound(0), IsMsgPrepare()),
	)

	validatorProposals, err := network.FinalizeSequence(101, 100*time.Millisecond)
	require.NoError(t, err)
	assert.True(t, AllValidProposals(network, validatorProposals))
}

func Test_Sequence_Finalized_In_Round_1_3(t *testing.T) {
	network := NewIBFTNetwork(
		NewIBFTValidator(),
		NewIBFTValidator(),
		NewIBFTValidator(),
		NewIBFTValidator(),
	).WithTransport(
		ExcludeMsgIf(HasRound(0), IsMsgCommit()),
	)

	validatorProposals, err := network.FinalizeSequence(101, 100*time.Millisecond)
	require.NoError(t, err)
	assert.True(t, AllValidProposals(network, validatorProposals))
}

func Test_Sequence_Two_Bad_Proposers(t *testing.T) {
	validator := NewIBFTValidator()
	network := NewIBFTNetwork(
		validator,
		NewIBFTValidator(),
		NewIBFTValidator(),
		NewIBFTValidator(),
	).WithTransport(
		ExcludeMsgIf(HasRound(0), IsMsgProposal()),
		ExcludeMsgIf(HasRound(1), IsMsgProposal()),
	)

	proposals, err := network.FinalizeSequence(101, 50*time.Millisecond)
	require.NoError(t, err)

	assert.True(t, AllValidProposals(network, proposals))
	assert.Equal(t, 2, int(proposals.From(validator.ID()).Round))
}
