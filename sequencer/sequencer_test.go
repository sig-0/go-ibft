package sequencer

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
	"github.com/madz-lab/go-ibft/test/mock"
)

var (
	Alice = mock.NewValidatorID("alice")
	Bob   = mock.NewValidatorID("bob")
	Chris = mock.NewValidatorID("chris")
	Dani  = mock.NewValidatorID("dani")

	ValidatorSet = mock.NewValidatorSet(Alice, Bob, Chris, Dani)
)

func Test_Sequencer_Finalize_Sequence_Cancelled(t *testing.T) {
	t.Parallel()

	ctx, cancelCtx := context.WithCancel(context.Background())
	c := make(chan *types.FinalizedProposal)

	go func(ctx context.Context) {
		defer close(c)

		seq := New(
			mock.Validator{
				IDFn: Alice.ID,
				Verifier: mock.Verifier{
					IsProposerFn: func(_ []byte, _, _ uint64) bool { return false },
				},
			},
			10*time.Millisecond,
		)

		c <- seq.Finalize(NewContext(ctx, WithMessageFeed(mock.NewSingleRoundFeed(nil))), 101)
	}(ctx)

	cancelCtx()

	assert.Nil(t, <-c)
}

//nolint:dupl // consensus messages are not entirely different among cases
func Test_Sequencer_Finalize_Sequence(t *testing.T) {
	t.Parallel()

	table := []struct {
		validator ibft.Validator
		quorum    ibft.Quorum
		feed      MessageFeed
		expected  *types.FinalizedProposal
		name      string
	}{
		{
			name:   "validator is not the proposer",
			quorum: mock.NonZeroQuorum,
			validator: mock.Validator{
				IDFn:   Alice.ID,
				Signer: Alice.Signer(),
				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      ValidatorSet.IsValidator,
					IsProposerFn:       mock.ProposersInRounds(mock.Proposer{ID: Bob, Round: 0}),
				},
			},

			feed: mock.NewMessageFeed([]types.Message{
				&types.MsgProposal{
					From:          Bob,
					View:          &types.View{Sequence: 101, Round: 0},
					BlockHash:     []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 0},
				},

				&types.MsgPrepare{
					From:      Alice,
					View:      &types.View{Sequence: 101, Round: 0},
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					From:       Alice,
					View:       &types.View{Sequence: 101, Round: 0},
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),

			expected: &types.FinalizedProposal{
				Proposal: []byte("block"),
				Round:    0,
				Seals: []types.FinalizedSeal{
					{
						From:       Alice,
						CommitSeal: []byte("commit seal"),
					},
				},
			},
		},

		{
			name:   "validator is the proposer",
			quorum: mock.NonZeroQuorum,
			validator: mock.Validator{
				IDFn:            Alice.ID,
				Signer:          Alice.Signer(),
				BuildProposalFn: func(_ uint64) []byte { return []byte("block") },
				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsProposerFn:       mock.ProposersInRounds(mock.Proposer{ID: Alice, Round: 0}),
					IsValidatorFn:      ValidatorSet.IsValidator,
				},
			},

			feed: mock.NewMessageFeed([]types.Message{
				&types.MsgPrepare{
					From:      Bob,
					View:      &types.View{Sequence: 101, Round: 0},
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					From:       Bob,
					View:       &types.View{Sequence: 101, Round: 0},
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),

			expected: &types.FinalizedProposal{
				Proposal: []byte("block"),
				Round:    0,
				Seals: []types.FinalizedSeal{
					{
						From:       Bob,
						CommitSeal: []byte("commit seal"),
					},
				},
			},
		},

		{
			name:   "round 1 proposal is valid with empty PB and PC",
			quorum: mock.NonZeroQuorum,
			validator: mock.Validator{
				IDFn:   Alice.ID,
				Signer: Alice.Signer(),
				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      ValidatorSet.IsValidator,
					IsProposerFn:       mock.ProposersInRounds(mock.Proposer{ID: Bob, Round: 1}),
				},
			},

			feed: mock.NewMessageFeed([]types.Message{
				&types.MsgProposal{
					From:          Bob,
					View:          &types.View{Sequence: 101, Round: 1},
					BlockHash:     []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							View: &types.View{Sequence: 101, Round: 1},
							From: Alice,
						},
					}},
				},

				&types.MsgPrepare{
					From:      Alice,
					View:      &types.View{Sequence: 101, Round: 1},
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					From:       Alice,
					View:       &types.View{Sequence: 101, Round: 1},
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),

			expected: &types.FinalizedProposal{
				Proposal: []byte("block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       Alice,
						CommitSeal: []byte("commit seal"),
					},
				},
			},
		},

		{
			name:   "round 1 proposal is valid with non-nil PB and PC",
			quorum: mock.NonZeroQuorum,
			validator: mock.Validator{
				IDFn:   Alice.ID,
				Signer: Alice.Signer(),
				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      ValidatorSet.IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: Chris, Round: 0},
						mock.Proposer{ID: Bob, Round: 1},
					),
				},
			},

			feed: mock.NewMessageFeed([]types.Message{
				&types.MsgProposal{
					From:          Bob,
					View:          &types.View{Sequence: 101, Round: 1},
					BlockHash:     []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{Block: []byte("round 0 block"), Round: 1},
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							From: Chris,
							View: &types.View{Sequence: 101, Round: 1},
							LatestPreparedCertificate: &types.PreparedCertificate{
								ProposalMessage: &types.MsgProposal{
									From:      Chris,
									View:      &types.View{Sequence: 101, Round: 0},
									BlockHash: []byte("block hash"),
									ProposedBlock: &types.ProposedBlock{
										Block: []byte("round 0 block"),
										Round: 0,
									},
								},
								PrepareMessages: []*types.MsgPrepare{
									{
										From:      Alice,
										View:      &types.View{Sequence: 101, Round: 0},
										BlockHash: []byte("block hash"),
									},
								},
							},
						},
					}},
				},

				&types.MsgPrepare{
					From:      Alice,
					View:      &types.View{Sequence: 101, Round: 1},
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					From:       Alice,
					View:       &types.View{Sequence: 101, Round: 1},
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),

			expected: &types.FinalizedProposal{
				Proposal: []byte("round 0 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       Alice,
						CommitSeal: []byte("commit seal"),
					},
				},
			},
		},

		{
			name:   "propose new block in round 1",
			quorum: mock.NonZeroQuorum,
			validator: mock.Validator{
				IDFn:            Alice.ID,
				Signer:          Alice.Signer(),
				BuildProposalFn: func(_ uint64) []byte { return []byte("round 1 block") },
				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      ValidatorSet.IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: Bob, Round: 0},
						mock.Proposer{ID: Alice, Round: 1},
					),
				},
			},

			feed: mock.NewMessageFeed([]types.Message{
				// need to justify Alice's proposal for round 1
				&types.MsgRoundChange{
					View: &types.View{Sequence: 101, Round: 1},
					From: Alice,
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      Bob,
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 1},
					From:       Bob,
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),

			expected: &types.FinalizedProposal{
				Proposal: []byte("round 1 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       Bob,
						CommitSeal: []byte("commit seal"),
					},
				},
			},
		},

		{
			name:   "propose old block in round 1",
			quorum: mock.NonZeroQuorum,
			validator: mock.Validator{
				IDFn:   Alice.ID,
				Signer: Alice.Signer(),
				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      ValidatorSet.IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: Bob, Round: 0},
						mock.Proposer{ID: Alice, Round: 1},
					),
				},
			},

			feed: mock.NewMessageFeed([]types.Message{
				&types.MsgRoundChange{
					From:                        Chris,
					View:                        &types.View{Sequence: 101, Round: 1},
					LatestPreparedProposedBlock: &types.ProposedBlock{Block: []byte("round 0 block"), Round: 0},
					LatestPreparedCertificate: &types.PreparedCertificate{
						ProposalMessage: &types.MsgProposal{
							From:          Bob,
							View:          &types.View{Sequence: 101, Round: 0},
							BlockHash:     []byte("block hash"),
							ProposedBlock: &types.ProposedBlock{Block: []byte("round 0 block"), Round: 0},
						},

						PrepareMessages: []*types.MsgPrepare{
							{
								From:      Chris,
								View:      &types.View{Sequence: 101, Round: 0},
								BlockHash: []byte("block hash"),
							},
						},
					},
				},

				&types.MsgPrepare{
					From:      Alice,
					View:      &types.View{Sequence: 101, Round: 1},
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					From:       Alice,
					View:       &types.View{Sequence: 101, Round: 1},
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),

			expected: &types.FinalizedProposal{
				Proposal: []byte("round 0 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       Alice,
						CommitSeal: []byte("commit seal"),
					},
				},
			},
		},

		{
			name:   "future rcc triggers round jump and new block is finalized",
			quorum: mock.NonZeroQuorum,
			validator: mock.Validator{
				IDFn:   Alice.ID,
				Signer: Alice.Signer(),
				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      ValidatorSet.IsValidator,
					IsProposerFn:       mock.ProposersInRounds(mock.Proposer{ID: Chris, Round: 3}),
				},
			},

			feed: mock.NewMessageFeed([]types.Message{
				&types.MsgProposal{
					From:          Chris,
					View:          &types.View{Sequence: 101, Round: 3},
					BlockHash:     []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 3},
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							From: Bob,
							View: &types.View{Sequence: 101, Round: 3},
						},
					}},
				},

				&types.MsgPrepare{
					From:      Alice,
					View:      &types.View{Sequence: 101, Round: 3},
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					From:       Alice,
					View:       &types.View{Sequence: 101, Round: 3},
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),

			expected: &types.FinalizedProposal{
				Proposal: []byte("block"),
				Round:    3,
				Seals: []types.FinalizedSeal{
					{
						From:       Alice,
						CommitSeal: []byte("commit seal"),
					},
				},
			},
		},

		{
			name:   "future proposal triggers round jump and new block is finalized",
			quorum: mock.NonZeroQuorum,
			validator: mock.Validator{
				IDFn:   Alice.ID,
				Signer: Alice.Signer(),
				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      ValidatorSet.IsValidator,
					IsProposerFn:       mock.ProposersInRounds(mock.Proposer{ID: Dani, Round: 5}),
				},
			},

			feed: mock.NewMessageFeed([]types.Message{
				&types.MsgProposal{
					From:          Dani,
					View:          &types.View{Sequence: 101, Round: 5},
					ProposedBlock: &types.ProposedBlock{Block: []byte("round 5 block"), Round: 5},
					BlockHash:     []byte("block hash"),
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							View: &types.View{Sequence: 101, Round: 5},
							From: Chris,
						},
					}},
				},

				&types.MsgPrepare{
					From:      Alice,
					View:      &types.View{Sequence: 101, Round: 5},
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					From:       Alice,
					View:       &types.View{Sequence: 101, Round: 5},
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),

			expected: &types.FinalizedProposal{
				Proposal: []byte("round 5 block"),
				Seals: []types.FinalizedSeal{
					{
						From:       Alice,
						CommitSeal: []byte("commit seal"),
					},
				},
				Round: 5,
			},
		},

		{
			name:   "round timer triggers round jump and proposer finalizes new block",
			quorum: mock.NonZeroQuorum,
			validator: mock.Validator{
				IDFn:            Alice.ID,
				Signer:          Alice.Signer(),
				BuildProposalFn: func(_ uint64) []byte { return []byte("round 1 block") },
				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      ValidatorSet.IsValidator,
					IsProposerFn:       mock.ProposersInRounds(mock.Proposer{ID: Alice, Round: 1}),
				},
			},

			feed: mock.NewSingleRoundFeed([]types.Message{
				&types.MsgRoundChange{
					From: Bob,
					View: &types.View{Sequence: 101, Round: 1},
				},

				&types.MsgPrepare{
					From:      Bob,
					View:      &types.View{Sequence: 101, Round: 1},
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					From:       Bob,
					View:       &types.View{Sequence: 101, Round: 1},
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),

			expected: &types.FinalizedProposal{
				Proposal: []byte("round 1 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       Bob,
						CommitSeal: []byte("commit seal"),
					},
				},
			},
		},

		{
			name:   "no prepare msgs in round 0 so new block is finalized in round 1",
			quorum: mock.NonZeroQuorum,
			validator: mock.Validator{
				IDFn:   Alice.ID,
				Signer: Alice.Signer(),
				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      ValidatorSet.IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: Bob, Round: 0},
						mock.Proposer{ID: Chris, Round: 1},
					),
				},
			},

			feed: mock.NewSingleRoundFeed([]types.Message{
				&types.MsgProposal{
					From:          Bob,
					View:          &types.View{Sequence: 101, Round: 0},
					BlockHash:     []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{Block: []byte("round 0 block"), Round: 0},
				},

				&types.MsgProposal{
					From:          Chris,
					View:          &types.View{Sequence: 101, Round: 1},
					BlockHash:     []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{Block: []byte("round 1 block"), Round: 1},
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							View: &types.View{Sequence: 101, Round: 1},
							From: Alice,
						},
					}},
				},

				&types.MsgPrepare{
					From:      Alice,
					View:      &types.View{Sequence: 101, Round: 1},
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					From:       Alice,
					View:       &types.View{Sequence: 101, Round: 1},
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),

			expected: &types.FinalizedProposal{
				Proposal: []byte("round 1 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       Alice,
						CommitSeal: []byte("commit seal"),
					},
				},
			},
		},

		{
			name:   "no quorum prepare msgs in round 0 so new block is finalized in round 1",
			quorum: mock.QuorumOf(2),
			validator: mock.Validator{
				IDFn:   Alice.ID,
				Signer: Alice.Signer(),
				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      ValidatorSet.IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: Bob, Round: 0},
						mock.Proposer{ID: Chris, Round: 1},
					),
				},
			},

			feed: mock.NewSingleRoundFeed([]types.Message{
				/* round 0 */

				&types.MsgProposal{
					From:          Bob,
					View:          &types.View{Sequence: 101, Round: 0},
					BlockHash:     []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{Block: []byte("round 0 block"), Round: 0},
				},

				&types.MsgPrepare{
					From:      Alice,
					View:      &types.View{Sequence: 101, Round: 0},
					BlockHash: []byte("block hash"),
				},

				/* round 1 */

				&types.MsgProposal{
					From:          Chris,
					View:          &types.View{Sequence: 101, Round: 1},
					BlockHash:     []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{Block: []byte("round 1 block"), Round: 1},
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							From: Alice,
							View: &types.View{Sequence: 101, Round: 1},
						},
						{
							From: Chris,
							View: &types.View{Sequence: 101, Round: 1},
						},
					}},
				},

				&types.MsgPrepare{
					From:      Alice,
					View:      &types.View{Sequence: 101, Round: 1},
					BlockHash: []byte("block hash"),
				},

				&types.MsgPrepare{
					From:      Bob,
					View:      &types.View{Sequence: 101, Round: 1},
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					From:       Alice,
					View:       &types.View{Sequence: 101, Round: 1},
					CommitSeal: []byte("commit seal"),
					BlockHash:  []byte("block hash"),
				},

				&types.MsgCommit{
					From:       Bob,
					View:       &types.View{Sequence: 101, Round: 1},
					CommitSeal: []byte("other commit seal"),
					BlockHash:  []byte("block hash"),
				},
			}),

			expected: &types.FinalizedProposal{
				Proposal: []byte("round 1 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       Alice,
						CommitSeal: []byte("commit seal"),
					},
					{
						From:       Bob,
						CommitSeal: []byte("other commit seal"),
					},
				},
			},
		},

		{
			name:   "no commit msgs in round 0 so new block is finalized in round 1",
			quorum: mock.QuorumOf(2),
			validator: mock.Validator{
				IDFn:   Alice.ID,
				Signer: Alice.Signer(),
				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      ValidatorSet.IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: Bob, Round: 0},
						mock.Proposer{ID: Chris, Round: 1},
					),
				},
			},

			feed: mock.NewSingleRoundFeed([]types.Message{
				/* round 0 */

				&types.MsgProposal{
					From:          Bob,
					View:          &types.View{Sequence: 101, Round: 0},
					BlockHash:     []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{Block: []byte("round 0 block"), Round: 0},
				},

				&types.MsgPrepare{
					From:      Alice,
					View:      &types.View{Sequence: 101, Round: 0},
					BlockHash: []byte("block hash"),
				},

				&types.MsgPrepare{
					From:      Bob,
					View:      &types.View{Sequence: 101, Round: 0},
					BlockHash: []byte("block hash"),
				},

				/* round 1 */

				&types.MsgProposal{
					From:          Chris,
					View:          &types.View{Sequence: 101, Round: 1},
					BlockHash:     []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{Block: []byte("round 1 block"), Round: 1},
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							From: Alice,
							View: &types.View{Sequence: 101, Round: 1},
						},
						{
							From: Bob,
							View: &types.View{Sequence: 101, Round: 1},
						},
					}},
				},

				&types.MsgPrepare{
					From:      Alice,
					View:      &types.View{Sequence: 101, Round: 1},
					BlockHash: []byte("block hash"),
				},

				&types.MsgPrepare{
					From:      Bob,
					View:      &types.View{Sequence: 101, Round: 1},
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					From:       Alice,
					View:       &types.View{Sequence: 101, Round: 1},
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},

				&types.MsgCommit{
					From:       Bob,
					View:       &types.View{Sequence: 101, Round: 1},
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("other commit seal"),
				},
			}),

			expected: &types.FinalizedProposal{
				Proposal: []byte("round 1 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       Alice,
						CommitSeal: []byte("commit seal"),
					},
					{
						From:       Bob,
						CommitSeal: []byte("other commit seal"),
					},
				},
			},
		},

		{
			name:   "no quorum commit msgs in round 0 so new block is finalized in round 1",
			quorum: mock.QuorumOf(2),
			validator: mock.Validator{
				IDFn:   Alice.ID,
				Signer: Alice.Signer(),
				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      ValidatorSet.IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: Bob, Round: 0},
						mock.Proposer{ID: Chris, Round: 1},
					),
				},
			},

			feed: mock.NewSingleRoundFeed([]types.Message{
				/* round 0 */

				&types.MsgProposal{
					From:          Bob,
					View:          &types.View{Sequence: 101, Round: 0},
					BlockHash:     []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{Block: []byte("round 0 block"), Round: 0},
				},

				&types.MsgPrepare{
					From:      Alice,
					View:      &types.View{Sequence: 101, Round: 0},
					BlockHash: []byte("block hash"),
				},

				&types.MsgPrepare{
					From:      Chris,
					View:      &types.View{Sequence: 101, Round: 0},
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					From:       Alice,
					View:       &types.View{Sequence: 101, Round: 0},
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},

				/* round 1 */

				&types.MsgProposal{
					From:          Chris,
					View:          &types.View{Sequence: 101, Round: 1},
					BlockHash:     []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{Block: []byte("round 1 block"), Round: 1},
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							From: Alice,
							View: &types.View{Sequence: 101, Round: 1},
						},
						{
							From: Chris,
							View: &types.View{Sequence: 101, Round: 1},
						},
					}},
				},

				&types.MsgPrepare{
					From:      Alice,
					View:      &types.View{Sequence: 101, Round: 1},
					BlockHash: []byte("block hash"),
				},
				&types.MsgPrepare{
					From:      Chris,
					View:      &types.View{Sequence: 101, Round: 1},
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					From:       Alice,
					View:       &types.View{Sequence: 101, Round: 1},
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},

				&types.MsgCommit{
					From:       Chris,
					View:       &types.View{Sequence: 101, Round: 1},
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("other commit seal"),
				},
			}),

			expected: &types.FinalizedProposal{
				Proposal: []byte("round 1 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       Alice,
						CommitSeal: []byte("commit seal"),
					},
					{
						From:       Chris,
						CommitSeal: []byte("other commit seal"),
					},
				},
			},
		},

		{
			name:   "round 1 proposer fails to build block so new block is finalized in round 2",
			quorum: mock.QuorumOf(2),
			validator: mock.Validator{
				IDFn:   Alice.ID,
				Signer: Alice.Signer(),
				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      ValidatorSet.IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: Bob, Round: 0},
						mock.Proposer{ID: Alice, Round: 1},
						mock.Proposer{ID: Chris, Round: 2},
					),
				},
			},

			feed: mock.NewSingleRoundFeed([]types.Message{
				&types.MsgRoundChange{
					From: Chris,
					View: &types.View{Sequence: 101, Round: 1},
				},

				&types.MsgProposal{
					From:          Chris,
					View:          &types.View{Sequence: 101, Round: 2},
					ProposedBlock: &types.ProposedBlock{Block: []byte("round 2 block"), Round: 2},
					BlockHash:     []byte("block hash"),
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							From: Bob,
							View: &types.View{Sequence: 101, Round: 2},
						},
						{
							From: Chris,
							View: &types.View{Sequence: 101, Round: 2},
						},
					}},
				},

				&types.MsgPrepare{
					From:      Bob,
					View:      &types.View{Sequence: 101, Round: 2},
					BlockHash: []byte("block hash"),
				},
				&types.MsgPrepare{
					From:      Alice,
					View:      &types.View{Sequence: 101, Round: 2},
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					From:       Bob,
					View:       &types.View{Sequence: 101, Round: 2},
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
				&types.MsgCommit{
					From:       Alice,
					View:       &types.View{Sequence: 101, Round: 2},
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("other commit seal"),
				},
			}),

			expected: &types.FinalizedProposal{
				Proposal: []byte("round 2 block"),
				Round:    2,
				Seals: []types.FinalizedSeal{
					{
						From:       Bob,
						CommitSeal: []byte("commit seal"),
					},
					{
						From:       Alice,
						CommitSeal: []byte("other commit seal"),
					},
				},
			},
		},
	}

	for _, tt := range table {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := NewContext(context.Background(),
				WithKeccak(mock.DummyKeccak),
				WithQuorum(tt.quorum),
				WithMessageFeed(tt.feed),
				WithMessageTransport(MessageTransport{
					Proposal:    mock.DummyMsgProposalTransport,
					Prepare:     mock.DummyMsgPrepareTransport,
					Commit:      mock.DummyMsgCommitTransport,
					RoundChange: mock.DummyMsgRoundChangeTransport,
				}),
			)

			s := New(tt.validator, time.Millisecond*50)
			assert.True(t, reflect.DeepEqual(tt.expected, s.Finalize(ctx, 101)))
		})
	}
}
