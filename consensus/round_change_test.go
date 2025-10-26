package consensus

import (
	"context"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sig-0/go-ibft/message"
	"github.com/sig-0/go-ibft/sequencer"
)

func Test_AwaitRoundChange(t *testing.T) {
	t.Parallel()

	t.Run("no incoming messages", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var (
				vs = theRealMockVS{checkQuorumFn: func(_ context.Context, _ uint64, _ [][]byte) (bool, error) {
					return false, nil
				}}
			)

			cons := New(vs, nil, nil)
			ctx, cancel := context.WithCancel(context.Background())

			var err error
			go func() {
				sequence := sequencer.Sequence{Number: 101, Round: 0}
				_, err = cons.AwaitRoundChange(ctx, sequence, message.NewStore())
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
				vs = theRealMockVS{checkQuorumFn: func(_ context.Context, _ uint64, _ [][]byte) (bool, error) {
					return false, nil
				}}
			)
			cons := New(vs, nil, nil)
			ctx, cancel := context.WithCancel(context.Background())

			var (
				err   error
				store = message.NewStore()
			)

			store.RoundChangeMessages.Add(&message.RoundChange{
				Sequence: 101,
				Round:    99999,
			})

			go func() {
				sequence := sequencer.Sequence{Number: 101, Round: 0}
				_, err = cons.AwaitRoundChange(ctx, sequence, store)
			}()

			synctest.Wait()
			require.NoError(t, err)

			cancel()

			synctest.Wait()
			assert.ErrorIs(t, err, context.Canceled)
		})
	})

	t.Run("awaited quorum of of round change messages", func(t *testing.T) {
		var (
			vs = theRealMockVS{
				checkQuorumFn: func(_ context.Context, _ uint64, _ [][]byte) (bool, error) {
					return true, nil
				},
				getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
					return [][]byte{[]byte("alice"), []byte("bob")}, nil
				},
			}

			sig = mockDeriver{addr: []byte("alice")}
		)
		cons := New(vs, nil, sig)
		require.NoError(t, cons.InitSequence(context.Background(), 101))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var (
			err   error
			store = message.NewStore()
		)

		sequence := sequencer.Sequence{
			Number: 101,
			Round:  0,
		}

		msg := &message.RoundChange{
			Sender:   []byte("alice"),
			Sequence: 101,
			Round:    0,
		}

		store.RoundChangeMessages.Add(msg)

		messages, err := cons.AwaitRoundChange(ctx, sequence, store)
		require.NoError(t, err)
		require.Len(t, messages, 1)

		assert.Equal(t, msg, messages[0])
	})
}

func Test_AwaitFutureRoundChange(t *testing.T) {
	t.Parallel()

	t.Run("no incoming messages", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var (
				vs = theRealMockVS{checkQuorumFn: func(_ context.Context, _ uint64, _ [][]byte) (bool, error) {
					return false, nil
				}}
			)

			cons := New(vs, nil, nil)
			ctx, cancel := context.WithCancel(context.Background())

			var err error
			go func() {
				sequence := sequencer.Sequence{Number: 101, Round: 0}
				_, err = cons.AwaitFutureRoundChange(ctx, sequence, message.NewStore())
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
				vs = theRealMockVS{checkQuorumFn: func(_ context.Context, _ uint64, _ [][]byte) (bool, error) {
					return false, nil
				}}
			)
			cons := New(vs, nil, nil)
			ctx, cancel := context.WithCancel(context.Background())

			var (
				err   error
				store = message.NewStore()
			)

			store.RoundChangeMessages.Add(&message.RoundChange{
				Sequence: 101,
				Round:    99999,
			})

			go func() {
				sequence := sequencer.Sequence{Number: 101, Round: 0}
				_, err = cons.AwaitFutureRoundChange(ctx, sequence, store)
			}()

			synctest.Wait()
			require.NoError(t, err)

			cancel()

			synctest.Wait()
			assert.ErrorIs(t, err, context.Canceled)
		})
	})

	t.Run("awaited quorum of of round change messages", func(t *testing.T) {
		var (
			vs = theRealMockVS{
				checkQuorumFn: func(_ context.Context, _ uint64, _ [][]byte) (bool, error) {
					return true, nil
				},
				getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
					return [][]byte{[]byte("alice"), []byte("bob")}, nil
				},
			}

			sig = mockDeriver{addr: []byte("alice")}
		)
		cons := New(vs, nil, sig)
		require.NoError(t, cons.InitSequence(context.Background(), 101))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var (
			err   error
			store = message.NewStore()
		)

		sequence := sequencer.Sequence{
			Number: 101,
			Round:  0,
		}

		msg := &message.RoundChange{
			Sender:   []byte("alice"),
			Sequence: 101,
			Round:    1,
		}

		store.RoundChangeMessages.Add(msg)

		messages, err := cons.AwaitFutureRoundChange(ctx, sequence, store)
		require.NoError(t, err)
		require.Len(t, messages, 1)

		assert.Equal(t, msg, messages[0])
	})
}

func Test_IsValidRoundChangeMessage(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name       string
		msg        *message.RoundChange
		validators [][]byte
		valid      bool
	}{
		{
			name:       "sender is not a validator",
			msg:        &message.RoundChange{Sender: []byte("not a validator")},
			validators: [][]byte{[]byte("alice"), []byte("bob")},
		},
		{
			name:       "ok with no prepared block and certificate",
			valid:      true,
			msg:        &message.RoundChange{Sender: []byte("alice")},
			validators: [][]byte{[]byte("alice"), []byte("bob")},
		},
		{
			name: "missing prepared certificate but not proposed block",
			msg: &message.RoundChange{
				Sender:                    []byte("alice"),
				LatestPreparedCertificate: &message.PreparedCertificate{},
			},
			validators: [][]byte{[]byte("alice"), []byte("bob")},
		},
		{
			name: "invalid prepared certificate",
			msg: &message.RoundChange{
				Sender:                      []byte("alice"),
				LatestPreparedCertificate:   &message.PreparedCertificate{},
				LatestPreparedProposedBlock: &message.ProposedBlock{},
			},
			validators: [][]byte{[]byte("alice"), []byte("bob")},
		},
		{
			name: "invalid prepared certificate",
			msg: &message.RoundChange{
				Sequence: 101,
				Round:    1,
				Sender:   []byte("alice"),
				LatestPreparedCertificate: &message.PreparedCertificate{
					ProposalMessage: &message.Proposal{
						Sender:    []byte("proposer"),
						Sequence:  101,
						Round:     0,
						BlockHash: []byte("block hash"),
					},
					PrepareMessages: []*message.Prepare{
						{
							Sender:    []byte("alice"),
							Sequence:  101,
							BlockHash: []byte("block hash"),
						},
					},
				},
				LatestPreparedProposedBlock: &message.ProposedBlock{},
			},
			validators: [][]byte{[]byte("alice"), []byte("bob")},
		},
		{
			name:  "ok",
			valid: true,
			msg: &message.RoundChange{
				Sequence: 101,
				Round:    1,
				Sender:   []byte("alice"),
				LatestPreparedCertificate: &message.PreparedCertificate{
					ProposalMessage: &message.Proposal{
						Sender:   []byte("proposer"),
						Sequence: 101,
						Round:    0,
						BlockHash: message.GetProposalHash(&message.ProposedBlock{
							Block: []byte("block"),
							Round: 0,
						}),
					},
					PrepareMessages: []*message.Prepare{
						{
							Sender:   []byte("alice"),
							Sequence: 101,
							BlockHash: message.GetProposalHash(&message.ProposedBlock{
								Block: []byte("block"),
								Round: 0,
							}),
						},
					},
				},
				LatestPreparedProposedBlock: &message.ProposedBlock{
					Block: []byte("block"),
					Round: 0,
				},
			},
			validators: [][]byte{[]byte("alice"), []byte("bob")},
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			vs := theRealMockVS{
				getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
					return tt.validators, nil
				},
				getProposersFn: func(_ context.Context, _, _ uint64) ([]byte, error) {
					return []byte("proposer"), nil
				},
				checkQuorumFn: func(_ context.Context, _ uint64, _ [][]byte) (bool, error) {
					return true, nil
				},
			}

			cons := New(vs, nil, nil)
			require.NoError(t, cons.InitSequence(ctx, 101))

			sequence := sequencer.Sequence{Number: 101, Round: 0}
			assert.Equal(t, tt.valid, cons.isValidRoundChange(ctx, sequence, tt.msg))
		})
	}
}
