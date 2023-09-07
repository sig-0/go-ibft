package sequencer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	ibft "github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

func TestIsValidMsgProposal(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name    string
		msg     *types.MsgProposal
		isValid bool

		// setup
		validator ibft.Validator
		verifier  ibft.Verifier
		quorum    ibft.Quorum
		keccak    ibft.Keccak
	}{
		{
			name: "invalid round in proposed block",
			msg: &types.MsgProposal{
				View:          &types.View{Round: 5},
				ProposedBlock: &types.ProposedBlock{Round: 0},
			},

			validator: mockValidator{idFn: func() []byte {
				return []byte("my validator")
			}},
		},

		{
			name: "we are the proposer",
			msg: &types.MsgProposal{
				View: &types.View{Sequence: 101, Round: 0},
				From: []byte("my validator"),
				ProposedBlock: &types.ProposedBlock{
					Block: []byte("block"),
					Round: 0,
				},
			},

			validator: mockValidator{
				idFn: func() []byte {
					return []byte("my validator")
				},
			},
		},

		{
			name: "invalid proposer",
			msg: &types.MsgProposal{
				View: &types.View{Sequence: 101, Round: 0},
				From: []byte("invalid proposer"),
				ProposedBlock: &types.ProposedBlock{
					Block: []byte("block"),
					Round: 0,
				},
			},

			validator: mockValidator{
				idFn: func() []byte { return []byte("my validator") },
			},
			verifier: mockVerifier{
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
			},
		},

		{
			name: "invalid block hash",
			msg: &types.MsgProposal{
				View: &types.View{Sequence: 101, Round: 0},
				From: []byte("proposer"),
				ProposedBlock: &types.ProposedBlock{
					Block: []byte("block"),
					Round: 0,
				},
				BlockHash: []byte("invalid block hash"),
			},

			validator: mockValidator{
				idFn: func() []byte { return []byte("my validator") },
			},
			verifier: mockVerifier{
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
			},

			keccak: KeccakFn(func(_ []byte) []byte { return []byte("block hash") }),
		},

		{
			name: "invalid round 0 block",
			msg: &types.MsgProposal{
				View: &types.View{Sequence: 101, Round: 0},
				From: []byte("proposer"),
				ProposedBlock: &types.ProposedBlock{
					Block: []byte("invalid round 0 block"),
					Round: 0,
				},
				BlockHash: []byte("block hash"),
			},

			validator: mockValidator{
				idFn: func() []byte { return []byte("my validator") },
			},
			verifier: mockVerifier{
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				isValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},

			keccak: KeccakFn(func(_ []byte) []byte { return []byte("block hash") }),
		},

		{
			name:    "valid proposal msg",
			isValid: true,
			msg: &types.MsgProposal{
				View: &types.View{Sequence: 101, Round: 0},
				From: []byte("proposer"),
				ProposedBlock: &types.ProposedBlock{
					Block: []byte("block"),
					Round: 0,
				},
				BlockHash: []byte("block hash"),
			},

			validator: mockValidator{
				idFn: func() []byte { return []byte("my validator") },
			},
			verifier: mockVerifier{
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				isValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},

			keccak: KeccakFn(func(_ []byte) []byte { return []byte("block hash") }),
		},

		{
			name: "(non zero round): nil rcc",
			msg: &types.MsgProposal{
				View: &types.View{Sequence: 101, Round: 1},
				From: []byte("proposer"),
				ProposedBlock: &types.ProposedBlock{
					Block: []byte("block"),
					Round: 1,
				},
				BlockHash:              []byte("block hash"),
				RoundChangeCertificate: nil,
			},

			validator: mockValidator{
				idFn: func() []byte { return []byte("my validator") },
			},
			verifier: mockVerifier{
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				isValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},

			keccak: KeccakFn(func(_ []byte) []byte { return []byte("block hash") }),
		},

		{
			name: "(non zero round): empty rcc",
			msg: &types.MsgProposal{
				View: &types.View{Sequence: 101, Round: 1},
				From: []byte("proposer"),
				ProposedBlock: &types.ProposedBlock{
					Block: []byte("block"),
					Round: 1,
				},
				BlockHash:              []byte("block hash"),
				RoundChangeCertificate: &types.RoundChangeCertificate{},
			},

			validator: mockValidator{
				idFn: func() []byte { return []byte("my validator") },
			},
			verifier: mockVerifier{
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				isValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},

			keccak: KeccakFn(func(_ []byte) []byte { return []byte("block hash") }),
		},

		{
			name: "(non zero round): invalid sequence in rcc",
			msg: &types.MsgProposal{
				View: &types.View{Sequence: 101, Round: 1},
				From: []byte("proposer"),
				ProposedBlock: &types.ProposedBlock{
					Block: []byte("block"),
					Round: 1,
				},
				BlockHash: []byte("block hash"),
				RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
					{
						View: &types.View{Sequence: 100},
					},
				}},
			},

			validator: mockValidator{
				idFn: func() []byte { return []byte("my validator") },
			},
			verifier: mockVerifier{
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				isValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},

			keccak: KeccakFn(func(_ []byte) []byte { return []byte("block hash") }),
		},

		{
			name: "(non zero round): invalid round in rcc",
			msg: &types.MsgProposal{
				View: &types.View{Sequence: 101, Round: 1},
				From: []byte("proposer"),
				ProposedBlock: &types.ProposedBlock{
					Block: []byte("block"),
					Round: 1,
				},
				BlockHash: []byte("block hash"),
				RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
					{
						View: &types.View{Sequence: 101, Round: 0},
					},
				}},
			},

			validator: mockValidator{
				idFn: func() []byte { return []byte("my validator") },
			},
			verifier: mockVerifier{
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				isValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},

			keccak: KeccakFn(func(_ []byte) []byte { return []byte("block hash") }),
		},

		{
			name: "(non zero round): invalid sender in rcc",
			msg: &types.MsgProposal{
				View: &types.View{Sequence: 101, Round: 1},
				From: []byte("proposer"),
				ProposedBlock: &types.ProposedBlock{
					Block: []byte("block"),
					Round: 1,
				},
				BlockHash: []byte("block hash"),
				RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
					{
						View: &types.View{Sequence: 101, Round: 1},
						From: []byte("not a validator"),
					},
				}},
			},

			validator: mockValidator{
				idFn: func() []byte { return []byte("my validator") },
			},
			verifier: mockVerifier{
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				isValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
				isValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
			},

			keccak: KeccakFn(func(_ []byte) []byte { return []byte("block hash") }),
		},

		{
			name: "(non zero round): duplicate sender in rcc",
			msg: &types.MsgProposal{
				View: &types.View{Sequence: 101, Round: 1},
				From: []byte("proposer"),
				ProposedBlock: &types.ProposedBlock{
					Block: []byte("block"),
					Round: 1,
				},
				BlockHash: []byte("block hash"),
				RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
					{
						View: &types.View{Sequence: 101, Round: 1},
						From: []byte("validator"),
					},
					{
						View: &types.View{Sequence: 101, Round: 1},
						From: []byte("validator"),
					},
				}},
			},

			validator: mockValidator{
				idFn: func() []byte { return []byte("my validator") },
			},
			verifier: mockVerifier{
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				isValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
				isValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
			},

			keccak: KeccakFn(func(_ []byte) []byte { return []byte("block hash") }),
		},

		{
			name: "(non zero round): no quorum in rcc",
			msg: &types.MsgProposal{
				View: &types.View{Sequence: 101, Round: 1},
				From: []byte("proposer"),
				ProposedBlock: &types.ProposedBlock{
					Block: []byte("block"),
					Round: 1,
				},
				BlockHash: []byte("block hash"),
				RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
					{
						View: &types.View{Sequence: 101, Round: 1},
						From: []byte("validator"),
					},
				}},
			},

			validator: mockValidator{
				idFn: func() []byte { return []byte("my validator") },
			},
			verifier: mockVerifier{
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				isValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
				isValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
			},

			keccak: KeccakFn(func(_ []byte) []byte { return []byte("block hash") }),
			quorum: QuorumFn(func(_ uint64, _ []types.Msg) bool { return false }),
		},

		{
			name: "(non zero round): invalid block in rcc",
			msg: &types.MsgProposal{
				View: &types.View{Sequence: 101, Round: 1},
				From: []byte("proposer"),
				ProposedBlock: &types.ProposedBlock{
					Block: []byte("invalid block" +
						""),
					Round: 1,
				},
				BlockHash: []byte("block hash"),
				RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
					{
						View: &types.View{Sequence: 101, Round: 1},
						From: []byte("validator"),
					},
				}},
			},

			validator: mockValidator{
				idFn: func() []byte { return []byte("my validator") },
			},
			verifier: mockVerifier{
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				isValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
				isValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
			},

			keccak: KeccakFn(func(_ []byte) []byte { return []byte("block hash") }),
			quorum: QuorumFn(func(_ uint64, _ []types.Msg) bool { return true }),
		},

		{
			name:    "(non zero round): valid proposal msg",
			isValid: true,
			msg: &types.MsgProposal{
				View: &types.View{Sequence: 101, Round: 1},
				From: []byte("proposer"),
				ProposedBlock: &types.ProposedBlock{
					Block: []byte("block"),
					Round: 1,
				},
				BlockHash: []byte("block hash"),
				RoundChangeCertificate: &types.RoundChangeCertificate{Messages: []*types.MsgRoundChange{
					{
						View: &types.View{Sequence: 101, Round: 1},
						From: []byte("validator"),
					},
				}},
			},

			validator: mockValidator{
				idFn: func() []byte { return []byte("my validator") },
			},
			verifier: mockVerifier{
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				isValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
				isValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
			},

			keccak: KeccakFn(func(_ []byte) []byte { return []byte("block hash") }),
			quorum: QuorumFn(func(_ uint64, _ []types.Msg) bool { return true }),
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seq := New(tt.validator, tt.verifier, 0)
			assert.Equal(t, tt.isValid, seq.isValidMsgProposal(tt.msg, tt.quorum, tt.keccak))
		})
	}
}

func TestIsValidMsgPrepare(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name    string
		msg     *types.MsgPrepare
		isValid bool

		// setup
		verifier         ibft.Verifier
		acceptedProposal *types.MsgProposal
	}{
		{
			name: "invalid sender",
			msg: &types.MsgPrepare{
				View: &types.View{Sequence: 101},
				From: []byte("not a validator"),
			},
			verifier: mockVerifier{isValidatorFn: func(from []byte, _ uint64) bool {
				return bytes.Equal(from, []byte("validator"))
			}},
		},

		{
			name: "invalid block hash",
			msg: &types.MsgPrepare{
				View:      &types.View{Sequence: 101},
				From:      []byte("validator"),
				BlockHash: []byte("invalid block hash"),
			},
			verifier: mockVerifier{isValidatorFn: func(from []byte, _ uint64) bool {
				return bytes.Equal(from, []byte("validator"))
			}},
			acceptedProposal: &types.MsgProposal{
				BlockHash: []byte("block hash"),
			},
		},

		{
			name:    "valid prepare msg",
			isValid: true,
			msg: &types.MsgPrepare{
				View:      &types.View{Sequence: 101},
				From:      []byte("validator"),
				BlockHash: []byte("block hash"),
			},
			verifier: mockVerifier{isValidatorFn: func(from []byte, _ uint64) bool {
				return bytes.Equal(from, []byte("validator"))
			}},
			acceptedProposal: &types.MsgProposal{
				BlockHash: []byte("block hash"),
			},
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seq := New(nil, tt.verifier, 0)
			seq.state.acceptedProposal = tt.acceptedProposal

			assert.Equal(t, tt.isValid, seq.isValidMsgPrepare(tt.msg))
		})
	}
}

func TestIsValidMsgCommit(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name    string
		msg     *types.MsgCommit
		isValid bool

		// setup
		verifier         ibft.Verifier
		recover          ibft.SigRecover
		acceptedProposal *types.MsgProposal
	}{
		{
			name: "invalid block hash",
			msg: &types.MsgCommit{
				BlockHash: []byte("invalid block hash"),
			},

			acceptedProposal: &types.MsgProposal{
				BlockHash: []byte("block hash"),
			},
		},

		{
			name: "invalid sender",
			msg: &types.MsgCommit{
				View:      &types.View{Sequence: 101},
				From:      []byte("not a validator"),
				BlockHash: []byte("block hash"),
			},

			verifier: mockVerifier{isValidatorFn: func(from []byte, _ uint64) bool {
				return bytes.Equal(from, []byte("validator"))
			}},
			acceptedProposal: &types.MsgProposal{
				BlockHash: []byte("block hash"),
			},
		},

		{
			name: "invalid commit seal",
			msg: &types.MsgCommit{
				View:       &types.View{Sequence: 101},
				BlockHash:  []byte("block hash"),
				From:       []byte("validator"),
				CommitSeal: []byte("invalid commit seal"),
			},

			verifier: mockVerifier{isValidatorFn: func(from []byte, _ uint64) bool {
				return bytes.Equal(from, []byte("validator"))
			}},
			acceptedProposal: &types.MsgProposal{
				BlockHash: []byte("block hash"),
			},

			recover: SigRecoverFn(func(_ []byte, _ []byte) []byte {
				return []byte("commit seal")
			}),
		},

		{
			name:    "valid commit msg",
			isValid: true,
			msg: &types.MsgCommit{
				View:       &types.View{Sequence: 101},
				BlockHash:  []byte("block hash"),
				From:       []byte("validator"),
				CommitSeal: []byte("commit seal"),
			},

			verifier: mockVerifier{isValidatorFn: func(from []byte, _ uint64) bool {
				return bytes.Equal(from, []byte("validator"))
			}},
			acceptedProposal: &types.MsgProposal{
				BlockHash: []byte("block hash"),
			},

			recover: SigRecoverFn(func(_ []byte, _ []byte) []byte {
				return []byte("validator")
			}),
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := New(nil, tt.verifier, 0)
			s.state.acceptedProposal = tt.acceptedProposal

			assert.Equal(t, tt.isValid, s.isValidCommit(tt.msg, tt.recover))
		})
	}
}

func TestIsValidMsgRoundChange(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name    string
		msg     *types.MsgRoundChange
		isValid bool

		// setup
		verifier ibft.Verifier
		quorum   ibft.Quorum
		keccak   ibft.Keccak
	}{
		{
			name: "invalid sender",
			msg: &types.MsgRoundChange{
				View: &types.View{Sequence: 101},
				From: []byte("not a validator"),
			},
			verifier: mockVerifier{isValidatorFn: func(from []byte, _ uint64) bool {
				return bytes.Equal(from, []byte("validator"))
			}},
		},

		{
			name:    "valid msg (pb and pc are nil)",
			isValid: true,
			msg: &types.MsgRoundChange{
				View: &types.View{Sequence: 101},
				From: []byte("validator"),
			},
			verifier: mockVerifier{isValidatorFn: func(from []byte, _ uint64) bool {
				return bytes.Equal(from, []byte("validator"))
			}},
		},

		{
			name: "pb is nil but pc is not",
			msg: &types.MsgRoundChange{
				View:                      &types.View{Sequence: 101},
				From:                      []byte("validator"),
				LatestPreparedCertificate: &types.PreparedCertificate{},
			},
			verifier: mockVerifier{isValidatorFn: func(from []byte, _ uint64) bool {
				return bytes.Equal(from, []byte("validator"))
			}},
		},

		{
			name: "pc is nil but pb is not",
			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101},
				From:                        []byte("validator"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
			},
			verifier: mockVerifier{isValidatorFn: func(from []byte, _ uint64) bool {
				return bytes.Equal(from, []byte("validator"))
			}},
		},

		{
			name: "(invalid pc) nil proposal msg",
			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101},
				From:                        []byte("validator"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: nil,
				},
			},
			verifier: mockVerifier{isValidatorFn: func(from []byte, _ uint64) bool {
				return bytes.Equal(from, []byte("validator"))
			}},
		},

		{
			name: "(invalid pc) nil prepare messages",
			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101},
				From:                        []byte("validator"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{},
					PrepareMessages: nil,
				},
			},
			verifier: mockVerifier{isValidatorFn: func(from []byte, _ uint64) bool {
				return bytes.Equal(from, []byte("validator"))
			}},
		},

		{
			name: "(invalid pc) invalid sequence in proposal msg",
			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101},
				From:                        []byte("validator"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						View: &types.View{Sequence: 100},
					},
					PrepareMessages: []*types.MsgPrepare{},
				},
			},
			verifier: mockVerifier{isValidatorFn: func(from []byte, _ uint64) bool {
				return bytes.Equal(from, []byte("validator"))
			}},
		},

		{
			name: "(invalid pc) invalid round in proposal msg",
			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 1},
				From:                        []byte("validator"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						View: &types.View{
							Sequence: 101,
							Round:    5,
						},
					},
					PrepareMessages: []*types.MsgPrepare{},
				},
			},
			verifier: mockVerifier{isValidatorFn: func(from []byte, _ uint64) bool {
				return bytes.Equal(from, []byte("validator"))
			}},
		},

		{
			name: "(invalid pc) invalid proposer in proposal msg",
			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 1},
				From:                        []byte("validator"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						From:      []byte("not a proposer"),
						View:      &types.View{Sequence: 101, Round: 0},
						BlockHash: []byte("block hash"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							View:      &types.View{Sequence: 101, Round: 0},
							From:      []byte("validator"),
							BlockHash: []byte("block hash"),
						},
					},
				},
			},
			verifier: mockVerifier{
				isValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
			},
		},

		{
			name: "(invalid pc) proposal and prepare sequence mismatch",
			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 2},
				From:                        []byte("validator"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						From: []byte("proposer"),
						View: &types.View{Sequence: 101, Round: 1},
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							View: &types.View{Sequence: 102},
						},
					},
				},
			},
			verifier: mockVerifier{
				isValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
			},
		},

		{
			name: "(invalid pc) proposal and prepare round mismatch",
			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 2},
				From:                        []byte("validator"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						From: []byte("proposer"),
						View: &types.View{Sequence: 101, Round: 1},
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							View: &types.View{Sequence: 101, Round: 0},
						},
					},
				},
			},
			verifier: mockVerifier{
				isValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
			},
		},

		{
			name: "(invalid pc) invalid block hash in prepare msg",
			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 2},
				From:                        []byte("validator"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						From:      []byte("proposer"),
						View:      &types.View{Sequence: 101, Round: 0},
						BlockHash: []byte("block hash"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							View:      &types.View{Sequence: 101, Round: 0},
							BlockHash: []byte("invalid block hash"),
						},
					},
				},
			},
			verifier: mockVerifier{
				isValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
			},
		},

		{
			name: "(invalid pc) invalid sender in prepare msg",
			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 1},
				From:                        []byte("validator"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						View:      &types.View{Sequence: 101, Round: 0},
						From:      []byte("proposer"),
						BlockHash: []byte("block hash"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							View:      &types.View{Sequence: 101, Round: 0},
							From:      []byte("not a validator"),
							BlockHash: []byte("block hash"),
						},
					},
				},
			},
			verifier: mockVerifier{
				isValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
			},
		},

		{
			name: "(invalid pc) duplicate sender in prepare msgs",
			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 1},
				From:                        []byte("validator"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						From:      []byte("proposer"),
						View:      &types.View{Sequence: 101, Round: 0},
						BlockHash: []byte("block hash"),
					},
					PrepareMessages: []*types.MsgPrepare{
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
				},
			},
			verifier: mockVerifier{
				isValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
			},
		},

		{
			name: "(invalid pc) no quorum",
			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 1},
				From:                        []byte("validator"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						From:      []byte("proposer"),
						View:      &types.View{Sequence: 101, Round: 0},
						BlockHash: []byte("block hash"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							View:      &types.View{Sequence: 101, Round: 0},
							From:      []byte("validator"),
							BlockHash: []byte("block hash"),
						},
					},
				},
			},
			verifier: mockVerifier{
				isValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
			},

			quorum: QuorumFn(func(_ uint64, _ []types.Msg) bool { return false }),
		},

		{
			name: "latest ppb hash does not match proposal block hash",
			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 1},
				From:                        []byte("validator"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						View:      &types.View{Sequence: 101, Round: 0},
						From:      []byte("proposer"),
						BlockHash: []byte("invalid block hash"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							View:      &types.View{Sequence: 101, Round: 0},
							From:      []byte("validator"),
							BlockHash: []byte("invalid block hash"),
						},
					},
				},
			},
			verifier: mockVerifier{
				isValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
			},

			quorum: QuorumFn(func(_ uint64, _ []types.Msg) bool { return true }),
			keccak: KeccakFn(func(_ []byte) []byte { return []byte("block hash") }),
		},

		{
			name:    "valid round change msg",
			isValid: true,
			msg: &types.MsgRoundChange{
				View:                        &types.View{Sequence: 101, Round: 1},
				From:                        []byte("validator"),
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						View:      &types.View{Sequence: 101, Round: 0},
						From:      []byte("proposer"),
						BlockHash: []byte("block hash"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							View:      &types.View{Sequence: 101, Round: 0},
							From:      []byte("validator"),
							BlockHash: []byte("block hash"),
						},
					},
				},
			},

			verifier: mockVerifier{
				isValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
			},
			quorum: QuorumFn(func(_ uint64, _ []types.Msg) bool { return true }),
			keccak: KeccakFn(func(_ []byte) []byte { return []byte("block hash") }),
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := New(nil, tt.verifier, 0)
			assert.Equal(t, tt.isValid, s.isValidMsgRoundChange(tt.msg, tt.quorum, tt.keccak))
		})
	}
}
