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

			previouslyAcceptedProposal := &types.MsgProposal{ProposalHash: tt.proposalHash}
			seq := Sequencer{state: state{acceptedProposal: previouslyAcceptedProposal}}

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
			name:         "proposal hash mismatch",
			proposalHash: []byte("proposal hash"),
			view:         &types.View{Sequence: 101, Round: 0},
			commit: &types.MsgCommit{
				View:         &types.View{Sequence: 101, Round: 0},
				ProposalHash: []byte("invalid proposal hash"),
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
			verifier: mockVerifier{recoverFromFn: func(_ []byte, _ []byte) []byte {
				return []byte("some other validator")
			}},
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
			verifier: mockVerifier{recoverFromFn: func(_ []byte, _ []byte) []byte {
				return []byte("some validator")
			}},
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
