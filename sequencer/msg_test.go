package sequencer

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sig-0/go-ibft/message"
)

var (
	alice = []byte("alice")
	bob   = []byte("bob")
	chris = []byte("chris")
)

func Test_IsValidMsgProposal(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name      string
		sequencer *Sequencer
		msg       *message.MsgProposal
		expected  bool
	}{
		{
			name:      "proposed block round and current round do not match",
			sequencer: &Sequencer{validator: MockValidator{address: alice}},
			msg: &message.MsgProposal{
				Info:          &message.MsgInfo{View: &message.View{Round: 0}},
				ProposedBlock: &message.ProposedBlock{Round: 5},
			},
		},

		{
			name:      "cannot verify own proposal",
			sequencer: &Sequencer{validator: MockValidator{address: alice}},
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					Sender: alice,
					View:   &message.View{Sequence: 101, Round: 0},
				},
				ProposedBlock: &message.ProposedBlock{Round: 0},
			},
		},

		{
			name: "bad proposer",
			sequencer: &Sequencer{
				validator: MockValidator{address: alice},
				validatorSet: MockValidatorSet{isProposerFn: func(from []byte, _ uint64, round uint64) bool {
					return bytes.Equal(from, bob) && round == 0
				}},
			},
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101, Round: 0},
					Sender: []byte("definitely not bob"),
				},
				ProposedBlock: &message.ProposedBlock{
					Block: []byte("block"),
					Round: 0,
				},
			},
		},

		{
			name: "invalid block hash",
			sequencer: &Sequencer{
				keccak:    MockKeccak(func(_ []byte) []byte { return []byte("keccak") }),
				validator: MockValidator{address: alice},
				validatorSet: MockValidatorSet{isProposerFn: func(from []byte, _ uint64, round uint64) bool {
					return bytes.Equal(from, bob) && round == 0
				}},
			},
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101, Round: 0},
					Sender: bob,
				},
				ProposedBlock: &message.ProposedBlock{Block: []byte("block"), Round: 0},
				BlockHash:     []byte("definitely not keccak"),
			},
		},

		{
			name: "invalid round 0 proposal",
			sequencer: &Sequencer{
				keccak:    MockKeccak(func(_ []byte) []byte { return []byte("keccak") }),
				validator: MockValidator{address: alice},
				validatorSet: MockValidatorSet{isProposerFn: func(from []byte, _ uint64, round uint64) bool {
					return bytes.Equal(from, bob) && round == 0
				}},
			},
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101, Round: 0},
					Sender: bob,
				},
				ProposedBlock: &message.ProposedBlock{Block: []byte("invalid block"), Round: 0},
				BlockHash:     []byte("bad keccak"),
			},
		},

		{
			name:     "valid round 0 proposal",
			expected: true,
			sequencer: &Sequencer{
				keccak: MockKeccak(func(_ []byte) []byte { return []byte("keccak") }),
				validator: MockValidator{
					address: alice,
					isValidProposalFn: func(_ uint64, block []byte) bool {
						return bytes.Equal(block, []byte("block"))
					},
				},
				validatorSet: MockValidatorSet{isProposerFn: func(from []byte, _ uint64, round uint64) bool {
					return bytes.Equal(from, bob) && round == 0
				}},
			},
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					Sender: bob,
					View:   &message.View{Sequence: 101, Round: 0},
				},
				ProposedBlock: &message.ProposedBlock{Block: []byte("block"), Round: 0},
				BlockHash:     []byte("keccak"),
			},
		},

		{
			name: "(non zero round): empty rcc",
			sequencer: &Sequencer{
				keccak: MockKeccak(func(_ []byte) []byte { return []byte("keccak") }),
				validator: MockValidator{
					address: alice,
					isValidProposalFn: func(_ uint64, block []byte) bool {
						return bytes.Equal(block, []byte("block"))
					},
				},
				validatorSet: MockValidatorSet{isProposerFn: func(from []byte, _ uint64, round uint64) bool {
					return bytes.Equal(from, bob) && round == 0
				}},
			},
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					Sender: bob,
					View:   &message.View{Sequence: 101, Round: 1},
				},
				ProposedBlock:          &message.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:              []byte("keccak"),
				RoundChangeCertificate: nil,
			},
		},

		{
			name: "(non zero round): bad round change message sequence in rcc",
			sequencer: &Sequencer{
				keccak: MockKeccak(func(_ []byte) []byte { return []byte("keccak") }),
				validator: MockValidator{
					address: alice,
					isValidProposalFn: func(_ uint64, block []byte) bool {
						return bytes.Equal(block, []byte("block"))
					},
				},
				validatorSet: MockValidatorSet{isProposerFn: func(from []byte, _ uint64, round uint64) bool {
					return bytes.Equal(from, bob) && round == 0
				}},
			},
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					Sender: bob,
					View:   &message.View{Sequence: 101, Round: 1},
				},
				ProposedBlock: &message.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &message.RoundChangeCertificate{
					Messages: []*message.MsgRoundChange{
						{
							Info: &message.MsgInfo{View: &message.View{Sequence: 99}},
						},
					},
				},
			},
		},

		{
			name: "(non zero round): bad round change message round in rcc",
			sequencer: &Sequencer{
				keccak: MockKeccak(func(_ []byte) []byte { return []byte("keccak") }),
				validator: MockValidator{
					address: alice,
					isValidProposalFn: func(_ uint64, block []byte) bool {
						return bytes.Equal(block, []byte("block"))
					},
				},
				validatorSet: MockValidatorSet{isProposerFn: func(from []byte, _ uint64, round uint64) bool {
					return bytes.Equal(from, bob) && round == 0
				}},
			},
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					Sender: bob,
					View:   &message.View{Sequence: 101, Round: 1},
				},
				ProposedBlock: &message.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &message.RoundChangeCertificate{
					Messages: []*message.MsgRoundChange{
						{
							Info: &message.MsgInfo{View: &message.View{Sequence: 101, Round: 0}},
						},
					},
				},
			},
		},

		{
			name: "(non zero round): bad validator in rcc",
			sequencer: &Sequencer{
				keccak: MockKeccak(func(_ []byte) []byte { return []byte("keccak") }),
				validator: MockValidator{
					address: alice,
					isValidProposalFn: func(_ uint64, block []byte) bool {
						return bytes.Equal(block, []byte("block"))
					},
				},
				validatorSet: MockValidatorSet{
					isValidatorFn: func(from []byte, _ uint64) bool {
						return !bytes.Equal(from, []byte("definitely not a validator"))
					},
					isProposerFn: func(from []byte, _ uint64, round uint64) bool {
						return bytes.Equal(from, bob) && round == 0
					},
				},
			},
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					Sender: bob,
					View:   &message.View{Sequence: 101, Round: 1},
				},
				ProposedBlock: &message.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &message.RoundChangeCertificate{
					Messages: []*message.MsgRoundChange{
						{
							Info: &message.MsgInfo{
								Sender: []byte("definitely not a validator"),
								View:   &message.View{Sequence: 101, Round: 1},
							},
						},
					},
				},
			},
		},

		{
			name: "(non zero round): duplicate sender in rcc",
			sequencer: &Sequencer{
				keccak: MockKeccak(func(_ []byte) []byte { return []byte("keccak") }),
				validator: MockValidator{
					address: alice,
					isValidProposalFn: func(_ uint64, block []byte) bool {
						return bytes.Equal(block, []byte("block"))
					},
				},
				validatorSet: MockValidatorSet{
					isValidatorFn: func(from []byte, _ uint64) bool {
						return !bytes.Equal(from, []byte("definitely not a validator"))
					},
					isProposerFn: func(from []byte, _ uint64, round uint64) bool {
						return bytes.Equal(from, bob) && round == 0
					},
				},
			},
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					Sender: bob,
					View:   &message.View{Sequence: 101, Round: 1},
				},
				ProposedBlock: &message.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &message.RoundChangeCertificate{
					Messages: []*message.MsgRoundChange{
						{
							Info: &message.MsgInfo{
								Sender: chris,
								View:   &message.View{Sequence: 101, Round: 1},
							},
						},

						{
							Info: &message.MsgInfo{
								Sender: chris,
								View:   &message.View{Sequence: 101, Round: 1},
							},
						},
					},
				},
			},
		},

		{
			name: "(non zero round): no quorum in rcc",
			sequencer: &Sequencer{
				keccak: MockKeccak(func(_ []byte) []byte { return []byte("keccak") }),
				validator: MockValidator{
					address: alice,
					isValidProposalFn: func(_ uint64, block []byte) bool {
						return bytes.Equal(block, []byte("block"))
					},
				},
				validatorSet: MockValidatorSet{
					isValidatorFn: func(from []byte, _ uint64) bool {
						return !bytes.Equal(from, []byte("definitely not a validator"))
					},
					isProposerFn: func(from []byte, _ uint64, round uint64) bool {
						return bytes.Equal(from, bob) && round == 0
					},
				},
			},
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					Sender: bob,
					View:   &message.View{Sequence: 101, Round: 1},
				},
				ProposedBlock: &message.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &message.RoundChangeCertificate{
					Messages: []*message.MsgRoundChange{
						{
							Info: &message.MsgInfo{
								Sender: chris,
								View:   &message.View{Sequence: 101, Round: 1},
							},
						},
					},
				},
			},
		},

		{
			name: "(non zero round): invalid block in rcc",
			sequencer: &Sequencer{
				keccak: MockKeccak(func(_ []byte) []byte { return []byte("keccak") }),
				validator: MockValidator{
					address: alice,
					isValidProposalFn: func(_ uint64, block []byte) bool {
						return bytes.Equal(block, []byte("block"))
					},
				},
				validatorSet: MockValidatorSet{
					isValidatorFn: func(from []byte, _ uint64) bool {
						return !bytes.Equal(from, []byte("definitely not a validator"))
					},
					isProposerFn: func(from []byte, _ uint64, round uint64) bool {
						return bytes.Equal(from, bob) && round == 0
					},
				},
			},

			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					Sender: bob,
					View:   &message.View{Sequence: 101, Round: 1},
				},
				ProposedBlock: &message.ProposedBlock{Block: []byte("invalid block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &message.RoundChangeCertificate{
					Messages: []*message.MsgRoundChange{
						{

							Info: &message.MsgInfo{
								Sender: chris,
								View:   &message.View{Sequence: 101, Round: 1},
							},
						},
					},
				},
			},
		},

		{
			name: "(non zero round): highest round block hash does not match derived hash",
			sequencer: &Sequencer{
				keccak: MockKeccak(func(_ []byte) []byte { return []byte("keccak") }),
				validator: MockValidator{
					address: alice,
					isValidProposalFn: func(_ uint64, block []byte) bool {
						return bytes.Equal(block, []byte("block"))
					},
				},
				validatorSet: MockValidatorSet{
					isValidatorFn: func(from []byte, _ uint64) bool {
						return bytes.Equal(from, []byte(alice)) || bytes.Equal(from, chris)
					},
					isProposerFn: func(id []byte, _ uint64, round uint64) bool {
						switch round {
						case 0:
							return bytes.Equal(id, []byte(alice))
						case 1:
							return bytes.Equal(id, bob)
						default:
							return false
						}
					},
					hasQuorumFn: func(_ []message.Message) bool {
						return false
					},
				},
			},
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					Sender: bob,
					View:   &message.View{Sequence: 101, Round: 1},
				},
				ProposedBlock: &message.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &message.RoundChangeCertificate{
					Messages: []*message.MsgRoundChange{
						{
							Info: &message.MsgInfo{
								Sender: chris,
								View:   &message.View{Sequence: 101, Round: 1},
							},
							LatestPreparedCertificate: &message.PreparedCertificate{
								ProposalMessage: &message.MsgProposal{
									Info: &message.MsgInfo{
										Sender: []byte(alice),
										View:   &message.View{Sequence: 101, Round: 0},
									},
									BlockHash: []byte("invalid keccak"),
								},
								PrepareMessages: []*message.MsgPrepare{
									{
										Info: &message.MsgInfo{
											View:   &message.View{Sequence: 101, Round: 0},
											Sender: chris,
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
			name:     "(non zero round): ok",
			expected: true,
			sequencer: &Sequencer{
				keccak: MockKeccak(func(_ []byte) []byte { return []byte("keccak") }),
				validator: MockValidator{
					address: alice,
					isValidProposalFn: func(_ uint64, block []byte) bool {
						return bytes.Equal(block, []byte("block"))
					},
				},
				validatorSet: MockValidatorSet{
					isValidatorFn: func(from []byte, _ uint64) bool {
						return bytes.Equal(from, []byte(alice)) || bytes.Equal(from, chris)
					},
					isProposerFn: func(id []byte, _ uint64, round uint64) bool {
						switch round {
						case 0:
							return bytes.Equal(id, []byte(alice))
						case 1:
							return bytes.Equal(id, bob)
						default:
							return false
						}
					},
					hasQuorumFn: func(_ []message.Message) bool {
						return true
					},
				},
			},

			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					Sender: bob,
					View:   &message.View{Sequence: 101, Round: 1},
				},
				ProposedBlock: &message.ProposedBlock{Block: []byte("block"), Round: 1},
				BlockHash:     []byte("keccak"),
				RoundChangeCertificate: &message.RoundChangeCertificate{
					Messages: []*message.MsgRoundChange{
						{
							Info: &message.MsgInfo{
								Sender: chris,
								View:   &message.View{Sequence: 101, Round: 1},
							},
							LatestPreparedProposedBlock: &message.ProposedBlock{},
							LatestPreparedCertificate: &message.PreparedCertificate{
								ProposalMessage: &message.MsgProposal{
									Info: &message.MsgInfo{
										Sender: []byte(alice),
										View:   &message.View{Sequence: 101, Round: 0},
									},

									BlockHash: []byte("keccak"),
								},
								PrepareMessages: []*message.MsgPrepare{
									{
										Info: &message.MsgInfo{
											View:   &message.View{Sequence: 101, Round: 0},
											Sender: chris,
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

			assert.Equal(t, tt.expected, tt.sequencer.isValidMsgProposal(tt.msg))
		})
	}
}

func Test_IsValidMsgPrepare(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		msg       *message.MsgPrepare
		sequencer *Sequencer
		name      string
		expected  bool
	}{
		{
			name: "invalid sender",
			sequencer: &Sequencer{validatorSet: MockValidatorSet{isValidatorFn: func(_ []byte, _ uint64) bool {
				return false
			}}},
			msg: &message.MsgPrepare{
				Info: &message.MsgInfo{
					Sender: []byte("definitely not a validator"),
					View:   &message.View{Sequence: 101},
				},
			},
		},

		{
			name: "invalid block hash",
			sequencer: &Sequencer{
				validatorSet: MockValidatorSet{isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
				}},
				state: state{proposal: &message.MsgProposal{
					BlockHash: []byte("keccak")},
				},
			},
			msg: &message.MsgPrepare{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101},
					Sender: chris,
				},
				BlockHash: []byte("definitely not keccak"),
			},
		},

		{
			name:     "ok",
			expected: true,
			sequencer: &Sequencer{
				validatorSet: MockValidatorSet{isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
				}},
				state: state{proposal: &message.MsgProposal{BlockHash: []byte("keccak")}},
			},
			msg: &message.MsgPrepare{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101},
					Sender: chris,
				},

				BlockHash: []byte("keccak"),
			},
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.sequencer.isValidMsgPrepare(tt.msg))
		})
	}
}

func Test_IsValidMsgCommit(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		msg       *message.MsgCommit
		sequencer *Sequencer
		name      string
		expected  bool
	}{
		{
			name: "invalid sender",
			sequencer: &Sequencer{validatorSet: MockValidatorSet{isValidatorFn: func(_ []byte, _ uint64) bool {
				return false
			}}},
			msg: &message.MsgCommit{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101},
					Sender: []byte("definitely not a validator"),
				},
			},
		},

		{
			name: "invalid block hash",
			sequencer: &Sequencer{
				validatorSet: MockValidatorSet{isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
				}},
				state: state{proposal: &message.MsgProposal{BlockHash: []byte("keccak")}},
			},
			msg: &message.MsgCommit{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101},
					Sender: chris,
				},
				BlockHash: []byte("definitely not keccak"),
			},
		},

		{
			name: "invalid commit seal",
			sequencer: &Sequencer{
				validatorSet: MockValidatorSet{isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
				}},
				state: state{proposal: &message.MsgProposal{BlockHash: []byte("keccak")}},
				sig: MockSignatureVerifier(func(_ []byte, _ []byte, _ []byte) error {
					return errors.New("bad sig")
				}),
			},
			msg: &message.MsgCommit{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101},
					Sender: chris,
				},

				BlockHash:  []byte("keccak"),
				CommitSeal: []byte("doesn't matter"),
			},
		},

		{
			name:     "ok",
			expected: true,
			sequencer: &Sequencer{
				validatorSet: MockValidatorSet{isValidatorFn: func(_ []byte, _ uint64) bool {
					return true
				}},
				state: state{proposal: &message.MsgProposal{BlockHash: []byte("keccak")}},
				sig: MockSignatureVerifier(func(_ []byte, _ []byte, _ []byte) error {
					return nil
				}),
			},
			msg: &message.MsgCommit{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101},
					Sender: chris,
				},

				BlockHash:  []byte("keccak"),
				CommitSeal: []byte("doesn't matter"),
			},
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.sequencer.isValidMsgCommit(tt.msg))
		})
	}
}

func TestIsValidMsgRoundChange(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		sequencer *Sequencer
		//validator Validator
		//quorum    Quorum
		//keccak    Keccak
		msg      *message.MsgRoundChange
		name     string
		expected bool
	}{
		{
			name: "invalid sender",
			sequencer: &Sequencer{validatorSet: MockValidatorSet{isValidatorFn: func(_ []byte, _ uint64) bool {
				return false
			}}},
			msg: &message.MsgRoundChange{
				Info: &message.MsgInfo{
					Sender: []byte("definitely not a validator"),
					View:   &message.View{Sequence: 101},
				},
			},
		},

		{
			name:     "valid message (pb and pc are nil)",
			expected: true,
			sequencer: &Sequencer{validatorSet: MockValidatorSet{isValidatorFn: func(_ []byte, _ uint64) bool {
				return true
			}}},
			msg: &message.MsgRoundChange{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101},
					Sender: chris,
				},
			},
		},

		{
			name: "pb and pc are not both present",
			sequencer: &Sequencer{validatorSet: MockValidatorSet{isValidatorFn: func(_ []byte, _ uint64) bool {
				return true
			}}},
			msg: &message.MsgRoundChange{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101},
					Sender: chris,
				},
				LatestPreparedCertificate: &message.PreparedCertificate{},
			},
		},

		{
			name: "(invalid pc) proposal message and prepare messages are not both present",
			sequencer: &Sequencer{validatorSet: MockValidatorSet{isValidatorFn: func(_ []byte, _ uint64) bool {
				return true
			}}},
			msg: &message.MsgRoundChange{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101},
					Sender: chris,
				},
				LatestPreparedProposedBlock: &message.ProposedBlock{},
				LatestPreparedCertificate: &message.PreparedCertificate{
					ProposalMessage: nil,
				},
			},
		},

		{
			name: "(invalid pc) invalid sequence in proposal message",
			sequencer: &Sequencer{validatorSet: MockValidatorSet{isValidatorFn: func(_ []byte, _ uint64) bool {
				return true
			}}},
			msg: &message.MsgRoundChange{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101},
					Sender: chris,
				},

				LatestPreparedProposedBlock: &message.ProposedBlock{},
				LatestPreparedCertificate: &message.PreparedCertificate{
					ProposalMessage: &message.MsgProposal{
						Info: &message.MsgInfo{
							View: &message.View{Sequence: 100},
						},
					},
					PrepareMessages: []*message.MsgPrepare{},
				},
			},
		},

		{
			name: "(invalid pc) invalid round in proposal message",
			sequencer: &Sequencer{validatorSet: MockValidatorSet{isValidatorFn: func(_ []byte, _ uint64) bool {
				return true
			}}},
			msg: &message.MsgRoundChange{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101},
					Sender: chris,
				},

				LatestPreparedProposedBlock: &message.ProposedBlock{},
				LatestPreparedCertificate: &message.PreparedCertificate{
					ProposalMessage: &message.MsgProposal{
						Info: &message.MsgInfo{
							View: &message.View{Sequence: 101, Round: 5},
						},
					},
					PrepareMessages: []*message.MsgPrepare{},
				},
			},
		},

		{
			name: "(invalid pc) invalid proposer in proposal message",
			sequencer: &Sequencer{validatorSet: MockValidatorSet{
				isValidatorFn: func(_ []byte, _ uint64) bool {
					return false
				},
				isProposerFn: func(_ []byte, _ uint64, _ uint64) bool {
					return false
				},
			}},
			msg: &message.MsgRoundChange{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101, Round: 1},
					Sender: chris,
				},

				LatestPreparedProposedBlock: &message.ProposedBlock{},
				LatestPreparedCertificate: &message.PreparedCertificate{
					ProposalMessage: &message.MsgProposal{
						Info: &message.MsgInfo{
							Sender: []byte("doesn't matter"),
							View:   &message.View{Sequence: 101, Round: 0},
						},
					},
					PrepareMessages: []*message.MsgPrepare{},
				},
			},
		},

		{
			name: "(invalid pc) proposal and prepare sequence mismatch",
			sequencer: &Sequencer{validatorSet: MockValidatorSet{
				isValidatorFn: func(_ []byte, _ uint64) bool { return true },
				isProposerFn:  func(_ []byte, _ uint64, _ uint64) bool { return true },
			}},
			msg: &message.MsgRoundChange{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101, Round: 2},
					Sender: chris,
				},

				LatestPreparedProposedBlock: &message.ProposedBlock{},
				LatestPreparedCertificate: &message.PreparedCertificate{
					ProposalMessage: &message.MsgProposal{
						Info: &message.MsgInfo{
							Sender: bob,
							View:   &message.View{Sequence: 101, Round: 1},
						},
					},
					PrepareMessages: []*message.MsgPrepare{
						{
							Info: &message.MsgInfo{
								View: &message.View{Sequence: 99},
							},
						},
					},
				},
			},
		},

		{
			name: "(invalid pc) proposal and prepare round mismatch",
			sequencer: &Sequencer{validatorSet: MockValidatorSet{
				isValidatorFn: func(_ []byte, _ uint64) bool { return true },
				isProposerFn:  func(_ []byte, _ uint64, _ uint64) bool { return true },
			}},
			msg: &message.MsgRoundChange{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101, Round: 2},
					Sender: chris,
				},

				LatestPreparedProposedBlock: &message.ProposedBlock{},
				LatestPreparedCertificate: &message.PreparedCertificate{
					ProposalMessage: &message.MsgProposal{
						Info: &message.MsgInfo{
							Sender: bob,
							View:   &message.View{Sequence: 101, Round: 1},
						},
					},
					PrepareMessages: []*message.MsgPrepare{
						{
							Info: &message.MsgInfo{
								View: &message.View{Sequence: 101, Round: 0},
							},
						},
					},
				},
			},
		},

		{
			name: "(invalid pc) invalid block hash in proposal and prepare",
			sequencer: &Sequencer{validatorSet: MockValidatorSet{
				isValidatorFn: func(_ []byte, _ uint64) bool { return true },
				isProposerFn:  func(_ []byte, _ uint64, _ uint64) bool { return true },
			}},
			msg: &message.MsgRoundChange{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101, Round: 2},
					Sender: chris,
				},

				LatestPreparedProposedBlock: &message.ProposedBlock{},
				LatestPreparedCertificate: &message.PreparedCertificate{
					ProposalMessage: &message.MsgProposal{
						Info: &message.MsgInfo{
							Sender: bob,
							View:   &message.View{Sequence: 101, Round: 1},
						},

						BlockHash: []byte("keccak"),
					},
					PrepareMessages: []*message.MsgPrepare{
						{
							Info: &message.MsgInfo{
								View: &message.View{Sequence: 101, Round: 1},
							},

							BlockHash: []byte("some other keccak"),
						},
					},
				},
			},
		},

		{
			name: "(invalid pc) invalid sender in prepare message",
			sequencer: &Sequencer{validatorSet: MockValidatorSet{
				isValidatorFn: func(_ []byte, _ uint64) bool { return true },
				isProposerFn: func(from []byte, _ uint64, _ uint64) bool {
					return bytes.Equal(from, chris)
				},
			}},
			msg: &message.MsgRoundChange{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101, Round: 2},
					Sender: chris,
				},

				LatestPreparedProposedBlock: &message.ProposedBlock{},
				LatestPreparedCertificate: &message.PreparedCertificate{
					ProposalMessage: &message.MsgProposal{
						Info: &message.MsgInfo{
							Sender: bob,
							View:   &message.View{Sequence: 101, Round: 1},
						},

						BlockHash: []byte("keccak"),
					},
					PrepareMessages: []*message.MsgPrepare{
						{
							Info: &message.MsgInfo{
								Sender: []byte("definitely not a validator"),
								View:   &message.View{Sequence: 101, Round: 1},
							},
							BlockHash: []byte("keccak"),
						},
					},
				},
			},
		},

		{
			name: "(invalid pc) duplicate sender in prepare messages",
			sequencer: &Sequencer{validatorSet: MockValidatorSet{
				isValidatorFn: func(_ []byte, _ uint64) bool { return true },
				isProposerFn:  func(_ []byte, _ uint64, _ uint64) bool { return true },
			}},
			msg: &message.MsgRoundChange{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101, Round: 2},
					Sender: chris,
				},

				LatestPreparedProposedBlock: &message.ProposedBlock{},
				LatestPreparedCertificate: &message.PreparedCertificate{
					ProposalMessage: &message.MsgProposal{
						Info: &message.MsgInfo{
							Sender: bob,
							View:   &message.View{Sequence: 101, Round: 1},
						},

						BlockHash: []byte("keccak"),
					},
					PrepareMessages: []*message.MsgPrepare{
						{
							Info: &message.MsgInfo{
								Sender: chris,
								View:   &message.View{Sequence: 101, Round: 1},
							},

							BlockHash: []byte("keccak"),
						},

						{
							Info: &message.MsgInfo{
								Sender: chris,
								View:   &message.View{Sequence: 101, Round: 1},
							},

							BlockHash: []byte("keccak"),
						},
					},
				},
			},
		},

		{
			name: "(invalid pc) no quorum messages",
			sequencer: &Sequencer{validatorSet: MockValidatorSet{
				isValidatorFn: func(_ []byte, _ uint64) bool { return true },
				isProposerFn:  func(_ []byte, _ uint64, _ uint64) bool { return true },
				hasQuorumFn: func(_ []message.Message) bool {
					return false
				},
			}},
			msg: &message.MsgRoundChange{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101, Round: 2},
					Sender: chris,
				},

				LatestPreparedProposedBlock: &message.ProposedBlock{},
				LatestPreparedCertificate: &message.PreparedCertificate{
					ProposalMessage: &message.MsgProposal{
						Info: &message.MsgInfo{
							Sender: bob,
							View:   &message.View{Sequence: 101, Round: 1},
						},

						BlockHash: []byte("keccak"),
					},
					PrepareMessages: []*message.MsgPrepare{
						{
							Info: &message.MsgInfo{
								Sender: chris,
								View:   &message.View{Sequence: 101, Round: 1},
							},

							BlockHash: []byte("keccak"),
						},
					},
				},
			},
		},

		{
			name: "latest ppb hash does not match proposal block hash",
			sequencer: &Sequencer{
				keccak: MockKeccak(func(_ []byte) []byte {
					return []byte("keccak")
				}),
				validatorSet: MockValidatorSet{
					isValidatorFn: func(_ []byte, _ uint64) bool { return true },
					isProposerFn:  func(_ []byte, _ uint64, _ uint64) bool { return true },
					hasQuorumFn: func(_ []message.Message) bool {
						return true
					},
				},
			},
			msg: &message.MsgRoundChange{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101, Round: 2},
					Sender: chris,
				},

				LatestPreparedProposedBlock: &message.ProposedBlock{},
				LatestPreparedCertificate: &message.PreparedCertificate{
					ProposalMessage: &message.MsgProposal{
						Info: &message.MsgInfo{
							Sender: bob,
							View:   &message.View{Sequence: 101, Round: 1},
						},

						BlockHash: []byte("some other keccak"),
					},
					PrepareMessages: []*message.MsgPrepare{
						{
							Info: &message.MsgInfo{
								Sender: chris,
								View:   &message.View{Sequence: 101, Round: 1},
							},

							BlockHash: []byte("some other keccak"),
						},
					},
				},
			},
		},

		{
			name:     "ok",
			expected: true,
			sequencer: &Sequencer{
				keccak: MockKeccak(func(_ []byte) []byte {
					return []byte("keccak")
				}),
				validatorSet: MockValidatorSet{
					isValidatorFn: func(_ []byte, _ uint64) bool { return true },
					isProposerFn:  func(_ []byte, _ uint64, _ uint64) bool { return true },
					hasQuorumFn:   func(_ []message.Message) bool { return true },
				},
			},
			msg: &message.MsgRoundChange{
				Info: &message.MsgInfo{
					View:   &message.View{Sequence: 101, Round: 2},
					Sender: chris,
				},

				LatestPreparedProposedBlock: &message.ProposedBlock{},
				LatestPreparedCertificate: &message.PreparedCertificate{
					ProposalMessage: &message.MsgProposal{
						Info: &message.MsgInfo{
							Sender: bob,
							View:   &message.View{Sequence: 101, Round: 1},
						},

						BlockHash: []byte("keccak"),
					},
					PrepareMessages: []*message.MsgPrepare{
						{
							Info: &message.MsgInfo{
								Sender: chris,
								View:   &message.View{Sequence: 101, Round: 1},
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

			assert.Equal(t, tt.expected, tt.sequencer.isValidMsgRoundChange(tt.msg))
		})
	}
}
