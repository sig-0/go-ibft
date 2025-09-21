package consensus

//
//import (
//	"bytes"
//	"errors"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/sig-0/go-ibft/message"
//)
//
//func Test_IsValidMsgProposal_Round0(t *testing.T) {
//	t.Parallel()
//
//	testTable := []struct {
//		sequencer *Sequencer
//		msg       *message.Proposal
//		name      string
//		expected  bool
//	}{
//		{
//			name:      "proposed block round and current round do not match",
//			sequencer: &Sequencer{validator: mockValidator{address: Alice}},
//			msg: &message.Proposal{
//				Round:         0,
//				ProposedBlock: &message.ProposedBlock{Round: 5},
//			},
//		},
//
//		{
//			name:      "cannot verify own proposal",
//			sequencer: &Sequencer{validator: mockValidator{address: Alice}},
//			msg: &message.Proposal{
//				Sender:        Alice,
//				Sequence:      101,
//				Round:         0,
//				ProposedBlock: &message.ProposedBlock{Round: 0},
//			},
//		},
//
//		{
//			name: "bad proposer",
//			sequencer: &Sequencer{
//				validator: mockValidator{address: Alice},
//				verifier: mockVerifier{isProposerFn: func(v []byte, _ uint64, round uint64) bool {
//					return bytes.Equal(v, Bob) && round == 0
//				}},
//			},
//			msg: &message.Proposal{
//				Sender:   []byte("definitely not Bob"),
//				Sequence: 101,
//				Round:    0,
//				ProposedBlock: &message.ProposedBlock{
//					Block: []byte("block"),
//					Round: 0,
//				},
//			},
//		},
//
//		{
//			name: "invalid block hash",
//			sequencer: &Sequencer{
//				keccak:    DummyKeccak,
//				validator: mockValidator{address: Alice},
//				verifier: mockVerifier{isProposerFn: func(v []byte, _, round uint64) bool {
//					return bytes.Equal(v, Bob) && round == 0
//				}},
//			},
//			msg: &message.Proposal{
//				Sender:   Bob,
//				Sequence: 101,
//				Round:    0,
//				ProposedBlock: &message.ProposedBlock{
//					Block: []byte("block"),
//					Round: 0,
//				},
//				BlockHash: []byte("definitely not keccak"),
//			},
//		},
//
//		{
//			name: "invalid round 0 proposal",
//			sequencer: &Sequencer{
//				keccak: DummyKeccak,
//				validator: mockValidator{
//					address: Alice,
//				},
//				verifier: mockVerifier{isProposerFn: func(v []byte, _ uint64, round uint64) bool {
//					return bytes.Equal(v, Bob) && round == 0
//				},
//					isValidProposalFn: func(_ []byte, _ uint64) bool {
//						return false
//					}},
//			},
//			msg: &message.Proposal{
//				Sender:   Bob,
//				Sequence: 101,
//				Round:    0,
//				ProposedBlock: &message.ProposedBlock{
//					Block: []byte("doesn't really matter"),
//					Round: 0,
//				},
//				BlockHash: DummyKeccakValue,
//			},
//		},
//
//		{
//			name:     "ok",
//			expected: true,
//			sequencer: &Sequencer{
//				keccak: DummyKeccak,
//				validator: mockValidator{
//					address: Alice,
//				},
//				verifier: mockVerifier{
//					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
//						return bytes.Equal(v, Bob) && round == 0
//					},
//					isValidProposalFn: AlwaysValidProposal,
//				},
//			},
//			msg: &message.Proposal{
//				Sender:   Bob,
//				Sequence: 101,
//				Round:    0,
//				ProposedBlock: &message.ProposedBlock{
//					Block: []byte("block"),
//					Round: 0,
//				},
//				BlockHash: DummyKeccakValue,
//			},
//		},
//	}
//
//	for _, tt := range testTable {
//		t.Run(tt.name, func(t *testing.T) {
//			t.Parallel()
//
//			assert.Equal(t, tt.expected, tt.sequencer.isValidMsgProposal(tt.msg))
//		})
//	}
//}
//
//func Test_IsValidMsgProposal_Higher_Rounds(t *testing.T) {
//	t.Parallel()
//
//	testTable := []struct {
//		sequencer *Sequencer
//		msg       *message.Proposal
//		name      string
//		expected  bool
//	}{
//		{
//			name: "missing rcc",
//			sequencer: &Sequencer{
//				keccak: DummyKeccak,
//				validator: mockValidator{
//					address: Alice,
//				},
//				verifier: mockVerifier{
//					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
//						return bytes.Equal(v, Bob) && round == 1
//					},
//					isValidProposalFn: AlwaysValidProposal,
//				},
//			},
//			msg: &message.Proposal{
//				Sender:   Bob,
//				Sequence: 101,
//				Round:    1,
//				ProposedBlock: &message.ProposedBlock{
//					Block: []byte("block"),
//					Round: 1,
//				},
//				BlockHash:              DummyKeccakValue,
//				RoundChangeCertificate: nil,
//			},
//		},
//
//		{
//			name: "rcc msg sequence (round) does not match proposal msg sequence (round)",
//			sequencer: &Sequencer{
//				keccak: DummyKeccak,
//				validator: mockValidator{
//					address: Alice,
//				},
//				verifier: mockVerifier{isProposerFn: func(v []byte, _ uint64, round uint64) bool {
//					return bytes.Equal(v, Bob) && round == 1
//				},
//					isValidProposalFn: AlwaysValidProposal,
//				},
//			},
//			msg: &message.Proposal{
//				Sender:   Bob,
//				Sequence: 101,
//				Round:    1,
//				ProposedBlock: &message.ProposedBlock{
//					Block: []byte("block"),
//					Round: 1,
//				},
//				BlockHash: DummyKeccakValue,
//				RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
//					{
//						Sequence: 202,
//						Round:    10101,
//					},
//				}},
//			},
//		},
//
//		{
//			name: "bad sender in rcc msg",
//			sequencer: &Sequencer{
//				keccak: DummyKeccak,
//				validator: mockValidator{
//					address: Alice,
//				},
//				verifier: mockVerifier{
//					isValidatorFn: func(v []byte, _ uint64) bool {
//						return !bytes.Equal(v, []byte("definitely not a validator"))
//					},
//					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
//						return bytes.Equal(v, Bob) && round == 1
//					},
//					isValidProposalFn: AlwaysValidProposal,
//				},
//			},
//			msg: &message.Proposal{
//				Sender:   Bob,
//				Sequence: 101,
//				Round:    1,
//				ProposedBlock: &message.ProposedBlock{
//					Block: []byte("block"),
//					Round: 1,
//				},
//				BlockHash: DummyKeccakValue,
//				RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
//					{
//						Sender:   []byte("definitely not a validator"),
//						Sequence: 101,
//						Round:    1,
//					},
//				}},
//			},
//		},
//
//		{
//			name: "duplicate sender in rcc",
//			sequencer: &Sequencer{
//				keccak: DummyKeccak,
//				validator: mockValidator{
//					address: Alice,
//				},
//				verifier: mockVerifier{
//					isValidatorFn: func(v []byte, _ uint64) bool {
//						return !bytes.Equal(v, []byte("definitely not a validator"))
//					},
//					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
//						return bytes.Equal(v, Bob) && round == 1
//					},
//					isValidProposalFn: AlwaysValidProposal,
//				},
//			},
//			msg: &message.Proposal{
//				Sender:   Bob,
//				Sequence: 101,
//				Round:    1,
//				ProposedBlock: &message.ProposedBlock{
//					Block: []byte("block"),
//					Round: 1,
//				},
//				BlockHash: DummyKeccakValue,
//				RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
//					{
//						Sender:   Chris,
//						Sequence: 101,
//						Round:    1,
//					},
//					{
//						Sender:   Chris,
//						Sequence: 101,
//						Round:    1,
//					},
//				}},
//			},
//		},
//
//		{
//			name: "no quorum in rcc",
//			sequencer: &Sequencer{
//				keccak: DummyKeccak,
//				validator: mockValidator{
//					address: Alice,
//				},
//				verifier: mockVerifier{
//					isValidatorFn: func(v []byte, _ uint64) bool {
//						return !bytes.Equal(v, []byte("definitely not a validator"))
//					},
//					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
//						return bytes.Equal(v, Bob) && round == 1
//					},
//					hasQuorumFn: func([][]byte, uint64) bool {
//						return false
//					},
//					isValidProposalFn: AlwaysValidProposal,
//				},
//			},
//			msg: &message.Proposal{
//				Sender:   Bob,
//				Sequence: 101,
//				Round:    1,
//				ProposedBlock: &message.ProposedBlock{
//					Block: []byte("block"),
//					Round: 1,
//				},
//				BlockHash: DummyKeccakValue,
//				RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
//					{
//						Sender:   Chris,
//						Sequence: 101,
//						Round:    1,
//					},
//					{
//						Sender:   Nina,
//						Sequence: 101,
//						Round:    1,
//					},
//				}},
//			},
//		},
//
//		{
//			name: "invalid block in rcc",
//			sequencer: &Sequencer{
//				keccak: DummyKeccak,
//				validator: mockValidator{
//					address: Alice,
//				},
//				verifier: mockVerifier{
//					isValidProposalFn: func(p []byte, _ uint64) bool {
//						return bytes.Equal(p, []byte("block"))
//					},
//					isValidatorFn: func(v []byte, _ uint64) bool {
//						return !bytes.Equal(v, []byte("definitely not a validator"))
//					},
//					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
//						return bytes.Equal(v, Bob) && round == 1
//					},
//					hasQuorumFn: func(_ [][]byte, _ uint64) bool {
//						return true
//					},
//				},
//			},
//			msg: &message.Proposal{
//				Sender:   Bob,
//				Sequence: 101,
//				Round:    1,
//				ProposedBlock: &message.ProposedBlock{
//					Block: []byte("invalid block"),
//					Round: 1,
//				},
//				BlockHash: DummyKeccakValue,
//				RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
//					{
//						Sender:   Chris,
//						Sequence: 101,
//						Round:    1,
//					},
//					{
//						Sender:   Nina,
//						Sequence: 101,
//						Round:    1,
//					},
//				}},
//			},
//		},
//
//		{
//			name: "highest round block hash does not match derived hash",
//			sequencer: &Sequencer{
//				keccak: DummyKeccak,
//				validator: mockValidator{
//					address: Alice,
//					isValidProposalFn: func(_ uint64, block []byte) bool {
//						return bytes.Equal(block, []byte("block"))
//					},
//				},
//				verifier: mockVerifier{
//					isValidatorFn: func(v []byte, _ uint64) bool {
//						return !bytes.Equal(v, []byte("definitely not a validator"))
//					},
//					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
//						return bytes.Equal(v, Alice) && round == 0 ||
//							bytes.Equal(v, Bob) && round == 1
//					},
//					hasQuorumFn: func(_ [][]byte, _ uint64) bool {
//						return true
//					},
//				},
//			},
//			msg: &message.Proposal{
//				Sender:   Bob,
//				Sequence: 101,
//				Round:    1,
//				ProposedBlock: &message.ProposedBlock{
//					Block: []byte("block"),
//					Round: 1,
//				},
//				BlockHash: DummyKeccakValue,
//				RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
//					{
//						Sender:   Chris,
//						Sequence: 101,
//						Round:    1,
//						LatestPreparedCertificate: &message.PreparedCertificate{
//							ProposalMessage: &message.Proposal{
//								Sender:    Alice,
//								Sequence:  101,
//								Round:     0,
//								BlockHash: []byte("invalid keccak"),
//							},
//							PrepareMessages: []*message.Prepare{
//								{
//									Sender:    Chris,
//									Sequence:  101,
//									Round:     0,
//									BlockHash: []byte("invalid keccak"),
//								},
//							},
//						},
//					},
//				}},
//			},
//		},
//
//		{
//			name:     "ok",
//			expected: true,
//			sequencer: &Sequencer{
//				keccak: DummyKeccak,
//				validator: mockValidator{
//					address: Alice,
//				},
//				verifier: mockVerifier{
//					isValidatorFn: func(from []byte, _ uint64) bool {
//						return bytes.Equal(from, Alice) || bytes.Equal(from, Chris)
//					},
//					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
//						return bytes.Equal(v, Alice) && round == 0 ||
//							bytes.Equal(v, Bob) && round == 1
//					},
//					hasQuorumFn: func(_ [][]byte, _ uint64) bool {
//						return true
//					},
//					isValidProposalFn: AlwaysValidProposal,
//				},
//			},
//
//			msg: &message.Proposal{
//				Sender:   Bob,
//				Sequence: 101,
//				Round:    1,
//				ProposedBlock: &message.ProposedBlock{
//					Block: []byte("block"),
//					Round: 1,
//				},
//				BlockHash: DummyKeccakValue,
//				RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
//					{
//						Sender:                      Chris,
//						Sequence:                    101,
//						Round:                       1,
//						LatestPreparedProposedBlock: &message.ProposedBlock{},
//						LatestPreparedCertificate: &message.PreparedCertificate{
//							ProposalMessage: &message.Proposal{
//								Sender:    Alice,
//								Sequence:  101,
//								Round:     0,
//								BlockHash: DummyKeccakValue,
//							},
//							PrepareMessages: []*message.Prepare{
//								{
//									Sender:    Chris,
//									Sequence:  101,
//									Round:     0,
//									BlockHash: DummyKeccakValue,
//								},
//							},
//						},
//					},
//				}},
//			},
//		},
//	}
//
//	for _, tt := range testTable {
//		t.Run(tt.name, func(t *testing.T) {
//			t.Parallel()
//
//			assert.Equal(t, tt.expected, tt.sequencer.isValidMsgProposal(tt.msg))
//		})
//	}
//}
//
//func Test_IsValidMsgPrepare(t *testing.T) {
//	t.Parallel()
//
//	testTable := []struct {
//		msg       *message.Prepare
//		sequencer *Sequencer
//		name      string
//		expected  bool
//	}{
//		{
//			name: "invalid sender",
//			sequencer: &Sequencer{verifier: mockVerifier{isValidatorFn: func(_ []byte, _ uint64) bool {
//				return false
//			}}},
//			msg: &message.Prepare{
//				Sender:   []byte("definitely not a validator"),
//				Sequence: 101,
//			},
//		},
//
//		{
//			name: "invalid block hash",
//			sequencer: &Sequencer{
//				verifier: mockVerifier{isValidatorFn: func(_ []byte, _ uint64) bool {
//					return true
//				}},
//				state: Sequence{
//					proposal: &message.Proposal{
//						BlockHash: []byte("keccak"),
//					},
//				},
//			},
//			msg: &message.Prepare{
//				Sequence:  101,
//				Sender:    Chris,
//				BlockHash: []byte("definitely not keccak"),
//			},
//		},
//
//		{
//			name:     "ok",
//			expected: true,
//			sequencer: &Sequencer{
//				verifier: mockVerifier{isValidatorFn: func(_ []byte, _ uint64) bool {
//					return true
//				}},
//				state: Sequence{proposal: &message.Proposal{BlockHash: DummyKeccakValue}},
//			},
//			msg: &message.Prepare{
//				Sequence:  101,
//				Sender:    Chris,
//				BlockHash: DummyKeccakValue,
//			},
//		},
//	}
//
//	for _, tt := range testTable {
//		t.Run(tt.name, func(t *testing.T) {
//			t.Parallel()
//
//			assert.Equal(t, tt.expected, tt.sequencer.isValidMsgPrepare(tt.msg))
//		})
//	}
//}
//
//func Test_IsValidMsgCommit(t *testing.T) {
//	t.Parallel()
//
//	testTable := []struct {
//		msg       *message.Commit
//		sequencer *Sequencer
//		name      string
//		expected  bool
//	}{
//		{
//			name: "invalid sender",
//			sequencer: &Sequencer{verifier: mockVerifier{isValidatorFn: func(_ []byte, _ uint64) bool {
//				return false
//			}}},
//			msg: &message.Commit{
//				Sequence: 101,
//				Sender:   []byte("definitely not a validator"),
//			},
//		},
//
//		{
//			name: "invalid block hash",
//			sequencer: &Sequencer{
//				verifier: mockVerifier{isValidatorFn: AlwaysAValidator},
//				state:    Sequence{proposal: &message.Proposal{BlockHash: []byte("keccak")}},
//			},
//			msg: &message.Commit{
//				Sequence:  101,
//				Sender:    Chris,
//				BlockHash: []byte("definitely not keccak"),
//			},
//		},
//
//		{
//			name: "invalid commit seal",
//			sequencer: &Sequencer{
//				verifier: mockVerifier{
//					isValidatorFn: AlwaysAValidator,
//					isValidSignatureFn: func(_ []byte, _ []byte, _ []byte) error {
//						return errors.New("bad sig")
//					},
//				},
//				state: Sequence{proposal: &message.Proposal{BlockHash: DummyKeccakValue}},
//			},
//			msg: &message.Commit{
//				Sequence: 101,
//				Sender:   Chris,
//
//				BlockHash:  []byte("keccak"),
//				CommitSeal: []byte("doesn't matter"),
//			},
//		},
//
//		{
//			name:     "ok",
//			expected: true,
//			sequencer: &Sequencer{
//				verifier: mockVerifier{
//					isValidatorFn:      AlwaysAValidator,
//					isValidSignatureFn: AlwaysValidSignature,
//				},
//				state: Sequence{proposal: &message.Proposal{BlockHash: DummyKeccakValue}},
//			},
//			msg: &message.Commit{
//				Sequence: 101,
//				Sender:   Chris,
//
//				BlockHash:  []byte("keccak"),
//				CommitSeal: []byte("doesn't matter"),
//			},
//		},
//	}
//
//	for _, tt := range testTable {
//		t.Run(tt.name, func(t *testing.T) {
//			t.Parallel()
//
//			assert.Equal(t, tt.expected, tt.sequencer.isValidMsgCommit(tt.msg))
//		})
//	}
//}
//
//func TestIsValidMsgRoundChange(t *testing.T) {
//	t.Parallel()
//
//	testTable := []struct {
//		sequencer *Sequencer
//		msg       *message.RoundChange
//		name      string
//		expected  bool
//	}{
//		{
//			name: "invalid sender",
//			sequencer: &Sequencer{verifier: mockVerifier{isValidatorFn: func(_ []byte, _ uint64) bool {
//				return false
//			}}},
//			msg: &message.RoundChange{
//				Sender:   []byte("definitely not a validator"),
//				Sequence: 101,
//			},
//		},
//
//		{
//			name:      "ok (pb and pc are nil)",
//			expected:  true,
//			sequencer: &Sequencer{verifier: mockVerifier{isValidatorFn: AlwaysAValidator}},
//			msg: &message.RoundChange{
//				Sequence: 101,
//				Sender:   Chris,
//			},
//		},
//
//		{
//			name:      "proposed block and prepared certificate must both be set",
//			sequencer: &Sequencer{verifier: mockVerifier{isValidatorFn: AlwaysAValidator}},
//			msg: &message.RoundChange{
//				Sequence:                  101,
//				Sender:                    Chris,
//				LatestPreparedCertificate: &message.PreparedCertificate{},
//			},
//		},
//
//		{
//			name:      "(invalid pc) proposal message and prepare messages are not included",
//			sequencer: &Sequencer{verifier: mockVerifier{isValidatorFn: AlwaysAValidator}},
//			msg: &message.RoundChange{
//				Sequence:                    101,
//				Sender:                      Chris,
//				LatestPreparedProposedBlock: &message.ProposedBlock{},
//				LatestPreparedCertificate: &message.PreparedCertificate{
//					ProposalMessage: nil,
//				},
//			},
//		},
//
//		{
//			name:      "(invalid pc) bad proposal message sequence",
//			sequencer: &Sequencer{verifier: mockVerifier{isValidatorFn: AlwaysAValidator}},
//			msg: &message.RoundChange{
//				Sequence:                    101,
//				Sender:                      Chris,
//				LatestPreparedProposedBlock: &message.ProposedBlock{},
//				LatestPreparedCertificate: &message.PreparedCertificate{
//					ProposalMessage: &message.Proposal{},
//					PrepareMessages: []*message.Prepare{},
//				},
//			},
//		},
//
//		{
//			name:      "(invalid pc) bad proposal message round",
//			sequencer: &Sequencer{verifier: mockVerifier{isValidatorFn: AlwaysAValidator}},
//			msg: &message.RoundChange{
//				Sequence:                    101,
//				Sender:                      Chris,
//				LatestPreparedProposedBlock: &message.ProposedBlock{},
//				LatestPreparedCertificate: &message.PreparedCertificate{
//					ProposalMessage: &message.Proposal{
//						Sequence: 101,
//						Round:    5,
//					},
//					PrepareMessages: []*message.Prepare{},
//				},
//			},
//		},
//
//		{
//			name: "(invalid pc) bad proposer",
//			sequencer: &Sequencer{verifier: mockVerifier{
//				isValidatorFn: func(v []byte, _ uint64) bool {
//					return bytes.Equal(v, Chris)
//				},
//				isProposerFn: func(_ []byte, _ uint64, _ uint64) bool {
//					return false
//				},
//			}},
//			msg: &message.RoundChange{
//				Sender:                      Chris,
//				Sequence:                    101,
//				Round:                       1,
//				LatestPreparedProposedBlock: &message.ProposedBlock{},
//				LatestPreparedCertificate: &message.PreparedCertificate{
//					ProposalMessage: &message.Proposal{
//						Sender:   []byte("bad proposer"),
//						Sequence: 101,
//						Round:    0,
//					},
//					PrepareMessages: []*message.Prepare{},
//				},
//			},
//		},
//
//		{
//			name: "(invalid pc) proposal msg sequence (round) and prepare msg sequence (round) do not match",
//			sequencer: &Sequencer{verifier: mockVerifier{
//				isValidatorFn: AlwaysAValidator,
//				isProposerFn:  func(_ []byte, _ uint64, _ uint64) bool { return true },
//			}},
//			msg: &message.RoundChange{
//				Sender:                      Chris,
//				Sequence:                    101,
//				Round:                       2,
//				LatestPreparedProposedBlock: &message.ProposedBlock{},
//				LatestPreparedCertificate: &message.PreparedCertificate{
//					ProposalMessage: &message.Proposal{
//						Sender:   Bob,
//						Sequence: 101,
//						Round:    1,
//					},
//					PrepareMessages: []*message.Prepare{
//						{
//							Sequence: 99,
//							Round:    3212,
//						},
//					},
//				},
//			},
//		},
//
//		{
//			name: "(invalid pc) proposal msg block hash and prepare msg block hash do not match",
//			sequencer: &Sequencer{verifier: mockVerifier{
//				isValidatorFn: AlwaysAValidator,
//				isProposerFn:  func(_ []byte, _ uint64, _ uint64) bool { return true },
//			}},
//			msg: &message.RoundChange{
//				Sender:                      Chris,
//				Sequence:                    101,
//				Round:                       2,
//				LatestPreparedProposedBlock: &message.ProposedBlock{},
//				LatestPreparedCertificate: &message.PreparedCertificate{
//					ProposalMessage: &message.Proposal{
//						Sender:    Bob,
//						Sequence:  101,
//						Round:     1,
//						BlockHash: []byte("xxx"),
//					},
//					PrepareMessages: []*message.Prepare{
//						{
//							Sequence:  101,
//							Round:     1,
//							BlockHash: []byte("yyy"),
//						},
//					},
//				},
//			},
//		},
//
//		{
//			name: "(invalid pc) bad prepare msg sender",
//			sequencer: &Sequencer{verifier: mockVerifier{
//				isValidatorFn: func(v []byte, _ uint64) bool {
//					return bytes.Equal(v, Bob) || bytes.Equal(v, Chris)
//				},
//				isProposerFn: func(_ []byte, _ uint64, _ uint64) bool {
//					return true
//				},
//			}},
//			msg: &message.RoundChange{
//				Sender:                      Chris,
//				Sequence:                    101,
//				Round:                       2,
//				LatestPreparedProposedBlock: &message.ProposedBlock{},
//				LatestPreparedCertificate: &message.PreparedCertificate{
//					ProposalMessage: &message.Proposal{
//						Sender:   Bob,
//						Sequence: 101, Round: 1,
//						BlockHash: DummyKeccakValue,
//					},
//					PrepareMessages: []*message.Prepare{
//						{
//							Sender:    []byte("definitely not a validator"),
//							Sequence:  101,
//							Round:     1,
//							BlockHash: DummyKeccakValue,
//						},
//					},
//				},
//			},
//		},
//
//		{
//			name: "(invalid pc) duplicate sender in prepare messages",
//			sequencer: &Sequencer{verifier: mockVerifier{
//				isValidatorFn: func(v []byte, _ uint64) bool {
//					return bytes.Equal(v, Bob) || bytes.Equal(v, Chris)
//				},
//				isProposerFn: func(_ []byte, _ uint64, _ uint64) bool {
//					return true
//				},
//			}},
//			msg: &message.RoundChange{
//				Sender:                      Chris,
//				Sequence:                    101,
//				Round:                       2,
//				LatestPreparedProposedBlock: &message.ProposedBlock{},
//				LatestPreparedCertificate: &message.PreparedCertificate{
//					ProposalMessage: &message.Proposal{
//						Sender:   Bob,
//						Sequence: 101, Round: 1,
//						BlockHash: DummyKeccakValue,
//					},
//					PrepareMessages: []*message.Prepare{
//						{
//							Sender:    Chris,
//							Sequence:  101,
//							Round:     1,
//							BlockHash: DummyKeccakValue,
//						},
//						{
//							Sender:    Chris,
//							Sequence:  101,
//							Round:     1,
//							BlockHash: DummyKeccakValue,
//						},
//					},
//				},
//			},
//		},
//
//		{
//			name: "(invalid pc) no quorum messages in pc",
//			sequencer: &Sequencer{verifier: mockVerifier{
//				isValidatorFn: AlwaysAValidator,
//				isProposerFn: func(_ []byte, _ uint64, _ uint64) bool {
//					return true
//				},
//				hasQuorumFn: func(_ [][]byte, _ uint64) bool {
//					return false
//				},
//			}},
//			msg: &message.RoundChange{
//				Sender:                      Chris,
//				Sequence:                    101,
//				Round:                       2,
//				LatestPreparedProposedBlock: &message.ProposedBlock{},
//				LatestPreparedCertificate: &message.PreparedCertificate{
//					ProposalMessage: &message.Proposal{
//						Sender:    Bob,
//						Sequence:  101,
//						Round:     1,
//						BlockHash: DummyKeccakValue,
//					},
//					PrepareMessages: []*message.Prepare{
//						{
//							Sender:    Chris,
//							Sequence:  101,
//							Round:     1,
//							BlockHash: DummyKeccakValue,
//						},
//						{
//							Sender:    Alice,
//							Sequence:  101,
//							Round:     1,
//							BlockHash: DummyKeccakValue,
//						},
//					},
//				},
//			},
//		},
//
//		{
//			name: "block hash in pc and block hash in pb do not match",
//			sequencer: &Sequencer{
//				keccak: DummyKeccak,
//				verifier: mockVerifier{
//					isValidatorFn: AlwaysAValidator,
//					isProposerFn:  func(_ []byte, _ uint64, _ uint64) bool { return true },
//					hasQuorumFn:   func(_ [][]byte, _ uint64) bool { return true },
//				},
//			},
//			msg: &message.RoundChange{
//				Sender:                      Chris,
//				Sequence:                    101,
//				Round:                       2,
//				LatestPreparedProposedBlock: &message.ProposedBlock{},
//				LatestPreparedCertificate: &message.PreparedCertificate{
//					ProposalMessage: &message.Proposal{
//						Sender:    Bob,
//						Sequence:  101,
//						Round:     1,
//						BlockHash: []byte("definitely not keccak"),
//					},
//					PrepareMessages: []*message.Prepare{
//						{
//							Sender:    Chris,
//							Sequence:  101,
//							Round:     1,
//							BlockHash: []byte("definitely not keccak"),
//						},
//						{
//							Sender:    Alice,
//							Sequence:  101,
//							Round:     1,
//							BlockHash: []byte("definitely not keccak"),
//						},
//					},
//				},
//			},
//		},
//
//		{
//			name:     "ok",
//			expected: true,
//			sequencer: &Sequencer{
//				keccak: DummyKeccak,
//				verifier: mockVerifier{
//					isValidatorFn: AlwaysAValidator,
//					isProposerFn:  func(_ []byte, _ uint64, _ uint64) bool { return true },
//					hasQuorumFn:   func(_ [][]byte, _ uint64) bool { return true },
//				},
//			},
//			msg: &message.RoundChange{
//				Sender:                      Chris,
//				Sequence:                    101,
//				Round:                       2,
//				LatestPreparedProposedBlock: &message.ProposedBlock{},
//				LatestPreparedCertificate: &message.PreparedCertificate{
//					ProposalMessage: &message.Proposal{
//						Sender:    Bob,
//						Sequence:  101,
//						Round:     1,
//						BlockHash: DummyKeccakValue,
//					},
//					PrepareMessages: []*message.Prepare{
//						{
//							Sender:    Chris,
//							Sequence:  101,
//							Round:     1,
//							BlockHash: DummyKeccakValue,
//						},
//						{
//							Sender:    Alice,
//							Sequence:  101,
//							Round:     1,
//							BlockHash: DummyKeccakValue,
//						},
//					},
//				},
//			},
//		},
//	}
//
//	for _, tt := range testTable {
//		t.Run(tt.name, func(t *testing.T) {
//			t.Parallel()
//
//			assert.Equal(t, tt.expected, tt.sequencer.isValidMsgRoundChange(tt.msg))
//		})
//	}
//}
