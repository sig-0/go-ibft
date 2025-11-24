//nolint:dupl // test cases not identical
package sequencer

import (
	"context"
	"log/slog"
	"slices"
	"testing"
	"time"

	"github.com/sig-0/go-ibft/message"

	"github.com/stretchr/testify/assert"
)

func Test_SequencerFinalizeCancelled(t *testing.T) {
	t.Parallel()

	var (
		v              = Alice
		round0Duration = 10 * time.Millisecond
		consensus      = allGoodConsensus{getProposer: func(_ context.Context, _, _ uint64) ([]byte, error) {
			return Bob.Address(), nil
		}}
	)

	s := NewSequencer(slog.Default(), consensus, v, nil, round0Duration)

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan *SequenceResult)

	go func(ctx context.Context) {
		defer close(ch)

		ch <- s.Finalize(ctx, 101, message.NewStore())
	}(ctx)

	cancel()

	assert.Nil(t, <-ch)
}

func Test_SequencerFinalize(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name      string
		validator Validator
		consensus Consensus
		messages  []any
		expected  *SequenceResult
	}{
		{
			name: "proposal is accepted in round 0",
			consensus: allGoodConsensus{getProposer: func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				return Bob.Address(), nil
			}},
			validator: Alice,
			expected: &SequenceResult{
				Round:    0,
				Proposal: []byte("bob_proposal"),
				Seals: []CommitSeal{
					{
						From: Alice.Address(),
						Seal: []byte("alice_sig"),
					},
					{
						From: Chris.Address(),
						Seal: []byte("chris_sig"),
					},
				},
			},
			messages: []any{
				&message.Proposal{
					Sender:   Bob.Address(),
					Sequence: 101,
					Round:    0,
					ProposedBlock: &message.ProposedBlock{
						Block: []byte("bob_proposal"),
						Round: 0,
					},
				},
				&message.Prepare{
					Sender:   Alice.Address(),
					Sequence: 101,
					Round:    0,
				},
				&message.Prepare{
					Sender:   Chris.Address(),
					Sequence: 101,
					Round:    0,
				},

				&message.Commit{
					Sender:     Alice.Address(),
					Sequence:   101,
					Round:      0,
					CommitSeal: []byte("alice_sig"),
				},
				&message.Commit{
					Sender:     Chris.Address(),
					Sequence:   101,
					Round:      0,
					CommitSeal: []byte("chris_sig"),
				},
			},
		},

		{
			name: "Bob and Chris accept Alice's proposal in round 0",
			consensus: allGoodConsensus{getProposer: func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				return Alice.Address(), nil
			}},
			validator: Alice,
			expected: &SequenceResult{
				Round:    0,
				Proposal: []byte("alice_proposal"),
				Seals: []CommitSeal{
					{
						From: Alice.Address(),
						Seal: []byte("alice_sig"),
					},
					{
						From: Bob.Address(),
						Seal: []byte("bob_sig"),
					},
					{
						From: Chris.Address(),
						Seal: []byte("chris_sig"),
					},
				},
			},

			messages: []any{
				&message.Prepare{
					Sender:    Bob.Address(),
					Sequence:  101,
					Round:     0,
					Signature: Alice.Sign(nil),
				},

				&message.Prepare{
					Sender:    Chris.Address(),
					Sequence:  101,
					Round:     0,
					Signature: Chris.Sign(nil),
				},

				&message.Commit{
					Sender:     Bob.Address(),
					Sequence:   101,
					Round:      0,
					CommitSeal: []byte("bob_sig"),
					Signature:  Bob.Sign(nil),
				},

				&message.Commit{
					Sender:     Chris.Address(),
					Sequence:   101,
					Round:      0,
					CommitSeal: []byte("chris_sig"),
					Signature:  Chris.Sign(nil),
				},
			},
		},

		{
			name: "Alice and Chris accept Bob's proposal in round 1 due to round change",
			consensus: allGoodConsensus{blockFutureProposal: true, getProposer: func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				return Bob.Address(), nil
			}},
			validator: Alice,
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("bob_proposal"),
				Seals: []CommitSeal{
					{
						From: Alice.Address(),
						Seal: []byte("alice_sig"),
					},
					{
						From: Chris.Address(),
						Seal: []byte("chris_sig"),
					},
				},
			},

			messages: []any{
				&message.Proposal{
					Sender:   Bob.Address(),
					Sequence: 101,
					Round:    1,
					ProposedBlock: &message.ProposedBlock{
						Block: []byte("bob_proposal"),
						Round: 1,
					},
					RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
						{
							Sender:   Alice.Address(),
							Sequence: 101,
							Round:    1,
						},

						{
							Sender:   Chris.Address(),
							Sequence: 101,
							Round:    1,
						},
					}},
				},

				&message.Prepare{
					Sender:   Alice.Address(),
					Sequence: 101,
					Round:    1,
				},

				&message.Prepare{
					Sender:   Chris.Address(),
					Sequence: 101,
					Round:    1,
				},

				&message.Commit{
					Sender:     Alice.Address(),
					Sequence:   101,
					Round:      1,
					CommitSeal: []byte("alice_sig"),
				},

				&message.Commit{
					Sender:     Chris.Address(),
					Sequence:   101,
					Round:      1,
					CommitSeal: []byte("chris_sig"),
				},
			},
		},

		{
			name: "Alice jumps to round 1 proposal and accepts it",
			consensus: allGoodConsensus{getProposer: func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				return Chris.Address(), nil
			}},
			validator: Alice,
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("chris_proposal"),
				Seals: []CommitSeal{
					{
						From: Alice.Address(),
						Seal: []byte("alice_sig"),
					},
					{
						From: Chris.Address(),
						Seal: []byte("chris_sig"),
					},
				},
			},

			messages: []any{
				&message.Proposal{
					Sender:   Bob.Address(),
					Sequence: 101,
					Round:    1,
					ProposedBlock: &message.ProposedBlock{
						Block: []byte("chris_proposal"),
						Round: 1,
					},
					RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
						{
							Sender:   Bob.Address(),
							Sequence: 101,
							Round:    1,
							LatestPreparedCertificate: &message.PreparedCertificate{
								ProposalMessage: &message.Proposal{
									Sender:   Chris.Address(),
									Sequence: 101,
									Round:    0,
									ProposedBlock: &message.ProposedBlock{
										Block: []byte("chris_proposal"),
										Round: 0,
									},
								},
								PrepareMessages: []*message.Prepare{
									{
										Sender:   Bob.Address(),
										Sequence: 101,
										Round:    0,
									},
									{
										Sender:   Nina.Address(),
										Sequence: 101,
										Round:    0,
									},
								},
							},
						},
						{
							Sender:   Nina.Address(),
							Sequence: 101,
							Round:    1,
							LatestPreparedCertificate: &message.PreparedCertificate{
								ProposalMessage: &message.Proposal{
									Sender:   Chris.Address(),
									Sequence: 101,
									Round:    0,
									ProposedBlock: &message.ProposedBlock{
										Block: []byte("chris_proposal"),
										Round: 0,
									},
								},
								PrepareMessages: []*message.Prepare{
									{
										Sender:   Bob.Address(),
										Sequence: 101,
										Round:    0,
									},
									{
										Sender:   Nina.Address(),
										Sequence: 101,
										Round:    0,
									},
								},
							},
						},
					}},
				},

				&message.Prepare{
					Sender:   Alice.Address(),
					Sequence: 101,
					Round:    1,
				},

				&message.Prepare{
					Sender:   Chris.Address(),
					Sequence: 101,
					Round:    1,
				},

				&message.Commit{
					Sender:     Alice.Address(),
					Sequence:   101,
					Round:      1,
					CommitSeal: []byte("alice_sig"),
				},

				&message.Commit{
					Sender:     Chris.Address(),
					Sequence:   101,
					Round:      1,
					CommitSeal: []byte("chris_sig"),
				},
			},
		},

		{
			name: "block proposed in round 1",
			consensus: allGoodConsensus{blockFutureRCC: true, getProposer: func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				if round == 1 {
					return Alice.Address(), nil
				}

				return Nina.Address(), nil
			}},
			validator: Alice,
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("alice_proposal"),
				Seals: []CommitSeal{
					{
						From: Alice.Address(),
						Seal: []byte("alice_sig"),
					},
					{
						From: Bob.Address(),
						Seal: []byte("bob_sig"),
					},

					{
						From: Nina.Address(),
						Seal: []byte("nina_sig"),
					},
				},
			},

			messages: []any{
				// need to justify Alice's proposal for round 1
				&message.RoundChange{
					Sender:    Alice.Address(),
					Sequence:  101,
					Round:     1,
					Signature: []byte("alice_sig"),
				},

				&message.RoundChange{
					Sender:    Nina.Address(),
					Sequence:  101,
					Round:     1,
					Signature: []byte("nina_sig"),
				},

				&message.Prepare{
					Sender:    Bob.Address(),
					Sequence:  101,
					Round:     1,
					Signature: []byte("bob_sig"),
				},

				&message.Prepare{
					Sender:    Nina.Address(),
					Sequence:  101,
					Round:     1,
					Signature: []byte("nina_sig"),
				},

				&message.Commit{
					Sender:     Bob.Address(),
					Sequence:   101,
					Round:      1,
					CommitSeal: []byte("bob_sig"),
					Signature:  []byte("bob_sig"),
				},

				&message.Commit{
					Sender:     Nina.Address(),
					Sequence:   101,
					Round:      1,
					CommitSeal: []byte("nina_sig"),
					Signature:  []byte("nina_sig"),
				},
			},
		},

		{
			name: "old block proposed in round 1",
			consensus: allGoodConsensus{blockFutureRCC: true, getProposer: func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				if round == 1 {
					return Alice.Address(), nil
				}

				return Nina.Address(), nil
			}},
			validator: Alice,
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("bob_proposal"),
				Seals: []CommitSeal{
					{
						From: Alice.Address(),
						Seal: []byte("alice_sig"),
					},
					{
						From: Bob.Address(),
						Seal: []byte("bob_sig"),
					},
					{
						From: Chris.Address(),
						Seal: []byte("chris_sig"),
					},
				},
			},

			messages: []any{
				&message.RoundChange{
					Sender:    Chris.Address(),
					Sequence:  101,
					Round:     1,
					Signature: []byte("chris_sig"),
					LatestPreparedProposedBlock: &message.ProposedBlock{
						Block: []byte("bob_proposal"),
						Round: 0,
					},
					LatestPreparedCertificate: &message.PreparedCertificate{
						ProposalMessage: &message.Proposal{
							Sender:    Bob.Address(),
							Sequence:  101,
							Round:     0,
							Signature: []byte("bob_sig"),
							ProposedBlock: &message.ProposedBlock{
								Block: []byte("bob_proposal"),
								Round: 0,
							},
						},

						PrepareMessages: []*message.Prepare{
							{
								Sender:    Chris.Address(),
								Sequence:  101,
								Round:     0,
								Signature: []byte("chris_sig"),
							},
							{
								Sender:    Nina.Address(),
								Sequence:  101,
								Round:     0,
								Signature: []byte("nina_sig"),
							},
						},
					},
				},

				&message.RoundChange{
					Sender:   Nina.Address(),
					Sequence: 101,
					Round:    1,
					LatestPreparedProposedBlock: &message.ProposedBlock{
						Block: []byte("bob_proposal"),
						Round: 0,
					},
					LatestPreparedCertificate: &message.PreparedCertificate{
						ProposalMessage: &message.Proposal{
							Sender:    Bob.Address(),
							Sequence:  101,
							Round:     0,
							Signature: []byte("bob_sig"),
							ProposedBlock: &message.ProposedBlock{
								Block: []byte("bob_proposal"),
								Round: 0,
							},
						},

						PrepareMessages: []*message.Prepare{
							{
								Sender:    Chris.Address(),
								Sequence:  101,
								Round:     0,
								Signature: []byte("chris_sig"),
							},
							{
								Sender:    Nina.Address(),
								Sequence:  101,
								Round:     0,
								Signature: []byte("nina_sig"),
							},
						},
					},
				},

				&message.Prepare{
					Sender:    Bob.Address(),
					Sequence:  101,
					Round:     1,
					Signature: []byte("bob_sig"),
				},

				&message.Prepare{
					Sender:    Chris.Address(),
					Sequence:  101,
					Round:     1,
					Signature: []byte("chris_sig"),
				},

				&message.Commit{
					Sender:     Bob.Address(),
					Sequence:   101,
					Round:      1,
					CommitSeal: []byte("bob_sig"),
					Signature:  []byte("bob_sig"),
				},

				&message.Commit{
					Sender:     Chris.Address(),
					Sequence:   101,
					Round:      1,
					CommitSeal: []byte("chris_sig"),
					Signature:  []byte("chris_sig"),
				},
			},
		},

		{
			name: "future rcc triggers round jump",
			consensus: allGoodConsensus{getProposer: func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				if round == 3 {
					return Alice.Address(), nil
				}

				return Nina.Address(), nil
			}},
			validator: Alice,
			expected: &SequenceResult{
				Round:    3,
				Proposal: []byte("alice_proposal"),
				Seals: []CommitSeal{
					{
						From: Alice.Address(),
						Seal: []byte("alice_sig"),
					},
					{
						From: Bob.Address(),
						Seal: []byte("bob_sig"),
					},
					{
						From: Chris.Address(),
						Seal: []byte("chris_sig"),
					},
				},
			},

			messages: []any{
				&message.RoundChange{
					Sender:    Bob.Address(),
					Sequence:  101,
					Round:     3,
					Signature: []byte("bob_sig"),
				},
				&message.RoundChange{
					Sender:    Chris.Address(),
					Sequence:  101,
					Round:     3,
					Signature: []byte("chris_sig"),
				},
				&message.Prepare{
					Sender:    Bob.Address(),
					Sequence:  101,
					Round:     3,
					Signature: []byte("bob_sig"),
				},
				&message.Prepare{
					Sender:    Chris.Address(),
					Sequence:  101,
					Round:     3,
					Signature: []byte("chris_sig"),
				},
				&message.Commit{
					Sender:     Bob.Address(),
					Sequence:   101,
					Round:      3,
					CommitSeal: []byte("bob_sig"),
					Signature:  []byte("bob_sig"),
				},
				&message.Commit{
					Sender:     Chris.Address(),
					Sequence:   101,
					Round:      3,
					CommitSeal: []byte("chris_sig"),
					Signature:  []byte("chris_sig"),
				},
			},
		},

		{
			name: "future proposal triggers round jump",
			consensus: allGoodConsensus{getProposer: func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				if round == 5 {
					return Nina.Address(), nil
				}

				return Bob.Address(), nil
			}},
			validator: Alice,
			expected: &SequenceResult{
				Round:    5,
				Proposal: []byte("nina_proposal"),
				Seals: []CommitSeal{
					{
						From: Alice.Address(),
						Seal: []byte("alice_sig"),
					},
					{
						From: Nina.Address(),
						Seal: []byte("nina_sig"),
					},
				},
			},

			messages: []any{
				&message.Proposal{
					Sender:    Nina.Address(),
					Sequence:  101,
					Round:     5,
					Signature: []byte("nina_sig"),
					ProposedBlock: &message.ProposedBlock{
						Block: []byte("nina_proposal"),
						Round: 5,
					},
					RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
						{
							Sender:    Chris.Address(),
							Sequence:  101,
							Round:     5,
							Signature: []byte("chris_sig"),
						},
						{
							Sender:    Bob.Address(),
							Sequence:  101,
							Round:     5,
							Signature: []byte("bob_sig"),
						},
					}},
				},

				&message.Prepare{
					Sender:    Alice.Address(),
					Sequence:  101,
					Round:     5,
					Signature: []byte("alice_sig"),
				},

				&message.Prepare{
					Sender:    Nina.Address(),
					Sequence:  101,
					Round:     5,
					Signature: []byte("nina_sig"),
				},

				&message.Commit{
					Sender:     Alice.Address(),
					Sequence:   101,
					Round:      5,
					CommitSeal: []byte("alice_sig"),
					Signature:  []byte("alice_sig"),
				},

				&message.Commit{
					Sender:     Nina.Address(),
					Sequence:   101,
					Round:      5,
					CommitSeal: []byte("nina_sig"),
					Signature:  []byte("nina_sig"),
				},
			},
		},

		{
			name: "round timer triggers round jump",
			consensus: allGoodConsensus{getProposer: func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				if round == 1 {
					return Alice.Address(), nil
				}

				return Nina.Address(), nil
			}},
			validator: Alice,
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("alice_proposal"),
				Seals: []CommitSeal{
					{
						From: Alice.Address(),
						Seal: []byte("alice_sig"),
					},
					{
						From: Bob.Address(),
						Seal: []byte("bob_sig"),
					},
					{
						From: Chris.Address(),
						Seal: []byte("chris_sig"),
					},
				},
			},
			messages: []any{
				//&message.RoundChange{
				//	// no way to mock the
				//	Sender:   Alice.Address(),
				//	Sequence: 101,
				//	Round:    1,
				//},

				//&message.RoundChange{
				//	Sender: Chris.Address(), Sequence: 101, Round: 1,
				//},

				&message.Prepare{
					Sender:    Bob.Address(),
					Sequence:  101,
					Round:     1,
					Signature: []byte("bob_sig"),
				},

				&message.Prepare{
					Sender:    Chris.Address(),
					Sequence:  101,
					Round:     1,
					Signature: []byte("chris_sig"),
				},

				&message.Commit{
					Sender:     Bob.Address(),
					Sequence:   101,
					Round:      1,
					CommitSeal: []byte("bob_sig"),
					Signature:  []byte("bob_sig"),
				},

				&message.Commit{
					Sender:     Chris.Address(),
					Sequence:   101,
					Round:      1,
					CommitSeal: []byte("chris_sig"),
					Signature:  []byte("chris_sig"),
				},
			},
		},

		{
			name: "no prepare messages in round 0",
			consensus: consensusOfTwo{allGoodConsensus{blockFutureProposal: true, getProposer: func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				if round == 0 {
					return Bob.Address(), nil
				}

				if round == 1 {
					return Chris.Address(), nil
				}

				return Nina.Address(), nil
			}}},
			validator: Alice,
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("chris_proposal"),
				Seals: []CommitSeal{
					{
						From: Alice.Address(),
						Seal: []byte("alice_sig"),
					},
					{
						From: Nina.Address(),
						Seal: []byte("nina_sig"),
					},
				},
			},

			messages: []any{
				&message.Proposal{
					Sender:   Bob.Address(),
					Sequence: 101,
					Round:    0,
					ProposedBlock: &message.ProposedBlock{
						Block: []byte("bob_proposal"),
						Round: 0,
					},
				},

				//&message.Prepare{
				//	Sender:   Alice.Address(),
				//	Sequence: 101,
				//	Round:    0,
				//},

				&message.Proposal{
					Sender:        Chris.Address(),
					Sequence:      101,
					Round:         1,
					ProposedBlock: &message.ProposedBlock{Block: []byte("chris_proposal"), Round: 1},
					RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
						{
							Sender:   Alice.Address(),
							Sequence: 101,
							Round:    1,
						},

						{
							Sender:   Chris.Address(),
							Sequence: 101,
							Round:    1,
						},
					}},
				},

				&message.Prepare{
					Sender:   Alice.Address(),
					Sequence: 101,
					Round:    1,
				},

				&message.Prepare{
					Sender:   Nina.Address(),
					Sequence: 101,
					Round:    1,
				},

				&message.Commit{
					Sender:     Alice.Address(),
					Sequence:   101,
					Round:      1,
					CommitSeal: []byte("alice_sig"),
				},

				&message.Commit{
					Sender:     Nina.Address(),
					Sequence:   101,
					Round:      1,
					CommitSeal: []byte("nina_sig"),
				},
			},
		},

		{
			name: "no commit messages in round 0",
			consensus: consensusOfTwo{allGoodConsensus{blockFutureProposal: true, getProposer: func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				if round == 0 {
					return Bob.Address(), nil
				}

				if round == 1 {
					return Chris.Address(), nil
				}

				return Nina.Address(), nil
			}}},
			validator: Alice,
			expected: &SequenceResult{
				Round:    1,
				Proposal: []byte("chris_proposal"),
				Seals: []CommitSeal{
					{
						From: Alice.Address(),
						Seal: []byte("alice_sig"),
					},
					{
						From: Bob.Address(),
						Seal: []byte("bob_sig"),
					},
				},
			},

			messages: []any{
				// round 0
				&message.Proposal{
					Sender:        Bob.Address(),
					Sequence:      101,
					Round:         0,
					ProposedBlock: &message.ProposedBlock{Block: []byte("bob_proposal"), Round: 0},
				},

				&message.Prepare{
					Sender:   Alice.Address(),
					Sequence: 101,
					Round:    0,
				},
				&message.Prepare{
					Sender:   Bob.Address(),
					Sequence: 101,
					Round:    0,
				},

				// round 1

				&message.Proposal{
					Sender:   Chris.Address(),
					Sequence: 101,
					Round:    1,
					ProposedBlock: &message.ProposedBlock{
						Block: []byte("chris_proposal"),
						Round: 1,
					},
					RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
						{
							Sender:   Alice.Address(),
							Sequence: 101,
							Round:    1,
						},
						{
							Sender:   Bob.Address(),
							Sequence: 101,
							Round:    1,
						},
					}},
				},

				&message.Prepare{
					Sender:   Alice.Address(),
					Sequence: 101,
					Round:    1,
				},
				&message.Prepare{
					Sender:   Bob.Address(),
					Sequence: 101,
					Round:    1,
				},

				&message.Commit{
					Sender:     Bob.Address(),
					Sequence:   101,
					Round:      1,
					CommitSeal: []byte("bob_sig"),
				},
			},
		},

		{
			name: "round 0 proposer fails to build block",
			consensus: allGoodConsensus{blockFutureProposal: true, blockFutureRCC: true, getProposer: func(ctx context.Context, sequence, round uint64) ([]byte, error) {
				if round == 0 {
					return Bob.Address(), nil
				}

				if round == 1 {
					return Alice.Address(), nil
				}

				return Chris.Address(), nil
			}},
			validator: Alice,
			expected: &SequenceResult{
				Round:    2,
				Proposal: []byte("chris_proposal"),
				Seals: []CommitSeal{
					{
						From: Bob.Address(),
						Seal: []byte("bob_sig"),
					},
					{
						From: Alice.Address(),
						Seal: []byte("alice_sig"),
					},
				},
			},

			messages: []any{
				//&message.RoundChange{
				//	Sender:   Chris.Address(),
				//	Sequence: 101,
				//	Round:    1,
				//},

				&message.Proposal{
					Sender:    Chris.Address(),
					Sequence:  101,
					Round:     2,
					Signature: []byte("chris_sig"),
					ProposedBlock: &message.ProposedBlock{
						Block: []byte("chris_proposal"),
						Round: 2,
					},
					RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
						{
							Sender:    Bob.Address(),
							Sequence:  101,
							Round:     2,
							Signature: []byte("bob_sig"),
						},
						{
							Sender:    Chris.Address(),
							Sequence:  101,
							Round:     2,
							Signature: []byte("chris_sig"),
						},
					}},
				},

				&message.Prepare{
					Sender:    Bob.Address(),
					Sequence:  101,
					Round:     2,
					Signature: []byte("bob_sig"),
				},
				&message.Prepare{
					Sender:    Alice.Address(),
					Sequence:  101,
					Round:     2,
					Signature: []byte("alice_sig"),
				},
				&message.Commit{
					Sender:     Bob.Address(),
					Sequence:   101,
					Round:      2,
					CommitSeal: []byte("bob_sig"),
					Signature:  []byte("bob_sig"),
				},
				&message.Commit{
					Sender:     Alice.Address(),
					Sequence:   101,
					Round:      2,
					CommitSeal: []byte("alice_sig"),
					Signature:  []byte("alice_sig"),
				},
			},
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store := message.NewStore()
			for _, m := range tt.messages {
				switch m := m.(type) {
				case *message.RoundChange:
					store.RoundChangeMessages.Add(m)
				case *message.Prepare:
					store.PrepareMessages.Add(m)
				case *message.Commit:
					store.CommitMessages.Add(m)
				case *message.Proposal:
					store.ProposalMessages.Add(m)
				}
			}

			s := NewSequencer(slog.Default(), tt.consensus, tt.validator, dummyTransport{}, 10*time.Millisecond)
			res := s.Finalize(context.Background(), 101, store)

			//assert.True(t, reflect.DeepEqual(tt.expected, res), "expected %#v, got %#v", tt.expected, res)
			assert.EqualValues(t, tt.expected.Round, res.Round)
			assert.Equal(t, tt.expected.Proposal, res.Proposal, "expected: %s actual: %s", tt.expected.Proposal, res.Proposal)

			slices.SortFunc(tt.expected.Seals, func(a, b CommitSeal) int {
				return slices.Compare(a.From, b.From)
			})
			slices.SortFunc(res.Seals, func(a, b CommitSeal) int {
				return slices.Compare(a.From, b.From)
			})

			assert.Equal(t, tt.expected.Seals, res.Seals)
		})
	}
}
