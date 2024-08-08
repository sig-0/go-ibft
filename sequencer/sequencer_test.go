package sequencer

import (
	"bytes"
	"context"
	"github.com/sig-0/go-ibft/message"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	alice = []byte("alice")
	bob   = []byte("bob")
	chris = []byte("chris")
	nina  = []byte("nina")
)

func Test_SequencerFinalizeCancelled(t *testing.T) {
	t.Parallel()

	s := NewSequencer(NewConfig(
		WithValidator(MockValidator{address: alice}),
		WithValidatorSet(MockValidatorSet{isProposerFn: func(_ []byte, _ uint64, _ uint64) bool {
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
		name     string
		cfg      Config
		expected *SequenceResult
	}{
		{
			name: "alice and chris accept bob's proposal in round 0",
			expected: &SequenceResult{
				Round:    0,
				Proposal: []byte("bob's proposal"),
				Seals: []CommitSeal{
					{
						From: alice,
						Seal: []byte("alice seal"),
					},
					{
						From: chris,
						Seal: []byte("chris seal"),
					},
				},
			},
			cfg: NewConfig(
				WithValidator(MockValidator{
					address:           alice,
					signFn:            func(_ []byte) []byte { return nil },
					isValidProposalFn: func(_ uint64, _ []byte) bool { return true },
				}),
				WithValidatorSet(MockValidatorSet{
					isValidatorFn: func(v []byte, _ uint64) bool {
						return true
					},
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, bob) && round == 0
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(NewMockFeed([]message.Message{
					&message.MsgProposal{
						Info:          &message.MsgInfo{Sender: bob, Sequence: 101, Round: 0},
						ProposedBlock: &message.ProposedBlock{Block: []byte("bob's proposal"), Round: 0},
						BlockHash:     []byte("block hash"),
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: alice, Sequence: 101, Round: 0},
						BlockHash: []byte("block hash"),
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: chris, Sequence: 101, Round: 0},
						BlockHash: []byte("block hash"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: alice, Sequence: 101, Round: 0},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("alice seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: chris, Sequence: 101, Round: 0},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("chris seal"),
					},
				})),
				WithKeccak(MockKeccak(func(_ []byte) []byte {
					return []byte("block hash")
				})),
				WithTransport(DummyTransport()),
				WithSignatureVerifier(MockSignatureVerifier(func(_ []byte, _ []byte, _ []byte) error {
					return nil
				})),
				WithRound0Duration(10*time.Millisecond),
			),
		},

		{
			name: "bob and chris accept alice's proposal in round 0",
			expected: &SequenceResult{
				Round:    0,
				Proposal: []byte("alice's proposal"),
				Seals: []CommitSeal{
					{
						From: bob,
						Seal: []byte("bob seal"),
					},
					{
						From: chris,
						Seal: []byte("chris seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(MockKeccak(func(_ []byte) []byte {
					return []byte("block hash")
				})),
				WithSignatureVerifier(MockSignatureVerifier(func(_ []byte, _ []byte, _ []byte) error {
					return nil
				})),
				WithValidator(MockValidator{
					address: alice,
					signFn:  func(_ []byte) []byte { return nil },
					buildProposalFn: func(_ uint64) []byte {
						return []byte("alice's proposal")
					},
					isValidProposalFn: func(_ uint64, _ []byte) bool { return true },
				}),
				WithValidatorSet(MockValidatorSet{
					isValidatorFn: func(v []byte, _ uint64) bool {
						return true
					},
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, alice) && round == 0
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(NewMockFeed([]message.Message{
					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: bob, Sequence: 101, Round: 0},
						BlockHash: []byte("block hash"),
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: chris, Sequence: 101, Round: 0},
						BlockHash: []byte("block hash"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: bob, Sequence: 101, Round: 0},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("bob seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: chris, Sequence: 101, Round: 0},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("chris seal"),
					},
				})),
			),
		},
		{
			name: "alice and chris accept bob's proposal in round 1 due to round change",
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("bob's round 1 proposal"),
				Seals: []CommitSeal{
					{
						From: alice,
						Seal: []byte("alice seal"),
					},
					{
						From: chris,
						Seal: []byte("chris seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(MockKeccak(func(_ []byte) []byte {
					return []byte("block hash")
				})),
				WithSignatureVerifier(MockSignatureVerifier(func(_ []byte, _ []byte, _ []byte) error {
					return nil
				})),
				WithValidator(MockValidator{
					address:           alice,
					signFn:            func(_ []byte) []byte { return nil },
					isValidProposalFn: func(_ uint64, _ []byte) bool { return true },
				}),
				WithValidatorSet(MockValidatorSet{
					isValidatorFn: func(v []byte, _ uint64) bool {
						return true
					},
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, bob) && round == 1
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(NewMockFeed([]message.Message{
					&message.MsgProposal{
						Info:          &message.MsgInfo{Sender: bob, Sequence: 101, Round: 1},
						BlockHash:     []byte("block hash"),
						ProposedBlock: &message.ProposedBlock{Block: []byte("bob's round 1 proposal"), Round: 1},
						RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.MsgRoundChange{
							{
								Info: &message.MsgInfo{Sequence: 101, Round: 1, Sender: alice},
							},

							{
								Info: &message.MsgInfo{Sequence: 101, Round: 1, Sender: chris},
							},
						}},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: alice, Sequence: 101, Round: 1},
						BlockHash: []byte("block hash"),
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: chris, Sequence: 101, Round: 1},
						BlockHash: []byte("block hash"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: alice, Sequence: 101, Round: 1},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("alice seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: chris, Sequence: 101, Round: 1},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("chris seal"),
					},
				})),
			),
		},

		{
			name: "alice jumps to round 1 proposal and accepts it",
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("chris' proposal"),
				Seals: []CommitSeal{
					{
						From: alice,
						Seal: []byte("alice seal"),
					},
					{
						From: chris,
						Seal: []byte("chris seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(MockKeccak(func(_ []byte) []byte {
					return []byte("block hash")
				})),
				WithSignatureVerifier(MockSignatureVerifier(func(_ []byte, _ []byte, _ []byte) error {
					return nil
				})),
				WithValidator(MockValidator{
					address:           alice,
					signFn:            func(_ []byte) []byte { return nil },
					isValidProposalFn: func(_ uint64, _ []byte) bool { return true },
				}),
				WithValidatorSet(MockValidatorSet{
					isValidatorFn: func(v []byte, _ uint64) bool {
						return true
					},
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, chris) && round == 0 || bytes.Equal(v, bob) && round == 1
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(NewMockFeed([]message.Message{
					&message.MsgProposal{
						Info:          &message.MsgInfo{Sender: bob, Sequence: 101, Round: 1},
						BlockHash:     []byte("block hash"),
						ProposedBlock: &message.ProposedBlock{Block: []byte("chris' proposal"), Round: 1},
						RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.MsgRoundChange{
							{
								Info: &message.MsgInfo{Sender: bob, Sequence: 101, Round: 1},
								LatestPreparedCertificate: &message.PreparedCertificate{
									ProposalMessage: &message.MsgProposal{
										Info:      &message.MsgInfo{Sender: chris, Sequence: 101, Round: 0},
										BlockHash: []byte("block hash"),
										ProposedBlock: &message.ProposedBlock{
											Block: []byte("chris' proposal"),
											Round: 0,
										},
									},
									PrepareMessages: []*message.MsgPrepare{
										{
											Info:      &message.MsgInfo{Sender: bob, Sequence: 101, Round: 0},
											BlockHash: []byte("block hash"),
										},
										{
											Info:      &message.MsgInfo{Sender: nina, Sequence: 101, Round: 0},
											BlockHash: []byte("block hash"),
										},
									},
								},
							},
							{
								Info: &message.MsgInfo{Sender: nina, Sequence: 101, Round: 1},
								LatestPreparedCertificate: &message.PreparedCertificate{
									ProposalMessage: &message.MsgProposal{
										Info:      &message.MsgInfo{Sender: chris, Sequence: 101, Round: 0},
										BlockHash: []byte("block hash"),
										ProposedBlock: &message.ProposedBlock{
											Block: []byte("chris' proposal"),
											Round: 0,
										},
									},
									PrepareMessages: []*message.MsgPrepare{
										{
											Info:      &message.MsgInfo{Sender: bob, Sequence: 101, Round: 0},
											BlockHash: []byte("block hash"),
										},
										{
											Info:      &message.MsgInfo{Sender: nina, Sequence: 101, Round: 0},
											BlockHash: []byte("block hash"),
										},
									},
								},
							},
						}},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: alice, Sequence: 101, Round: 1},
						BlockHash: []byte("block hash"),
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: chris, Sequence: 101, Round: 1},
						BlockHash: []byte("block hash"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: alice, Sequence: 101, Round: 1},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("alice seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: chris, Sequence: 101, Round: 1},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("chris seal"),
					},
				})),
			),
		},

		{
			name: "block proposed in round 1",
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("alice's round 1 proposal"),
				Seals: []CommitSeal{
					{
						From: bob,
						Seal: []byte("bob seal"),
					},

					{
						From: nina,
						Seal: []byte("nina seal"),
					},
				},
			},

			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(MockKeccak(func(_ []byte) []byte {
					return []byte("block hash")
				})),
				WithSignatureVerifier(MockSignatureVerifier(func(_ []byte, _ []byte, _ []byte) error {
					return nil
				})),
				WithValidator(MockValidator{
					address:           alice,
					signFn:            func(_ []byte) []byte { return nil },
					isValidProposalFn: func(_ uint64, _ []byte) bool { return true },
					buildProposalFn: func(_ uint64) []byte {
						return []byte("alice's round 1 proposal")
					},
				}),
				WithValidatorSet(MockValidatorSet{
					isValidatorFn: func(v []byte, _ uint64) bool {
						return true
					},
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, bob) && round == 0 || bytes.Equal(v, alice) && round == 1
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(NewMockFeed([]message.Message{
					// need to justify alice's proposal for round 1
					&message.MsgRoundChange{
						Info: &message.MsgInfo{Sequence: 101, Round: 1, Sender: alice},
					},

					&message.MsgRoundChange{
						Info: &message.MsgInfo{Sequence: 101, Round: 1, Sender: nina},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sequence: 101, Round: 1, Sender: bob},
						BlockHash: []byte("block hash"),
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sequence: 101, Round: 1, Sender: nina},
						BlockHash: []byte("block hash"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sequence: 101, Round: 1, Sender: bob},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("bob seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sequence: 101, Round: 1, Sender: nina},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("nina seal"),
					},
				})),
			),
		},

		{
			name: "old block proposed in round 1",
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("bob's round 0 proposal"),
				Seals: []CommitSeal{
					{
						From: bob,
						Seal: []byte("bob seal"),
					},
					{
						From: chris,
						Seal: []byte("chris seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(MockKeccak(func(_ []byte) []byte {
					return []byte("block hash")
				})),
				WithSignatureVerifier(MockSignatureVerifier(func(_ []byte, _ []byte, _ []byte) error {
					return nil
				})),
				WithValidator(MockValidator{
					address:           alice,
					signFn:            func(_ []byte) []byte { return nil },
					isValidProposalFn: func(_ uint64, _ []byte) bool { return true },
				}),
				WithValidatorSet(MockValidatorSet{
					isValidatorFn: func(v []byte, _ uint64) bool {
						return true
					},
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, bob) && round == 0 || bytes.Equal(v, alice) && round == 1
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(NewMockFeed([]message.Message{
					&message.MsgRoundChange{
						Info: &message.MsgInfo{Sender: chris, Sequence: 101, Round: 1},
						LatestPreparedProposedBlock: &message.ProposedBlock{
							Block: []byte("bob's round 0 proposal"),
							Round: 0,
						},
						LatestPreparedCertificate: &message.PreparedCertificate{
							ProposalMessage: &message.MsgProposal{
								Info:      &message.MsgInfo{Sender: bob, Sequence: 101, Round: 0},
								BlockHash: []byte("block hash"),
								ProposedBlock: &message.ProposedBlock{
									Block: []byte("bob's round 0 proposal"),
									Round: 0,
								},
							},

							PrepareMessages: []*message.MsgPrepare{
								{
									Info:      &message.MsgInfo{Sender: chris, Sequence: 101, Round: 0},
									BlockHash: []byte("block hash"),
								},
								{
									Info:      &message.MsgInfo{Sender: nina, Sequence: 101, Round: 0},
									BlockHash: []byte("block hash"),
								},
							},
						},
					},

					&message.MsgRoundChange{
						Info: &message.MsgInfo{Sender: nina, Sequence: 101, Round: 1},
						LatestPreparedProposedBlock: &message.ProposedBlock{
							Block: []byte("bob's round 0 proposal"),
							Round: 0,
						},
						LatestPreparedCertificate: &message.PreparedCertificate{
							ProposalMessage: &message.MsgProposal{
								Info:      &message.MsgInfo{Sender: bob, Sequence: 101, Round: 0},
								BlockHash: []byte("block hash"),
								ProposedBlock: &message.ProposedBlock{
									Block: []byte("bob's round 0 proposal"),
									Round: 0,
								},
							},

							PrepareMessages: []*message.MsgPrepare{
								{
									Info:      &message.MsgInfo{Sender: chris, Sequence: 101, Round: 0},
									BlockHash: []byte("block hash"),
								},
								{
									Info:      &message.MsgInfo{Sender: nina, Sequence: 101, Round: 0},
									BlockHash: []byte("block hash"),
								},
							},
						},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: bob, Sequence: 101, Round: 1},
						BlockHash: []byte("block hash"),
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: chris, Sequence: 101, Round: 1},
						BlockHash: []byte("block hash"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: bob, Sequence: 101, Round: 1},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("bob seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: chris, Sequence: 101, Round: 1},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("chris seal"),
					},
				})),
			),
		},

		{
			name: "future rcc triggers round jump",
			expected: &SequenceResult{
				Round:    3,
				Proposal: []byte("chris' round 3 proposal"),
				Seals: []CommitSeal{
					{
						From: alice,
						Seal: []byte("alice seal"),
					},
					{
						From: bob,
						Seal: []byte("bob seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(MockKeccak(func(_ []byte) []byte {
					return []byte("block hash")
				})),
				WithSignatureVerifier(MockSignatureVerifier(func(_ []byte, _ []byte, _ []byte) error {
					return nil
				})),
				WithValidator(MockValidator{
					address:           alice,
					signFn:            func(_ []byte) []byte { return nil },
					isValidProposalFn: func(_ uint64, _ []byte) bool { return true },
				}),
				WithValidatorSet(MockValidatorSet{
					isValidatorFn: func(v []byte, _ uint64) bool {
						return true
					},
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, chris) && round == 3
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(NewMockFeed([]message.Message{
					&message.MsgProposal{
						Info:      &message.MsgInfo{Sender: chris, Sequence: 101, Round: 3},
						BlockHash: []byte("block hash"),
						ProposedBlock: &message.ProposedBlock{
							Block: []byte("chris' round 3 proposal"),
							Round: 3,
						},
						RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.MsgRoundChange{
							{
								Info: &message.MsgInfo{Sender: bob, Sequence: 101, Round: 3},
							},
							{
								Info: &message.MsgInfo{Sender: chris, Sequence: 101, Round: 3},
							},
						}},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: alice, Sequence: 101, Round: 3},
						BlockHash: []byte("block hash"),
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: bob, Sequence: 101, Round: 3},
						BlockHash: []byte("block hash"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: alice, Sequence: 101, Round: 3},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("alice seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: bob, Sequence: 101, Round: 3},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("bob seal"),
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
						From: alice,
						Seal: []byte("alice seal"),
					},
					{
						From: nina,
						Seal: []byte("nina seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(MockKeccak(func(_ []byte) []byte {
					return []byte("block hash")
				})),
				WithSignatureVerifier(MockSignatureVerifier(func(_ []byte, _ []byte, _ []byte) error {
					return nil
				})),
				WithValidator(MockValidator{
					address:           alice,
					signFn:            func(_ []byte) []byte { return nil },
					isValidProposalFn: func(_ uint64, _ []byte) bool { return true },
				}),
				WithValidatorSet(MockValidatorSet{
					isValidatorFn: func(v []byte, _ uint64) bool {
						return true
					},
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, nina) && round == 5
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(NewMockFeed([]message.Message{
					&message.MsgProposal{
						Info:          &message.MsgInfo{Sender: nina, Sequence: 101, Round: 5},
						ProposedBlock: &message.ProposedBlock{Block: []byte("round 5 block"), Round: 5},
						BlockHash:     []byte("block hash"),
						RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.MsgRoundChange{
							{
								Info: &message.MsgInfo{Sequence: 101, Round: 5, Sender: chris},
							},

							{
								Info: &message.MsgInfo{Sequence: 101, Round: 5, Sender: bob},
							},
						}},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: alice, Sequence: 101, Round: 5},
						BlockHash: []byte("block hash"),
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: nina, Sequence: 101, Round: 5},
						BlockHash: []byte("block hash"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: alice, Sequence: 101, Round: 5},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("alice seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: nina, Sequence: 101, Round: 5},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("nina seal"),
					},
				})),
			),
		},

		{
			name: "round timer triggers round jump",
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("alice round 1 proposal"),
				Seals: []CommitSeal{
					{
						From: bob,
						Seal: []byte("bob seal"),
					},
					{
						From: chris,
						Seal: []byte("chris seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(MockKeccak(func(_ []byte) []byte {
					return []byte("block hash")
				})),
				WithSignatureVerifier(MockSignatureVerifier(func(_ []byte, _ []byte, _ []byte) error {
					return nil
				})),
				WithValidator(MockValidator{
					address:           alice,
					signFn:            func(_ []byte) []byte { return nil },
					isValidProposalFn: func(_ uint64, _ []byte) bool { return true },
					buildProposalFn: func(_ uint64) []byte {
						return []byte("alice round 1 proposal")
					},
				}),
				WithValidatorSet(MockValidatorSet{
					isValidatorFn: func(v []byte, _ uint64) bool {
						return true
					},
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, alice) && round == 1
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(NewMockFeed([]message.Message{
					&message.MsgRoundChange{
						Info: &message.MsgInfo{Sender: bob, Sequence: 101, Round: 1},
					},

					&message.MsgRoundChange{
						Info: &message.MsgInfo{Sender: chris, Sequence: 101, Round: 1},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: bob, Sequence: 101, Round: 1},
						BlockHash: []byte("block hash"),
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: chris, Sequence: 101, Round: 1},
						BlockHash: []byte("block hash"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: bob, Sequence: 101, Round: 1},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("bob seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: chris, Sequence: 101, Round: 1},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("chris seal"),
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
						From: alice,
						Seal: []byte("alice seal"),
					},

					{
						From: nina,
						Seal: []byte("nina seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(MockKeccak(func(_ []byte) []byte {
					return []byte("block hash")
				})),
				WithSignatureVerifier(MockSignatureVerifier(func(_ []byte, _ []byte, _ []byte) error {
					return nil
				})),
				WithValidator(MockValidator{
					address:           alice,
					signFn:            func(_ []byte) []byte { return nil },
					isValidProposalFn: func(_ uint64, _ []byte) bool { return true },
				}),
				WithValidatorSet(MockValidatorSet{
					isValidatorFn: func(v []byte, _ uint64) bool {
						return true
					},
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, bob) && round == 0 || bytes.Equal(v, chris) && round == 1
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(NewMockFeed([]message.Message{
					&message.MsgProposal{
						Info:          &message.MsgInfo{Sender: bob, Sequence: 101, Round: 0},
						BlockHash:     []byte("block hash"),
						ProposedBlock: &message.ProposedBlock{Block: []byte("round 0 block"), Round: 0},
					},

					&message.MsgProposal{
						Info:          &message.MsgInfo{Sender: chris, Sequence: 101, Round: 1},
						BlockHash:     []byte("block hash"),
						ProposedBlock: &message.ProposedBlock{Block: []byte("round 1 block"), Round: 1},
						RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.MsgRoundChange{
							{
								Info: &message.MsgInfo{Sequence: 101, Round: 1, Sender: alice},
							},

							{
								Info: &message.MsgInfo{Sequence: 101, Round: 1, Sender: chris},
							},
						}},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: alice, Sequence: 101, Round: 1},
						BlockHash: []byte("block hash"),
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: nina, Sequence: 101, Round: 1},
						BlockHash: []byte("block hash"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: alice, Sequence: 101, Round: 1},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("alice seal"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: nina, Sequence: 101, Round: 1},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("nina seal"),
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
						From: alice,
						Seal: []byte("alice seal"),
					},
					{
						From: bob,
						Seal: []byte("bob seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(MockKeccak(func(_ []byte) []byte {
					return []byte("block hash")
				})),
				WithSignatureVerifier(MockSignatureVerifier(func(_ []byte, _ []byte, _ []byte) error {
					return nil
				})),
				WithValidator(MockValidator{
					address:           alice,
					signFn:            func(_ []byte) []byte { return nil },
					isValidProposalFn: func(_ uint64, _ []byte) bool { return true },
				}),
				WithValidatorSet(MockValidatorSet{
					isValidatorFn: func(v []byte, _ uint64) bool {
						return true
					},
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, bob) && round == 0 || bytes.Equal(v, chris) && round == 1
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(NewSingleRoundMockFeed([]message.Message{
					/* round 0 */
					&message.MsgProposal{
						Info:          &message.MsgInfo{Sender: bob, Sequence: 101, Round: 0},
						BlockHash:     []byte("block hash"),
						ProposedBlock: &message.ProposedBlock{Block: []byte("round 0 block"), Round: 0},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: alice, Sequence: 101, Round: 0},
						BlockHash: []byte("block hash"),
					},
					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: bob, Sequence: 101, Round: 0},
						BlockHash: []byte("block hash"),
					},

					/* round 1 */

					&message.MsgProposal{
						Info:          &message.MsgInfo{Sender: chris, Sequence: 101, Round: 1},
						BlockHash:     []byte("block hash"),
						ProposedBlock: &message.ProposedBlock{Block: []byte("round 1 block"), Round: 1},
						RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.MsgRoundChange{
							{
								Info: &message.MsgInfo{Sender: alice, Sequence: 101, Round: 1},
							},
							{
								Info: &message.MsgInfo{Sender: bob, Sequence: 101, Round: 1},
							},
						}},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: alice, Sequence: 101, Round: 1},
						BlockHash: []byte("block hash"),
					},
					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: bob, Sequence: 101, Round: 1},
						BlockHash: []byte("block hash"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: alice, Sequence: 101, Round: 1},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("alice seal"),
					},
					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: bob, Sequence: 101, Round: 1},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("bob seal"),
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
						From: bob,
						Seal: []byte("bob seal"),
					},
					{
						From: alice,
						Seal: []byte("alice seal"),
					},
				},
			},
			cfg: NewConfig(
				WithRound0Duration(10*time.Millisecond),
				WithTransport(DummyTransport()),
				WithKeccak(MockKeccak(func(_ []byte) []byte {
					return []byte("block hash")
				})),
				WithSignatureVerifier(MockSignatureVerifier(func(_ []byte, _ []byte, _ []byte) error {
					return nil
				})),
				WithValidator(MockValidator{
					address:           alice,
					signFn:            func(_ []byte) []byte { return nil },
					isValidProposalFn: func(_ uint64, _ []byte) bool { return true },
				}),
				WithValidatorSet(MockValidatorSet{
					isValidatorFn: func(v []byte, _ uint64) bool {
						return true
					},
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, bob) && round == 0 ||
							bytes.Equal(v, alice) && round == 1 ||
							bytes.Equal(v, chris) && round == 2
					},
					hasQuorumFn: func(messages []message.Message) bool {
						return len(messages) >= 2
					},
				}),
				WithFeed(NewSingleRoundMockFeed([]message.Message{
					&message.MsgRoundChange{
						Info: &message.MsgInfo{Sender: chris, Sequence: 101, Round: 1},
					},

					&message.MsgProposal{
						Info:          &message.MsgInfo{Sender: chris, Sequence: 101, Round: 2},
						ProposedBlock: &message.ProposedBlock{Block: []byte("round 2 block"), Round: 2},
						BlockHash:     []byte("block hash"),
						RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.MsgRoundChange{
							{
								Info: &message.MsgInfo{Sender: bob, Sequence: 101, Round: 2},
							},
							{
								Info: &message.MsgInfo{Sender: chris, Sequence: 101, Round: 2},
							},
						}},
					},

					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: bob, Sequence: 101, Round: 2},
						BlockHash: []byte("block hash"),
					},
					&message.MsgPrepare{
						Info:      &message.MsgInfo{Sender: alice, Sequence: 101, Round: 2},
						BlockHash: []byte("block hash"),
					},

					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: bob, Sequence: 101, Round: 2},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("bob seal"),
					},
					&message.MsgCommit{
						Info:       &message.MsgInfo{Sender: alice, Sequence: 101, Round: 2},
						BlockHash:  []byte("block hash"),
						CommitSeal: []byte("alice seal"),
					},
				})),
			),
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := NewSequencer(tt.cfg).Finalize(context.Background(), 101)
			assert.True(t, reflect.DeepEqual(tt.expected, res))
		})
	}
}
