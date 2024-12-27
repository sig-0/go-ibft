//nolint:dupl // test cases not identical
package sequencer

import (
	"bytes"
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/sig-0/go-ibft/message"

	"github.com/stretchr/testify/assert"
)

func Test_SequencerFinalizeCancelled(t *testing.T) {
	t.Parallel()

	s := NewSequencer(NewConfig(
		WithValidator(mockValidator{address: Alice}),
		WithValidatorSet(mockValidatorSet{isProposerFn: func(_ []byte, _ uint64, _ uint64) bool {
			return false
		}}),
		WithRound0Duration(10*time.Millisecond),
		WithFeed(NewSingleRoundMockFeed(nil)),
	))

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan *SequenceResult)

	go func(ctx context.Context) {
		defer close(ch)

		ch <- s.Finalize(ctx, 101)
	}(ctx)

	cancel()

	assert.Nil(t, <-ch)
}

func Test_SequencerFinalize(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		expected *SequenceResult
		cfg      Config
		name     string
	}{
		{
			name: "Alice and Chris accept Bob's proposal in round 0",
			expected: &SequenceResult{
				Round:    0,
				Proposal: []byte("Bob's proposal"),
				Seals: []CommitSeal{
					{
						From: Alice,
						Seal: []byte("Alice seal"),
					},
					{
						From: Chris,
						Seal: []byte("Chris seal"),
					},
				},
			},
			cfg: NewConfig(
				WithValidator(mockValidator{
					address:           Alice,
					signFn:            DummySignFn,
					isValidProposalFn: AlwaysValidProposal,
				}),
				WithValidatorSet(mockValidatorSet{
					isValidatorFn: AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Bob) && round == 0
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(newMockFeed([]message.Message{
					&message.MsgProposal{
						Info:          &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 0},
						ProposedBlock: &message.ProposedBlock{Block: []byte("Bob's proposal"), Round: 0},
						BlockHash:     DummyKeccakValue,
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Alice, Sequence: 101, Round: 0},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 0},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Alice, Sequence: 101, Round: 0},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Alice seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 0},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Chris seal"),
					},
				})),
				WithKeccak(DummyKeccak),
				WithTransport(DummyTransport()),
				WithSignatureVerifier(AlwaysValidSignature),
				WithRound0Duration(10*time.Millisecond),
			),
		},

		{
			name: "Bob and Chris accept Alice's proposal in round 0",
			expected: &SequenceResult{
				Round:    0,
				Proposal: []byte("Alice's proposal"),
				Seals: []CommitSeal{
					{
						From: Bob,
						Seal: []byte("Bob seal"),
					},
					{
						From: Chris,
						Seal: []byte("Chris seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(DummyKeccak),
				WithSignatureVerifier(AlwaysValidSignature),
				WithValidator(mockValidator{
					address: Alice,
					signFn:  DummySignFn,
					buildProposalFn: func(_ uint64) []byte {
						return []byte("Alice's proposal")
					},
					isValidProposalFn: AlwaysValidProposal,
				}),
				WithValidatorSet(mockValidatorSet{
					isValidatorFn: AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Alice) && round == 0
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(newMockFeed([]message.Message{
					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 0},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 0},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 0},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Bob seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 0},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Chris seal"),
					},
				})),
			),
		},
		{
			name: "Alice and Chris accept Bob's proposal in round 1 due to round change",
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("Bob's round 1 proposal"),
				Seals: []CommitSeal{
					{
						From: Alice,
						Seal: []byte("Alice seal"),
					},
					{
						From: Chris,
						Seal: []byte("Chris seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(DummyKeccak),
				WithSignatureVerifier(AlwaysValidSignature),
				WithValidator(mockValidator{
					address:           Alice,
					signFn:            DummySignFn,
					isValidProposalFn: AlwaysValidProposal,
				}),
				WithValidatorSet(mockValidatorSet{
					isValidatorFn: AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Bob) && round == 1
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(newMockFeed([]message.Message{
					&message.MsgProposal{
						Info:          &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 1},
						BlockHash:     DummyKeccakValue,
						ProposedBlock: &message.ProposedBlock{Block: []byte("Bob's round 1 proposal"), Round: 1},
						RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.MsgRoundChange{
							{
								Info: &message.MsgInfo{Sequence: 101, Round: 1, Sender: Alice},
							},

							{
								Info: &message.MsgInfo{Sequence: 101, Round: 1, Sender: Chris},
							},
						}},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Alice, Sequence: 101, Round: 1},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 1},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Alice, Sequence: 101, Round: 1},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Alice seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 1},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Chris seal"),
					},
				})),
			),
		},

		{
			name: "Alice jumps to round 1 proposal and accepts it",
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("Chris' proposal"),
				Seals: []CommitSeal{
					{
						From: Alice,
						Seal: []byte("Alice seal"),
					},
					{
						From: Chris,
						Seal: []byte("Chris seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(DummyKeccak),
				WithSignatureVerifier(AlwaysValidSignature),
				WithValidator(mockValidator{
					address:           Alice,
					signFn:            DummySignFn,
					isValidProposalFn: AlwaysValidProposal,
				}),
				WithValidatorSet(mockValidatorSet{
					isValidatorFn: AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Chris) && round == 0 || bytes.Equal(v, Bob) && round == 1
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(newMockFeed([]message.Message{
					&message.MsgProposal{
						Info:          &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 1},
						BlockHash:     DummyKeccakValue,
						ProposedBlock: &message.ProposedBlock{Block: []byte("Chris' proposal"), Round: 1},
						RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.MsgRoundChange{
							{
								Info: &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 1},
								LatestPreparedCertificate: &message.PreparedCertificate{
									ProposalMessage: &message.MsgProposal{
										Info:      &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 0},
										BlockHash: DummyKeccakValue,
										ProposedBlock: &message.ProposedBlock{
											Block: []byte("Chris' proposal"),
											Round: 0,
										},
									},
									PrepareMessages: []*message.MsgPrepare{
										{
											Info:      &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 0},
											BlockHash: DummyKeccakValue,
										},
										{
											Info:      &message.MsgInfo{Sender: Nina, Sequence: 101, Round: 0},
											BlockHash: DummyKeccakValue,
										},
									},
								},
							},
							{
								Info: &message.MsgInfo{Sender: Nina, Sequence: 101, Round: 1},
								LatestPreparedCertificate: &message.PreparedCertificate{
									ProposalMessage: &message.MsgProposal{
										Info:      &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 0},
										BlockHash: DummyKeccakValue,
										ProposedBlock: &message.ProposedBlock{
											Block: []byte("Chris' proposal"),
											Round: 0,
										},
									},
									PrepareMessages: []*message.MsgPrepare{
										{
											Info:      &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 0},
											BlockHash: DummyKeccakValue,
										},
										{
											Info:      &message.MsgInfo{Sender: Nina, Sequence: 101, Round: 0},
											BlockHash: DummyKeccakValue,
										},
									},
								},
							},
						}},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Alice, Sequence: 101, Round: 1},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 1},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Alice, Sequence: 101, Round: 1},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Alice seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 1},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Chris seal"),
					},
				})),
			),
		},

		{
			name: "block proposed in round 1",
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("Alice's round 1 proposal"),
				Seals: []CommitSeal{
					{
						From: Bob,
						Seal: []byte("Bob seal"),
					},

					{
						From: Nina,
						Seal: []byte("Nina seal"),
					},
				},
			},

			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(DummyKeccak),
				WithSignatureVerifier(AlwaysValidSignature),
				WithValidator(mockValidator{
					address:           Alice,
					signFn:            DummySignFn,
					isValidProposalFn: AlwaysValidProposal,
					buildProposalFn: func(_ uint64) []byte {
						return []byte("Alice's round 1 proposal")
					},
				}),
				WithValidatorSet(mockValidatorSet{
					isValidatorFn: AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Bob) && round == 0 || bytes.Equal(v, Alice) && round == 1
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(newMockFeed([]message.Message{
					// need to justify Alice's proposal for round 1
					&message.MsgRoundChange{
						Info: &message.MsgInfo{Sequence: 101, Round: 1, Sender: Alice},
					},

					&message.MsgRoundChange{
						Info: &message.MsgInfo{Sequence: 101, Round: 1, Sender: Nina},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sequence: 101, Round: 1, Sender: Bob},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sequence: 101, Round: 1, Sender: Nina},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sequence: 101, Round: 1, Sender: Bob},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Bob seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sequence: 101, Round: 1, Sender: Nina},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Nina seal"),
					},
				})),
			),
		},

		{
			name: "old block proposed in round 1",
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("Bob's round 0 proposal"),
				Seals: []CommitSeal{
					{
						From: Bob,
						Seal: []byte("Bob seal"),
					},
					{
						From: Chris,
						Seal: []byte("Chris seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(DummyKeccak),
				WithSignatureVerifier(AlwaysValidSignature),
				WithValidator(mockValidator{
					address:           Alice,
					signFn:            DummySignFn,
					isValidProposalFn: AlwaysValidProposal,
				}),
				WithValidatorSet(mockValidatorSet{
					isValidatorFn: AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Bob) && round == 0 || bytes.Equal(v, Alice) && round == 1
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(newMockFeed([]message.Message{
					&message.MsgRoundChange{
						Info: &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 1},
						LatestPreparedProposedBlock: &message.ProposedBlock{
							Block: []byte("Bob's round 0 proposal"),
							Round: 0,
						},
						LatestPreparedCertificate: &message.PreparedCertificate{
							ProposalMessage: &message.MsgProposal{
								Info:      &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 0},
								BlockHash: DummyKeccakValue,
								ProposedBlock: &message.ProposedBlock{
									Block: []byte("Bob's round 0 proposal"),
									Round: 0,
								},
							},

							PrepareMessages: []*message.MsgPrepare{
								{
									Info:      &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 0},
									BlockHash: DummyKeccakValue,
								},
								{
									Info:      &message.MsgInfo{Sender: Nina, Sequence: 101, Round: 0},
									BlockHash: DummyKeccakValue,
								},
							},
						},
					},

					&message.MsgRoundChange{
						Info: &message.MsgInfo{Sender: Nina, Sequence: 101, Round: 1},
						LatestPreparedProposedBlock: &message.ProposedBlock{
							Block: []byte("Bob's round 0 proposal"),
							Round: 0,
						},
						LatestPreparedCertificate: &message.PreparedCertificate{
							ProposalMessage: &message.MsgProposal{
								Info:      &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 0},
								BlockHash: DummyKeccakValue,
								ProposedBlock: &message.ProposedBlock{
									Block: []byte("Bob's round 0 proposal"),
									Round: 0,
								},
							},

							PrepareMessages: []*message.MsgPrepare{
								{
									Info:      &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 0},
									BlockHash: DummyKeccakValue,
								},
								{
									Info:      &message.MsgInfo{Sender: Nina, Sequence: 101, Round: 0},
									BlockHash: DummyKeccakValue,
								},
							},
						},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 1},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 1},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 1},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Bob seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 1},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Chris seal"),
					},
				})),
			),
		},

		{
			name: "future rcc triggers round jump",
			expected: &SequenceResult{
				Round:    3,
				Proposal: []byte("Alice round 3 proposal"),
				Seals: []CommitSeal{
					{
						From: Bob,
						Seal: []byte("Bob seal"),
					},
					{
						From: Chris,
						Seal: []byte("Chris seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(DummyKeccak),
				WithSignatureVerifier(AlwaysValidSignature),
				WithValidator(mockValidator{
					address:           Alice,
					signFn:            DummySignFn,
					isValidProposalFn: AlwaysValidProposal,
					buildProposalFn: func(_ uint64) []byte {
						return []byte("Alice round 3 proposal")
					},
				}),
				WithValidatorSet(mockValidatorSet{
					isValidatorFn: AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Alice) && round == 3
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(newMockFeed([]message.Message{
					&message.MsgRoundChange{Info: &message.MsgInfo{
						Sender:   Bob,
						Sequence: 101,
						Round:    3,
					}},

					&message.MsgRoundChange{Info: &message.MsgInfo{
						Sender:   Chris,
						Sequence: 101,
						Round:    3,
					}},

					&message.MsgPrepare{
						Info: &message.MsgInfo{
							Sender:   Bob,
							Sequence: 101,
							Round:    3,
						},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgPrepare{
						Info: &message.MsgInfo{
							Sender:   Chris,
							Sequence: 101,
							Round:    3,
						},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgCommit{
						Info: &message.MsgInfo{
							Sender:   Bob,
							Sequence: 101,
							Round:    3,
						},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Bob seal"),
					},

					&message.MsgCommit{
						Info: &message.MsgInfo{
							Sender:   Chris,
							Sequence: 101,
							Round:    3,
						},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Chris seal"),
					},
				})),
			),
		},

		{
			name: "future proposal triggers round jump",
			expected: &SequenceResult{
				Round:    5,
				Proposal: []byte("round 5 block"),
				Seals: []CommitSeal{
					{
						From: Alice,
						Seal: []byte("Alice seal"),
					},
					{
						From: Nina,
						Seal: []byte("Nina seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(DummyKeccak),
				WithSignatureVerifier(AlwaysValidSignature),
				WithValidator(mockValidator{
					address:           Alice,
					signFn:            DummySignFn,
					isValidProposalFn: AlwaysValidProposal,
				}),
				WithValidatorSet(mockValidatorSet{
					isValidatorFn: AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Nina) && round == 5
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(newMockFeed([]message.Message{
					&message.MsgProposal{
						Info:          &message.MsgInfo{Sender: Nina, Sequence: 101, Round: 5},
						ProposedBlock: &message.ProposedBlock{Block: []byte("round 5 block"), Round: 5},
						BlockHash:     DummyKeccakValue,
						RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.MsgRoundChange{
							{
								Info: &message.MsgInfo{Sequence: 101, Round: 5, Sender: Chris},
							},

							{
								Info: &message.MsgInfo{Sequence: 101, Round: 5, Sender: Bob},
							},
						}},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Alice, Sequence: 101, Round: 5},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Nina, Sequence: 101, Round: 5},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Alice, Sequence: 101, Round: 5},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Alice seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Nina, Sequence: 101, Round: 5},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Nina seal"),
					},
				})),
			),
		},

		{
			name: "round timer triggers round jump",
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("Alice round 1 proposal"),
				Seals: []CommitSeal{
					{
						From: Bob,
						Seal: []byte("Bob seal"),
					},
					{
						From: Chris,
						Seal: []byte("Chris seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(DummyKeccak),
				WithSignatureVerifier(AlwaysValidSignature),
				WithValidator(mockValidator{
					address:           Alice,
					signFn:            DummySignFn,
					isValidProposalFn: AlwaysValidProposal,
					buildProposalFn: func(_ uint64) []byte {
						return []byte("Alice round 1 proposal")
					},
				}),
				WithValidatorSet(mockValidatorSet{
					isValidatorFn: AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Alice) && round == 1
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(NewSingleRoundMockFeed([]message.Message{
					&message.MsgRoundChange{
						Info: &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 1},
					},

					&message.MsgRoundChange{
						Info: &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 1},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 1},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 1},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 1},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Bob seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 1},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Chris seal"),
					},
				})),
			),
		},

		{
			name: "no prepare messages in round 0",
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("round 1 block"),
				Seals: []CommitSeal{
					{
						From: Alice,
						Seal: []byte("Alice seal"),
					},

					{
						From: Nina,
						Seal: []byte("Nina seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(DummyKeccak),
				WithSignatureVerifier(AlwaysValidSignature),
				WithValidator(mockValidator{
					address:           Alice,
					signFn:            DummySignFn,
					isValidProposalFn: AlwaysValidProposal,
				}),
				WithValidatorSet(mockValidatorSet{
					isValidatorFn: AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Bob) && round == 0 || bytes.Equal(v, Chris) && round == 1
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(NewSingleRoundMockFeed([]message.Message{
					&message.MsgProposal{
						Info: &message.MsgInfo{
							Sender:   Bob,
							Sequence: 101,
							Round:    0,
						},
						BlockHash: DummyKeccakValue,
						ProposedBlock: &message.ProposedBlock{
							Block: []byte("round 0 block"),
							Round: 0,
						},
					},

					&message.MsgPrepare{
						Info: &message.MsgInfo{
							Sender:   Alice,
							Sequence: 101,
							Round:    0,
						},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgProposal{
						Info:          &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 1},
						BlockHash:     DummyKeccakValue,
						ProposedBlock: &message.ProposedBlock{Block: []byte("round 1 block"), Round: 1},
						RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.MsgRoundChange{
							{
								Info: &message.MsgInfo{Sequence: 101, Round: 1, Sender: Alice},
							},

							{
								Info: &message.MsgInfo{Sequence: 101, Round: 1, Sender: Chris},
							},
						}},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Alice, Sequence: 101, Round: 1},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Nina, Sequence: 101, Round: 1},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Alice, Sequence: 101, Round: 1},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Alice seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Nina, Sequence: 101, Round: 1},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Nina seal"),
					},
				})),
			),
		},

		{
			name: "no commit messages in round 0",
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("round 1 block"),
				Seals: []CommitSeal{
					{
						From: Alice,
						Seal: []byte("Alice seal"),
					},
					{
						From: Bob,
						Seal: []byte("Bob seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(DummyKeccak),
				WithSignatureVerifier(AlwaysValidSignature),
				WithValidator(mockValidator{
					address:           Alice,
					signFn:            DummySignFn,
					isValidProposalFn: AlwaysValidProposal,
				}),
				WithValidatorSet(mockValidatorSet{
					isValidatorFn: AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Bob) && round == 0 || bytes.Equal(v, Chris) && round == 1
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(NewSingleRoundMockFeed([]message.Message{
					/* round 0 */
					&message.MsgProposal{
						Info:          &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 0},
						BlockHash:     DummyKeccakValue,
						ProposedBlock: &message.ProposedBlock{Block: []byte("round 0 block"), Round: 0},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Alice, Sequence: 101, Round: 0},
						BlockHash: DummyKeccakValue,
					},
					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 0},
						BlockHash: DummyKeccakValue,
					},

					/* round 1 */

					&message.MsgProposal{
						Info:          &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 1},
						BlockHash:     DummyKeccakValue,
						ProposedBlock: &message.ProposedBlock{Block: []byte("round 1 block"), Round: 1},
						RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.MsgRoundChange{
							{
								Info: &message.MsgInfo{Sender: Alice, Sequence: 101, Round: 1},
							},
							{
								Info: &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 1},
							},
						}},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Alice, Sequence: 101, Round: 1},
						BlockHash: DummyKeccakValue,
					},
					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 1},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Alice, Sequence: 101, Round: 1},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Alice seal"),
					},
					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 1},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Bob seal"),
					},
				})),
			),
		},

		{
			name: "round 0 proposer fails to build block",
			expected: &SequenceResult{
				Round:    2,
				Proposal: []byte("round 2 block"),
				Seals: []CommitSeal{
					{
						From: Bob,
						Seal: []byte("Bob seal"),
					},
					{
						From: Alice,
						Seal: []byte("Alice seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(DummyKeccak),
				WithSignatureVerifier(AlwaysValidSignature),
				WithValidator(mockValidator{
					address:           Alice,
					signFn:            DummySignFn,
					isValidProposalFn: AlwaysValidProposal,
				}),
				WithValidatorSet(mockValidatorSet{
					isValidatorFn: AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Bob) && round == 0 ||
							bytes.Equal(v, Alice) && round == 1 ||
							bytes.Equal(v, Chris) && round == 2
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(NewSingleRoundMockFeed([]message.Message{
					&message.MsgRoundChange{
						Info: &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 1},
					},

					&message.MsgProposal{
						Info:          &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 2},
						ProposedBlock: &message.ProposedBlock{Block: []byte("round 2 block"), Round: 2},
						BlockHash:     DummyKeccakValue,
						RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.MsgRoundChange{
							{
								Info: &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 2},
							},
							{
								Info: &message.MsgInfo{Sender: Chris, Sequence: 101, Round: 2},
							},
						}},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 2},
						BlockHash: DummyKeccakValue,
					},
					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: Alice, Sequence: 101, Round: 2},
						BlockHash: DummyKeccakValue,
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Bob, Sequence: 101, Round: 2},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Bob seal"),
					},
					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: Alice, Sequence: 101, Round: 2},
						BlockHash:  DummyKeccakValue,
						CommitSeal: []byte("Alice seal"),
					},
				})),
			),
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := NewSequencer(tt.cfg).Finalize(context.Background(), 101)
			assert.True(t, reflect.DeepEqual(tt.expected, res))
		})
	}
}
