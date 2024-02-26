package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madz-lab/go-ibft"
)

func Test_Finalize_Sequence_4_Validators(t *testing.T) {
	validators := []ibft.Validator{
		NewIBFTValidator(),
		NewIBFTValidator(),
		NewIBFTValidator(),
		NewIBFTValidator(),
	}

	table := []struct {
		name                   string
		round0Timeout          time.Duration
		expectedFinalizedRound uint64
		opts                   []MessageOption
	}{
		{
			name:                   "proposal finalized in round 0",
			round0Timeout:          50 * time.Millisecond,
			expectedFinalizedRound: 0,
		},

		{
			name:                   "proposal finalized in round 1 because proposal was missed",
			round0Timeout:          100 * time.Millisecond,
			expectedFinalizedRound: 1,
			opts: []MessageOption{
				ExcludeMsgIf(IsMsgProposal(), HasRound(0)),
			},
		},

		{
			name:                   "proposal finalized in round 1 because all prepare messages were missed",
			round0Timeout:          100 * time.Millisecond,
			expectedFinalizedRound: 1,
			opts: []MessageOption{
				ExcludeMsgIf(HasRound(0), IsMsgPrepare()),
			},
		},

		{
			name:                   "proposal finalized in round 1 because all commit messages were missed",
			round0Timeout:          100 * time.Millisecond,
			expectedFinalizedRound: 1,
			opts: []MessageOption{
				ExcludeMsgIf(HasRound(0), IsMsgCommit()),
			},
		},

		{
			name:                   "proposal finalized in round 2 because first 2 proposals were missed",
			round0Timeout:          100 * time.Millisecond,
			expectedFinalizedRound: 2,
			opts: []MessageOption{
				ExcludeMsgIf(HasRound(0), IsMsgProposal()),
				ExcludeMsgIf(HasRound(1), IsMsgProposal()),
			},
		},
	}

	for _, tt := range table {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			network := NewIBFTNetwork(validators...).WithTransport(tt.opts...)
			proposals, err := network.FinalizeSequence(101, tt.round0Timeout)

			require.NoError(t, err)
			require.True(t, AllValidProposals(network, proposals))
			assert.Equal(t, tt.expectedFinalizedRound, proposals[0].Round)
		})
	}
}
