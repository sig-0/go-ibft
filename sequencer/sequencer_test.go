package sequencer

import "github.com/sig-0/go-ibft/test/mock"

//
//import (
//	"context"
//	"reflect"
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/sig-0/go-ibft"
//	"github.com/sig-0/go-ibft/message/types"
//	"github.com/sig-0/go-ibft/test/mock"
//)

var (
	Alice = mock.NewValidatorID("alice")
	Bob   = mock.NewValidatorID("bob")
	Chris = mock.NewValidatorID("chris")
	Dani  = mock.NewValidatorID("dani")

	//ValidatorSet = mock.NewValidatorSet(Alice, Bob, Chris, Dani)
)

//func TestSequencerFinalizeCancelled(t *testing.T) {
//	t.Parallel()
//
//	v := mock.Validator{
//		IDFn:        Alice.ID,
//		SigVerifier: mock.SigVerifier{IsProposerFn: func(_ []byte, _, _ uint64) bool { return false }},
//	}
//
//	seq := NewSequencer(v, 10*time.Millisecond)
//	c, cancel := context.WithCancel(context.Background())
//	ch := make(chan *types.FinalizedProposal)
//
//	go func(c context.Context) {
//		defer close(ch)
//
//		ctx := NewContext(c)
//		ctx = ctx.WithMsgFeed(mock.NewSingleRoundFeed(nil))
//		ch <- seq.Finalize(ctx, 101)
//	}(c)
//
//	cancel()
//
//	assert.Nil(t, <-ch)
//}
//
//func TestSequencerFinalize(t *testing.T) {
//	t.Parallel()
//
//	testTable := []struct {
//		validator ibft.Validator
//		quorum    ibft.Quorum
//		feed      MsgFeed
//		expected  *types.FinalizedProposal
//		name      string
//	}{
//		{
//			name:   "validator is not the proposer",
//			quorum: mock.NonZeroQuorum,
//			validator: mock.Validator{
//				IDFn:   Alice.ID,
//				Signer: Alice.Signer(),
//				SigVerifier: mock.SigVerifier{
//					IsValidSignatureFn: mock.OkSignature,
//					IsValidProposalFn:  mock.OkBlock,
//					IsValidatorFn:      ValidatorSet.IsValidator,
//					IsProposerFn:       mock.ProposersInRounds(mock.Proposer{ID: Bob, Round: 0}),
//				},
//			},
//
//			feed: mock.NewMessageFeed([]types.Message{
//				&types.MsgProposal{
//					Metadata: &types.MsgMetadata{
//						Sender: Bob,
//						View:   &types.View{Sequence: 101, Round: 0},
//					},
//					BlockHash:     []byte("block hash"),
//					ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 0},
//				},
//
//				&types.MsgPrepare{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 0},
//					},
//
//					BlockHash: []byte("block hash"),
//				},
//
//				&types.MsgCommit{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 0},
//					},
//
//					BlockHash:  []byte("block hash"),
//					CommitSeal: []byte("commit seal"),
//				},
//			}),
//
//			expected: &types.FinalizedProposal{
//				Proposal: []byte("block"),
//				Round:    0,
//				Seals: []types.FinalizedSeal{
//					{
//						From:       Alice,
//						CommitSeal: []byte("commit seal"),
//					},
//				},
//			},
//		},
//
//		{
//			name:   "validator is the proposer",
//			quorum: mock.NonZeroQuorum,
//			validator: mock.Validator{
//				IDFn:            Alice.ID,
//				Signer:          Alice.Signer(),
//				BuildProposalFn: func(_ uint64) []byte { return []byte("block") },
//				SigVerifier: mock.SigVerifier{
//					IsValidSignatureFn: mock.OkSignature,
//					IsProposerFn:       mock.ProposersInRounds(mock.Proposer{ID: Alice, Round: 0}),
//					IsValidatorFn:      ValidatorSet.IsValidator,
//				},
//			},
//
//			feed: mock.NewMessageFeed([]types.Message{
//				&types.MsgPrepare{
//					Metadata: &types.MsgMetadata{
//						Sender: Bob,
//						View:   &types.View{Sequence: 101, Round: 0},
//					},
//
//					BlockHash: []byte("block hash"),
//				},
//
//				&types.MsgCommit{
//					Metadata: &types.MsgMetadata{
//						Sender: Bob,
//						View:   &types.View{Sequence: 101, Round: 0},
//					},
//					BlockHash:  []byte("block hash"),
//					CommitSeal: []byte("commit seal"),
//				},
//			}),
//
//			expected: &types.FinalizedProposal{
//				Proposal: []byte("block"),
//				Round:    0,
//				Seals: []types.FinalizedSeal{
//					{
//						From:       Bob,
//						CommitSeal: []byte("commit seal"),
//					},
//				},
//			},
//		},
//
//		{
//			name:   "round 1 proposal is valid with empty PB and PC",
//			quorum: mock.NonZeroQuorum,
//			validator: mock.Validator{
//				IDFn:   Alice.ID,
//				Signer: Alice.Signer(),
//				SigVerifier: mock.SigVerifier{
//					IsValidSignatureFn: mock.OkSignature,
//					IsValidProposalFn:  mock.OkBlock,
//					IsValidatorFn:      ValidatorSet.IsValidator,
//					IsProposerFn:       mock.ProposersInRounds(mock.Proposer{ID: Bob, Round: 1}),
//				},
//			},
//
//			feed: mock.NewMessageFeed([]types.Message{
//				&types.MsgProposal{
//					Metadata: &types.MsgMetadata{
//						Sender: Bob,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//					BlockHash:     []byte("block hash"),
//					ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
//					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
//						{
//							Metadata: &types.MsgMetadata{
//								View:   &types.View{Sequence: 101, Round: 1},
//								Sender: Alice,
//							},
//						},
//					}},
//				},
//
//				&types.MsgPrepare{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//					BlockHash: []byte("block hash"),
//				},
//
//				&types.MsgCommit{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//					BlockHash:  []byte("block hash"),
//					CommitSeal: []byte("commit seal"),
//				},
//			}),
//
//			expected: &types.FinalizedProposal{
//				Proposal: []byte("block"),
//				Round:    1,
//				Seals: []types.FinalizedSeal{
//					{
//						From:       Alice,
//						CommitSeal: []byte("commit seal"),
//					},
//				},
//			},
//		},
//
//		{
//			name:   "round 1 proposal is valid with non-nil PB and PC",
//			quorum: mock.NonZeroQuorum,
//			validator: mock.Validator{
//				IDFn:   Alice.ID,
//				Signer: Alice.Signer(),
//				SigVerifier: mock.SigVerifier{
//					IsValidSignatureFn: mock.OkSignature,
//					IsValidProposalFn:  mock.OkBlock,
//					IsValidatorFn:      ValidatorSet.IsValidator,
//					IsProposerFn: mock.ProposersInRounds(
//						mock.Proposer{ID: Chris, Round: 0},
//						mock.Proposer{ID: Bob, Round: 1},
//					),
//				},
//			},
//
//			feed: mock.NewMessageFeed([]types.Message{
//				&types.MsgProposal{
//					Metadata: &types.MsgMetadata{
//						Sender: Bob,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//
//					BlockHash:     []byte("block hash"),
//					ProposedBlock: &types.ProposedBlock{Block: []byte("round 0 block"), Round: 1},
//					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
//						{
//							Metadata: &types.MsgMetadata{
//								Sender: Chris,
//								View:   &types.View{Sequence: 101, Round: 1},
//							},
//
//							LatestPreparedCertificate: &types.PreparedCertificate{
//								ProposalMessage: &types.MsgProposal{
//									Metadata: &types.MsgMetadata{
//										Sender: Chris,
//										View:   &types.View{Sequence: 101, Round: 0},
//									},
//
//									BlockHash: []byte("block hash"),
//									ProposedBlock: &types.ProposedBlock{
//										Block: []byte("round 0 block"),
//										Round: 0,
//									},
//								},
//								PrepareMessages: []*types.MsgPrepare{
//									{
//										Metadata: &types.MsgMetadata{
//											Sender: Alice,
//											View:   &types.View{Sequence: 101, Round: 0},
//										},
//
//										BlockHash: []byte("block hash"),
//									},
//								},
//							},
//						},
//					}},
//				},
//
//				&types.MsgPrepare{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//
//					BlockHash: []byte("block hash"),
//				},
//
//				&types.MsgCommit{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//
//					BlockHash:  []byte("block hash"),
//					CommitSeal: []byte("commit seal"),
//				},
//			}),
//
//			expected: &types.FinalizedProposal{
//				Proposal: []byte("round 0 block"),
//				Round:    1,
//				Seals: []types.FinalizedSeal{
//					{
//						From:       Alice,
//						CommitSeal: []byte("commit seal"),
//					},
//				},
//			},
//		},
//
//		{
//			name:   "block proposed in round 1",
//			quorum: mock.NonZeroQuorum,
//			validator: mock.Validator{
//				IDFn:            Alice.ID,
//				Signer:          Alice.Signer(),
//				BuildProposalFn: func(_ uint64) []byte { return []byte("round 1 block") },
//				SigVerifier: mock.SigVerifier{
//					IsValidSignatureFn: mock.OkSignature,
//					IsValidProposalFn:  mock.OkBlock,
//					IsValidatorFn:      ValidatorSet.IsValidator,
//					IsProposerFn: mock.ProposersInRounds(
//						mock.Proposer{ID: Bob, Round: 0},
//						mock.Proposer{ID: Alice, Round: 1},
//					),
//				},
//			},
//
//			feed: mock.NewMessageFeed([]types.Message{
//				// need to justify Alice's proposal for round 1
//				&types.MsgRoundChange{
//					Metadata: &types.MsgMetadata{
//						View:   &types.View{Sequence: 101, Round: 1},
//						Sender: Alice,
//					},
//				},
//
//				&types.MsgPrepare{
//					Metadata: &types.MsgMetadata{
//						View:   &types.View{Sequence: 101, Round: 1},
//						Sender: Bob,
//					},
//
//					BlockHash: []byte("block hash"),
//				},
//
//				&types.MsgCommit{
//					Metadata: &types.MsgMetadata{
//						View:   &types.View{Sequence: 101, Round: 1},
//						Sender: Bob,
//					},
//
//					BlockHash:  []byte("block hash"),
//					CommitSeal: []byte("commit seal"),
//				},
//			}),
//
//			expected: &types.FinalizedProposal{
//				Proposal: []byte("round 1 block"),
//				Round:    1,
//				Seals: []types.FinalizedSeal{
//					{
//						From:       Bob,
//						CommitSeal: []byte("commit seal"),
//					},
//				},
//			},
//		},
//
//		{
//			name:   "old block proposed in round 1",
//			quorum: mock.NonZeroQuorum,
//			validator: mock.Validator{
//				IDFn:   Alice.ID,
//				Signer: Alice.Signer(),
//				SigVerifier: mock.SigVerifier{
//					IsValidSignatureFn: mock.OkSignature,
//					IsValidProposalFn:  mock.OkBlock,
//					IsValidatorFn:      ValidatorSet.IsValidator,
//					IsProposerFn: mock.ProposersInRounds(
//						mock.Proposer{ID: Bob, Round: 0},
//						mock.Proposer{ID: Alice, Round: 1},
//					),
//				},
//			},
//
//			feed: mock.NewMessageFeed([]types.Message{
//				&types.MsgRoundChange{
//					Metadata: &types.MsgMetadata{
//						Sender: Chris,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//
//					LatestPreparedProposedBlock: &types.ProposedBlock{Block: []byte("round 0 block"), Round: 0},
//					LatestPreparedCertificate: &types.PreparedCertificate{
//						ProposalMessage: &types.MsgProposal{
//							Metadata: &types.MsgMetadata{
//								Sender: Bob,
//								View:   &types.View{Sequence: 101, Round: 0},
//							},
//
//							BlockHash:     []byte("block hash"),
//							ProposedBlock: &types.ProposedBlock{Block: []byte("round 0 block"), Round: 0},
//						},
//
//						PrepareMessages: []*types.MsgPrepare{
//							{
//								Metadata: &types.MsgMetadata{
//									Sender: Chris,
//									View:   &types.View{Sequence: 101, Round: 0},
//								},
//								BlockHash: []byte("block hash"),
//							},
//						},
//					},
//				},
//
//				&types.MsgPrepare{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//					BlockHash: []byte("block hash"),
//				},
//
//				&types.MsgCommit{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//
//					BlockHash:  []byte("block hash"),
//					CommitSeal: []byte("commit seal"),
//				},
//			}),
//
//			expected: &types.FinalizedProposal{
//				Proposal: []byte("round 0 block"),
//				Round:    1,
//				Seals: []types.FinalizedSeal{
//					{
//						From:       Alice,
//						CommitSeal: []byte("commit seal"),
//					},
//				},
//			},
//		},
//
//		{
//			name:   "future rcc triggers round jump",
//			quorum: mock.NonZeroQuorum,
//			validator: mock.Validator{
//				IDFn:   Alice.ID,
//				Signer: Alice.Signer(),
//				SigVerifier: mock.SigVerifier{
//					IsValidSignatureFn: mock.OkSignature,
//					IsValidProposalFn:  mock.OkBlock,
//					IsValidatorFn:      ValidatorSet.IsValidator,
//					IsProposerFn:       mock.ProposersInRounds(mock.Proposer{ID: Chris, Round: 3}),
//				},
//			},
//
//			feed: mock.NewMessageFeed([]types.Message{
//				&types.MsgProposal{
//					Metadata: &types.MsgMetadata{
//						Sender: Chris,
//						View:   &types.View{Sequence: 101, Round: 3},
//					},
//
//					BlockHash:     []byte("block hash"),
//					ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 3},
//					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
//						{
//							Metadata: &types.MsgMetadata{
//								Sender: Bob,
//								View:   &types.View{Sequence: 101, Round: 3},
//							},
//						},
//					}},
//				},
//
//				&types.MsgPrepare{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 3},
//					},
//
//					BlockHash: []byte("block hash"),
//				},
//
//				&types.MsgCommit{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 3},
//					},
//
//					BlockHash:  []byte("block hash"),
//					CommitSeal: []byte("commit seal"),
//				},
//			}),
//
//			expected: &types.FinalizedProposal{
//				Proposal: []byte("block"),
//				Round:    3,
//				Seals: []types.FinalizedSeal{
//					{
//						From:       Alice,
//						CommitSeal: []byte("commit seal"),
//					},
//				},
//			},
//		},
//
//		{
//			name:   "future proposal triggers round jump",
//			quorum: mock.NonZeroQuorum,
//			validator: mock.Validator{
//				IDFn:   Alice.ID,
//				Signer: Alice.Signer(),
//				SigVerifier: mock.SigVerifier{
//					IsValidSignatureFn: mock.OkSignature,
//					IsValidProposalFn:  mock.OkBlock,
//					IsValidatorFn:      ValidatorSet.IsValidator,
//					IsProposerFn:       mock.ProposersInRounds(mock.Proposer{ID: Dani, Round: 5}),
//				},
//			},
//
//			feed: mock.NewMessageFeed([]types.Message{
//				&types.MsgProposal{
//					Metadata: &types.MsgMetadata{
//						Sender: Dani,
//						View:   &types.View{Sequence: 101, Round: 5},
//					},
//
//					ProposedBlock: &types.ProposedBlock{Block: []byte("round 5 block"), Round: 5},
//					BlockHash:     []byte("block hash"),
//					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
//						{
//							Metadata: &types.MsgMetadata{
//								View:   &types.View{Sequence: 101, Round: 5},
//								Sender: Chris,
//							},
//						},
//					}},
//				},
//
//				&types.MsgPrepare{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 5},
//					},
//
//					BlockHash: []byte("block hash"),
//				},
//
//				&types.MsgCommit{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 5},
//					},
//
//					BlockHash:  []byte("block hash"),
//					CommitSeal: []byte("commit seal"),
//				},
//			}),
//
//			expected: &types.FinalizedProposal{
//				Proposal: []byte("round 5 block"),
//				Seals: []types.FinalizedSeal{
//					{
//						From:       Alice,
//						CommitSeal: []byte("commit seal"),
//					},
//				},
//				Round: 5,
//			},
//		},
//
//		{
//			name:   "round timer triggers round jump",
//			quorum: mock.NonZeroQuorum,
//			validator: mock.Validator{
//				IDFn:            Alice.ID,
//				Signer:          Alice.Signer(),
//				BuildProposalFn: func(_ uint64) []byte { return []byte("round 1 block") },
//				SigVerifier: mock.SigVerifier{
//					IsValidSignatureFn: mock.OkSignature,
//					IsValidProposalFn:  mock.OkBlock,
//					IsValidatorFn:      ValidatorSet.IsValidator,
//					IsProposerFn:       mock.ProposersInRounds(mock.Proposer{ID: Alice, Round: 1}),
//				},
//			},
//
//			feed: mock.NewSingleRoundFeed([]types.Message{
//				&types.MsgRoundChange{
//					Metadata: &types.MsgMetadata{
//						Sender: Bob,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//				},
//
//				&types.MsgPrepare{
//					Metadata: &types.MsgMetadata{
//						Sender: Bob,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//
//					BlockHash: []byte("block hash"),
//				},
//
//				&types.MsgCommit{
//					Metadata: &types.MsgMetadata{
//						Sender: Bob,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//
//					BlockHash:  []byte("block hash"),
//					CommitSeal: []byte("commit seal"),
//				},
//			}),
//
//			expected: &types.FinalizedProposal{
//				Proposal: []byte("round 1 block"),
//				Round:    1,
//				Seals: []types.FinalizedSeal{
//					{
//						From:       Bob,
//						CommitSeal: []byte("commit seal"),
//					},
//				},
//			},
//		},
//
//		{
//			name:   "no prepare messages in round 0",
//			quorum: mock.NonZeroQuorum,
//			validator: mock.Validator{
//				IDFn:   Alice.ID,
//				Signer: Alice.Signer(),
//				SigVerifier: mock.SigVerifier{
//					IsValidSignatureFn: mock.OkSignature,
//					IsValidProposalFn:  mock.OkBlock,
//					IsValidatorFn:      ValidatorSet.IsValidator,
//					IsProposerFn: mock.ProposersInRounds(
//						mock.Proposer{ID: Bob, Round: 0},
//						mock.Proposer{ID: Chris, Round: 1},
//					),
//				},
//			},
//
//			feed: mock.NewSingleRoundFeed([]types.Message{
//				&types.MsgProposal{
//					Metadata: &types.MsgMetadata{
//						Sender: Bob,
//						View:   &types.View{Sequence: 101, Round: 0},
//					},
//
//					BlockHash:     []byte("block hash"),
//					ProposedBlock: &types.ProposedBlock{Block: []byte("round 0 block"), Round: 0},
//				},
//
//				&types.MsgProposal{
//					Metadata: &types.MsgMetadata{
//						Sender: Chris,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//
//					BlockHash:     []byte("block hash"),
//					ProposedBlock: &types.ProposedBlock{Block: []byte("round 1 block"), Round: 1},
//					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
//						{
//							Metadata: &types.MsgMetadata{
//								View:   &types.View{Sequence: 101, Round: 1},
//								Sender: Alice,
//							},
//						},
//					}},
//				},
//
//				&types.MsgPrepare{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//
//					BlockHash: []byte("block hash"),
//				},
//
//				&types.MsgCommit{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//					BlockHash:  []byte("block hash"),
//					CommitSeal: []byte("commit seal"),
//				},
//			}),
//
//			expected: &types.FinalizedProposal{
//				Proposal: []byte("round 1 block"),
//				Round:    1,
//				Seals: []types.FinalizedSeal{
//					{
//						From:       Alice,
//						CommitSeal: []byte("commit seal"),
//					},
//				},
//			},
//		},
//
//		{
//			name:   "no commit messages in round 0",
//			quorum: mock.QuorumOf(2),
//			validator: mock.Validator{
//				IDFn:   Alice.ID,
//				Signer: Alice.Signer(),
//				SigVerifier: mock.SigVerifier{
//					IsValidSignatureFn: mock.OkSignature,
//					IsValidProposalFn:  mock.OkBlock,
//					IsValidatorFn:      ValidatorSet.IsValidator,
//					IsProposerFn: mock.ProposersInRounds(
//						mock.Proposer{ID: Bob, Round: 0},
//						mock.Proposer{ID: Chris, Round: 1},
//					),
//				},
//			},
//
//			feed: mock.NewSingleRoundFeed([]types.Message{
//				/* round 0 */
//
//				&types.MsgProposal{
//					Metadata: &types.MsgMetadata{
//						Sender: Bob,
//						View:   &types.View{Sequence: 101, Round: 0},
//					},
//
//					BlockHash:     []byte("block hash"),
//					ProposedBlock: &types.ProposedBlock{Block: []byte("round 0 block"), Round: 0},
//				},
//
//				&types.MsgPrepare{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 0},
//					},
//
//					BlockHash: []byte("block hash"),
//				},
//				&types.MsgPrepare{
//					Metadata: &types.MsgMetadata{
//						Sender: Bob,
//						View:   &types.View{Sequence: 101, Round: 0},
//					},
//					BlockHash: []byte("block hash"),
//				},
//
//				/* round 1 */
//
//				&types.MsgProposal{
//					Metadata: &types.MsgMetadata{
//						Sender: Chris,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//
//					BlockHash:     []byte("block hash"),
//					ProposedBlock: &types.ProposedBlock{Block: []byte("round 1 block"), Round: 1},
//					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
//						{
//							Metadata: &types.MsgMetadata{
//								Sender: Alice,
//								View:   &types.View{Sequence: 101, Round: 1},
//							},
//						},
//						{
//							Metadata: &types.MsgMetadata{
//								Sender: Bob,
//								View:   &types.View{Sequence: 101, Round: 1},
//							},
//						},
//					}},
//				},
//
//				&types.MsgPrepare{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//
//					BlockHash: []byte("block hash"),
//				},
//				&types.MsgPrepare{
//					Metadata: &types.MsgMetadata{
//						Sender: Bob,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//
//					BlockHash: []byte("block hash"),
//				},
//
//				&types.MsgCommit{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//					BlockHash:  []byte("block hash"),
//					CommitSeal: []byte("commit seal"),
//				},
//				&types.MsgCommit{
//					Metadata: &types.MsgMetadata{
//						Sender: Bob,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//					BlockHash:  []byte("block hash"),
//					CommitSeal: []byte("other commit seal"),
//				},
//			}),
//
//			expected: &types.FinalizedProposal{
//				Proposal: []byte("round 1 block"),
//				Round:    1,
//				Seals: []types.FinalizedSeal{
//					{
//						From:       Alice,
//						CommitSeal: []byte("commit seal"),
//					},
//					{
//						From:       Bob,
//						CommitSeal: []byte("other commit seal"),
//					},
//				},
//			},
//		},
//
//		{
//			name:   "round 0 proposer fails to build block",
//			quorum: mock.QuorumOf(2),
//			validator: mock.Validator{
//				IDFn:   Alice.ID,
//				Signer: Alice.Signer(),
//				SigVerifier: mock.SigVerifier{
//					IsValidSignatureFn: mock.OkSignature,
//					IsValidProposalFn:  mock.OkBlock,
//					IsValidatorFn:      ValidatorSet.IsValidator,
//					IsProposerFn: mock.ProposersInRounds(
//						mock.Proposer{ID: Bob, Round: 0},
//						mock.Proposer{ID: Alice, Round: 1},
//						mock.Proposer{ID: Chris, Round: 2},
//					),
//				},
//			},
//
//			feed: mock.NewSingleRoundFeed([]types.Message{
//				&types.MsgRoundChange{
//					Metadata: &types.MsgMetadata{
//						Sender: Chris,
//						View:   &types.View{Sequence: 101, Round: 1},
//					},
//				},
//
//				&types.MsgProposal{
//					Metadata: &types.MsgMetadata{
//						Sender: Chris,
//						View:   &types.View{Sequence: 101, Round: 2},
//					},
//
//					ProposedBlock: &types.ProposedBlock{Block: []byte("round 2 block"), Round: 2},
//					BlockHash:     []byte("block hash"),
//					RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
//						{
//							Metadata: &types.MsgMetadata{
//								Sender: Bob,
//								View:   &types.View{Sequence: 101, Round: 2},
//							},
//						},
//						{
//							Metadata: &types.MsgMetadata{
//								Sender: Chris,
//								View:   &types.View{Sequence: 101, Round: 2},
//							},
//						},
//					}},
//				},
//
//				&types.MsgPrepare{
//					Metadata: &types.MsgMetadata{
//						Sender: Bob,
//						View:   &types.View{Sequence: 101, Round: 2},
//					},
//
//					BlockHash: []byte("block hash"),
//				},
//				&types.MsgPrepare{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 2},
//					},
//
//					BlockHash: []byte("block hash"),
//				},
//
//				&types.MsgCommit{
//					Metadata: &types.MsgMetadata{
//						Sender: Bob,
//						View:   &types.View{Sequence: 101, Round: 2},
//					},
//					BlockHash:  []byte("block hash"),
//					CommitSeal: []byte("commit seal"),
//				},
//				&types.MsgCommit{
//					Metadata: &types.MsgMetadata{
//						Sender: Alice,
//						View:   &types.View{Sequence: 101, Round: 2},
//					},
//					BlockHash:  []byte("block hash"),
//					CommitSeal: []byte("other commit seal"),
//				},
//			}),
//
//			expected: &types.FinalizedProposal{
//				Proposal: []byte("round 2 block"),
//				Round:    2,
//				Seals: []types.FinalizedSeal{
//					{
//						From:       Bob,
//						CommitSeal: []byte("commit seal"),
//					},
//					{
//						From:       Alice,
//						CommitSeal: []byte("other commit seal"),
//					},
//				},
//			},
//		},
//	}
//
//	for _, tt := range testTable {
//		tt := tt
//		t.Run(tt.name, func(t *testing.T) {
//			t.Parallel()
//
//			ctx := NewContext(context.Background())
//			ctx = ctx.WithQuorum(tt.quorum)
//			ctx = ctx.WithMsgFeed(tt.feed)
//			ctx = ctx.WithKeccak(mock.DummyKeccak("block hash"))
//			ctx = ctx.WithTransport(mock.DummyTransport())
//
//			s := NewSequencer(tt.validator, time.Millisecond*10)
//			assert.True(t, reflect.DeepEqual(tt.expected, s.Finalize(ctx, 101)))
//		})
//	}
//}
