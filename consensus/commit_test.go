package consensus

import (
	"context"
	"testing"
	"testing/synctest"

	"github.com/sig-0/go-ibft/message"
	"github.com/sig-0/go-ibft/sequencer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_AwaitCommit(t *testing.T) {
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
				_, err = cons.AwaitCommit(ctx, sequence, message.NewStore())
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

			store.CommitMessages.Add(&message.Commit{
				Sequence: 101,
				Round:    99999,
			})

			go func() {
				sequence := sequencer.Sequence{Number: 101, Round: 0}
				_, err = cons.AwaitCommit(ctx, sequence, store)
			}()

			synctest.Wait()
			require.NoError(t, err)

			cancel()

			synctest.Wait()
			assert.ErrorIs(t, err, context.Canceled)
		})
	})

	t.Run("awaited quorum of commit messages", func(t *testing.T) {
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

		blockHash := message.GetProposalHash(&message.ProposedBlock{
			Block: []byte("block"),
			Round: 0,
		})

		sequence := sequencer.Sequence{
			Number:   101,
			Round:    0,
			Proposal: &message.Proposal{BlockHash: blockHash},
		}

		msg := &message.Commit{
			Sender:    []byte("alice"),
			Sequence:  101,
			Round:     0,
			BlockHash: blockHash,
		}

		store.CommitMessages.Add(msg)

		messages, err := cons.AwaitCommit(ctx, sequence, store)
		require.NoError(t, err)
		require.Len(t, messages, 1)

		assert.Equal(t, msg, messages[0])
	})
}

func Test_IsValidCommitMessage(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name       string
		msg        *message.Commit
		validSig   bool
		validators [][]byte
		blockHash  []byte
		valid      bool
	}{
		{
			name:       "sender is not a validator",
			msg:        &message.Commit{Sender: []byte("not a validator")},
			validators: [][]byte{[]byte("alice"), []byte("bob")},
		},
		{
			name: "invalid block hash",
			msg: &message.Commit{
				Sender:    []byte("alice"),
				BlockHash: []byte("invalid block hash"),
			},
			blockHash:  []byte("block hash"),
			validators: [][]byte{[]byte("alice"), []byte("bob")},
		},
		{
			name:     "ok",
			valid:    true,
			validSig: true,
			msg: &message.Commit{
				Sender:    []byte("alice"),
				BlockHash: []byte("block hash"),
			},
			blockHash:  []byte("block hash"),
			validators: [][]byte{[]byte("alice"), []byte("bob")},
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			vs := theRealMockVS{getValidatorsFn: func(_ context.Context, _ uint64) ([][]byte, error) {
				return tt.validators, nil
			}}

			sequence := sequencer.Sequence{Proposal: &message.Proposal{
				BlockHash: tt.blockHash,
			}}

			sig := mockDeriver{addr: []byte("alice")}

			cons := New(vs, nil, sig)
			require.NoError(t, cons.InitSequence(ctx, 101))

			assert.Equal(t, tt.valid, cons.isValidCommit(ctx, sequence, tt.msg))
		})
	}
}
