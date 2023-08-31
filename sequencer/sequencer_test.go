package sequencer

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/madz-lab/go-ibft/message/types"
)

func TestFinalizeSequence(t *testing.T) {
	t.Parallel()

	type testTable struct {
		name                   string
		expectedFinalizedBlock []byte
		expectedFinalizedRound uint64
		expectedCommitSeals    [][]byte

		validator Validator
		verifier  Verifier
		transport Transport
		quorum    Quorum
		feed      MessageFeed
	}

	t.Run("happy flows (round 0)", func(t *testing.T) {
		t.Parallel()

		testTable := []testTable{
			{
				name:                   "validator is not the proposer",
				expectedFinalizedBlock: []byte("block"),
				expectedCommitSeals:    [][]byte{[]byte("commit seal")},
				expectedFinalizedRound: 0,

				validator: mockValidator{
					idFn:   func() []byte { return []byte("validator id") },
					signFn: func(_ []byte) []byte { return nil },
				},

				verifier: mockVerifier{
					keccakFn:       func(_ []byte) []byte { return []byte("proposal hash") },
					isValidBlockFn: func(_ []byte) bool { return true },
					isProposerFn: func(_ *types.View, from []byte) bool {
						return bytes.Equal(from, []byte("proposer"))
					},
					recoverFromFn: func(_ []byte, _ []byte) []byte { return []byte("some validator") },
					isValidatorFn: func(_ []byte, _ uint64) bool { return true },
				},

				feed: mockMessageeFeed{
					proposalsByView: map[uint64]map[uint64][]*types.MsgProposal{
						101: {
							0: []*types.MsgProposal{
								{
									View:         &types.View{Sequence: 101, Round: 0},
									From:         []byte("proposer"),
									ProposalHash: []byte("proposal hash"),
									ProposedBlock: &types.ProposedBlock{
										Data:  []byte("block"),
										Round: 0,
									},
								},
							},
						},
					},
					preparesByView: map[uint64]map[uint64][]*types.MsgPrepare{
						101: {
							0: []*types.MsgPrepare{
								{
									View:         &types.View{Sequence: 101, Round: 0},
									From:         []byte("some validator"),
									ProposalHash: []byte("proposal hash"),
								},
							},
						},
					},
					commitsByView: map[uint64]map[uint64][]*types.MsgCommit{
						101: {
							0: []*types.MsgCommit{
								{
									View:         &types.View{Sequence: 101, Round: 0},
									From:         []byte("some validator"),
									ProposalHash: []byte("proposal hash"),
									CommitSeal:   []byte("commit seal"),
								},
							},
						},
					},
				},

				quorum: mockQuorum{
					quorumPrepare: func(_ ...*types.MsgPrepare) bool {
						return true
					},
					quorumCommit: func(_ ...*types.MsgCommit) bool {
						return true
					},
				},

				transport: DummyTransport,
			},

			{
				name:                   "validator is the proposer",
				expectedFinalizedBlock: []byte("block"),
				expectedCommitSeals:    [][]byte{[]byte("commit seal")},
				expectedFinalizedRound: 0,

				validator: mockValidator{
					idFn:         func() []byte { return []byte("validator id") },
					signFn:       func(_ []byte) []byte { return nil },
					buildBlockFn: func() []byte { return []byte("block") },
				},

				verifier: mockVerifier{
					keccakFn:      func(_ []byte) []byte { return []byte("proposal hash") },
					isProposerFn:  func(_ *types.View, _ []byte) bool { return true },
					recoverFromFn: func(_ []byte, _ []byte) []byte { return []byte("some validator") },
					isValidatorFn: func(_ []byte, _ uint64) bool { return true },
				},

				feed: mockMessageeFeed{
					preparesByView: map[uint64]map[uint64][]*types.MsgPrepare{
						101: {
							0: []*types.MsgPrepare{
								{
									View:         &types.View{Sequence: 101, Round: 0},
									From:         []byte("some validator"),
									ProposalHash: []byte("proposal hash"),
								},
							},
						},
					},
					commitsByView: map[uint64]map[uint64][]*types.MsgCommit{
						101: {
							0: []*types.MsgCommit{
								{
									View:         &types.View{Sequence: 101, Round: 0},
									From:         []byte("some validator"),
									ProposalHash: []byte("proposal hash"),
									CommitSeal:   []byte("commit seal"),
								},
							},
						},
					},
				},

				quorum: mockQuorum{
					quorumPrepare: func(_ ...*types.MsgPrepare) bool {
						return true
					},
					quorumCommit: func(_ ...*types.MsgCommit) bool {
						return true
					},
				},

				transport: DummyTransport,
			},
		}

		for _, tt := range testTable {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				s := New(tt.validator, tt.verifier, 1*time.Second)
				s.WithTransport(tt.transport)
				s.WithQuorum(tt.quorum)

				fb := s.FinalizeSequence(context.Background(), 101, tt.feed)
				assert.NotNil(t, fb)
				assert.Equal(t, tt.expectedFinalizedRound, fb.Round)
				assert.Equal(t, tt.expectedFinalizedBlock, fb.Block)
				assert.Equal(t, tt.expectedCommitSeals, fb.CommitSeals)
			})
		}
	})

	t.Run("unhappy flow: round 0 expires and block is finalized in round 1", func(t *testing.T) {
		t.Parallel()

		testTable := []testTable{
			{
				name:                   "round 1 proposal is valid with empty PB and PC",
				expectedFinalizedBlock: []byte("block"),
				expectedCommitSeals:    [][]byte{[]byte("commit seal")},
				expectedFinalizedRound: 1,

				validator: mockValidator{
					idFn:   func() []byte { return []byte("validator id") },
					signFn: func(_ []byte) []byte { return nil },
				},

				verifier: mockVerifier{
					keccakFn:       func(_ []byte) []byte { return []byte("proposal hash") },
					isValidBlockFn: func(_ []byte) bool { return true },
					isProposerFn: func(_ *types.View, from []byte) bool {
						return bytes.Equal(from, []byte("proposer"))
					},
					recoverFromFn: func(_ []byte, _ []byte) []byte { return []byte("some validator") },
					isValidatorFn: func(_ []byte, _ uint64) bool { return true },
				},

				feed: mockMessageeFeed{
					proposalsByView: map[uint64]map[uint64][]*types.MsgProposal{
						101: {
							0: nil,
							1: []*types.MsgProposal{
								{
									View:          &types.View{Sequence: 101, Round: 1},
									From:          []byte("proposer"),
									ProposedBlock: &types.ProposedBlock{Data: []byte("block"), Round: 1},
									ProposalHash:  []byte("proposal hash"),
									RoundChangeCertificate: &types.RoundChangeCertificate{
										Messages: []*types.MsgRoundChange{
											{
												View: &types.View{Sequence: 101, Round: 1},
												From: []byte("some validator"),
											},
										},
									},
								},
							},
						},
					},

					preparesByView: map[uint64]map[uint64][]*types.MsgPrepare{
						101: {
							1: []*types.MsgPrepare{
								{
									View:         &types.View{Sequence: 101, Round: 1},
									From:         []byte("some validator"),
									ProposalHash: []byte("proposal hash"),
								},
							},
						},
					},

					commitsByView: map[uint64]map[uint64][]*types.MsgCommit{
						101: {
							1: []*types.MsgCommit{
								{
									View:         &types.View{Sequence: 101, Round: 1},
									From:         []byte("some validator"),
									ProposalHash: []byte("proposal hash"),
									CommitSeal:   []byte("commit seal"),
								},
							},
						},
					},
				},

				quorum: mockQuorum{
					quorumPrepare: func(_ ...*types.MsgPrepare) bool {
						return true
					},
					quorumCommit: func(_ ...*types.MsgCommit) bool {
						return true
					},
					quorumRoundChange: func(_ ...*types.MsgRoundChange) bool {
						return true
					},
				},

				transport: DummyTransport,
			},

			{
				name:                   "round 1 proposal is valid with non-nil PB and PC",
				expectedFinalizedBlock: []byte("round 0 block"),
				expectedCommitSeals:    [][]byte{[]byte("commit seal")},
				expectedFinalizedRound: 1,

				validator: mockValidator{
					idFn:   func() []byte { return []byte("validator id") },
					signFn: func(_ []byte) []byte { return nil },
				},

				verifier: mockVerifier{
					keccakFn:       func(_ []byte) []byte { return []byte("proposal hash") },
					isValidBlockFn: func(_ []byte) bool { return true },
					isProposerFn: func(_ *types.View, from []byte) bool {
						return bytes.Equal(from, []byte("proposer"))
					},
					recoverFromFn: func(_ []byte, _ []byte) []byte { return []byte("some validator") },
					isValidatorFn: func(_ []byte, _ uint64) bool { return true },
				},

				feed: mockMessageeFeed{
					proposalsByView: map[uint64]map[uint64][]*types.MsgProposal{
						101: {
							0: nil,
							1: []*types.MsgProposal{
								{
									View:          &types.View{Sequence: 101, Round: 1},
									From:          []byte("proposer"),
									ProposedBlock: &types.ProposedBlock{Data: []byte("round 0 block"), Round: 1},
									ProposalHash:  []byte("proposal hash"),
									RoundChangeCertificate: &types.RoundChangeCertificate{
										Messages: []*types.MsgRoundChange{
											{
												View: &types.View{Sequence: 101, Round: 1},
												From: []byte("some validator"),
												LatestPreparedCertificate: &types.PreparedCertificate{
													ProposalMessage: &types.MsgProposal{
														View:          &types.View{Sequence: 101, Round: 0},
														From:          []byte("proposer"),
														ProposedBlock: &types.ProposedBlock{Data: []byte("round 0 block"), Round: 0},
														ProposalHash:  []byte("proposal hash"),
													},
													PrepareMessages: []*types.MsgPrepare{
														{
															View: &types.View{
																Sequence: 101,
																Round:    0,
															},
															From:         []byte("some validator"),
															ProposalHash: []byte("proposal hash"),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},

					preparesByView: map[uint64]map[uint64][]*types.MsgPrepare{
						101: {
							1: []*types.MsgPrepare{
								{
									View:         &types.View{Sequence: 101, Round: 1},
									From:         []byte("some validator"),
									ProposalHash: []byte("proposal hash"),
								},
							},
						},
					},

					commitsByView: map[uint64]map[uint64][]*types.MsgCommit{
						101: {
							1: []*types.MsgCommit{
								{
									View:         &types.View{Sequence: 101, Round: 1},
									From:         []byte("some validator"),
									ProposalHash: []byte("proposal hash"),
									CommitSeal:   []byte("commit seal"),
								},
							},
						},
					},
				},

				quorum: mockQuorum{
					quorumPrepare: func(_ ...*types.MsgPrepare) bool {
						return true
					},
					quorumCommit: func(_ ...*types.MsgCommit) bool {
						return true
					},
					quorumRoundChange: func(_ ...*types.MsgRoundChange) bool {
						return true
					},
				},

				transport: DummyTransport,
			},

			{
				name:                   "propose new block in round 1",
				expectedFinalizedBlock: []byte("round 1 block"),
				expectedCommitSeals:    [][]byte{[]byte("commit seal")},
				expectedFinalizedRound: 1,

				validator: mockValidator{
					idFn:   func() []byte { return []byte("validator id") },
					signFn: func(_ []byte) []byte { return nil },
					buildBlockFn: func() []byte {
						return []byte("round 1 block")
					},
				},

				verifier: mockVerifier{
					keccakFn:       func(_ []byte) []byte { return []byte("proposal hash") },
					isValidBlockFn: func(_ []byte) bool { return true },
					isProposerFn: func(view *types.View, from []byte) bool {
						switch view.Round {
						case 0:
							return bytes.Equal(from, []byte("proposer"))
						case 1:
							return bytes.Equal(from, []byte("validator id"))
						default:
							return false
						}
					},
					recoverFromFn: func(_ []byte, _ []byte) []byte { return []byte("some validator") },
					isValidatorFn: func(_ []byte, _ uint64) bool { return true },
				},

				feed: mockMessageeFeed{
					proposalsByView: map[uint64]map[uint64][]*types.MsgProposal{
						101: {
							0: nil,
						},
					},

					roundChangesByView: map[uint64]map[uint64][]*types.MsgRoundChange{
						101: {
							1: []*types.MsgRoundChange{
								{
									View: &types.View{Sequence: 101, Round: 1},
									From: []byte("some validator"),
								},
							},
						},
					},

					preparesByView: map[uint64]map[uint64][]*types.MsgPrepare{
						101: {
							1: []*types.MsgPrepare{
								{
									View:         &types.View{Sequence: 101, Round: 1},
									From:         []byte("some validator"),
									ProposalHash: []byte("proposal hash"),
								},
							},
						},
					},

					commitsByView: map[uint64]map[uint64][]*types.MsgCommit{
						101: {
							1: []*types.MsgCommit{
								{
									View:         &types.View{Sequence: 101, Round: 1},
									From:         []byte("some validator"),
									ProposalHash: []byte("proposal hash"),
									CommitSeal:   []byte("commit seal"),
								},
							},
						},
					},
				},

				quorum: mockQuorum{
					quorumPrepare: func(_ ...*types.MsgPrepare) bool {
						return true
					},
					quorumCommit: func(_ ...*types.MsgCommit) bool {
						return true
					},
					quorumRoundChange: func(_ ...*types.MsgRoundChange) bool {
						return true
					},
				},

				transport: DummyTransport,
			},

			{
				name:                   "propose new block in round 1",
				expectedFinalizedBlock: []byte("round 0 block"),
				expectedCommitSeals:    [][]byte{[]byte("commit seal")},
				expectedFinalizedRound: 1,

				validator: mockValidator{
					idFn:   func() []byte { return []byte("validator id") },
					signFn: func(_ []byte) []byte { return nil },
				},

				verifier: mockVerifier{
					keccakFn:       func(_ []byte) []byte { return []byte("proposal hash") },
					isValidBlockFn: func(_ []byte) bool { return true },
					isProposerFn: func(view *types.View, from []byte) bool {
						switch view.Round {
						case 0:
							return bytes.Equal(from, []byte("proposer"))
						case 1:
							return bytes.Equal(from, []byte("validator id"))
						default:
							return false
						}
					},
					recoverFromFn: func(data []byte, _ []byte) []byte {
						if bytes.Equal(data, []byte("proposal hash")) {
							return []byte("some validator")
						}

						return []byte("proposer")
					},
					isValidatorFn: func(_ []byte, _ uint64) bool { return true },
				},

				feed: mockMessageeFeed{
					proposalsByView: map[uint64]map[uint64][]*types.MsgProposal{
						101: {
							0: nil,
						},
					},

					roundChangesByView: map[uint64]map[uint64][]*types.MsgRoundChange{
						101: {
							1: []*types.MsgRoundChange{
								{
									View: &types.View{Sequence: 101, Round: 1},
									From: []byte("some validator"),
									LatestPreparedProposedBlock: &types.ProposedBlock{
										Data:  []byte("round 0 block"),
										Round: 0,
									},
									LatestPreparedCertificate: &types.PreparedCertificate{
										ProposalMessage: &types.MsgProposal{
											View: &types.View{
												Sequence: 101,
												Round:    0,
											},
											From: []byte("proposer"),
											ProposedBlock: &types.ProposedBlock{
												Data:  []byte("round 0 block"),
												Round: 0,
											},
											ProposalHash: []byte("proposal hash"),
										},
										PrepareMessages: []*types.MsgPrepare{}, // empty for simplicity
									},
								},
							},
						},
					},

					preparesByView: map[uint64]map[uint64][]*types.MsgPrepare{
						101: {
							1: []*types.MsgPrepare{
								{
									View:         &types.View{Sequence: 101, Round: 1},
									From:         []byte("some validator"),
									ProposalHash: []byte("proposal hash"),
								},
							},
						},
					},

					commitsByView: map[uint64]map[uint64][]*types.MsgCommit{
						101: {
							1: []*types.MsgCommit{
								{
									View:         &types.View{Sequence: 101, Round: 1},
									From:         []byte("some validator"),
									ProposalHash: []byte("proposal hash"),
									CommitSeal:   []byte("commit seal"),
								},
							},
						},
					},
				},

				quorum: mockQuorum{
					quorumPrepare: func(_ ...*types.MsgPrepare) bool {
						return true
					},
					quorumCommit: func(_ ...*types.MsgCommit) bool {
						return true
					},
					quorumRoundChange: func(_ ...*types.MsgRoundChange) bool {
						return true
					},
				},

				transport: DummyTransport,
			},
		}

		for _, tt := range testTable {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				s := New(tt.validator, tt.verifier, time.Millisecond*100)
				s.WithTransport(tt.transport)
				s.WithQuorum(tt.quorum)

				fb := s.FinalizeSequence(context.Background(), 101, tt.feed)
				assert.NotNil(t, fb)
				assert.Equal(t, tt.expectedFinalizedRound, fb.Round)
				assert.Equal(t, tt.expectedFinalizedBlock, fb.Block)
				assert.Equal(t, tt.expectedCommitSeals, fb.CommitSeals)
			})
		}
	})
}
