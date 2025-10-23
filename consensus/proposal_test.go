package consensus

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sig-0/go-ibft/message"
	"github.com/sig-0/go-ibft/sequencer"
)

func Test_AwaitProposal(t *testing.T) {
	t.Parallel()

	t.Run("no incoming messages", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var (
				vs               = mockValidatorSet{}
				proposalVerifier = mockProposalVerifier{}
				sigVerifier      = mockSignatureVerifier{}
			)

			cons := New(vs, proposalVerifier, sigVerifier)
			ctx, cancel := context.WithCancel(context.Background())

			var err error
			go func() {
				sequence := sequencer.Sequence{Number: 101, Round: 0}
				_, err = cons.AwaitProposal(ctx, sequence, message.NewStore())
			}()

			synctest.Wait()
			require.NoError(t, err)

			cancel()

			synctest.Wait()
			assert.ErrorIs(t, err, context.Canceled)
		})
	})

	t.Run("no messages matching round", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var (
				vs               = mockValidatorSet{}
				proposalVerifier = mockProposalVerifier{}
				sigVerifier      = mockSignatureVerifier{}
			)

			cons := New(vs, proposalVerifier, sigVerifier)
			ctx, cancel := context.WithCancel(context.Background())

			var (
				err   error
				store = message.NewStore()
			)

			store.ProposalMessages.Add(&message.Proposal{
				Sequence: 101,
				Round:    99999,
			})

			go func() {
				sequence := sequencer.Sequence{Number: 101, Round: 0}
				_, err = cons.AwaitProposal(ctx, sequence, store)
			}()

			synctest.Wait()
			require.NoError(t, err)

			cancel()

			synctest.Wait()
			assert.ErrorIs(t, err, context.Canceled)
		})
	})

	t.Run("no valid messages", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var (
				vs = mockValidatorSet{
					proposer: []byte("the proposer"),
				}
				proposalVerifier = mockProposalVerifier{valid: false}
				sigVerifier      = mockSignatureVerifier{}
			)

			cons := New(vs, proposalVerifier, sigVerifier)
			ctx, cancel := context.WithCancel(context.Background())

			var (
				err   error
				store = message.NewStore()
			)

			pb := &message.ProposedBlock{
				Block: []byte("block"),
				Round: 0,
			}

			store.ProposalMessages.Add(&message.Proposal{
				Sender:        []byte("the proposer"),
				Sequence:      101,
				Round:         0,
				ProposedBlock: pb,
				BlockHash:     message.GetProposalHash(pb),
			})

			go func() {
				sequence := sequencer.Sequence{Number: 101, Round: 0}
				_, err = cons.AwaitProposal(ctx, sequence, store)
			}()

			synctest.Wait()
			require.NoError(t, err)

			cancel()

			synctest.Wait()
			assert.ErrorIs(t, err, context.Canceled)
		})
	})

	t.Run("awaited proposal", func(t *testing.T) {
		var (
			vs = mockValidatorSet{
				proposer: []byte("the proposer"),
			}
			proposalVerifier = mockProposalVerifier{valid: true}
		)

		cons := New(vs, proposalVerifier, nil)

		pb := &message.ProposedBlock{
			Block: []byte("block"),
			Round: 0,
		}

		msg := &message.Proposal{
			Sender:        []byte("the proposer"),
			Sequence:      101,
			Round:         0,
			ProposedBlock: pb,
			BlockHash:     message.GetProposalHash(pb),
		}

		store := message.NewStore()
		store.ProposalMessages.Add(msg)

		sequence := sequencer.Sequence{Number: 101, Round: 0}
		awaitedProposal, err := cons.AwaitProposal(context.Background(), sequence, store)
		require.NoError(t, err)

		assert.Equal(t, msg, awaitedProposal)
	})
}

func Test_IsValidProposalMessage(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name     string
		sequence sequencer.Sequence
		msg      *message.Proposal

		vs         ValidatorSet
		vrf        Verifier
		sig        message.SignatureVerifier
		validators [][]byte

		expected bool
	}{

		{
			name: "round in message does not match round in proposed block",
			vs: theRealMockVS{getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
				return [][]byte{[]byte("alice"), []byte("bob")}, nil
			}},
			msg: &message.Proposal{
				Round:         1,
				ProposedBlock: &message.ProposedBlock{Round: 2},
			},
		},
		{
			name: "external error when getting validators",
			msg: &message.Proposal{
				Round:         1,
				ProposedBlock: &message.ProposedBlock{Round: 1},
			},
			vs: theRealMockVS{
				getProposersFn: func(_ context.Context, _, _ uint64) ([]byte, error) {
					return nil, errors.New("external error")
				},
				getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
					return [][]byte{[]byte("alice"), []byte("bob")}, nil
				},
			},
		},
		{
			name: "proposal not coming from proposer",
			msg: &message.Proposal{
				Sender:        []byte("not alice"),
				Round:         1,
				ProposedBlock: &message.ProposedBlock{Round: 1},
			},
			vs: theRealMockVS{
				getProposersFn: func(_ context.Context, _, _ uint64) ([]byte, error) {
					return []byte("alice"), nil
				},
				getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
					return [][]byte{[]byte("alice"), []byte("bob")}, nil
				},
			},
		},
		{
			name: "invalid block hash in proposal",
			msg: &message.Proposal{
				Sender:    []byte("alice"),
				Round:     1,
				BlockHash: []byte("not the block hash"),
				ProposedBlock: &message.ProposedBlock{
					Block: []byte("block"),
					Round: 1,
				},
			},
			vs: theRealMockVS{
				getProposersFn: func(_ context.Context, _, _ uint64) ([]byte, error) {
					return []byte("alice"), nil
				},
				getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
					return [][]byte{[]byte("alice"), []byte("bob")}, nil
				},
			},
		},
		{
			name: "bad round 0 proposal",
			msg: &message.Proposal{
				Sender: []byte("alice"),
				Round:  0,
				BlockHash: message.GetProposalHash(&message.ProposedBlock{
					Block: []byte("block"),
					Round: 0,
				}),
				ProposedBlock: &message.ProposedBlock{
					Block: []byte("block"),
					Round: 0,
				},
			},
			vs: theRealMockVS{
				getProposersFn: func(_ context.Context, _, _ uint64) ([]byte, error) {
					return []byte("alice"), nil
				},
				getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
					return [][]byte{[]byte("alice"), []byte("bob")}, nil
				},
			},
			vrf: mockProposalVerifier{valid: false},
		},
		{
			name: "invalid rcc",
			msg: &message.Proposal{
				Sender: []byte("alice"),
				Round:  3,
				BlockHash: message.GetProposalHash(&message.ProposedBlock{
					Block: []byte("block"),
					Round: 3,
				}),
				ProposedBlock: &message.ProposedBlock{
					Block: []byte("block"),
					Round: 3,
				},
			},
			vs: theRealMockVS{
				getProposersFn: func(_ context.Context, _, _ uint64) ([]byte, error) {
					return []byte("alice"), nil
				},
				getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
					return [][]byte{[]byte("alice"), []byte("bob")}, nil
				},
			},
		},
		{
			name: "invalid block in rcc",
			msg: &message.Proposal{
				Sender: []byte("alice"),
				Round:  3,
				BlockHash: message.GetProposalHash(&message.ProposedBlock{
					Block: []byte("block"),
					Round: 3,
				}),
				ProposedBlock: &message.ProposedBlock{
					Block: []byte("block"),
					Round: 3,
				},
				RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
					{
						Sender: []byte("bob"),
						Round:  3,
					},
				}},
			},
			vs: theRealMockVS{
				getProposersFn: func(_ context.Context, _, _ uint64) ([]byte, error) {
					return []byte("alice"), nil
				},
				getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
					return [][]byte{[]byte("alice"), []byte("bob")}, nil
				},
				checkQuorumFn: func(_ context.Context, _ uint64, _ [][]byte) (bool, error) {
					return true, nil
				},
			},
			vrf: mockProposalVerifier{valid: false},
		},
		{
			name: "invalid block in rcc",
			msg: &message.Proposal{
				Sender: []byte("alice"),
				Round:  3,
				BlockHash: message.GetProposalHash(&message.ProposedBlock{
					Block: []byte("block"),
					Round: 3,
				}),
				ProposedBlock: &message.ProposedBlock{
					Block: []byte("block"),
					Round: 3,
				},
				RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
					{
						Sender: []byte("bob"),
						Round:  2,
						LatestPreparedCertificate: &message.PreparedCertificate{
							ProposalMessage: &message.Proposal{
								Sender: []byte("alice"),
								Round:  2,
								BlockHash: message.GetProposalHash(&message.ProposedBlock{
									Block: []byte("not block"),
									Round: 2,
								}),
							},
						},
					},
				}},
			},
			vs: theRealMockVS{
				getProposersFn: func(_ context.Context, _, _ uint64) ([]byte, error) {
					return []byte("alice"), nil
				},
				getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
					return [][]byte{[]byte("alice"), []byte("bob")}, nil
				},
				checkQuorumFn: func(_ context.Context, _ uint64, _ [][]byte) (bool, error) {
					return true, nil
				},
			},
			vrf: mockProposalVerifier{valid: false},
		},
		{
			name:     "ok",
			expected: true,
			msg: &message.Proposal{
				Sender: []byte("alice"),
				Round:  3,
				BlockHash: message.GetProposalHash(&message.ProposedBlock{
					Block: []byte("block"),
					Round: 3,
				}),
				ProposedBlock: &message.ProposedBlock{
					Block: []byte("block"),
					Round: 3,
				},
				RoundChangeCertificate: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
					{
						Sender: []byte("bob"),
						Round:  3,
						LatestPreparedCertificate: &message.PreparedCertificate{
							ProposalMessage: &message.Proposal{
								Sender: []byte("alice"),
								Round:  2,
								BlockHash: message.GetProposalHash(&message.ProposedBlock{
									Block: []byte("block"),
									Round: 2,
								}),
							},
							PrepareMessages: []*message.Prepare{
								{
									Sender: []byte("bob"),
									Round:  2,
									BlockHash: message.GetProposalHash(&message.ProposedBlock{
										Block: []byte("block"),
										Round: 2,
									}),
								},
							},
						},
					},
				}},
			},
			vs: theRealMockVS{
				getProposersFn: func(_ context.Context, _, _ uint64) ([]byte, error) {
					return []byte("alice"), nil
				},
				getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
					return [][]byte{
						[]byte("alice"),
						[]byte("bob"),
					}, nil
				},
				checkQuorumFn: func(_ context.Context, _ uint64, _ [][]byte) (bool, error) {
					return true, nil
				},
			},
			validators: [][]byte{
				[]byte("alice"),
				[]byte("bob"),
			},
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sequence := uint64(101)
			ctx := context.Background()
			c := New(tt.vs, tt.vrf, tt.sig)
			require.NoError(t, c.InitSequence(ctx, sequence))

			assert.Equal(t, tt.expected, c.isValidProposal(ctx, tt.sequence, tt.msg))
		})
	}
}

func Test_IsValidRCC(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name     string
		rcc      *message.RoundChangeCertificate
		proposal *message.Proposal
		vs       ValidatorSet
		expected bool
	}{
		{
			name: "missing rcc",
			rcc:  &message.RoundChangeCertificate{},
			vs: theRealMockVS{getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
				return [][]byte{[]byte("alice"), []byte("bob")}, nil
			}},
		},
		{
			name: "sequence (round) does not match proposal msg sequence (round)",
			proposal: &message.Proposal{
				Sequence: 101,
				Round:    1,
			},
			rcc: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
				{
					Sequence: 202,
					Round:    2,
				},
			}},
			vs: theRealMockVS{getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
				return [][]byte{[]byte("alice"), []byte("bob")}, nil
			}},
		},
		{
			name: "sender in rcc is not a validator",
			proposal: &message.Proposal{
				Sequence: 101,
				Round:    1,
			},
			rcc: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
				{
					Sequence: 101,
					Round:    1,
					Sender:   []byte("definitely not a validator"),
				},
			}},
			vs: theRealMockVS{getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
				return [][]byte{[]byte("alice"), []byte("bob")}, nil
			}},
		},
		{
			name: "duplicate sender in rcc",
			proposal: &message.Proposal{
				Sequence: 101,
				Round:    1,
			},
			rcc: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
				{
					Sequence: 101,
					Round:    1,
					Sender:   []byte("bob"),
				},
				{
					Sequence: 101,
					Round:    1,
					Sender:   []byte("bob"),
				},
			}},
			vs: theRealMockVS{getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
				return [][]byte{[]byte("alice"), []byte("bob")}, nil
			}},
		},
		{
			name: "no quorum in rcc",
			proposal: &message.Proposal{
				Sequence: 101,
				Round:    1,
			},
			rcc: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
				{
					Sequence: 101,
					Round:    1,
					Sender:   []byte("alice"),
				},
				{
					Sequence: 101,
					Round:    1,
					Sender:   []byte("bob"),
				},
			}},
			vs: theRealMockVS{
				getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
					return [][]byte{[]byte("alice"), []byte("bob")}, nil
				},
				checkQuorumFn: func(_ context.Context, _ uint64, _ [][]byte) (bool, error) {
					return false, nil
				},
			},
		},
		{
			name:     "ok",
			expected: true,
			proposal: &message.Proposal{
				Sequence: 101,
				Round:    1,
			},
			rcc: &message.RoundChangeCertificate{Messages: []*message.RoundChange{
				{
					Sequence: 101,
					Round:    1,
					Sender:   []byte("alice"),
				},
				{
					Sequence: 101,
					Round:    1,
					Sender:   []byte("bob"),
				},
			}},
			vs: theRealMockVS{
				getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
					return [][]byte{[]byte("alice"), []byte("bob")}, nil
				},
				checkQuorumFn: func(_ context.Context, _ uint64, _ [][]byte) (bool, error) {
					return true, nil
				},
			},
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sequence := uint64(101)
			ctx := context.Background()
			c := New(tt.vs, nil, nil)
			require.NoError(t, c.InitSequence(ctx, sequence))

			assert.Equal(t, tt.expected, c.isValidRCC(ctx, tt.rcc, tt.proposal))
		})
	}
}
