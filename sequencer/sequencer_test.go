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

func Test_Sequencer_Finalize_Sequence_Cancelled(t *testing.T) {
	t.Parallel()

	validator := mock.Validator{
		IDFn: mock.DummyValidatorIDFn,
		Verifier: mock.Verifier{
			IsProposerFn: func(_ []byte, _, _ uint64) bool { return false },
		},
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	c := make(chan *types.FinalizedProposal)

	go func(ctx context.Context) {
		defer close(c)

		seq := New(validator, 10*time.Millisecond)
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
			name: "validator is not the proposer",
			expected: &types.FinalizedProposal{
				Proposal: []byte("block"),
				Round:    0,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: mock.Validator{
				Signer: mock.DummySigner,
				IDFn:   mock.DummyValidatorIDFn,

				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsProposerFn:       mock.ProposersInRounds(mock.Proposer{ID: []byte("proposer"), Round: 0}),
					IsValidatorFn:      mock.AlwaysValidator,
				},
			},

			quorum: mock.NonZeroQuorum,

			feed: mock.NewMessageFeed([]ibft.Message{
				&types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("proposer"),
					BlockHash: []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{
						Block: []byte("block"),
						Round: 0,
					},
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("some validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 0},
					From:       []byte("some validator"),
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),
		},

		{
			name: "validator is the proposer",
			expected: &types.FinalizedProposal{
				Proposal: []byte("block"),
				Round:    0,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("some validator"),
					},
				},
			},

			validator: mock.Validator{
				Signer:          mock.DummySigner,
				IDFn:            mock.DummyValidatorIDFn,
				BuildProposalFn: func(_ uint64) []byte { return []byte("block") },

				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsProposerFn:       func(_ []byte, _, _ uint64) bool { return true },
					IsValidatorFn:      mock.AlwaysValidator,
				},
			},

			quorum: mock.NonZeroQuorum,

			feed: mock.NewMessageFeed([]ibft.Message{
				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("some validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 0},
					From:       []byte("some validator"),
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("some validator"),
				},
			}),
		},

		{
			name: "round 1 proposal is valid with empty PB and PC",
			expected: &types.FinalizedProposal{
				Proposal: []byte("block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("my validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: mock.Validator{
				Signer: mock.DummySigner,
				IDFn:   mock.DummyValidatorIDFn,

				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      mock.AlwaysValidator,
					IsProposerFn:       mock.ProposersInRounds(mock.Proposer{ID: []byte("proposer"), Round: 1}),
				},
			},

			quorum: mock.NonZeroQuorum,

			feed: mock.NewMessageFeed([]ibft.Message{
				&types.MsgProposal{
					View:          &types.View{Sequence: 101, Round: 1},
					From:          []byte("proposer"),
					BlockHash:     []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							View: &types.View{Sequence: 101, Round: 1},
							From: []byte("my validator"),
						},
					}},
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("my validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 1},
					From:       []byte("my validator"),
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),
		},

		{
			name: "round 1 proposal is valid with non-nil PB and PC",
			expected: &types.FinalizedProposal{
				Proposal: []byte("round 0 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: mock.Validator{
				Signer: mock.DummySigner,
				IDFn:   mock.DummyValidatorIDFn,

				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      mock.AlwaysValidator,
					IsProposerFn:       mock.ProposersInRounds(mock.Proposer{ID: []byte("proposer"), Round: 1}),
				},
			},

			quorum: mock.NonZeroQuorum,

			feed: mock.NewMessageFeed([]ibft.Message{
				&types.MsgProposal{
					View:          &types.View{Sequence: 101, Round: 1},
					From:          []byte("proposer"),
					BlockHash:     []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{Block: []byte("round 0 block"), Round: 1},
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							View: &types.View{Sequence: 101, Round: 1},
							From: []byte("some validator"),
							LatestPreparedCertificate: &types.PreparedCertificate{
								ProposalMessage: &types.MsgProposal{
									View:      &types.View{Sequence: 101, Round: 0},
									From:      []byte("proposer"),
									BlockHash: []byte("block hash"),
									ProposedBlock: &types.ProposedBlock{
										Block: []byte("round 0 block"),
										Round: 0,
									},
								},
								PrepareMessages: []*types.MsgPrepare{
									{
										View:      &types.View{Sequence: 101, Round: 0},
										From:      []byte("some validator"),
										BlockHash: []byte("block hash"),
									},
								},
							},
						},
					}},
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("some validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 1},
					From:       []byte("some validator"),
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),
		},

		{
			name: "propose new block in round 1",
			expected: &types.FinalizedProposal{
				Proposal: []byte("round 1 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: mock.Validator{
				IDFn:            mock.DummyValidatorIDFn,
				Signer:          mock.DummySigner,
				BuildProposalFn: func(_ uint64) []byte { return []byte("round 1 block") },

				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      mock.AlwaysValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: []byte("proposer"), Round: 0},
						mock.Proposer{ID: mock.DummyValidatorIDFn(), Round: 1},
					),
				},
			},

			quorum: mock.NonZeroQuorum,

			feed: mock.NewMessageFeed([]ibft.Message{
				&types.MsgRoundChange{
					View: &types.View{Sequence: 101, Round: 1},
					From: []byte("some validator"),
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("some validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 1},
					From:       []byte("some validator"),
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),
		},

		{
			name: "propose old block in round 1",
			expected: &types.FinalizedProposal{
				Proposal: []byte("round 0 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: mock.Validator{
				IDFn:   mock.DummyValidatorIDFn,
				Signer: mock.DummySigner,

				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      mock.AlwaysValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: []byte("proposer"), Round: 0},
						mock.Proposer{ID: mock.DummyValidatorIDFn(), Round: 1},
					),
				},
			},

			quorum: mock.NonZeroQuorum,

			feed: mock.NewMessageFeed([]ibft.Message{
				&types.MsgRoundChange{
					View: &types.View{Sequence: 101, Round: 1},
					From: []byte("some validator"),
					LatestPreparedProposedBlock: &types.ProposedBlock{
						Block: []byte("round 0 block"),
						Round: 0,
					},
					LatestPreparedCertificate: &types.PreparedCertificate{
						ProposalMessage: &types.MsgProposal{
							View:      &types.View{Sequence: 101, Round: 0},
							From:      []byte("proposer"),
							BlockHash: []byte("block hash"),
							ProposedBlock: &types.ProposedBlock{
								Block: []byte("round 0 block"),
								Round: 0,
							},
						},
						PrepareMessages: []*types.MsgPrepare{
							{
								View:      &types.View{Sequence: 101, Round: 0},
								From:      []byte("some validator"),
								BlockHash: []byte("block hash"),
							},
						},
					},
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("some validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 1},
					From:       []byte("some validator"),
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),
		},

		{
			name: "future rcc triggers round jump and new block is finalized",
			expected: &types.FinalizedProposal{
				Proposal: []byte("block"),
				Round:    3,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: mock.Validator{
				IDFn:   mock.DummyValidatorIDFn,
				Signer: mock.DummySigner,

				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      mock.AlwaysValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: []byte("proposer"), Round: 3},
					),
				},
			},

			quorum: mock.NonZeroQuorum,

			feed: mock.NewMessageFeed([]ibft.Message{
				&types.MsgRoundChange{
					View: &types.View{Sequence: 101, Round: 3},
					From: []byte("some validator"),
				},

				&types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 3},
					From:      []byte("proposer"),
					BlockHash: []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{
						Block: []byte("block"),
						Round: 3,
					},
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							View: &types.View{Sequence: 101, Round: 3},
							From: []byte("some validator"),
						},
					}},
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 3},
					From:      []byte("some validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 3},
					From:       []byte("some validator"),
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),
		},

		{
			name: "future proposal triggers round jump and new block is finalized",
			expected: &types.FinalizedProposal{
				Proposal: []byte("round 5 block"),
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
				Round: 5,
			},

			validator: mock.Validator{
				IDFn:   mock.DummyValidatorIDFn,
				Signer: mock.DummySigner,

				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsProposerFn:       mock.ProposersInRounds(mock.Proposer{ID: []byte("proposer"), Round: 5}),
					IsValidatorFn:      mock.AlwaysValidator,
				},
			},

			quorum: mock.NonZeroQuorum,

			feed: mock.NewMessageFeed([]ibft.Message{
				&types.MsgProposal{
					View: &types.View{Sequence: 101, Round: 5},
					From: []byte("proposer"),
					ProposedBlock: &types.ProposedBlock{
						Block: []byte("round 5 block"),
						Round: 5,
					},
					BlockHash: []byte("block hash"),
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							View: &types.View{Sequence: 101, Round: 5},
							From: []byte("some validator"),
						},
					}},
				},

				&types.MsgPrepare{
					View: &types.View{
						Sequence: 101,
						Round:    5,
					},
					From:      []byte("some validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					View: &types.View{
						Sequence: 101,
						Round:    5,
					},
					From:       []byte("some validator"),
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),
		},

		{
			name: "round timer triggers round jump and proposer finalizes new block",
			expected: &types.FinalizedProposal{
				Proposal: []byte("round 1 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: mock.Validator{
				IDFn:            mock.DummyValidatorIDFn,
				Signer:          mock.DummySigner,
				BuildProposalFn: func(_ uint64) []byte { return []byte("round 1 block") },

				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsProposerFn:       mock.ProposersInRounds(mock.Proposer{ID: mock.DummyValidatorIDFn(), Round: 1}),
					IsValidatorFn:      mock.AlwaysValidator,
				},
			},

			quorum: mock.NonZeroQuorum,

			feed: mock.NewSingleRoundFeed([]ibft.Message{
				&types.MsgRoundChange{
					View: &types.View{Sequence: 101, Round: 1},
					From: []byte("some validator"),
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("some validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 1},
					From:       []byte("some validator"),
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),
		},

		{
			name: "round timer triggers round jump and proposer finalizes new block",
			expected: &types.FinalizedProposal{
				Proposal: []byte("round 0 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: mock.Validator{
				IDFn:   mock.DummyValidatorIDFn,
				Signer: mock.DummySigner,

				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      mock.AlwaysValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: []byte("proposer"), Round: 0},
						mock.Proposer{ID: mock.DummyValidatorIDFn(), Round: 1},
					),
				},
			},

			quorum: mock.NonZeroQuorum,

			feed: mock.NewSingleRoundFeed([]ibft.Message{
				&types.MsgRoundChange{
					View: &types.View{Sequence: 101, Round: 1},
					From: []byte("some validator"),
					LatestPreparedProposedBlock: &types.ProposedBlock{
						Block: []byte("round 0 block"),
						Round: 0,
					},
					LatestPreparedCertificate: &types.PreparedCertificate{
						ProposalMessage: &types.MsgProposal{
							View:      &types.View{Sequence: 101, Round: 0},
							From:      []byte("proposer"),
							BlockHash: []byte("block hash"),
						},
						PrepareMessages: []*types.MsgPrepare{
							{
								View:      &types.View{Sequence: 101, Round: 0},
								From:      []byte("some validator"),
								BlockHash: []byte("block hash"),
							},
						},
					},
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("some validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 1},
					From:       []byte("some validator"),
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),
		},

		{
			name: "no prepare msgs in round 0 so new block is finalized in round 1",
			expected: &types.FinalizedProposal{
				Proposal: []byte("round 1 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: mock.Validator{
				IDFn:   mock.DummyValidatorIDFn,
				Signer: mock.DummySigner,

				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: []byte("proposer"), Round: 0},
						mock.Proposer{ID: []byte("proposer"), Round: 1},
					),
					IsValidatorFn: mock.AlwaysValidator,
				},
			},

			quorum: mock.NonZeroQuorum,

			feed: mock.NewSingleRoundFeed([]ibft.Message{
				&types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("proposer"),
					BlockHash: []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{
						Block: []byte("round 0 block"),
						Round: 0,
					},
				},

				&types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("proposer"),
					BlockHash: []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{
						Block: []byte("round 1 block"),
						Round: 1,
					},
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							View: &types.View{Sequence: 101, Round: 1},
							From: []byte("some validator"),
						},
					}},
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("some validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 1},
					From:       []byte("some validator"),
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
			}),
		},

		{
			name: "no quorum prepare msgs in round 0 so new block is finalized in round 1",
			expected: &types.FinalizedProposal{
				Proposal: []byte("round 1 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("validator"),
						CommitSeal: []byte("commit seal"),
					},
					{
						From:       []byte("other validator"),
						CommitSeal: []byte("other commit seal"),
					},
				},
			},

			validator: mock.Validator{
				IDFn:   mock.DummyValidatorIDFn,
				Signer: mock.DummySigner,

				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: []byte("proposer"), Round: 0},
						mock.Proposer{ID: []byte("proposer"), Round: 1},
					),
					IsValidatorFn: mock.AlwaysValidator,
				},
			},

			quorum: mock.QuorumOf(2),

			feed: mock.NewSingleRoundFeed([]ibft.Message{
				&types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("proposer"),
					BlockHash: []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{
						Block: []byte("round 0 block"),
						Round: 0,
					},
				},

				&types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("proposer"),
					BlockHash: []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{
						Block: []byte("round 1 block"),
						Round: 1,
					},
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							View: &types.View{Sequence: 101, Round: 1},
							From: []byte("validator"),
						},
						{
							View: &types.View{Sequence: 101, Round: 1},
							From: []byte("other validator"),
						},
					}},
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("other validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 1},
					From:       []byte("validator"),
					CommitSeal: []byte("commit seal"),
					BlockHash:  []byte("block hash"),
				},
				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 1},
					From:       []byte("other validator"),
					CommitSeal: []byte("other commit seal"),
					BlockHash:  []byte("block hash"),
				},
			}),
		},

		{
			name: "no commit msgs in round 0 so new block is finalized in round 1",
			expected: &types.FinalizedProposal{
				Proposal: []byte("round 1 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("validator"),
						CommitSeal: []byte("commit seal"),
					},
					{
						From:       []byte("other validator"),
						CommitSeal: []byte("other commit seal"),
					},
				},
			},

			validator: mock.Validator{
				IDFn:   mock.DummyValidatorIDFn,
				Signer: mock.DummySigner,

				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: []byte("proposer"), Round: 0},
						mock.Proposer{ID: []byte("proposer"), Round: 1},
					),
					IsValidatorFn: mock.AlwaysValidator,
				},
			},

			quorum: mock.QuorumOf(2),

			feed: mock.NewSingleRoundFeed([]ibft.Message{
				&types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("proposer"),
					BlockHash: []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{
						Block: []byte("round 0 block"),
						Round: 0,
					},
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("validator"),
					BlockHash: []byte("block hash"),
				},
				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("proposer"),
					BlockHash: []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{
						Block: []byte("round 1 block"),
						Round: 1,
					},
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							View: &types.View{Sequence: 101, Round: 1},
							From: []byte("validator"),
						},
						{
							View: &types.View{Sequence: 101, Round: 1},
							From: []byte("other validator"),
						},
					}},
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("validator"),
					BlockHash: []byte("block hash"),
				},
				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("other validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 1},
					From:       []byte("validator"),
					CommitSeal: []byte("commit seal"),
					BlockHash:  []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 1},
					From:       []byte("other validator"),
					CommitSeal: []byte("other commit seal"),
					BlockHash:  []byte("block hash"),
				},
			}),
		},

		{
			name: "no quorum commit msgs in round 0 so new block is finalized in round 1",
			expected: &types.FinalizedProposal{
				Proposal: []byte("round 1 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("validator"),
						CommitSeal: []byte("commit seal"),
					},
					{
						From:       []byte("other validator"),
						CommitSeal: []byte("other commit seal"),
					},
				},
			},

			validator: mock.Validator{
				IDFn:   mock.DummyValidatorIDFn,
				Signer: mock.DummySigner,

				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: []byte("proposer"), Round: 0},
						mock.Proposer{ID: []byte("proposer"), Round: 1},
					),
					IsValidatorFn: mock.AlwaysValidator,
				},
			},

			quorum: mock.QuorumOf(2),

			feed: mock.NewSingleRoundFeed([]ibft.Message{
				&types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("proposer"),
					BlockHash: []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{
						Block: []byte("round 0 block"),
						Round: 0,
					},
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("other validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 0},
					From:       []byte("validator"),
					CommitSeal: []byte("commit seal"),
					BlockHash:  []byte("block hash"),
				},

				&types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("proposer"),
					BlockHash: []byte("block hash"),
					ProposedBlock: &types.ProposedBlock{
						Block: []byte("round 1 block"),
						Round: 1,
					},
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							View: &types.View{Sequence: 101, Round: 1},
							From: []byte("validator"),
						},
						{
							View: &types.View{Sequence: 101, Round: 1},
							From: []byte("other validator"),
						},
					}},
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("validator"),
					BlockHash: []byte("block hash"),
				},
				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("other validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 1},
					From:       []byte("validator"),
					CommitSeal: []byte("commit seal"),
					BlockHash:  []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 1},
					From:       []byte("other validator"),
					CommitSeal: []byte("other commit seal"),
					BlockHash:  []byte("block hash"),
				},
			}),
		},

		{
			name: "round 1 proposer fails to build block so new block is finalized in round 2",
			expected: &types.FinalizedProposal{
				Proposal: []byte("round 2 block"),
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("validator"),
						CommitSeal: []byte("commit seal"),
					},
					{
						From:       []byte("other validator"),
						CommitSeal: []byte("other commit seal"),
					},
				},
				Round: 2,
			},

			validator: mock.Validator{
				IDFn:   mock.DummyValidatorIDFn,
				Signer: mock.DummySigner,

				Verifier: mock.Verifier{
					IsValidSignatureFn: mock.AlwaysValidSignature,
					IsValidProposalFn:  mock.AlwaysValidBlock,
					IsValidatorFn:      mock.AlwaysValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: []byte("proposer"), Round: 0},
						mock.Proposer{ID: mock.DummyValidatorIDFn(), Round: 1},
						mock.Proposer{ID: []byte("proposer"), Round: 2},
					),
				},
			},

			quorum: mock.QuorumOf(2),

			feed: mock.NewSingleRoundFeed([]ibft.Message{
				&types.MsgRoundChange{
					View: &types.View{Sequence: 101, Round: 1},
					From: []byte("validator"),
				},

				&types.MsgProposal{
					View: &types.View{
						Sequence: 101,
						Round:    2,
					},
					From: []byte("proposer"),
					ProposedBlock: &types.ProposedBlock{
						Block: []byte("round 2 block"),
						Round: 2,
					},
					BlockHash: []byte("block hash"),
					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
						{
							View: &types.View{Sequence: 101, Round: 2},
							From: []byte("validator"),
						},
						{
							View: &types.View{Sequence: 101, Round: 2},
							From: []byte("other validator"),
						},
					}},
				},

				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 2},
					From:      []byte("validator"),
					BlockHash: []byte("block hash"),
				},
				&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 2},
					From:      []byte("other validator"),
					BlockHash: []byte("block hash"),
				},

				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 2},
					From:       []byte("validator"),
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("commit seal"),
				},
				&types.MsgCommit{
					View:       &types.View{Sequence: 101, Round: 2},
					From:       []byte("other validator"),
					BlockHash:  []byte("block hash"),
					CommitSeal: []byte("other commit seal"),
				},
			}),
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
