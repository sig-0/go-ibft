package sequencer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/madz-lab/go-ibft/message/types"
)

func TestIsValidProposal(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name     string
		id       []byte
		view     *types.View
		verifier Verifier
		quorum   Quorum
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
			verifier: mockVerifier{
				isProposerFn: func(_ *types.View, from []byte) bool {
					return bytes.Equal([]byte("proposer"), from)
				},
				isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
				},
			},
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
			verifier: mockVerifier{
				isProposerFn: func(_ *types.View, from []byte) bool {
					return bytes.Equal([]byte("proposer"), from)
				},
				isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
				},
			},
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
				isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
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
				isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
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
				isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
				},
			},
			isValid: true,
		},

		{
			name: "no rcc in round 1 proposal",
			id:   []byte("validator"),
			view: &types.View{Sequence: 101, Round: 1},
			proposal: &types.MsgProposal{
				From:                   []byte("proposer"),
				View:                   &types.View{Sequence: 101, Round: 1},
				ProposedBlock:          &types.ProposedBlock{Data: []byte("block"), Round: 1},
				ProposalHash:           []byte("proposal hash"),
				RoundChangeCertificate: nil,
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
				isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
				},
			},
		},

		{
			name: "no messages in rcc of round 1 proposal",
			id:   []byte("validator"),
			view: &types.View{Sequence: 101, Round: 1},
			proposal: &types.MsgProposal{
				From:          []byte("proposer"),
				View:          &types.View{Sequence: 101, Round: 1},
				ProposedBlock: &types.ProposedBlock{Data: []byte("block"), Round: 1},
				ProposalHash:  []byte("proposal hash"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{},
				},
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
				isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
				},
			},
		},

		{
			name: "no quorum in rcc of round 1 proposal",
			id:   []byte("validator"),
			view: &types.View{Sequence: 101, Round: 1},
			proposal: &types.MsgProposal{
				From:          []byte("proposer"),
				View:          &types.View{Sequence: 101, Round: 1},
				ProposedBlock: &types.ProposedBlock{Data: []byte("block"), Round: 1},
				ProposalHash:  []byte("proposal hash"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{},
				},
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
				isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
				},
			},
			quorum: mockQuorum{
				quorumRoundChange: func(_ ...*types.MsgRoundChange) bool {
					return false
				},
			},
		},

		{
			name: "invalid sender in rcc of round 1 proposal",
			id:   []byte("validator"),
			view: &types.View{Sequence: 101, Round: 1},
			proposal: &types.MsgProposal{
				From:          []byte("proposer"),
				View:          &types.View{Sequence: 101, Round: 1},
				ProposedBlock: &types.ProposedBlock{Data: []byte("block"), Round: 1},
				ProposalHash:  []byte("proposal hash"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{
							View: &types.View{
								Sequence: 101,
								Round:    1,
							},
							From: []byte("invalid sender"),
						},
					},
				},
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
				isValidatorFn: func(from []byte, _ uint64) bool {
					if bytes.Equal(from, []byte("invalid sender")) {
						return false
					}

					return true
				},
			},
			quorum: mockQuorum{
				quorumRoundChange: func(_ ...*types.MsgRoundChange) bool {
					return true
				},
			},
		},

		{
			name: "valid round 1 proposal",
			id:   []byte("validator"),
			view: &types.View{Sequence: 101, Round: 1},
			proposal: &types.MsgProposal{
				From:          []byte("proposer"),
				View:          &types.View{Sequence: 101, Round: 1},
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
				isValidatorFn: func(from []byte, _ uint64) bool {
					if bytes.Equal(from, []byte("invalid sender")) {
						return false
					}

					return true
				},
			},
			quorum: mockQuorum{
				quorumRoundChange: func(_ ...*types.MsgRoundChange) bool {
					return true
				},
			},
			isValid: true,
		},

		{
			name: "valid round 1 proposal with block from round 0",
			id:   []byte("validator"),
			view: &types.View{Sequence: 101, Round: 1},
			proposal: &types.MsgProposal{
				From:          []byte("proposer"),
				View:          &types.View{Sequence: 101, Round: 1},
				ProposedBlock: &types.ProposedBlock{Data: []byte("block"), Round: 1},
				ProposalHash:  []byte("proposal hash"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{
							View: &types.View{Sequence: 101, Round: 1},
							From: []byte("some validator"),
							LatestPreparedCertificate: &types.PreparedCertificate{
								ProposalMessage: &types.MsgProposal{
									View: &types.View{
										Sequence: 101,
										Round:    0,
									},
									From:         []byte("proposer"),
									ProposalHash: []byte("proposal hash"),
								},
								PrepareMessages: []*types.MsgPrepare{
									{
										View:         &types.View{Sequence: 101, Round: 0},
										From:         []byte("some validator"),
										ProposalHash: []byte("proposal hash"),
									},
								},
							},
						},
					},
				},
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
				isValidatorFn: func(from []byte, _ uint64) bool {
					if bytes.Equal(from, []byte("invalid sender")) {
						return false
					}

					return true
				},
			},
			quorum: mockQuorum{
				quorumPrepare: func(_ ...*types.MsgPrepare) bool {
					return true
				},
				quorumRoundChange: func(_ ...*types.MsgRoundChange) bool {
					return true
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
				quorum:   tt.quorum,
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
		verifier     Verifier
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
			name: "invalid sender",
			view: &types.View{Sequence: 101, Round: 0},
			prepare: &types.MsgPrepare{
				View:         &types.View{Sequence: 101, Round: 0},
				ProposalHash: []byte("invalid proposal hash"),
			},
			verifier: mockVerifier{
				isValidatorFn: func(_ []byte, _ uint64) bool {
					return false
				},
			},
			proposalHash: []byte("proposal hash"),
		},

		{
			name: "invalid proposal hash",
			view: &types.View{Sequence: 101, Round: 0},
			prepare: &types.MsgPrepare{
				View:         &types.View{Sequence: 101, Round: 5},
				ProposalHash: []byte("invalid proposal hash"),
			},
			proposalHash: []byte("proposal hash"),
			verifier: mockVerifier{
				isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
				},
			},
		},

		{
			name: "valid prepare",
			view: &types.View{Sequence: 101, Round: 0},
			prepare: &types.MsgPrepare{
				View:         &types.View{Sequence: 101, Round: 0},
				ProposalHash: []byte("proposal hash"),
			},
			proposalHash: []byte("proposal hash"),
			verifier: mockVerifier{
				isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
				},
			},
			isValid: true,
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			previouslyAcceptedProposal := &types.MsgProposal{ProposalHash: tt.proposalHash}
			seq := Sequencer{
				verifier: tt.verifier,
				state:    state{acceptedProposal: previouslyAcceptedProposal},
			}

			assert.Equal(t, tt.isValid, seq.isValidPrepare(tt.view, tt.prepare))
		})
	}
}

func TestIsValidCommit(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name         string
		view         *types.View
		proposalHash []byte
		commit       *types.MsgCommit
		verifier     Verifier
		isValid      bool
	}{
		{
			name: "view sequence mismatch",
			view: &types.View{Sequence: 101, Round: 0},
			commit: &types.MsgCommit{
				View: &types.View{Sequence: 102, Round: 0},
			},
		},

		{
			name: "view round mismatch",
			view: &types.View{Sequence: 101, Round: 0},
			commit: &types.MsgCommit{
				View: &types.View{Sequence: 101, Round: 5},
			},
		},

		{
			name: "invalid sender",
			view: &types.View{Sequence: 101, Round: 0},
			commit: &types.MsgCommit{
				View: &types.View{Sequence: 101, Round: 5},
			},
			verifier: mockVerifier{isValidatorFn: func(_ []byte, _ uint64) bool {
				return false
			}},
		},

		{
			name:         "proposal hash mismatch",
			proposalHash: []byte("proposal hash"),
			view:         &types.View{Sequence: 101, Round: 0},
			commit: &types.MsgCommit{
				View:         &types.View{Sequence: 101, Round: 0},
				ProposalHash: []byte("invalid proposal hash"),
			},
			verifier: mockVerifier{
				isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
				},
			},
		},

		{
			name:         "invalid commit seal",
			proposalHash: []byte("proposal hash"),
			view:         &types.View{Sequence: 101, Round: 0},
			commit: &types.MsgCommit{
				View:         &types.View{Sequence: 101, Round: 0},
				From:         []byte("some validator"),
				ProposalHash: []byte("proposal hash"),
			},
			verifier: mockVerifier{
				isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
				},
				recoverFromFn: func(_ []byte, _ []byte) []byte {
					return []byte("some other validator")
				},
			},
		},

		{
			name:         "valid commit",
			proposalHash: []byte("proposal hash"),
			view:         &types.View{Sequence: 101, Round: 0},
			commit: &types.MsgCommit{
				View:         &types.View{Sequence: 101, Round: 0},
				From:         []byte("some validator"),
				ProposalHash: []byte("proposal hash"),
			},
			verifier: mockVerifier{
				isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
				},
				recoverFromFn: func(_ []byte, _ []byte) []byte {
					return []byte("some validator")
				},
			},
			isValid: true,
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			previouslyAcceptedProposal := &types.MsgProposal{ProposalHash: tt.proposalHash}

			seq := Sequencer{
				verifier: tt.verifier,
				state:    state{acceptedProposal: previouslyAcceptedProposal},
			}

			assert.Equal(t, tt.isValid, seq.isValidCommit(tt.view, tt.commit))
		})
	}
}
