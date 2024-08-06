package sequencer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sig-0/go-ibft"
	"github.com/sig-0/go-ibft/message/types"
	"github.com/sig-0/go-ibft/test/mock"
)

//nolint:dupl // messages are not entirely different among cases
func TestIsValidMsgProposal(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		validator    ibft.Validator
		validatorSet ibft.ValidatorSet
		quorum       ibft.Quorum
		sigVerifier  ibft.SignatureVerifier
		msg          *types.MsgProposal
		name         string
		isValid      bool
	}{
		{
			name:      "proposed block round and view round do not match",
			validator: mock.Validator{IDFn: Alice.ID},
			msg: &types.MsgProposal{
				Metadata:      &types.MsgMetadata{View: &types.View{Round: 0}},
				ProposedBlock: &types.ProposedBlock{Round: 5},
			},
		},

		{
			name:      "cannot verify own proposal",
			validator: mock.Validator{IDFn: Alice.ID},
			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					Sender: Alice,
					View:   &types.View{Sequence: 101, Round: 0},
				},
				ProposedBlock: &types.ProposedBlock{Round: 0},
			},
		},

		{
			name: "invalid proposer",
			validatorSet: mock.ValidatorSet{
				//IsValidatorFn: func(id []byte, _ uint64) bool {
				//	return bytes.Equal(id, Bob)
				//},
				IsProposerFn: func(id []byte, _, round uint64) bool {
					return bytes.Equal(id, Bob) && round == 0
				},
			},
			//validatorSet: mock.ProposersInRounds(mock.Proposer{ID: Bob, Round: 0}),
			validator: mock.Validator{IDFn: Alice.ID},
			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101, Round: 0},
					Sender: mock.NewValidatorID("definitely not bob"),
				},
				ProposedBlock: &types.ProposedBlock{
					Block: []byte("block"),
					Round: 0,
				},
			},
		},

		{
			name:      "invalid block hash",
			validator: mock.Validator{IDFn: Alice.ID},
			validatorSet: mock.ValidatorSet{IsProposerFn: func(id []byte, _ uint64, round uint64) bool {
				return bytes.Equal(id, Bob) && round == 0
			}},
			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101, Round: 0},
					Sender: Bob,
				},
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 0},
				BlockHash:     []byte("definitely not keccak"),
			},
		},

		{
			name: "invalid round 0 block",
			validator: mock.Validator{
				IDFn: Alice.ID,
				IsValidProposalFn: func(block []byte, _ uint64) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},
			validatorSet: mock.ValidatorSet{IsProposerFn: func(id []byte, _ uint64, round uint64) bool {
				return bytes.Equal(id, Bob) && round == 0
			}},
			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101, Round: 0},
					Sender: Bob,
				},
				ProposedBlock: &types.ProposedBlock{Block: []byte("invalid block"), Round: 0},
				BlockHash:     []byte("keccak"),
			},
		},

		{
			name: "valid round 0 proposal",
			validatorSet: mock.ValidatorSet{IsProposerFn: func(id []byte, _ uint64, round uint64) bool {
				return bytes.Equal(id, Bob) && round == 0
			}},
			validator: mock.Validator{
				IDFn: Alice.ID,
				IsValidProposalFn: func(block []byte, _ uint64) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},
			isValid: true,
			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					Sender: Bob,
					View:   &types.View{Sequence: 101, Round: 0},
				},
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 0},
				BlockHash:     []byte("keccak"),
			},
		},

		{
			name: "(non zero round): empty rcc",
			validatorSet: mock.ValidatorSet{IsProposerFn: func(id []byte, _ uint64, round uint64) bool {
				return bytes.Equal(id, Bob) && round == 1
			}},
			validator: mock.Validator{
				IDFn: Alice.ID,
				IsValidProposalFn: func(block []byte, _ uint64) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},
			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					Sender: Bob,
					View:   &types.View{Sequence: 101, Round: 1},
				},
				ProposedBlock:          &types.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:              []byte("keccak"),
				RoundChangeCertificate: nil,
			},
		},

		{
			name: "(non zero round): invalid message sequence in rcc",
			validatorSet: mock.ValidatorSet{IsProposerFn: func(id []byte, _ uint64, round uint64) bool {
				return bytes.Equal(id, Bob) && round == 1
			}},
			validator: mock.Validator{
				IDFn: Alice.ID,
				IsValidProposalFn: func(block []byte, _ uint64) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},
			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					Sender: Bob,
					View:   &types.View{Sequence: 101, Round: 1},
				},
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{
							Metadata: &types.MsgMetadata{View: &types.View{Sequence: 99}},
						},
					},
				},
			},
		},

		{
			name: "(non zero round): invalid message round in rcc",
			validatorSet: mock.ValidatorSet{IsProposerFn: func(id []byte, _ uint64, round uint64) bool {
				return bytes.Equal(id, Bob) && round == 1
			}},
			validator: mock.Validator{
				IDFn: Alice.ID,
				IsValidProposalFn: func(block []byte, _ uint64) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},

			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					Sender: Bob,
					View:   &types.View{Sequence: 101, Round: 1},
				},
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{
							Metadata: &types.MsgMetadata{View: &types.View{Sequence: 101, Round: 0}},
						},
					},
				},
			},
		},

		{
			name: "(non zero round): invalid sender in rcc",
			validatorSet: mock.ValidatorSet{
				IsValidatorFn: func(id []byte, _ uint64) bool {
					return !bytes.Equal(id, []byte("definitely not a validator"))
				},
				IsProposerFn: func(id []byte, _ uint64, round uint64) bool {
					return bytes.Equal(id, Bob) && round == 1
				},
			},
			validator: mock.Validator{
				IDFn: Alice.ID,
				IsValidProposalFn: func(block []byte, _ uint64) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},
			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					Sender: Bob,
					View:   &types.View{Sequence: 101, Round: 1},
				},
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{
							Metadata: &types.MsgMetadata{
								Sender: []byte("definitely not a validator"),
								View:   &types.View{Sequence: 101, Round: 1},
							},
						},
					},
				},
			},
		},

		{
			name: "(non zero round): duplicate sender in rcc",
			validatorSet: mock.ValidatorSet{
				IsValidatorFn: func(id []byte, _ uint64) bool {
					return bytes.Equal(id, Chris)
				},
				IsProposerFn: func(id []byte, _ uint64, round uint64) bool {
					return bytes.Equal(id, Bob) && round == 1
				},
			},
			validator: mock.Validator{
				IDFn: Alice.ID,
				IsValidProposalFn: func(block []byte, _ uint64) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},
			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					Sender: Bob,
					View:   &types.View{Sequence: 101, Round: 1},
				},
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{
							Metadata: &types.MsgMetadata{
								Sender: Chris,
								View:   &types.View{Sequence: 101, Round: 1},
							},
						},

						{
							Metadata: &types.MsgMetadata{
								Sender: Chris,
								View:   &types.View{Sequence: 101, Round: 1},
							},
						},
					},
				},
			},
		},

		{
			name:   "(non zero round): no quorum in rcc",
			quorum: mock.NoQuorum,
			validatorSet: mock.ValidatorSet{
				IsValidatorFn: func(id []byte, _ uint64) bool {
					return bytes.Equal(id, Chris)
				},
				IsProposerFn: func(id []byte, _ uint64, round uint64) bool {
					return bytes.Equal(id, Bob) && round == 1
				},
			},
			validator: mock.Validator{
				IDFn: Alice.ID,
				IsValidProposalFn: func(block []byte, _ uint64) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},

			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					Sender: Bob,
					View:   &types.View{Sequence: 101, Round: 1},
				},
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{
							Metadata: &types.MsgMetadata{
								Sender: Chris,
								View:   &types.View{Sequence: 101, Round: 1},
							},
						},
					},
				},
			},
		},

		{
			name:   "(non zero round): invalid block in rcc",
			quorum: mock.NonZeroQuorum,
			validatorSet: mock.ValidatorSet{
				IsValidatorFn: func(id []byte, _ uint64) bool {
					return bytes.Equal(id, Chris)
				},
				IsProposerFn: func(id []byte, _ uint64, round uint64) bool {
					return bytes.Equal(id, Bob) && round == 1
				},
			},
			validator: mock.Validator{
				IDFn: Alice.ID,
				IsValidProposalFn: func(block []byte, _ uint64) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},

			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					Sender: Bob,
					View:   &types.View{Sequence: 101, Round: 1},
				},
				ProposedBlock: &types.ProposedBlock{Block: []byte("invalid block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{

							Metadata: &types.MsgMetadata{
								Sender: Chris,
								View:   &types.View{Sequence: 101, Round: 1},
							},
						},
					},
				},
			},
		},

		{
			name:   "(non zero round): highest round block hash does not match derived hash",
			quorum: mock.NonZeroQuorum,
			validatorSet: mock.ValidatorSet{
				IsValidatorFn: func(id []byte, _ uint64) bool {
					return bytes.Equal(id, Alice) || bytes.Equal(id, Chris)
				},
				IsProposerFn: func(id []byte, _ uint64, round uint64) bool {
					switch round {
					case 0:
						return bytes.Equal(id, Alice)
					case 1:
						return bytes.Equal(id, Bob)
					default:
						return false
					}
				},
			},
			validator: mock.Validator{
				IDFn: Alice.ID,
				IsValidProposalFn: func(block []byte, _ uint64) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},

			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					Sender: Bob,
					View:   &types.View{Sequence: 101, Round: 1},
				},
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{
							Metadata: &types.MsgMetadata{
								Sender: Chris,
								View:   &types.View{Sequence: 101, Round: 1},
							},
							LatestPreparedCertificate: &types.PreparedCertificate{
								ProposalMessage: &types.MsgProposal{
									Metadata: &types.MsgMetadata{
										Sender: Alice,
										View:   &types.View{Sequence: 101, Round: 0},
									},
									BlockHash: []byte("invalid keccak"),
								},
								PrepareMessages: []*types.MsgPrepare{
									{
										Metadata: &types.MsgMetadata{
											View:   &types.View{Sequence: 101, Round: 0},
											Sender: Chris,
										},
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
			name:   "(non zero round): valid proposal message",
			quorum: mock.NonZeroQuorum,
			validatorSet: mock.ValidatorSet{
				IsValidatorFn: func(id []byte, _ uint64) bool {
					return bytes.Equal(id, Alice) || bytes.Equal(id, Chris)
				},
				IsProposerFn: func(id []byte, _ uint64, round uint64) bool {
					switch round {
					case 0:
						return bytes.Equal(id, Alice)
					case 1:
						return bytes.Equal(id, Bob)
					default:
						return false
					}
				},
			},
			validator: mock.Validator{
				IDFn: Alice.ID,
				IsValidProposalFn: func(block []byte, _ uint64) bool {
					return bytes.Equal(block, []byte("block"))
				},
			},

			isValid: true,
			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					Sender: Bob,
					View:   &types.View{Sequence: 101, Round: 1},
				},
				ProposedBlock: &types.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &types.RoundChangeCertificate{
					Messages: []*types.MsgRoundChange{
						{
							Metadata: &types.MsgMetadata{
								Sender: Chris,
								View:   &types.View{Sequence: 101, Round: 1},
							},
							LatestPreparedProposedBlock: &types.ProposedBlock{},
							LatestPreparedCertificate: &types.PreparedCertificate{
								ProposalMessage: &types.MsgProposal{
									Metadata: &types.MsgMetadata{
										Sender: Alice,
										View:   &types.View{Sequence: 101, Round: 0},
									},

									BlockHash: []byte("keccak"),
								},
								PrepareMessages: []*types.MsgPrepare{
									{
										Metadata: &types.MsgMetadata{
											View:   &types.View{Sequence: 101, Round: 0},
											Sender: Chris,
										},

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

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seq := NewSequencer(tt.validator, tt.validatorSet, 0)
			assert.Equal(t, tt.isValid, seq.isValidMsgProposal(tt.msg, tt.quorum, mock.DummyKeccak("keccak")))
		})
	}
}

func TestIsValidMsgPrepare(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		validator        ibft.Validator
		validatorSet     ibft.ValidatorSet
		sigVerifier      ibft.SignatureVerifier
		msg              *types.MsgPrepare
		acceptedProposal *types.MsgProposal
		name             string
		isValid          bool
	}{
		{
			name: "invalid sender",
			validatorSet: mock.ValidatorSet{IsValidatorFn: func(_ []byte, _ uint64) bool {
				return false
			}},
			msg: &types.MsgPrepare{
				Metadata: &types.MsgMetadata{
					Sender: []byte("definitely not a validator"),
					View:   &types.View{Sequence: 101},
				},
			},
		},

		{
			name: "invalid block hash",
			validatorSet: mock.ValidatorSet{IsValidatorFn: func(_ []byte, _ uint64) bool {
				return true
			}},
			acceptedProposal: &types.MsgProposal{BlockHash: []byte("keccak")},
			msg: &types.MsgPrepare{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101},
					Sender: Chris,
				},
				BlockHash: []byte("definitely not keccak"),
			},
		},

		{
			name: "ok",
			validatorSet: mock.ValidatorSet{IsValidatorFn: func(_ []byte, _ uint64) bool {
				return true
			}},
			isValid:          true,
			acceptedProposal: &types.MsgProposal{BlockHash: []byte("keccak")},
			msg: &types.MsgPrepare{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101},
					Sender: Chris,
				},

				BlockHash: []byte("keccak"),
			},
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			seq := NewSequencer(tt.validator, tt.validatorSet, 0)
			seq.state.proposal = tt.acceptedProposal

			assert.Equal(t, tt.isValid, seq.isValidMsgPrepare(tt.msg))
		})
	}
}

func TestIsValidMsgCommit(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		validator        ibft.Validator
		msg              *types.MsgCommit
		acceptedProposal *types.MsgProposal
		name             string
		isValid          bool
	}{
		{
			name: "invalid sender",
			validator: mock.Validator{
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
				},
			},
			msg: &types.MsgCommit{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101},
					Sender: []byte("definitely not a validator"),
				},
			},
		},

		{
			name:             "invalid block hash",
			acceptedProposal: &types.MsgProposal{BlockHash: []byte("keccak")},
			validator: mock.Validator{
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
				},
			},
			msg: &types.MsgCommit{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101},
					Sender: Chris,
				},

				BlockHash: []byte("definitely not keccak"),
			},
		},

		{
			name:             "invalid commit seal",
			acceptedProposal: &types.MsgProposal{BlockHash: []byte("keccak")},
			validator: mock.Validator{
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
					IsValidSignatureFn: func(_, _, seal []byte) bool {
						return bytes.Equal(seal, []byte("commit seal"))
					},
				},
			},
			msg: &types.MsgCommit{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101},
					Sender: Chris,
				},

				BlockHash:  []byte("keccak"),
				CommitSeal: []byte("invalid commit seal"),
			},
		},

		{
			name:             "valid message",
			acceptedProposal: &types.MsgProposal{BlockHash: []byte("keccak")},
			validator: mock.Validator{
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
					IsValidSignatureFn: func(_, _, seal []byte) bool {
						return bytes.Equal(seal, []byte("commit seal"))
					},
				},
			},
			isValid: true,
			msg: &types.MsgCommit{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101},
					Sender: Chris,
				},

				BlockHash:  []byte("keccak"),
				CommitSeal: []byte("commit seal"),
			},
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := NewSequencer(tt.validator, 0)
			s.state.proposal = tt.acceptedProposal

			assert.Equal(t, tt.isValid, s.isValidMsgCommit(tt.msg))
		})
	}
}

func TestIsValidMsgRoundChange(t *testing.T) {
	t.Parallel()

	testTable := []struct {
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
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
				},
			},
			msg: &types.MsgRoundChange{
				Metadata: &types.MsgMetadata{
					Sender: []byte("definitely not a validator"),
					View:   &types.View{Sequence: 101},
				},
			},
		},

		{
			name: "valid message (pb and pc are nil)",
			validator: mock.Validator{
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
				},
			},
			isValid: true,
			msg: &types.MsgRoundChange{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101},
					Sender: Chris,
				},
			},
		},

		{
			name: "pb and pc are not both present",
			validator: mock.Validator{
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
				},
			},
			msg: &types.MsgRoundChange{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101},
					Sender: Chris,
				},
				LatestPreparedCertificate: &types.PreparedCertificate{},
			},
		},

		{
			name: "(invalid pc) proposal message and prepare messages are not both present",
			validator: mock.Validator{
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
				},
			},
			msg: &types.MsgRoundChange{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101},
					Sender: Chris,
				},
				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: nil,
				},
			},
		},

		{
			name: "(invalid pc) invalid sequence in proposal message",
			validator: mock.Validator{
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
				},
			},
			msg: &types.MsgRoundChange{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101},
					Sender: Chris,
				},

				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						Metadata: &types.MsgMetadata{
							View: &types.View{Sequence: 100},
						},
					},
					PrepareMessages: []*types.MsgPrepare{},
				},
			},
		},

		{
			name: "(invalid pc) invalid round in proposal message",
			validator: mock.Validator{
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
				},
			},
			msg: &types.MsgRoundChange{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101},
					Sender: Chris,
				},

				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						Metadata: &types.MsgMetadata{
							View: &types.View{Sequence: 101, Round: 5},
						},
					},
					PrepareMessages: []*types.MsgPrepare{},
				},
			},
		},

		{
			name: "(invalid pc) invalid proposer in proposal message",
			validator: mock.Validator{
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: Bob, Round: 0}),
				},
			},
			msg: &types.MsgRoundChange{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101, Round: 1},
					Sender: Chris,
				},

				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						Metadata: &types.MsgMetadata{
							Sender: mock.NewValidatorID("dani"),
							View:   &types.View{Sequence: 101, Round: 0},
						},
					},
					PrepareMessages: []*types.MsgPrepare{},
				},
			},
		},

		{
			name: "(invalid pc) proposal and prepare sequence mismatch",
			validator: mock.Validator{
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: Bob, Round: 1}),
				},
			},
			msg: &types.MsgRoundChange{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101, Round: 2},
					Sender: Chris,
				},

				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						Metadata: &types.MsgMetadata{
							Sender: Bob,
							View:   &types.View{Sequence: 101, Round: 1},
						},
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							Metadata: &types.MsgMetadata{
								View: &types.View{Sequence: 99},
							},
						},
					},
				},
			},
		},

		{
			name: "(invalid pc) proposal and prepare round mismatch",
			validator: mock.Validator{
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: Bob, Round: 1}),
				},
			},
			msg: &types.MsgRoundChange{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101, Round: 2},
					Sender: Chris,
				},

				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						Metadata: &types.MsgMetadata{
							Sender: Bob,
							View:   &types.View{Sequence: 101, Round: 1},
						},
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							Metadata: &types.MsgMetadata{
								View: &types.View{Sequence: 101, Round: 0},
							},
						},
					},
				},
			},
		},

		{
			name: "(invalid pc) invalid block hash in proposal and prepare",
			validator: mock.Validator{
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: Bob, Round: 1}),
				},
			},
			msg: &types.MsgRoundChange{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101, Round: 2},
					Sender: Chris,
				},

				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						Metadata: &types.MsgMetadata{
							Sender: Bob,
							View:   &types.View{Sequence: 101, Round: 1},
						},

						BlockHash: []byte("keccak"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							Metadata: &types.MsgMetadata{
								View: &types.View{Sequence: 101, Round: 1},
							},

							BlockHash: []byte("some other keccak"),
						},
					},
				},
			},
		},

		{
			name: "(invalid pc) invalid sender in prepare message",
			validator: mock.Validator{
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: Bob, Round: 1}),
				},
			},
			msg: &types.MsgRoundChange{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101, Round: 2},
					Sender: Chris,
				},

				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						Metadata: &types.MsgMetadata{
							Sender: Bob,
							View:   &types.View{Sequence: 101, Round: 1},
						},

						BlockHash: []byte("keccak"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							Metadata: &types.MsgMetadata{
								Sender: []byte("definitely not a validator"),
								View:   &types.View{Sequence: 101, Round: 1},
							},
							BlockHash: []byte("keccak"),
						},
					},
				},
			},
		},

		{
			name: "(invalid pc) duplicate sender in prepare messages",
			validator: mock.Validator{
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: Bob, Round: 1}),
				},
			},
			msg: &types.MsgRoundChange{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101, Round: 2},
					Sender: Chris,
				},

				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						Metadata: &types.MsgMetadata{
							Sender: Bob,
							View:   &types.View{Sequence: 101, Round: 1},
						},

						BlockHash: []byte("keccak"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							Metadata: &types.MsgMetadata{
								Sender: Chris,
								View:   &types.View{Sequence: 101, Round: 1},
							},

							BlockHash: []byte("keccak"),
						},

						{
							Metadata: &types.MsgMetadata{
								Sender: Chris,
								View:   &types.View{Sequence: 101, Round: 1},
							},

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
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: Bob, Round: 1}),
				},
			},
			msg: &types.MsgRoundChange{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101, Round: 2},
					Sender: Chris,
				},

				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						Metadata: &types.MsgMetadata{
							Sender: Bob,
							View:   &types.View{Sequence: 101, Round: 1},
						},

						BlockHash: []byte("keccak"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							Metadata: &types.MsgMetadata{
								Sender: Chris,
								View:   &types.View{Sequence: 101, Round: 1},
							},

							BlockHash: []byte("keccak"),
						},
					},
				},
			},
		},

		{
			name:   "latest ppb hash does not match proposal block hash",
			quorum: mock.NonZeroQuorum,
			validator: mock.Validator{
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: Bob, Round: 1}),
				},
			},
			msg: &types.MsgRoundChange{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101, Round: 2},
					Sender: Chris,
				},

				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						Metadata: &types.MsgMetadata{
							Sender: Bob,
							View:   &types.View{Sequence: 101, Round: 1},
						},

						BlockHash: []byte("some other keccak"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							Metadata: &types.MsgMetadata{
								Sender: Chris,
								View:   &types.View{Sequence: 101, Round: 1},
							},

							BlockHash: []byte("some other keccak"),
						},
					},
				},
			},
		},

		{
			name:   "valid round change ibftMsg",
			quorum: mock.NonZeroQuorum,
			validator: mock.Validator{
				SigVerifier: mock.SigVerifier{
					IsValidatorFn: ValidatorSet.IsValidator,
					IsProposerFn: mock.ProposersInRounds(
						mock.Proposer{ID: Bob, Round: 1}),
				},
			},
			isValid: true,
			msg: &types.MsgRoundChange{
				Metadata: &types.MsgMetadata{
					View:   &types.View{Sequence: 101, Round: 2},
					Sender: Chris,
				},

				LatestPreparedProposedBlock: &types.ProposedBlock{},
				LatestPreparedCertificate: &types.PreparedCertificate{
					ProposalMessage: &types.MsgProposal{
						Metadata: &types.MsgMetadata{
							Sender: Bob,
							View:   &types.View{Sequence: 101, Round: 1},
						},

						BlockHash: []byte("keccak"),
					},
					PrepareMessages: []*types.MsgPrepare{
						{
							Metadata: &types.MsgMetadata{
								Sender: Chris,
								View:   &types.View{Sequence: 101, Round: 1},
							},

							BlockHash: []byte("keccak"),
						},
					},
				},
			},
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := NewSequencer(tt.validator, 0)
			assert.Equal(t, tt.isValid, s.isValidMsgRoundChange(tt.msg, tt.quorum, mock.DummyKeccak("keccak")))
		})
	}
}
