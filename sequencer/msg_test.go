package sequencer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
	"github.com/madz-lab/go-ibft/test/mock"
)

func TestIsValidMsgProposal(t *testing.T) {
	t.Parallel()

	table := []struct {
		validator ibft.Validator
		quorum    ibft.Quorum
		keccak    ibft.Keccak
		msg       *types.MsgProposal
		name      string
		isValid   bool
	}{
		{
			name:      "proposed block round and view round do not match",
			validator: mock.Validator{IDFn: mock.NewValidatorID("alice").ID},

			msg: &types.MsgProposal{
				View:          &types.View{Round: 0},
				ProposedBlock: &types.ProposedBlock{Round: 5},
			},
		},

		{
			name:      "cannot verify own proposal",
			validator: mock.Validator{IDFn: mock.NewValidatorID("alice").ID},

			msg: &types.MsgProposal{
				View:          &types.View{Sequence: 101, Round: 0},
				ProposedBlock: &types.ProposedBlock{Round: 0},
				From:          mock.NewValidatorID("alice"),
			},
		},

		{
			name: "invalid proposer",
			validator: mock.Validator{
				IDFn: mock.NewValidatorID("alice").ID,
				Verifier: mock.Verifier{
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 0},
					),
				},
			},

			msg: &types.MsgProposal{
				View: &types.View{Sequence: 101, Round: 0},
				From: mock.NewValidatorID("definitely not bob"),
				ProposedBlock: &types.ProposedBlock{
					Block: []byte("block"),
					Round: 0,
				},
			},
		},

		{
			name:   "invalid block hash",
			keccak: mock.NewDummyKeccak("keccak"),
			validator: mock.Validator{
				IDFn: mock.NewValidatorID("alice").ID,
				Verifier: mock.Verifier{
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 0},
					),
				},
			},

			msg: &types.MsgProposal{
				View:          &types.View{Sequence: 101, Round: 0},
				From:          mock.NewValidatorID("bob"),
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 0},
				BlockHash:     []byte("definitely not keccak"),
			},
		},

		{
			name:   "invalid round 0 block",
			keccak: mock.NewDummyKeccak("keccak"),
			validator: mock.Validator{
				IDFn: mock.NewValidatorID("alice").ID,
				Verifier: mock.Verifier{
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 0},
					),
					IsValidProposalFn: func(block []byte, _ uint64) bool {
						return bytes.Equal(block, []byte("block"))
					},
				},
			},

			msg: &types.MsgProposal{
				View:          &types.View{Sequence: 101, Round: 0},
				From:          mock.NewValidatorID("bob"),
				ProposedBlock: &types.ProposedBlock{Block: []byte("invalid block"), Round: 0},
				BlockHash:     []byte("keccak"),
			},
		},

		{
			name:   "valid round 0 proposal",
			keccak: mock.NewDummyKeccak("keccak"),
			validator: mock.Validator{
				IDFn: mock.NewValidatorID("alice").ID,
				Verifier: mock.Verifier{
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 0},
					),
					IsValidProposalFn: func(block []byte, _ uint64) bool {
						return bytes.Equal(block, []byte("block"))
					},
				},
			},

			isValid: true,
			msg: &types.MsgProposal{
				From:          mock.NewValidatorID("bob"),
				View:          &types.View{Sequence: 101, Round: 0},
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 0},
				BlockHash:     []byte("keccak"),
			},
		},

		{
			name:   "(non zero round): empty rcc",
			keccak: mock.NewDummyKeccak("keccak"),
			validator: mock.Validator{
				IDFn: mock.NewValidatorID("alice").ID,
				Verifier: mock.Verifier{
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 1},
					),
					IsValidProposalFn: func(block []byte, _ uint64) bool {
						return bytes.Equal(block, []byte("block"))
					},
				},
			},

			msg: &types.MsgProposal{
				From:                   mock.NewValidatorID("bob"),
				View:                   &types.View{Sequence: 101, Round: 1},
				ProposedBlock:          &types.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:              []byte("keccak"),
				RoundChangeCertificate: nil,
			},
		},

		{
			name:   "(non zero round): invalid msg sequence in rcc",
			keccak: mock.NewDummyKeccak("keccak"),
			validator: mock.Validator{
				IDFn: mock.NewValidatorID("alice").ID,
				Verifier: mock.Verifier{
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 1},
					),
					IsValidProposalFn: func(block []byte, _ uint64) bool {
						return bytes.Equal(block, []byte("block"))
					},
				},
			},

			msg: &types.MsgProposal{
				From:          mock.NewValidatorID("bob"),
				View:          &types.View{Sequence: 101, Round: 1},
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{
							View: &types.View{Sequence: 99},
						},
					},
				},
			},
		},

		{
			name:   "(non zero round): invalid msg round in rcc",
			keccak: mock.NewDummyKeccak("keccak"),
			validator: mock.Validator{
				IDFn: mock.NewValidatorID("alice").ID,
				Verifier: mock.Verifier{
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 1},
					),
					IsValidProposalFn: func(block []byte, _ uint64) bool {
						return bytes.Equal(block, []byte("block"))
					},
				},
			},

			msg: &types.MsgProposal{
				From:          mock.NewValidatorID("bob"),
				View:          &types.View{Sequence: 101, Round: 1},
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{
							View: &types.View{Sequence: 101, Round: 0},
						},
					},
				},
			},
		},

		{
			name:   "(non zero round): invalid sender in rcc",
			keccak: mock.NewDummyKeccak("keccak"),
			validator: mock.Validator{
				IDFn: mock.NewValidatorID("alice").ID,
				Verifier: mock.Verifier{
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 1},
					),
					IsValidProposalFn: func(block []byte, _ uint64) bool {
						return bytes.Equal(block, []byte("block"))
					},
					IsValidatorFn: mock.NewValidatorSet("alice", "bob").IsValidator,
				},
			},

			msg: &types.MsgProposal{
				From:          mock.NewValidatorID("bob"),
				View:          &types.View{Sequence: 101, Round: 1},
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{
							From: mock.NewValidatorID("chris"),
							View: &types.View{Sequence: 101, Round: 1},
						},
					},
				},
			},
		},

		{
			name:   "(non zero round): duplicate sender in rcc",
			keccak: mock.NewDummyKeccak("keccak"),
			validator: mock.Validator{
				IDFn: mock.NewValidatorID("alice").ID,
				Verifier: mock.Verifier{
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 1},
					),
					IsValidProposalFn: func(block []byte, _ uint64) bool {
						return bytes.Equal(block, []byte("block"))
					},
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
				},
			},

			msg: &types.MsgProposal{
				From:          mock.NewValidatorID("bob"),
				View:          &types.View{Sequence: 101, Round: 1},
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{
							From: mock.NewValidatorID("chris"),
							View: &types.View{Sequence: 101, Round: 1},
						},

						{
							From: mock.NewValidatorID("chris"),
							View: &types.View{Sequence: 101, Round: 1},
						},
					},
				},
			},
		},

		{
			name:   "(non zero round): no quorum in rcc",
			keccak: mock.NewDummyKeccak("keccak"),
			quorum: mock.NoQuorum,
			validator: mock.Validator{
				IDFn: mock.NewValidatorID("alice").ID,
				Verifier: mock.Verifier{
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 1},
					),
					IsValidProposalFn: func(block []byte, _ uint64) bool {
						return bytes.Equal(block, []byte("block"))
					},
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
				},
			},

			msg: &types.MsgProposal{
				From:          mock.NewValidatorID("bob"),
				View:          &types.View{Sequence: 101, Round: 1},
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{
							From: mock.NewValidatorID("chris"),
							View: &types.View{Sequence: 101, Round: 1},
						},
					},
				},
			},
		},

		{
			name:   "(non zero round): invalid block in rcc",
			keccak: mock.NewDummyKeccak("keccak"),
			quorum: mock.NonZeroQuorum,
			validator: mock.Validator{
				IDFn: mock.NewValidatorID("alice").ID,
				Verifier: mock.Verifier{
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 1},
					),
					IsValidProposalFn: func(block []byte, _ uint64) bool {
						return bytes.Equal(block, []byte("block"))
					},
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
				},
			},

			msg: &types.MsgProposal{
				From:          mock.NewValidatorID("bob"),
				View:          &types.View{Sequence: 101, Round: 1},
				ProposedBlock: &types.ProposedBlock{Block: []byte("invalid block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{
							From: mock.NewValidatorID("chris"),
							View: &types.View{Sequence: 101, Round: 1},
						},
					},
				},
			},
		},

		{
			name:   "(non zero round): highest round block hash does not match derived hash",
			keccak: mock.NewDummyKeccak("keccak"),
			quorum: mock.NonZeroQuorum,
			validator: mock.Validator{
				IDFn: mock.NewValidatorID("alice").ID,
				Verifier: mock.Verifier{
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("alice"), Round: 0},
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 1},
					),
					IsValidProposalFn: func(block []byte, _ uint64) bool {
						return bytes.Equal(block, []byte("block"))
					},
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
				},
			},

			msg: &types.MsgProposal{
				From:          mock.NewValidatorID("bob"),
				View:          &types.View{Sequence: 101, Round: 1},
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{
							From: mock.NewValidatorID("chris"),
							View: &types.View{Sequence: 101, Round: 1},
							LatestPreparedCertificate: &types.PreparedCertificate{
								ProposalMessage: &types.MsgProposal{
									From:      mock.NewValidatorID("alice"),
									View:      &types.View{Sequence: 101, Round: 0},
									BlockHash: []byte("invalid keccak"),
								},
								PrepareMessages: []*types.MsgPrepare{
									{
										View:      &types.View{Sequence: 101, Round: 0},
										From:      mock.NewValidatorID("chris"),
										BlockHash: []byte("invalid keccak"),
									},
								},
							},
						},
					},
				},
			},
		},

		{
			name:   "(non zero round): valid proposal msg",
			keccak: mock.NewDummyKeccak("keccak"),
			quorum: mock.NonZeroQuorum,
			validator: mock.Validator{
				IDFn: mock.NewValidatorID("alice").ID,
				Verifier: mock.Verifier{
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("alice"), Round: 0},
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 1},
					),
					IsValidProposalFn: func(block []byte, _ uint64) bool {
						return bytes.Equal(block, []byte("block"))
					},
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
				},
			},

			isValid: true,
			msg: &types.MsgProposal{
				From:          mock.NewValidatorID("bob"),
				View:          &types.View{Sequence: 101, Round: 1},
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{
							From:                        mock.NewValidatorID("chris"),
							View:                        &types.View{Sequence: 101, Round: 1},
							LatestPreparedProposedBlock: &types.ProposedBlock{},
							LatestPreparedCertificate: &types.PreparedCertificate{
								ProposalMessage: &types.MsgProposal{
									From:      mock.NewValidatorID("alice"),
									View:      &types.View{Sequence: 101, Round: 0},
									BlockHash: []byte("keccak"),
								},
								PrepareMessages: []*types.MsgPrepare{
									{
										View:      &types.View{Sequence: 101, Round: 0},
										From:      mock.NewValidatorID("chris"),
										BlockHash: []byte("keccak"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range table {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seq := New(tt.validator, 0)
			assert.Equal(t, tt.isValid, seq.isValidMsgProposal(tt.msg, tt.quorum, tt.keccak))
		})
	}
}

func TestIsValidMsgPrepare(t *testing.T) {
	t.Parallel()

	table := []struct {
		validator        ibft.Validator
		msg              *types.MsgPrepare
		acceptedProposal *types.MsgProposal
		name             string
		isValid          bool
	}{
		{
			name: "invalid sender",
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob").IsValidator,
				},
			},

			msg: &types.MsgPrepare{
				View: &types.View{Sequence: 101},
				From: mock.NewValidatorID("chris"),
			},
		},

		{
			name: "invalid block hash",
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
				},
			},

			acceptedProposal: &types.MsgProposal{BlockHash: []byte("keccak")},
			msg: &types.MsgPrepare{
				View:      &types.View{Sequence: 101},
				From:      mock.NewValidatorID("chris"),
				BlockHash: []byte("definitely not keccak"),
			},
		},

		{
			name: "invalid block hash",
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
				},
			},

			isValid:          true,
			acceptedProposal: &types.MsgProposal{BlockHash: []byte("keccak")},
			msg: &types.MsgPrepare{
				View:      &types.View{Sequence: 101},
				From:      mock.NewValidatorID("chris"),
				BlockHash: []byte("keccak"),
			},
		},
	}

	for _, tt := range table {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seq := New(tt.validator, 0)
			seq.state.proposal = tt.acceptedProposal

			assert.Equal(t, tt.isValid, seq.isValidMsgPrepare(tt.msg))
		})
	}
}

func TestIsValidMsgCommit(t *testing.T) {
	t.Parallel()

	table := []struct {
		validator        ibft.Validator
		msg              *types.MsgCommit
		acceptedProposal *types.MsgProposal
		name             string
		isValid          bool
	}{

		{
			name: "invalid sender",
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob").IsValidator,
				},
			},

			msg: &types.MsgCommit{
				View: &types.View{Sequence: 101},
				From: mock.NewValidatorID("chris"),
			},
		},

		{
			name:             "invalid block hash",
			acceptedProposal: &types.MsgProposal{BlockHash: []byte("keccak")},
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
				},
			},

			msg: &types.MsgCommit{
				View:      &types.View{Sequence: 101},
				From:      mock.NewValidatorID("chris"),
				BlockHash: []byte("definitely not keccak"),
			},
		},

		{
			name:             "invalid commit seal",
			acceptedProposal: &types.MsgProposal{BlockHash: []byte("keccak")},
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
					IsValidSignatureFn: func(_, _, seal []byte) bool {
						return bytes.Equal(seal, []byte("commit seal"))
					},
				},
			},

			msg: &types.MsgCommit{
				View:       &types.View{Sequence: 101},
				From:       mock.NewValidatorID("chris"),
				BlockHash:  []byte("keccak"),
				CommitSeal: []byte("invalid commit seal"),
			},
		},

		{
			name:             "valid msg",
			acceptedProposal: &types.MsgProposal{BlockHash: []byte("keccak")},
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
					IsValidSignatureFn: func(_, _, seal []byte) bool {
						return bytes.Equal(seal, []byte("commit seal"))
					},
				},
			},

			isValid: true,
			msg: &types.MsgCommit{
				View:       &types.View{Sequence: 101},
				From:       mock.NewValidatorID("chris"),
				BlockHash:  []byte("keccak"),
				CommitSeal: []byte("commit seal"),
			},
		},
	}

	for _, tt := range table {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := New(tt.validator, 0)
			s.state.proposal = tt.acceptedProposal

			assert.Equal(t, tt.isValid, s.isValidMsgCommit(tt.msg))
		})
	}
}

func TestIsValidMsgRoundChange(t *testing.T) {
	t.Parallel()

	table := []struct {
		validator ibft.Validator
		quorum    ibft.Quorum
		keccak    ibft.Keccak
		msg       *types.MsgRoundChange
		name      string
		isValid   bool
	}{
		{
			name: "invalid sender",
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob").IsValidator,
				},
			},

			msg: &types.MsgRoundChange{
				View: &types.View{Sequence: 101},
				From: mock.NewValidatorID("chris"),
			},
		},

		{
			name: "valid msg (pb and pc are nil)",
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
				},
			},

			isValid: true,
			msg: &types.MsgRoundChange{
				View: &types.View{Sequence: 101},
				From: mock.NewValidatorID("chris"),
			},
		},

		{
			name: "pb and pc are not both present",
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
				},
			},

			msg: &types.MsgRoundChange{
				View:                      &types.View{Sequence: 101},
				From:                      mock.NewValidatorID("chris"),
				LatestPreparedCertificate: &types.PreparedCertificate{},
			},
		},

		{
			name: "(invalid pc) proposal msg and prepare messages are not both present",
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
				},
			},

			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101},
				From:                        mock.NewValidatorID("chris"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: nil,
				},
			},
		},

		{
			name: "(invalid pc) invalid sequence in proposal msg",
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
				},
			},

			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101},
				From:                        mock.NewValidatorID("chris"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						View: &types.View{Sequence: 100},
					},
					PrepareMessages: []*types.MsgPrepare{},
				},
			},
		},

		{
			name: "(invalid pc) invalid round in proposal msg",
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
				},
			},

			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101},
				From:                        mock.NewValidatorID("chris"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						View: &types.View{Sequence: 101, Round: 5},
					},
					PrepareMessages: []*types.MsgPrepare{},
				},
			},
		},

		{
			name: "(invalid pc) invalid proposer in proposal msg",
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 0}),
				},
			},

			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 1},
				From:                        mock.NewValidatorID("chris"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						From: mock.NewValidatorID("dani"),
						View: &types.View{Sequence: 101, Round: 0},
					},
					PrepareMessages: []*types.MsgPrepare{},
				},
			},
		},

		{
			name: "(invalid pc) proposal and prepare sequence mismatch",
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 1}),
				},
			},

			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 2},
				From:                        mock.NewValidatorID("chris"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						From: mock.NewValidatorID("bob"),
						View: &types.View{Sequence: 101, Round: 1},
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							View: &types.View{Sequence: 99},
						},
					},
				},
			},
		},

		{
			name: "(invalid pc) proposal and prepare round mismatch",
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 1}),
				},
			},

			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 2},
				From:                        mock.NewValidatorID("chris"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						From: mock.NewValidatorID("bob"),
						View: &types.View{Sequence: 101, Round: 1},
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							View: &types.View{Sequence: 101, Round: 0},
						},
					},
				},
			},
		},

		{
			name: "(invalid pc) invalid block hash in proposal and prepare",
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 1}),
				},
			},

			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 2},
				From:                        mock.NewValidatorID("chris"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						From:      mock.NewValidatorID("bob"),
						View:      &types.View{Sequence: 101, Round: 1},
						BlockHash: []byte("keccak"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							View:      &types.View{Sequence: 101, Round: 1},
							BlockHash: []byte("some other keccak"),
						},
					},
				},
			},
		},

		{
			name: "(invalid pc) invalid sender in prepare msg",
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 1}),
				},
			},

			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 2},
				From:                        mock.NewValidatorID("chris"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						From:      mock.NewValidatorID("bob"),
						View:      &types.View{Sequence: 101, Round: 1},
						BlockHash: []byte("keccak"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							From:      mock.NewValidatorID("dani"),
							View:      &types.View{Sequence: 101, Round: 1},
							BlockHash: []byte("keccak"),
						},
					},
				},
			},
		},

		{
			name: "(invalid pc) duplicate sender in prepare msgs",
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 1}),
				},
			},

			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 2},
				From:                        mock.NewValidatorID("chris"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						From:      mock.NewValidatorID("bob"),
						View:      &types.View{Sequence: 101, Round: 1},
						BlockHash: []byte("keccak"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							From:      mock.NewValidatorID("chris"),
							View:      &types.View{Sequence: 101, Round: 1},
							BlockHash: []byte("keccak"),
						},

						{
							From:      mock.NewValidatorID("chris"),
							View:      &types.View{Sequence: 101, Round: 1},
							BlockHash: []byte("keccak"),
						},
					},
				},
			},
		},

		{
			name:   "(invalid pc) no quorum messages",
			quorum: mock.NoQuorum,
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 1}),
				},
			},

			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 2},
				From:                        mock.NewValidatorID("chris"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						From:      mock.NewValidatorID("bob"),
						View:      &types.View{Sequence: 101, Round: 1},
						BlockHash: []byte("keccak"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							From:      mock.NewValidatorID("chris"),
							View:      &types.View{Sequence: 101, Round: 1},
							BlockHash: []byte("keccak"),
						},
					},
				},
			},
		},

		{
			name:   "latest ppb hash does not match proposal block hash",
			quorum: mock.NonZeroQuorum,
			keccak: mock.NewDummyKeccak("keccak"),
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 1}),
				},
			},

			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 2},
				From:                        mock.NewValidatorID("chris"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						From:      mock.NewValidatorID("bob"),
						View:      &types.View{Sequence: 101, Round: 1},
						BlockHash: []byte("some other keccak"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							From:      mock.NewValidatorID("chris"),
							View:      &types.View{Sequence: 101, Round: 1},
							BlockHash: []byte("some other keccak"),
						},
					},
				},
			},
		},

		{
			name:   "valid round change msg",
			quorum: mock.NonZeroQuorum,
			keccak: mock.NewDummyKeccak("keccak"),
			validator: mock.Validator{
				Verifier: mock.Verifier{
					IsValidatorFn: mock.NewValidatorSet("alice", "bob", "chris").IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: mock.NewValidatorID("bob"), Round: 1}),
				},
			},

			isValid: true,
			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 2},
				From:                        mock.NewValidatorID("chris"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						From:      mock.NewValidatorID("bob"),
						View:      &types.View{Sequence: 101, Round: 1},
						BlockHash: []byte("keccak"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							From:      mock.NewValidatorID("chris"),
							View:      &types.View{Sequence: 101, Round: 1},
							BlockHash: []byte("keccak"),
						},
					},
				},
			},
		},
	}

	for _, tt := range table {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := New(tt.validator, 0)
			assert.Equal(t, tt.isValid, s.isValidMsgRoundChange(tt.msg, tt.quorum, tt.keccak))
		})
	}
}
