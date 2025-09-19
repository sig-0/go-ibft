//nolint:dupl // test cases not identical
package sequencer

import (
	"bytes"
	"context"
	"slices"
	"testing"
	"time"

	"github.com/sig-0/go-ibft/message"

	"github.com/stretchr/testify/assert"
)

func Test_SequencerFinalizeCancelled(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Validator: mockValidator{address: Alice},
		ValidatorSet: mockVerifier{isProposerFn: func(_ []byte, _ uint64, _ uint64) bool {
			return false
		}},
		Feed:           message.NewStore(),
		Round0Duration: 10 * time.Millisecond,
		Vrf:            allGoodVrf{},
	}

	s := NewSequencer(cfg)
	s.proposerAlgo = mockProposerAlgo(func(ctx context.Context, sequence, round uint64) ([]byte, error) {
		return Bob, nil
	})

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
		algo     ProposerSelector
		vrf      Vrf
		cfg      Config
		name     string
		messages []any
	}{
		{
			name: "proposal is accepted in round 0",
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

			vrf: allGoodVrf{},
			algo: mockProposerAlgo(func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				return Bob, nil
			}),
			messages: []any{
				&message.Proposal{
					Sender:   Bob,
					Sequence: 101,
					Round:    0,
					ProposedBlock: &message.ProposedBlock{
						Block: []byte("Bob's proposal"),
						Round: 0,
					},
				},
				&message.Prepare{
					Sender:   Alice,
					Sequence: 101,
					Round:    0,
				},

				&message.Prepare{
					Sender:   Chris,
					Sequence: 101,
					Round:    0,
				},

				&message.Commit{
					Sender:     Alice,
					Sequence:   101,
					Round:      0,
					CommitSeal: []byte("Alice seal"),
				},
				&message.Commit{
					Sender:     Chris,
					Sequence:   101,
					Round:      0,
					CommitSeal: []byte("Chris seal"),
				},
			},

			cfg: Config{
				Validator: mockValidator{
					address: Alice,
					signFn:  DummySignFn,
				},
				//ValidatorSet: mockVerifier{
				//	isValidatorFn: AlwaysAValidator,
				//	isProposerFn: func(v []byte, _ uint64, round uint64) bool {
				//		return bytes.Equal(v, Bob) && round == 0
				//	},
				//	hasQuorumFn: func(messages [][]byte, _ uint64) bool {
				//		return len(messages) >= 2
				//	},
				//	isValidProposalFn:  AlwaysValidProposal,
				//	isValidSignatureFn: AlwaysValidSignature,
				//},
				Transport:      dummyTransport{},
				Keccak:         DummyKeccak,
				Round0Duration: 10 * time.Millisecond,
				Feed:           message.NewStore(),
			},
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

			vrf: allGoodVrf{},

			messages: []any{
				&message.Prepare{
					Sender: Bob, Sequence: 101, Round: 0,
					BlockHash: DummyKeccakValue,
				},

				&message.Prepare{
					Sender: Chris, Sequence: 101, Round: 0,
					BlockHash: DummyKeccakValue,
				},

				&message.Commit{
					Sender: Bob, Sequence: 101, Round: 0,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Bob seal"),
				},

				&message.Commit{
					Sender: Chris, Sequence: 101, Round: 0,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Chris seal"),
				},
			},

			algo: mockProposerAlgo(func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				return Alice, nil
			}),
			cfg: Config{
				Validator: mockValidator{
					address: Alice,
					signFn:  DummySignFn,
					buildProposalFn: func(_ uint64) []byte {
						return []byte("Alice's proposal")
					},
				},
				ValidatorSet: mockVerifier{
					isValidatorFn: AlwaysAValidator,
					//isProposerFn: func(v []byte, _ uint64, round uint64) bool {
					//	return bytes.Equal(v, Alice) && round == 0
					//},
					hasQuorumFn: func(messages [][]byte, _ uint64) bool {
						return len(messages) >= 2
					},
					isValidProposalFn:  AlwaysValidProposal,
					isValidSignatureFn: AlwaysValidSignature,
				},
				Transport:      dummyTransport{},
				Keccak:         DummyKeccak,
				Round0Duration: 10 * time.Millisecond,
				Feed:           message.NewStore(),
			},
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

			algo: mockProposerAlgo(func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				return Bob, nil
			}),

			vrf: allGoodVrf{},

			cfg: Config{
				Validator: mockValidator{
					address: Alice,
					signFn:  DummySignFn,
				},
				ValidatorSet: mockVerifier{
					isValidProposalFn: AlwaysValidProposal,

					isValidatorFn: AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Bob) && round == 1
					},
					hasQuorumFn: func(messages [][]byte, _ uint64) bool {
						return len(messages) >= 2
					},
					isValidSignatureFn: AlwaysValidSignature,
				},
				Transport:      dummyTransport{},
				Feed:           message.NewStore(),
				Keccak:         DummyKeccak,
				Round0Duration: 10 * time.Millisecond,
			},

			messages: []any{
				&message.Proposal{
					Sender: Bob, Sequence: 101, Round: 1,
					BlockHash:     DummyKeccakValue,
					ProposedBlock: &message.ProposedBlock{Block: []byte("Bob's round 1 proposal"), Round: 1},
					RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
						{
							Sequence: 101, Round: 1, Sender: Alice,
						},

						{
							Sequence: 101, Round: 1, Sender: Chris,
						},
					}},
				},

				&message.Prepare{
					Sender: Alice, Sequence: 101, Round: 1,
					BlockHash: DummyKeccakValue,
				},

				&message.Prepare{
					Sender: Chris, Sequence: 101, Round: 1,
					BlockHash: DummyKeccakValue,
				},

				&message.Commit{
					Sender: Alice, Sequence: 101, Round: 1,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Alice seal"),
				},

				&message.Commit{
					Sender: Chris, Sequence: 101, Round: 1,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Chris seal"),
				},
			},
		},

		{
			name: "Alice jumps to round 1 proposal and accepts it",
			algo: mockProposerAlgo(func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				return Chris, nil
			}),
			vrf: allGoodVrf{},
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

			cfg: Config{
				Validator: mockValidator{
					address: Alice,
					signFn:  DummySignFn,
				},
				ValidatorSet: mockVerifier{
					isValidProposalFn: AlwaysValidProposal,

					isValidatorFn: AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Chris) && round == 0 || bytes.Equal(v, Bob) && round == 1
					},
					hasQuorumFn: func(messages [][]byte, _ uint64) bool {
						return len(messages) >= 2
					},
					isValidSignatureFn: AlwaysValidSignature,
				},
				Transport:      dummyTransport{},
				Feed:           message.NewStore(),
				Keccak:         DummyKeccak,
				Round0Duration: 10 * time.Millisecond,
			},

			messages: []any{
				&message.Proposal{
					Sender: Bob, Sequence: 101, Round: 1,
					BlockHash:     DummyKeccakValue,
					ProposedBlock: &message.ProposedBlock{Block: []byte("Chris' proposal"), Round: 1},
					RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
						{
							Sender: Bob, Sequence: 101, Round: 1,
							LatestPreparedCertificate: &message.PreparedCertificate{
								ProposalMessage: &message.Proposal{
									Sender: Chris, Sequence: 101, Round: 0,
									BlockHash: DummyKeccakValue,
									ProposedBlock: &message.ProposedBlock{
										Block: []byte("Chris' proposal"),
										Round: 0,
									},
								},
								PrepareMessages: []*message.Prepare{
									{
										Sender: Bob, Sequence: 101, Round: 0,
										BlockHash: DummyKeccakValue,
									},
									{
										Sender: Nina, Sequence: 101, Round: 0,
										BlockHash: DummyKeccakValue,
									},
								},
							},
						},
						{
							Sender: Nina, Sequence: 101, Round: 1,
							LatestPreparedCertificate: &message.PreparedCertificate{
								ProposalMessage: &message.Proposal{
									Sender: Chris, Sequence: 101, Round: 0,
									BlockHash: DummyKeccakValue,
									ProposedBlock: &message.ProposedBlock{
										Block: []byte("Chris' proposal"),
										Round: 0,
									},
								},
								PrepareMessages: []*message.Prepare{
									{
										Sender: Bob, Sequence: 101, Round: 0,
										BlockHash: DummyKeccakValue,
									},
									{
										Sender: Nina, Sequence: 101, Round: 0,
										BlockHash: DummyKeccakValue,
									},
								},
							},
						},
					}},
				},

				&message.Prepare{
					Sender: Alice, Sequence: 101, Round: 1,
					BlockHash: DummyKeccakValue,
				},

				&message.Prepare{
					Sender: Chris, Sequence: 101, Round: 1,
					BlockHash: DummyKeccakValue,
				},

				&message.Commit{
					Sender: Alice, Sequence: 101, Round: 1,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Alice seal"),
				},

				&message.Commit{
					Sender: Chris, Sequence: 101, Round: 1,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Chris seal"),
				},
			},
		},

		{
			name: "block proposed in round 1",

			algo: mockProposerAlgo(func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				if round == 1 {
					return Alice, nil
				}

				return Nina, nil
			}),
			vrf: allGoodVrf{},
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

			cfg: Config{
				Validator: mockValidator{
					address: Alice,
					signFn:  DummySignFn,
					buildProposalFn: func(_ uint64) []byte {
						return []byte("Alice's round 1 proposal")
					},
				},
				ValidatorSet: mockVerifier{
					isValidProposalFn: AlwaysValidProposal,

					isValidatorFn: AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Bob) && round == 0 || bytes.Equal(v, Alice) && round == 1
					},
					hasQuorumFn: func(messages [][]byte, _ uint64) bool {
						return len(messages) >= 2
					},
					isValidSignatureFn: AlwaysValidSignature,
				},
				Transport:      dummyTransport{},
				Feed:           message.NewStore(),
				Keccak:         DummyKeccak,
				Round0Duration: 10 * time.Millisecond,
			},

			messages: []any{
				// need to justify Alice's proposal for round 1
				&message.RoundChange{
					Sequence: 101, Round: 1, Sender: Alice,
				},

				&message.RoundChange{
					Sequence: 101, Round: 1, Sender: Nina,
				},

				&message.Prepare{
					Sequence: 101, Round: 1, Sender: Bob,
					BlockHash: DummyKeccakValue,
				},

				&message.Prepare{
					Sequence: 101, Round: 1, Sender: Nina,
					BlockHash: DummyKeccakValue,
				},

				&message.Commit{
					Sequence: 101, Round: 1, Sender: Bob,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Bob seal"),
				},

				&message.Commit{
					Sequence: 101, Round: 1, Sender: Nina,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Nina seal"),
				},
			},
		},

		{
			name: "old block proposed in round 1",
			algo: mockProposerAlgo(func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				if round == 1 {
					return Alice, nil
				}

				return Nina, nil
			}),
			vrf: allGoodVrf{},
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

			cfg: Config{
				Validator: mockValidator{
					address: Alice,
					signFn:  DummySignFn,
				},
				ValidatorSet: mockVerifier{
					isValidProposalFn: AlwaysValidProposal,

					isValidatorFn: AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Bob) && round == 0 || bytes.Equal(v, Alice) && round == 1
					},
					hasQuorumFn: func(messages [][]byte, _ uint64) bool {
						return len(messages) >= 2
					},
					isValidSignatureFn: AlwaysValidSignature,
				},
				Transport:      dummyTransport{},
				Feed:           message.NewStore(),
				Keccak:         DummyKeccak,
				Round0Duration: 10 * time.Millisecond,
			},
			messages: []any{
				&message.RoundChange{
					Sender:   Chris,
					Sequence: 101,
					Round:    1,
					LatestPreparedProposedBlock: &message.ProposedBlock{
						Block: []byte("Bob's round 0 proposal"),
						Round: 0,
					},
					LatestPreparedCertificate: &message.PreparedCertificate{
						ProposalMessage: &message.Proposal{
							Sender:    Bob,
							Sequence:  101,
							Round:     0,
							BlockHash: DummyKeccakValue,
							ProposedBlock: &message.ProposedBlock{
								Block: []byte("Bob's round 0 proposal"),
								Round: 0,
							},
						},

						PrepareMessages: []*message.Prepare{
							{
								Sender:    Chris,
								Sequence:  101,
								Round:     0,
								BlockHash: DummyKeccakValue,
							},
							{
								Sender:    Nina,
								Sequence:  101,
								Round:     0,
								BlockHash: DummyKeccakValue,
							},
						},
					},
				},

				&message.RoundChange{
					Sender:   Nina,
					Sequence: 101,
					Round:    1,
					LatestPreparedProposedBlock: &message.ProposedBlock{
						Block: []byte("Bob's round 0 proposal"),
						Round: 0,
					},
					LatestPreparedCertificate: &message.PreparedCertificate{
						ProposalMessage: &message.Proposal{
							Sender:    Bob,
							Sequence:  101,
							Round:     0,
							BlockHash: DummyKeccakValue,
							ProposedBlock: &message.ProposedBlock{
								Block: []byte("Bob's round 0 proposal"),
								Round: 0,
							},
						},

						PrepareMessages: []*message.Prepare{
							{
								Sender:    Chris,
								Sequence:  101,
								Round:     0,
								BlockHash: DummyKeccakValue,
							},
							{
								Sender:    Nina,
								Sequence:  101,
								Round:     0,
								BlockHash: DummyKeccakValue,
							},
						},
					},
				},

				&message.Prepare{
					Sender:    Bob,
					Sequence:  101,
					Round:     1,
					BlockHash: DummyKeccakValue,
				},

				&message.Prepare{
					Sender:    Chris,
					Sequence:  101,
					Round:     1,
					BlockHash: DummyKeccakValue,
				},

				&message.Commit{
					Sender:     Bob,
					Sequence:   101,
					Round:      1,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Bob seal"),
				},

				&message.Commit{
					Sender:     Chris,
					Sequence:   101,
					Round:      1,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Chris seal"),
				},
			},
		},

		{
			name: "future rcc triggers round jump",
			algo: mockProposerAlgo(func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				if round == 3 {
					return Alice, nil
				}

				return Nina, nil
			}),
			vrf: allGoodVrf{},
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

			cfg: Config{
				Validator: mockValidator{
					address: Alice,
					signFn:  DummySignFn,
					buildProposalFn: func(_ uint64) []byte {
						return []byte("Alice round 3 proposal")
					},
				},
				ValidatorSet: mockVerifier{
					isValidProposalFn: AlwaysValidProposal,
					isValidatorFn:     AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Alice) && round == 3
					},
					hasQuorumFn: func(messages [][]byte, _ uint64) bool {
						return len(messages) >= 2
					},
					isValidSignatureFn: AlwaysValidSignature,
				},
				Transport:      dummyTransport{},
				Feed:           message.NewStore(),
				Keccak:         DummyKeccak,
				Round0Duration: 10 * time.Millisecond,
			},

			messages: []any{
				&message.RoundChange{
					Sender:   Bob,
					Sequence: 101,
					Round:    3,
				},

				&message.RoundChange{
					Sender:   Chris,
					Sequence: 101,
					Round:    3,
				}, &message.Prepare{

					Sender:   Bob,
					Sequence: 101,
					Round:    3,

					BlockHash: DummyKeccakValue,
				},

				&message.Prepare{

					Sender:   Chris,
					Sequence: 101,
					Round:    3,

					BlockHash: DummyKeccakValue,
				},

				&message.Commit{

					Sender:   Bob,
					Sequence: 101,
					Round:    3,

					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Bob seal"),
				},

				&message.Commit{

					Sender:   Chris,
					Sequence: 101,
					Round:    3,

					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Chris seal"),
				},
			},
		},

		{
			name: "future proposal triggers round jump",
			algo: mockProposerAlgo(func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				if round == 5 {
					return Nina, nil
				}

				return Bob, nil
			}),
			vrf: allGoodVrf{},
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

			cfg: Config{
				Validator: mockValidator{
					address: Alice,
					signFn:  DummySignFn,
				},
				ValidatorSet: mockVerifier{
					isValidProposalFn: AlwaysValidProposal,
					isValidatorFn:     AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Nina) && round == 5
					},
					hasQuorumFn: func(messages [][]byte, _ uint64) bool {
						return len(messages) >= 2
					},
					isValidSignatureFn: AlwaysValidSignature,
				},
				Transport:      dummyTransport{},
				Feed:           message.NewStore(),
				Keccak:         DummyKeccak,
				Round0Duration: 10 * time.Millisecond,
			},

			messages: []any{
				&message.Proposal{
					Sender: Nina, Sequence: 101, Round: 5,
					ProposedBlock: &message.ProposedBlock{Block: []byte("round 5 block"), Round: 5},
					BlockHash:     DummyKeccakValue,
					RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
						{
							Sequence: 101, Round: 5, Sender: Chris,
						},

						{
							Sequence: 101, Round: 5, Sender: Bob,
						},
					}},
				},

				&message.Prepare{
					Sender: Alice, Sequence: 101, Round: 5,
					BlockHash: DummyKeccakValue,
				},

				&message.Prepare{
					Sender: Nina, Sequence: 101, Round: 5,
					BlockHash: DummyKeccakValue,
				},

				&message.Commit{
					Sender: Alice, Sequence: 101, Round: 5,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Alice seal"),
				},

				&message.Commit{
					Sender: Nina, Sequence: 101, Round: 5,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Nina seal"),
				},
			},
		},

		{
			name: "round timer triggers round jump",
			algo: mockProposerAlgo(func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				if round == 1 {
					return Alice, nil
				}

				return Nina, nil
			}),
			vrf: allGoodVrf{},
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

			cfg: Config{
				Validator: mockValidator{
					address: Alice,
					signFn:  DummySignFn,
					buildProposalFn: func(_ uint64) []byte {
						return []byte("Alice round 1 proposal")
					},
				},
				ValidatorSet: mockVerifier{
					isValidProposalFn: AlwaysValidProposal,
					isValidatorFn:     AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Alice) && round == 1
					},
					hasQuorumFn: func(messages [][]byte, _ uint64) bool {
						// set to 1 so Alice's own round change msg trigger the right build flow
						return len(messages) >= 1
					},
					isValidSignatureFn: AlwaysValidSignature,
				},
				Transport:      dummyTransport{},
				Feed:           message.NewStore(),
				Keccak:         DummyKeccak,
				Round0Duration: 10 * time.Millisecond,
			},

			messages: []any{
				//&message.RoundChange{
				//	// no way to mock the
				//	Sender:   Alice,
				//	Sequence: 101,
				//	Round:    1,
				//},

				//&message.RoundChange{
				//	Sender: Chris, Sequence: 101, Round: 1,
				//},

				&message.Prepare{
					Sender: Bob, Sequence: 101, Round: 1,
					BlockHash: DummyKeccakValue,
				},

				&message.Prepare{
					Sender: Chris, Sequence: 101, Round: 1,
					BlockHash: DummyKeccakValue,
				},

				&message.Commit{
					Sender: Bob, Sequence: 101, Round: 1,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Bob seal"),
				},

				&message.Commit{
					Sender: Chris, Sequence: 101, Round: 1,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Chris seal"),
				},
			},
		},

		{
			name: "no prepare messages in round 0",
			algo: mockProposerAlgo(func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				if round == 0 {
					return Bob, nil
				}

				if round == 1 {
					return Chris, nil
				}

				return Nina, nil
			}),
			vrf: allGoodVrf{},
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

			cfg: Config{
				Keccak: DummyKeccak,
				Validator: mockValidator{
					address: Alice,
					signFn:  DummySignFn,
				},
				ValidatorSet: mockVerifier{
					isValidProposalFn: AlwaysValidProposal,
					isValidatorFn:     AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Bob) && round == 0 || bytes.Equal(v, Chris) && round == 1
					},
					hasQuorumFn: func(messages [][]byte, _ uint64) bool {
						return len(messages) >= 2
					},
					isValidSignatureFn: AlwaysValidSignature,
				},
				Transport:      dummyTransport{},
				Feed:           message.NewStore(),
				Round0Duration: 10 * time.Millisecond,
			},

			messages: []any{
				&message.Proposal{
					Sender:    Bob,
					Sequence:  101,
					Round:     0,
					BlockHash: DummyKeccakValue,
					ProposedBlock: &message.ProposedBlock{
						Block: []byte("round 0 block"),
						Round: 0,
					},
				},

				&message.Prepare{
					Sender:    Alice,
					Sequence:  101,
					Round:     0,
					BlockHash: DummyKeccakValue,
				},

				&message.Proposal{
					Sender: Chris, Sequence: 101, Round: 1,
					BlockHash:     DummyKeccakValue,
					ProposedBlock: &message.ProposedBlock{Block: []byte("round 1 block"), Round: 1},
					RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
						{
							Sequence: 101, Round: 1, Sender: Alice,
						},

						{
							Sequence: 101, Round: 1, Sender: Chris,
						},
					}},
				},

				&message.Prepare{
					Sender: Alice, Sequence: 101, Round: 1,
					BlockHash: DummyKeccakValue,
				},

				&message.Prepare{
					Sender: Nina, Sequence: 101, Round: 1,
					BlockHash: DummyKeccakValue,
				},

				&message.Commit{
					Sender: Alice, Sequence: 101, Round: 1,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Alice seal"),
				},

				&message.Commit{
					Sender: Nina, Sequence: 101, Round: 1,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Nina seal"),
				},
			},
		},

		{
			name: "no commit messages in round 0",
			algo: mockProposerAlgo(func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				if round == 0 {
					return Bob, nil
				}

				if round == 1 {
					return Chris, nil
				}

				return Nina, nil
			}),
			vrf: allGoodVrf{},
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

			cfg: Config{
				Keccak:         DummyKeccak,
				Round0Duration: 10 * time.Millisecond,
				Validator: mockValidator{
					address: Alice,
					signFn:  DummySignFn,
				},
				ValidatorSet: mockVerifier{
					isValidSignatureFn: AlwaysValidSignature,
					isValidProposalFn:  AlwaysValidProposal,
					isValidatorFn:      AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Bob) && round == 0 || bytes.Equal(v, Chris) && round == 1
					},
					hasQuorumFn: func(messages [][]byte, _ uint64) bool {
						return len(messages) >= 2
					},
				},
				Transport: dummyTransport{},
				Feed:      message.NewStore(),
			},

			messages: []any{
				// round 0
				&message.Proposal{
					Sender: Bob, Sequence: 101, Round: 0,
					BlockHash:     DummyKeccakValue,
					ProposedBlock: &message.ProposedBlock{Block: []byte("round 0 block"), Round: 0},
				},

				&message.Prepare{
					Sender: Alice, Sequence: 101, Round: 0,
					BlockHash: DummyKeccakValue,
				},
				&message.Prepare{
					Sender: Bob, Sequence: 101, Round: 0,
					BlockHash: DummyKeccakValue,
				},

				// round 1

				&message.Proposal{
					Sender: Chris, Sequence: 101, Round: 1,
					BlockHash:     DummyKeccakValue,
					ProposedBlock: &message.ProposedBlock{Block: []byte("round 1 block"), Round: 1},
					RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
						{
							Sender: Alice, Sequence: 101, Round: 1,
						},
						{
							Sender: Bob, Sequence: 101, Round: 1,
						},
					}},
				},

				&message.Prepare{
					Sender: Alice, Sequence: 101, Round: 1,
					BlockHash: DummyKeccakValue,
				},
				&message.Prepare{
					Sender: Bob, Sequence: 101, Round: 1,
					BlockHash: DummyKeccakValue,
				},

				&message.Commit{
					Sender: Alice, Sequence: 101, Round: 1,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Alice seal"),
				},
				&message.Commit{
					Sender: Bob, Sequence: 101, Round: 1,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Bob seal"),
				},
			},
		},

		{
			name: "round 0 proposer fails to build block",
			algo: mockProposerAlgo(func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				if round == 0 {
					return Bob, nil
				}

				if round == 1 {
					return Alice, nil
				}

				return Chris, nil
			}),
			vrf: allGoodVrf{},
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

			cfg: Config{
				Validator: mockValidator{
					address: Alice,
					signFn:  DummySignFn,
				},
				ValidatorSet: mockVerifier{
					isValidProposalFn: AlwaysValidProposal,
					isValidatorFn:     AlwaysAValidator,
					isProposerFn: func(v []byte, _ uint64, round uint64) bool {
						return bytes.Equal(v, Bob) && round == 0 ||
							bytes.Equal(v, Nina) && round == 1 ||
							bytes.Equal(v, Chris) && round == 2
					},
					hasQuorumFn: func(messages [][]byte, _ uint64) bool {
						return len(messages) >= 2
					},
					isValidSignatureFn: AlwaysValidSignature,
				},
				Transport:      dummyTransport{},
				Feed:           message.NewStore(),
				Keccak:         DummyKeccak,
				Round0Duration: 10 * time.Millisecond,
			},

			messages: []any{
				//&message.RoundChange{
				//	Sender:   Chris,
				//	Sequence: 101,
				//	Round:    1,
				//},

				&message.Proposal{
					Sender:        Chris,
					Sequence:      101,
					Round:         2,
					ProposedBlock: &message.ProposedBlock{Block: []byte("round 2 block"), Round: 2},
					BlockHash:     DummyKeccakValue,
					RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
						{
							Sender:   Bob,
							Sequence: 101,
							Round:    2,
						},
						{
							Sender:   Chris,
							Sequence: 101,
							Round:    2,
						},
					}},
				},

				&message.Prepare{
					Sender: Bob, Sequence: 101, Round: 2,
					BlockHash: DummyKeccakValue,
				},
				&message.Prepare{
					Sender: Alice, Sequence: 101, Round: 2,
					BlockHash: DummyKeccakValue,
				},

				&message.Commit{
					Sender: Bob, Sequence: 101, Round: 2,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Bob seal"),
				},
				&message.Commit{
					Sender: Alice, Sequence: 101, Round: 2,
					BlockHash:  DummyKeccakValue,
					CommitSeal: []byte("Alice seal"),
				},
			},
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.cfg.Vrf = tt.vrf

			s := NewSequencer(tt.cfg)
			s.proposerAlgo = tt.algo

			for _, m := range tt.messages {
				switch m := m.(type) {
				case *message.RoundChange:
					s.feed.RoundChangeMessages.Add(m)
				case *message.Prepare:
					s.feed.PrepareMessages.Add(m)
				case *message.Commit:
					s.feed.CommitMessages.Add(m)
				case *message.Proposal:
					s.feed.ProposalMessages.Add(m)
				}
			}

			res := s.Finalize(context.Background(), 101)
			//assert.True(t, reflect.DeepEqual(tt.expected, res), "expected %#v, got %#v", tt.expected, res)
			assert.EqualValues(t, tt.expected.Round, res.Round)
			assert.Equal(t, tt.expected.Proposal, res.Proposal)

			slices.SortFunc(tt.expected.Seals, func(a, b CommitSeal) int {
				return slices.Compare(a.From, b.From)
			})
			slices.SortFunc(res.Seals, func(a, b CommitSeal) int {
				return slices.Compare(a.From, b.From)
			})

			assert.EqualValues(t, tt.expected.Seals, res.Seals)
		})
	}
}
