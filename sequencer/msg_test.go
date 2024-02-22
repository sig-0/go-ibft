package sequencer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

var (
	FalseQuorum = QuorumFn(func(_ uint64, _ []ibft.Message) bool { return false })
	TrueQuorum  = QuorumFn(func(_ uint64, _ []ibft.Message) bool { return true })
)

func TestIsValidMsgProposal(t *testing.T) {
	t.Parallel()

	table := []struct {
		validator ibft.Validator
		verifier  ibft.Verifier
		quorum    ibft.Quorum
		keccak    ibft.Keccak
		msg       *types.MsgProposal
		name      string
		isValid   bool
	}{
		{
			name: "invalid round in proposed block",
			msg: &types.MsgProposal{
				View:          &types.View{Round: 5},
				ProposedBlock: &types.ProposedBlock{Round: 0},
			},

			validator: MockValidator{IDFn: func() []byte {
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

			validator: MockValidator{
				IDFn: func() []byte {
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

			validator: MockValidator{
				IDFn: func() []byte { return []byte("my validator") },
			},
			verifier: MockVerifier{
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
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

			validator: MockValidator{
				IDFn: func() []byte { return []byte("my validator") },
			},
			verifier: MockVerifier{
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
			},

			keccak: BlockHashKeccak,
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

			validator: MockValidator{
				IDFn: func() []byte { return []byte("my validator") },
			},
			verifier: MockVerifier{
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},

			keccak: BlockHashKeccak,
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

			validator: MockValidator{
				IDFn: func() []byte { return []byte("my validator") },
			},
			verifier: MockVerifier{
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},

			keccak: BlockHashKeccak,
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

			validator: MockValidator{
				IDFn: func() []byte { return []byte("my validator") },
			},
			verifier: MockVerifier{
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},

			keccak: BlockHashKeccak,
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

			validator: MockValidator{
				IDFn: func() []byte { return []byte("my validator") },
			},
			verifier: MockVerifier{
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},

			keccak: BlockHashKeccak,
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

			validator: MockValidator{
				IDFn: func() []byte { return []byte("my validator") },
			},
			verifier: MockVerifier{
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},

			keccak: BlockHashKeccak,
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

			validator: MockValidator{
				IDFn: func() []byte { return []byte("my validator") },
			},
			verifier: MockVerifier{
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},

			keccak: BlockHashKeccak,
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

			validator: MockValidator{
				IDFn: func() []byte { return []byte("my validator") },
			},
			verifier: MockVerifier{
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
				IsValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
			},

			keccak: BlockHashKeccak,
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

			validator: MockValidator{
				IDFn: func() []byte { return []byte("my validator") },
			},
			verifier: MockVerifier{
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
				IsValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
			},

			keccak: BlockHashKeccak,
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

			validator: MockValidator{
				IDFn: func() []byte { return []byte("my validator") },
			},
			verifier: MockVerifier{
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
				IsValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
			},

			keccak: BlockHashKeccak,
			quorum: FalseQuorum,
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

			validator: MockValidator{
				IDFn: func() []byte { return []byte("my validator") },
			},
			verifier: MockVerifier{
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
				IsValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
			},

			keccak: BlockHashKeccak,
			quorum: FalseQuorum,
		},

		{
			name: "(non zero round): pb hash does not equal highest round block hash",
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
						View:                        &types.View{Sequence: 101, Round: 1},
						From:                        []byte("validator"),
						LatestPreparedProposedBlock: &types.ProposedBlock{},
						LatestPreparedCertificate: &types.PreparedCertificate{
							ProposalMessage: &types.MsgProposal{
								From:      []byte("proposer"),
								View:      &types.View{Sequence: 101, Round: 0},
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
				}},
			},

			validator: MockValidator{
				IDFn: func() []byte { return []byte("my validator") },
			},
			verifier: MockVerifier{
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
				IsValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
			},

			keccak: BlockHashKeccak,
			quorum: TrueQuorum,
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

			validator: MockValidator{
				IDFn: func() []byte { return []byte("my validator") },
			},
			verifier: MockVerifier{
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
				IsValidBlockFn: func(block []byte) bool {
					return bytes.Equal(block, []byte("block"))
				},
				IsValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
			},

			keccak: BlockHashKeccak,
			quorum: TrueQuorum,
		},
	}

	for _, tt := range table {
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

	table := []struct {
		verifier         ibft.Verifier
		msg              *types.MsgPrepare
		acceptedProposal *types.MsgProposal
		name             string
		isValid          bool
	}{
		{
			name: "invalid sender",
			msg: &types.MsgPrepare{
				View: &types.View{Sequence: 101},
				From: []byte("not a validator"),
			},
			verifier: MockVerifier{IsValidatorFn: func(from []byte, _ uint64) bool {
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
			verifier: MockVerifier{IsValidatorFn: func(from []byte, _ uint64) bool {
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
			verifier: MockVerifier{IsValidatorFn: func(from []byte, _ uint64) bool {
				return bytes.Equal(from, []byte("validator"))
			}},
			acceptedProposal: &types.MsgProposal{
				BlockHash: []byte("block hash"),
			},
		},
	}

	for _, tt := range table {
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

	table := []struct {
		verifier         ibft.Verifier
		recover          ibft.SigRecover
		msg              *types.MsgCommit
		acceptedProposal *types.MsgProposal
		name             string
		isValid          bool
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

			verifier: MockVerifier{IsValidatorFn: func(from []byte, _ uint64) bool {
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

			verifier: MockVerifier{IsValidatorFn: func(from []byte, _ uint64) bool {
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

			verifier: MockVerifier{IsValidatorFn: func(from []byte, _ uint64) bool {
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

	for _, tt := range table {
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

	table := []struct {
		verifier ibft.Verifier
		quorum   ibft.Quorum
		keccak   ibft.Keccak
		msg      *types.MsgRoundChange
		name     string
		isValid  bool
	}{
		{
			name: "invalid sender",
			msg: &types.MsgRoundChange{
				View: &types.View{Sequence: 101},
				From: []byte("not a validator"),
			},
			verifier: MockVerifier{IsValidatorFn: func(from []byte, _ uint64) bool {
				return bytes.Equal(from, []byte("validator"))
			}},
		},

		{
			name:    "valid msg (pb and pc are nil",
			isValid: true,
			msg: &types.MsgRoundChange{
				View: &types.View{Sequence: 101},
				From: []byte("validator"),
			},
			verifier: MockVerifier{IsValidatorFn: func(from []byte, _ uint64) bool {
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
			verifier: MockVerifier{IsValidatorFn: func(from []byte, _ uint64) bool {
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
			verifier: MockVerifier{IsValidatorFn: func(from []byte, _ uint64) bool {
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
			verifier: MockVerifier{IsValidatorFn: func(from []byte, _ uint64) bool {
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
			verifier: MockVerifier{IsValidatorFn: func(from []byte, _ uint64) bool {
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
			verifier: MockVerifier{IsValidatorFn: func(from []byte, _ uint64) bool {
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
			verifier: MockVerifier{IsValidatorFn: func(from []byte, _ uint64) bool {
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
			verifier: MockVerifier{
				IsValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
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
			verifier: MockVerifier{
				IsValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
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
			verifier: MockVerifier{
				IsValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
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
			verifier: MockVerifier{
				IsValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
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
			verifier: MockVerifier{
				IsValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
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
			verifier: MockVerifier{
				IsValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
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
			verifier: MockVerifier{
				IsValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
			},

			quorum: FalseQuorum,
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
			verifier: MockVerifier{
				IsValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
			},

			quorum: TrueQuorum,
			keccak: BlockHashKeccak,
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

			verifier: MockVerifier{
				IsValidatorFn: func(from []byte, _ uint64) bool {
					return bytes.Equal(from, []byte("validator"))
				},
				IsProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, []byte("proposer"))
				},
			},
			quorum: TrueQuorum,
			keccak: BlockHashKeccak,
		},
	}

	for _, tt := range table {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := New(nil, tt.verifier, 0)
			assert.Equal(t, tt.isValid, s.isValidMsgRoundChange(tt.msg, tt.quorum, tt.keccak))
		})
	}
}
