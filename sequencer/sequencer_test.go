package sequencer

import (
	"bytes"
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

var (
	/* Common test values */

	// ibft.Context fields
	DummyTransport  = TransportFn(func(_ ibft.Message) {})
	NonZeroQuorum   = QuorumFn(func(msgs []ibft.Message) bool { return len(msgs) != 0 })
	BlockHashKeccak = KeccakFn(func(_ []byte) []byte { return []byte("block hash") })

	// ibft.Verifier methods
	TrueBlock     = func(_ []byte) bool { return true }
	TrueSignature = func(_, _, _ []byte) bool { return true }
	TrueValidator = func(_ []byte, _ uint64) bool { return true }

	// ibft.Validator methods
	NilSignature = func(_ []byte) []byte { return nil }
	MyValidator  = func() []byte { return []byte("my validator") }
)

type testTable struct {
	validator ibft.Validator
	verifier  ibft.Verifier

	transport ibft.Transport
	msgFeed   ibft.MessageFeed
	quorumFn  ibft.Quorum
	keccakFn  ibft.Keccak

	expectedFinalizedBlock *types.FinalizedProposal
	name                   string
}

func Test_Sequencer_Finalize_Sequence_Cancelled(t *testing.T) {
	t.Parallel()

	var (
		validator = MockValidator{IDFn: func() []byte { return nil }}
		verifier  = MockVerifier{IsProposerFn: func(_ []byte, _, _ uint64) bool { return false }}
		msgFeed   = singleRoundFeed(MockFeed{
			proposal: messagesByView[*types.MsgProposal]{
				101: {},
			},
		})

		seq = New(validator, verifier, 10*time.Millisecond)
	)

	ctx, cancelCtx := context.WithCancel(context.Background())
	c := make(chan *types.FinalizedProposal)

	go func(ctx context.Context) {
		defer close(c)

		c <- seq.FinalizeSequence(ibft.NewIBFTContext(ctx).WithFeed(msgFeed), 101)
	}(ctx)

	cancelCtx()

	assert.Nil(t, <-c)
}

//nolint:dupl // consensus messages are not entirely different among cases
func Test_Sequencer_Finalize_Sequence(t *testing.T) {
	t.Parallel()

	table := []testTable{
		{
			name: "validator is not the proposer",
			expectedFinalizedBlock: &types.FinalizedProposal{
				Proposal: []byte("block"),
				Round:    0,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: MockValidator{
				IDFn:   MyValidator,
				SignFn: NilSignature,
			},

			verifier: MockVerifier{
				IsValidSignatureFn: TrueSignature,
				IsValidBlockFn:     TrueBlock,
				IsProposerFn:       func(from []byte, _, _ uint64) bool { return bytes.Equal(from, []byte("proposer")) },
				IsValidatorFn:      TrueValidator,
			},

			keccakFn:  BlockHashKeccak,
			transport: DummyTransport,
			quorumFn:  NonZeroQuorum,
			msgFeed: allRoundsFeed(MockFeed{
				proposal: messagesByView[*types.MsgProposal]{
					101: {
						0: {
							{
								View:      &types.View{Sequence: 101, Round: 0},
								From:      []byte("proposer"),
								BlockHash: []byte("block hash"),
								ProposedBlock: &types.ProposedBlock{
									Block: []byte("block"),
									Round: 0,
								},
							},
						},
					},
				},

				prepare: messagesByView[*types.MsgPrepare]{
					101: {
						0: {
							{
								View:      &types.View{Sequence: 101, Round: 0},
								From:      []byte("some validator"),
								BlockHash: []byte("block hash"),
							},
						},
					},
				},

				commit: messagesByView[*types.MsgCommit]{
					101: {
						0: {
							{
								View:       &types.View{Sequence: 101, Round: 0},
								From:       []byte("some validator"),
								BlockHash:  []byte("block hash"),
								CommitSeal: []byte("commit seal"),
							},
						},
					},
				},
			}),
		},

		{
			name: "validator is the proposer",
			expectedFinalizedBlock: &types.FinalizedProposal{
				Proposal: []byte("block"),
				Round:    0,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("some validator"),
					},
				},
			},

			validator: MockValidator{
				IDFn:            MyValidator,
				SignFn:          NilSignature,
				BuildProposalFn: func() []byte { return []byte("block") },
			},

			verifier: MockVerifier{
				IsValidSignatureFn: TrueSignature,
				IsProposerFn:       func(_ []byte, _, _ uint64) bool { return true },
				IsValidatorFn:      TrueValidator,
			},

			keccakFn:  BlockHashKeccak,
			transport: DummyTransport,
			quorumFn:  NonZeroQuorum,
			msgFeed: allRoundsFeed(MockFeed{
				prepare: messagesByView[*types.MsgPrepare]{
					101: {
						0: {
							{
								View:      &types.View{Sequence: 101, Round: 0},
								From:      []byte("some validator"),
								BlockHash: []byte("block hash"),
							},
						},
					},
				},

				commit: messagesByView[*types.MsgCommit]{
					101: {
						0: {
							{
								View:       &types.View{Sequence: 101, Round: 0},
								From:       []byte("some validator"),
								BlockHash:  []byte("block hash"),
								CommitSeal: []byte("some validator"),
							},
						},
					},
				},
			}),
		},

		{
			name: "round 1 proposal is valid with empty PB and PC",
			expectedFinalizedBlock: &types.FinalizedProposal{
				Proposal: []byte("block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("my validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: MockValidator{
				IDFn:   MyValidator,
				SignFn: func(_ []byte) []byte { return []byte("commit seal") },
			},

			verifier: MockVerifier{
				IsValidSignatureFn: TrueSignature,
				IsValidBlockFn:     TrueBlock,
				IsValidatorFn:      TrueValidator,
				IsProposerFn: func(from []byte, _, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
			},

			keccakFn:  BlockHashKeccak,
			transport: DummyTransport,
			quorumFn:  NonZeroQuorum,
			msgFeed: allRoundsFeed(MockFeed{
				proposal: messagesByView[*types.MsgProposal]{
					101: {
						0: nil,
						1: {
							{
								View:      &types.View{Sequence: 101, Round: 1},
								From:      []byte("proposer"),
								BlockHash: []byte("block hash"),
								ProposedBlock: &types.ProposedBlock{
									Block: []byte("block"),
									Round: 1,
								},
								RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
									{
										View: &types.View{Sequence: 101, Round: 1},
										From: []byte("my validator"),
									},
								}},
							},
						},
					},
				},

				prepare: messagesByView[*types.MsgPrepare]{
					101: {
						1: {
							{
								View:      &types.View{Sequence: 101, Round: 1},
								From:      []byte("my validator"),
								BlockHash: []byte("block hash"),
							},
						},
					},
				},

				commit: messagesByView[*types.MsgCommit]{
					101: {
						1: {
							{
								View:       &types.View{Sequence: 101, Round: 1},
								From:       []byte("my validator"),
								BlockHash:  []byte("block hash"),
								CommitSeal: []byte("commit seal"),
							},
						},
					},
				},
			}),
		},

		{
			name: "round 1 proposal is valid with non-nil PB and PC",
			expectedFinalizedBlock: &types.FinalizedProposal{
				Proposal: []byte("round 0 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: MockValidator{
				IDFn:   MyValidator,
				SignFn: NilSignature,
			},

			verifier: MockVerifier{
				IsValidSignatureFn: TrueSignature,
				IsValidBlockFn:     TrueBlock,
				IsProposerFn: func(from []byte, _, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidatorFn: TrueValidator,
			},

			keccakFn:  BlockHashKeccak,
			transport: DummyTransport,
			quorumFn:  NonZeroQuorum,
			msgFeed: allRoundsFeed(MockFeed{
				proposal: messagesByView[*types.MsgProposal]{
					101: {
						0: nil,
						1: {
							{
								View:      &types.View{Sequence: 101, Round: 1},
								From:      []byte("proposer"),
								BlockHash: []byte("block hash"),
								ProposedBlock: &types.ProposedBlock{
									Block: []byte("round 0 block"),
									Round: 1,
								},
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
						},
					},
				},

				prepare: messagesByView[*types.MsgPrepare]{
					101: {
						1: {
							{
								View:      &types.View{Sequence: 101, Round: 1},
								From:      []byte("some validator"),
								BlockHash: []byte("block hash"),
							},
						},
					},
				},

				commit: messagesByView[*types.MsgCommit]{
					101: {
						1: {
							{
								View:       &types.View{Sequence: 101, Round: 1},
								From:       []byte("some validator"),
								BlockHash:  []byte("block hash"),
								CommitSeal: []byte("commit seal"),
							},
						},
					},
				},
			}),
		},

		{
			name: "propose new block in round 1",
			expectedFinalizedBlock: &types.FinalizedProposal{
				Proposal: []byte("round 1 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: MockValidator{
				IDFn:            MyValidator,
				SignFn:          NilSignature,
				BuildProposalFn: func() []byte { return []byte("round 1 block") },
			},

			verifier: MockVerifier{
				IsValidSignatureFn: TrueSignature,
				IsValidBlockFn:     TrueBlock,
				IsValidatorFn:      TrueValidator,
				IsProposerFn: func(from []byte, _, round uint64) bool {
					if round == 0 {
						return false
					}

					return bytes.Equal(from, []byte("my validator"))
				},
			},

			keccakFn:  BlockHashKeccak,
			transport: DummyTransport,
			quorumFn:  NonZeroQuorum,
			msgFeed: allRoundsFeed(MockFeed{
				proposal: messagesByView[*types.MsgProposal]{
					101: {
						0: nil,
					},
				},

				prepare: messagesByView[*types.MsgPrepare]{
					101: {
						1: {
							{
								View:      &types.View{Sequence: 101, Round: 1},
								From:      []byte("some validator"),
								BlockHash: []byte("block hash"),
							},
						},
					},
				},

				commit: messagesByView[*types.MsgCommit]{
					101: {
						1: {
							{
								View:       &types.View{Sequence: 101, Round: 1},
								From:       []byte("some validator"),
								BlockHash:  []byte("block hash"),
								CommitSeal: []byte("commit seal"),
							},
						},
					},
				},

				roundChange: messagesByView[*types.MsgRoundChange]{
					101: {
						1: {
							{
								View: &types.View{Sequence: 101, Round: 1},
								From: []byte("some validator"),
							},
						},
					},
				},
			}),
		},

		{
			name: "propose old block in round 1",
			expectedFinalizedBlock: &types.FinalizedProposal{
				Proposal: []byte("round 0 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: MockValidator{
				IDFn:   MyValidator,
				SignFn: NilSignature,
			},

			verifier: MockVerifier{
				IsValidSignatureFn: TrueSignature,
				IsValidBlockFn:     TrueBlock,
				IsValidatorFn:      TrueValidator,
				IsProposerFn: func(from []byte, _, round uint64) bool {
					if round == 1 {
						return bytes.Equal(from, []byte("my validator"))
					}

					return bytes.Equal(from, []byte("proposer"))
				},
			},

			keccakFn:  BlockHashKeccak,
			transport: DummyTransport,
			quorumFn:  NonZeroQuorum,
			msgFeed: allRoundsFeed(MockFeed{
				proposal: messagesByView[*types.MsgProposal]{
					101: {
						0: nil,
					},
				},

				roundChange: messagesByView[*types.MsgRoundChange]{
					101: {
						1: {
							{
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
						},
					},
				},

				prepare: messagesByView[*types.MsgPrepare]{
					101: {
						1: {
							{
								View:      &types.View{Sequence: 101, Round: 1},
								From:      []byte("some validator"),
								BlockHash: []byte("block hash"),
							},
						},
					},
				},

				commit: messagesByView[*types.MsgCommit]{
					101: {
						1: {
							{
								View:       &types.View{Sequence: 101, Round: 1},
								From:       []byte("some validator"),
								BlockHash:  []byte("block hash"),
								CommitSeal: []byte("commit seal"),
							},
						},
					},
				},
			}),
		},

		{
			name: "future rcc triggers round jump and new block is finalized",
			expectedFinalizedBlock: &types.FinalizedProposal{
				Proposal: []byte("block"),
				Round:    3,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: MockValidator{
				IDFn:   MyValidator,
				SignFn: func(_ []byte) []byte { return []byte("commit seal") },
			},

			verifier: MockVerifier{
				IsValidSignatureFn: TrueSignature,
				IsValidBlockFn:     TrueBlock,
				IsValidatorFn:      TrueValidator,
				IsProposerFn:       func(from []byte, _, _ uint64) bool { return bytes.Equal(from, []byte("proposer")) },
			},

			keccakFn:  BlockHashKeccak,
			transport: DummyTransport,
			quorumFn:  NonZeroQuorum,
			msgFeed: allRoundsFeed(MockFeed{
				proposal: messagesByView[*types.MsgProposal]{
					101: {
						3: {
							{
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
						},
					},
				},

				prepare: messagesByView[*types.MsgPrepare]{
					101: {
						3: {
							{
								View:      &types.View{Sequence: 101, Round: 3},
								From:      []byte("some validator"),
								BlockHash: []byte("block hash"),
							},
						},
					},
				},

				commit: messagesByView[*types.MsgCommit]{
					101: {
						3: {
							{
								View:       &types.View{Sequence: 101, Round: 3},
								From:       []byte("some validator"),
								BlockHash:  []byte("block hash"),
								CommitSeal: []byte("commit seal"),
							},
						},
					},
				},

				roundChange: messagesByView[*types.MsgRoundChange]{
					101: {
						3: {
							{
								View: &types.View{Sequence: 101, Round: 3},
								From: []byte("some validator"),
							},
						},
					},
				},
			}),
		},

		{
			name: "future proposal triggers round jump and new block is finalized",
			expectedFinalizedBlock: &types.FinalizedProposal{
				Proposal: []byte("round 5 block"),
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
				Round: 5,
			},

			validator: MockValidator{
				IDFn:   MyValidator,
				SignFn: NilSignature,
			},

			verifier: MockVerifier{
				IsValidSignatureFn: TrueSignature,
				IsValidBlockFn:     TrueBlock,
				IsProposerFn: func(from []byte, _, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidatorFn: TrueValidator,
			},

			keccakFn:  BlockHashKeccak,
			transport: DummyTransport,
			quorumFn:  NonZeroQuorum,
			msgFeed: allRoundsFeed(MockFeed{
				proposal: messagesByView[*types.MsgProposal]{
					101: {
						5: {
							{
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
						},
					},
				},

				prepare: messagesByView[*types.MsgPrepare]{
					101: {
						5: {
							{
								View: &types.View{
									Sequence: 101,
									Round:    5,
								},
								From:      []byte("some validator"),
								BlockHash: []byte("block hash"),
							},
						},
					},
				},

				commit: messagesByView[*types.MsgCommit]{
					101: {
						5: {
							{
								View: &types.View{
									Sequence: 101,
									Round:    5,
								},
								From:       []byte("some validator"),
								BlockHash:  []byte("block hash"),
								CommitSeal: []byte("commit seal"),
							},
						},
					},
				},
			}),
		},

		{
			name: "round timer triggers round jump and proposer finalizes new block",
			expectedFinalizedBlock: &types.FinalizedProposal{
				Proposal: []byte("round 1 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: MockValidator{
				IDFn:            MyValidator,
				SignFn:          NilSignature,
				BuildProposalFn: func() []byte { return []byte("round 1 block") },
			},

			verifier: MockVerifier{
				IsValidSignatureFn: TrueSignature,
				IsValidBlockFn:     TrueBlock,
				IsProposerFn: func(from []byte, _, round uint64) bool {
					return bytes.Equal(from, []byte("my validator")) && round == 1
				},
				IsValidatorFn: TrueValidator,
			},

			keccakFn:  BlockHashKeccak,
			transport: DummyTransport,
			quorumFn:  NonZeroQuorum,
			msgFeed: singleRoundFeed(MockFeed{
				prepare: messagesByView[*types.MsgPrepare]{
					101: {
						1: {
							{
								View:      &types.View{Sequence: 101, Round: 1},
								From:      []byte("some validator"),
								BlockHash: []byte("block hash"),
							},
						},
					},
				},
				commit: messagesByView[*types.MsgCommit]{
					101: {
						1: {
							{
								View:       &types.View{Sequence: 101, Round: 1},
								From:       []byte("some validator"),
								BlockHash:  []byte("block hash"),
								CommitSeal: []byte("commit seal"),
							},
						},
					},
				},
				roundChange: messagesByView[*types.MsgRoundChange]{
					101: {
						1: {
							{
								View: &types.View{Sequence: 101, Round: 1},
								From: []byte("some validator"),
							},
						},
					},
				},
			}),
		},

		{
			name: "round timer triggers round jump and proposer finalizes new block",
			expectedFinalizedBlock: &types.FinalizedProposal{
				Proposal: []byte("round 0 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: MockValidator{
				IDFn:   MyValidator,
				SignFn: NilSignature,
			},

			verifier: MockVerifier{
				IsValidSignatureFn: TrueSignature,
				IsValidBlockFn:     TrueBlock,
				IsProposerFn: func(from []byte, _, round uint64) bool {
					return bytes.Equal(from, []byte("proposer")) ||
						bytes.Equal(from, []byte("my validator")) && round == 1
				},
				IsValidatorFn: TrueValidator,
			},

			keccakFn:  BlockHashKeccak,
			transport: DummyTransport,
			quorumFn:  NonZeroQuorum,
			msgFeed: singleRoundFeed(MockFeed{
				prepare: messagesByView[*types.MsgPrepare]{
					101: {
						1: {
							{
								View:      &types.View{Sequence: 101, Round: 1},
								From:      []byte("some validator"),
								BlockHash: []byte("block hash"),
							},
						},
					},
				},
				commit: messagesByView[*types.MsgCommit]{
					101: {
						1: {
							{
								View:       &types.View{Sequence: 101, Round: 1},
								From:       []byte("some validator"),
								BlockHash:  []byte("block hash"),
								CommitSeal: []byte("commit seal"),
							},
						},
					},
				},
				roundChange: messagesByView[*types.MsgRoundChange]{
					101: {
						1: {
							{
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
						},
					},
				},
			}),
		},

		{
			name: "no prepare msgs in round 0 so new block is finalized in round 1",
			expectedFinalizedBlock: &types.FinalizedProposal{
				Proposal: []byte("round 1 block"),
				Round:    1,
				Seals: []types.FinalizedSeal{
					{
						From:       []byte("some validator"),
						CommitSeal: []byte("commit seal"),
					},
				},
			},

			validator: MockValidator{
				IDFn:   MyValidator,
				SignFn: NilSignature,
			},

			verifier: MockVerifier{
				IsValidSignatureFn: TrueSignature,
				IsValidBlockFn:     TrueBlock,
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidatorFn: TrueValidator,
			},

			keccakFn:  BlockHashKeccak,
			transport: DummyTransport,
			quorumFn:  NonZeroQuorum,
			msgFeed: singleRoundFeed(MockFeed{
				proposal: messagesByView[*types.MsgProposal]{
					101: {
						0: {
							{
								View:      &types.View{Sequence: 101, Round: 0},
								From:      []byte("proposer"),
								BlockHash: []byte("block hash"),
								ProposedBlock: &types.ProposedBlock{
									Block: []byte("round 0 block"),
									Round: 0,
								},
							},
						},

						1: {
							{
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
						},
					},
				},

				prepare: messagesByView[*types.MsgPrepare]{
					101: {
						1: {
							{
								View:      &types.View{Sequence: 101, Round: 1},
								From:      []byte("some validator"),
								BlockHash: []byte("block hash"),
							},
						},
					},
				},

				commit: messagesByView[*types.MsgCommit]{
					101: {
						1: {
							{
								View:       &types.View{Sequence: 101, Round: 1},
								From:       []byte("some validator"),
								BlockHash:  []byte("block hash"),
								CommitSeal: []byte("commit seal"),
							},
						},
					},
				},
			}),
		},

		{
			name: "no quorum prepare msgs in round 0 so new block is finalized in round 1",
			expectedFinalizedBlock: &types.FinalizedProposal{
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

			validator: MockValidator{
				IDFn:   MyValidator,
				SignFn: NilSignature,
			},

			verifier: MockVerifier{
				IsValidSignatureFn: TrueSignature,
				IsValidBlockFn:     TrueBlock,
				IsProposerFn: func(from []byte, _, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidatorFn: TrueValidator,
			},

			keccakFn:  BlockHashKeccak,
			transport: DummyTransport,
			quorumFn:  QuorumFn(func(msgs []ibft.Message) bool { return len(msgs) == 2 }),
			msgFeed: singleRoundFeed(MockFeed{
				proposal: messagesByView[*types.MsgProposal]{
					101: {
						0: {
							{
								View:      &types.View{Sequence: 101, Round: 0},
								From:      []byte("proposer"),
								BlockHash: []byte("block hash"),
								ProposedBlock: &types.ProposedBlock{
									Block: []byte("round 0 block"),
									Round: 0,
								},
							},
						},

						1: {
							{
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
						},
					},
				},

				prepare: messagesByView[*types.MsgPrepare]{
					101: {
						0: {
							{
								View:      &types.View{Sequence: 101, Round: 0},
								From:      []byte("validator"),
								BlockHash: []byte("block hash"),
							},
						},

						1: {
							{
								View:      &types.View{Sequence: 101, Round: 1},
								From:      []byte("validator"),
								BlockHash: []byte("block hash"),
							},
							{
								View:      &types.View{Sequence: 101, Round: 1},
								From:      []byte("other validator"),
								BlockHash: []byte("block hash"),
							},
						},
					},
				},

				commit: messagesByView[*types.MsgCommit]{
					101: {
						1: {
							{
								View:       &types.View{Sequence: 101, Round: 1},
								From:       []byte("validator"),
								CommitSeal: []byte("commit seal"),
								BlockHash:  []byte("block hash"),
							},
							{
								View:       &types.View{Sequence: 101, Round: 1},
								From:       []byte("other validator"),
								CommitSeal: []byte("other commit seal"),
								BlockHash:  []byte("block hash"),
							},
						},
					},
				},
			}),
		},

		{
			name: "no commit msgs in round 0 so new block is finalized in round 1",
			expectedFinalizedBlock: &types.FinalizedProposal{
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

			validator: MockValidator{
				IDFn:   MyValidator,
				SignFn: NilSignature,
			},

			verifier: MockVerifier{
				IsValidSignatureFn: TrueSignature,
				IsValidBlockFn:     TrueBlock,
				IsProposerFn: func(from []byte, _, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidatorFn: TrueValidator,
			},

			keccakFn:  BlockHashKeccak,
			transport: DummyTransport,
			quorumFn:  QuorumFn(func(msgs []ibft.Message) bool { return len(msgs) == 2 }),

			msgFeed: singleRoundFeed(MockFeed{
				proposal: messagesByView[*types.MsgProposal]{
					101: {
						0: {
							{
								View:      &types.View{Sequence: 101, Round: 0},
								From:      []byte("proposer"),
								BlockHash: []byte("block hash"),
								ProposedBlock: &types.ProposedBlock{
									Block: []byte("round 0 block"),
									Round: 0,
								},
							},
						},

						1: {
							{
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
						},
					},
				},

				prepare: messagesByView[*types.MsgPrepare]{
					101: {
						0: {
							{
								View:      &types.View{Sequence: 101, Round: 0},
								From:      []byte("validator"),
								BlockHash: []byte("block hash"),
							},
							{
								View:      &types.View{Sequence: 101, Round: 0},
								From:      []byte("validator"),
								BlockHash: []byte("block hash"),
							},
						},

						1: {
							{
								View:      &types.View{Sequence: 101, Round: 1},
								From:      []byte("validator"),
								BlockHash: []byte("block hash"),
							},
							{
								View:      &types.View{Sequence: 101, Round: 1},
								From:      []byte("other validator"),
								BlockHash: []byte("block hash"),
							},
						},
					},
				},

				commit: messagesByView[*types.MsgCommit]{
					101: {
						1: {
							{
								View:       &types.View{Sequence: 101, Round: 1},
								From:       []byte("validator"),
								CommitSeal: []byte("commit seal"),
								BlockHash:  []byte("block hash"),
							},

							{
								View:       &types.View{Sequence: 101, Round: 1},
								From:       []byte("other validator"),
								CommitSeal: []byte("other commit seal"),
								BlockHash:  []byte("block hash"),
							},
						},
					},
				},
			}),
		},

		{
			name: "no quorum commit msgs in round 0 so new block is finalized in round 1",
			expectedFinalizedBlock: &types.FinalizedProposal{
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

			validator: MockValidator{
				IDFn:   MyValidator,
				SignFn: NilSignature,
			},

			verifier: MockVerifier{
				IsValidSignatureFn: TrueSignature,
				IsValidBlockFn:     TrueBlock,
				IsProposerFn: func(from []byte, _, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidatorFn: TrueValidator,
			},

			keccakFn:  BlockHashKeccak,
			transport: DummyTransport,
			quorumFn:  QuorumFn(func(msgs []ibft.Message) bool { return len(msgs) == 2 }),

			msgFeed: singleRoundFeed(MockFeed{
				proposal: messagesByView[*types.MsgProposal]{
					101: {
						0: {
							{
								View:      &types.View{Sequence: 101, Round: 0},
								From:      []byte("proposer"),
								BlockHash: []byte("block hash"),
								ProposedBlock: &types.ProposedBlock{
									Block: []byte("round 0 block"),
									Round: 0,
								},
							},
						},

						1: {
							{
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
						},
					},
				},

				prepare: messagesByView[*types.MsgPrepare]{
					101: {
						0: {
							{
								View:      &types.View{Sequence: 101, Round: 0},
								From:      []byte("validator"),
								BlockHash: []byte("block hash"),
							},
							{
								View:      &types.View{Sequence: 101, Round: 0},
								From:      []byte("other validator"),
								BlockHash: []byte("block hash"),
							},
						},

						1: {
							{
								View:      &types.View{Sequence: 101, Round: 1},
								From:      []byte("validator"),
								BlockHash: []byte("block hash"),
							},
							{
								View:      &types.View{Sequence: 101, Round: 1},
								From:      []byte("other validator"),
								BlockHash: []byte("block hash"),
							},
						},
					},
				},

				commit: messagesByView[*types.MsgCommit]{
					101: {
						0: {
							{
								View:       &types.View{Sequence: 101, Round: 0},
								From:       []byte("validator"),
								CommitSeal: []byte("commit seal"),
								BlockHash:  []byte("block hash"),
							},
						},

						1: {
							{
								View:       &types.View{Sequence: 101, Round: 1},
								From:       []byte("validator"),
								CommitSeal: []byte("commit seal"),
								BlockHash:  []byte("block hash"),
							},

							{
								View:       &types.View{Sequence: 101, Round: 1},
								From:       []byte("other validator"),
								CommitSeal: []byte("other commit seal"),
								BlockHash:  []byte("block hash"),
							},
						},
					},
				},
			}),
		},

		{
			name: "round 1 proposer fails to build block so new block is finalized in round 2",
			expectedFinalizedBlock: &types.FinalizedProposal{
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

			validator: MockValidator{
				IDFn:   MyValidator,
				SignFn: NilSignature,
			},
			verifier: MockVerifier{
				IsValidSignatureFn: TrueSignature,
				IsValidBlockFn:     TrueBlock,
				IsValidatorFn:      TrueValidator,
				IsProposerFn: func(from []byte, _, round uint64) bool {
					return bytes.Equal(from, []byte("proposer")) ||
						bytes.Equal(from, []byte("my validator")) && round == 1
				},
			},

			keccakFn:  BlockHashKeccak,
			transport: DummyTransport,
			quorumFn:  QuorumFn(func(msgs []ibft.Message) bool { return len(msgs) == 2 }),

			msgFeed: singleRoundFeed(MockFeed{
				proposal: messagesByView[*types.MsgProposal]{
					101: {
						2: {
							{
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
						},
					},
				},

				prepare: messagesByView[*types.MsgPrepare]{
					101: {
						2: {
							{
								View:      &types.View{Sequence: 101, Round: 2},
								From:      []byte("validator"),
								BlockHash: []byte("block hash"),
							},
							{
								View:      &types.View{Sequence: 101, Round: 2},
								From:      []byte("other validator"),
								BlockHash: []byte("block hash"),
							},
						},
					},
				},

				commit: messagesByView[*types.MsgCommit]{
					101: {
						2: {
							{
								View:       &types.View{Sequence: 101, Round: 2},
								From:       []byte("validator"),
								BlockHash:  []byte("block hash"),
								CommitSeal: []byte("commit seal"),
							},
							{
								View:       &types.View{Sequence: 101, Round: 2},
								From:       []byte("other validator"),
								BlockHash:  []byte("block hash"),
								CommitSeal: []byte("other commit seal"),
							},
						},
					},
				},

				roundChange: messagesByView[*types.MsgRoundChange]{
					101: {
						1: {
							{
								View: &types.View{Sequence: 101, Round: 1},
								From: []byte("validator"),
							},
						},
					},
				},
			}),
		},
	}

	for _, tt := range table {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := ibft.NewIBFTContext(context.Background())
			ctx = ctx.WithTransport(tt.transport)
			ctx = ctx.WithFeed(tt.msgFeed)
			ctx = ctx.WithQuorum(tt.quorumFn)
			ctx = ctx.WithKeccak(tt.keccakFn)

			s := New(tt.validator, tt.verifier, time.Millisecond*100)
			assert.True(t, reflect.DeepEqual(tt.expectedFinalizedBlock, s.FinalizeSequence(ctx, 101)))
		})
	}
}
