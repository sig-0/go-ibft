package sequencer

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/madz-lab/go-ibft/message/types"
)

func TestHappyFlow(t *testing.T) {
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

	t.Run("happy flow (round 0)", func(t *testing.T) {
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
				},

				feed: mockMessageeFeed{
					subProposalFn: func() []*types.MsgProposal {
						return []*types.MsgProposal{
							{
								View:         &types.View{Sequence: 101, Round: 0},
								From:         []byte("proposer"),
								ProposalHash: []byte("proposal hash"),
								ProposedBlock: &types.ProposedBlock{
									Data:  []byte("block"),
									Round: 0,
								},
							},
						}
					},
					subPrepareFn: func() []*types.MsgPrepare {
						return []*types.MsgPrepare{
							{
								View:         &types.View{Sequence: 101, Round: 0},
								From:         []byte("some validator"),
								ProposalHash: []byte("proposal hash"),
							},
						}
					},
					subCommitFn: func() []*types.MsgCommit {
						return []*types.MsgCommit{
							{
								View:         &types.View{Sequence: 101, Round: 0},
								From:         []byte("some validator"),
								ProposalHash: []byte("proposal hash"),
								CommitSeal:   []byte("commit seal"),
							},
						}
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
				},

				feed: mockMessageeFeed{
					subPrepareFn: func() []*types.MsgPrepare {
						return []*types.MsgPrepare{
							{
								View:         &types.View{Sequence: 101, Round: 0},
								From:         []byte("some validator"),
								ProposalHash: []byte("proposal hash"),
							},
						}
					},
					subCommitFn: func() []*types.MsgCommit {
						return []*types.MsgCommit{
							{
								View:         &types.View{Sequence: 101, Round: 0},
								From:         []byte("some validator"),
								ProposalHash: []byte("proposal hash"),
								CommitSeal:   []byte("commit seal"),
							},
						}
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
}

func TestIsValidProposal(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name     string
		id       []byte
		view     *types.View
		verifier Verifier
		proposal *types.MsgProposal
		isValid  bool
	}{
		{
			name: "view sequence mismatch",
			view: &types.View{Sequence: 101, Round: 0},
			proposal: &types.MsgProposal{
				View: &types.View{Sequence: 102, Round: 0},
			},
		},

		{
			name: "view round mismatch",
			view: &types.View{Sequence: 101, Round: 0},
			proposal: &types.MsgProposal{
				View: &types.View{Sequence: 101, Round: 5},
			},
		},

		{
			name: "proposed block round mismatch",
			view: &types.View{Sequence: 101, Round: 0},
			proposal: &types.MsgProposal{
				View:          &types.View{Sequence: 101, Round: 0},
				ProposedBlock: &types.ProposedBlock{Round: 5},
			},
		},

		{
			name: "validator is proposer",
			id:   []byte("proposer"),
			view: &types.View{Sequence: 101, Round: 0},
			proposal: &types.MsgProposal{
				View:          &types.View{Sequence: 101, Round: 0},
				ProposedBlock: &types.ProposedBlock{Round: 0},
			},
			verifier: mockVerifier{isProposerFn: func(_ *types.View, from []byte) bool {
				return bytes.Equal([]byte("proposer"), from)
			}},
		},

		{
			name: "invalid proposer in proposal",
			id:   []byte("validator"),
			view: &types.View{Sequence: 101, Round: 0},
			proposal: &types.MsgProposal{
				From:          []byte("invalid proposer"),
				View:          &types.View{Sequence: 101, Round: 0},
				ProposedBlock: &types.ProposedBlock{Round: 0},
			},
			verifier: mockVerifier{isProposerFn: func(_ *types.View, from []byte) bool {
				return bytes.Equal([]byte("proposer"), from)
			}},
		},

		{
			name: "invalid block",
			id:   []byte("validator"),
			view: &types.View{Sequence: 101, Round: 0},
			proposal: &types.MsgProposal{
				From:          []byte("proposer"),
				View:          &types.View{Sequence: 101, Round: 0},
				ProposedBlock: &types.ProposedBlock{Data: []byte("invalid block"), Round: 0},
			},
			verifier: mockVerifier{
				isProposerFn: func(_ *types.View, from []byte) bool {
					return bytes.Equal([]byte("proposer"), from)
				},
				isValidBlockFn: func(b []byte) bool {
					return bytes.Equal([]byte("block"), b)
				},
			},
		},

		{
			name: "invalid proposal hash",
			id:   []byte("validator"),
			view: &types.View{Sequence: 101, Round: 0},
			proposal: &types.MsgProposal{
				From:          []byte("proposer"),
				View:          &types.View{Sequence: 101, Round: 0},
				ProposedBlock: &types.ProposedBlock{Data: []byte("block"), Round: 0},
				ProposalHash:  []byte("invalid proposal hash"),
			},
			verifier: mockVerifier{
				isProposerFn: func(_ *types.View, from []byte) bool {
					return bytes.Equal([]byte("proposer"), from)
				},
				isValidBlockFn: func(b []byte) bool {
					return bytes.Equal([]byte("block"), b)
				},
				keccakFn: func(_ []byte) []byte {
					return []byte("proposal hash")
				},
			},
		},

		{
			name: "valid round 0 proposal",
			id:   []byte("validator"),
			view: &types.View{Sequence: 101, Round: 0},
			proposal: &types.MsgProposal{
				From:          []byte("proposer"),
				View:          &types.View{Sequence: 101, Round: 0},
				ProposedBlock: &types.ProposedBlock{Data: []byte("block"), Round: 0},
				ProposalHash:  []byte("proposal hash"),
			},
			verifier: mockVerifier{
				isProposerFn: func(_ *types.View, from []byte) bool {
					return bytes.Equal([]byte("proposer"), from)
				},
				isValidBlockFn: func(b []byte) bool {
					return bytes.Equal([]byte("block"), b)
				},
				keccakFn: func(_ []byte) []byte {
					return []byte("proposal hash")
				},
			},
			isValid: true,
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seq := Sequencer{
				id:       tt.id,
				verifier: tt.verifier,
			}

			assert.Equal(t, tt.isValid, seq.isValidProposal(tt.view, tt.proposal))
		})
	}
}

func TestIsValidPrepare(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name         string
		view         *types.View
		prepare      *types.MsgPrepare
		proposalHash []byte
		isValid      bool
	}{
		{
			name: "view sequence mismatch",
			view: &types.View{Sequence: 101, Round: 0},
			prepare: &types.MsgPrepare{
				View: &types.View{Sequence: 102, Round: 0},
			},
		},

		{
			name: "view round mismatch",
			view: &types.View{Sequence: 101, Round: 0},
			prepare: &types.MsgPrepare{
				View: &types.View{Sequence: 101, Round: 5},
			},
		},

		{
			name: "invalid proposal hash",
			view: &types.View{Sequence: 101, Round: 0},
			prepare: &types.MsgPrepare{
				View:         &types.View{Sequence: 101, Round: 5},
				ProposalHash: []byte("invalid proposal hash"),
			},
			proposalHash: []byte("proposal hash"),
		},

		{
			name: "valid prepare",
			view: &types.View{Sequence: 101, Round: 0},
			prepare: &types.MsgPrepare{
				View:         &types.View{Sequence: 101, Round: 0},
				ProposalHash: []byte("proposal hash"),
			},
			proposalHash: []byte("proposal hash"),
			isValid:      true,
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			acceptedProposal := &types.MsgProposal{ProposalHash: tt.proposalHash}
			seq := Sequencer{state: state{acceptedProposal: acceptedProposal}}

			assert.Equal(t, tt.isValid, seq.isValidPrepare(tt.view, tt.prepare))
		})
	}
}
