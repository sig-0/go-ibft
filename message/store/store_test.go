//nolint:all
package store

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madz-lab/go-ibft/message/types"
)

type testTable[M msg] struct {
	sigRecover sigRecoverFn
	msg        *M
	runTestFn  func(*Store, *M)
	name       string
}

type noMsg struct{}

func (n noMsg) GetFrom() []byte { return nil }

func (n noMsg) GetSignature() []byte { return nil }

func (n noMsg) Payload() []byte { return nil }

func TestUnknownMessageType(t *testing.T) {
	t.Parallel()

	store := New(sigRecoverFn(func(_ []byte, _ []byte) []byte { return nil }))
	assert.ErrorIs(t, store.AddMessage(noMsg{}), ErrUnknownType)
}

func TestStore_MsgProposal(t *testing.T) {
	t.Parallel()

	testTable := []testTable[types.MsgProposal]{
		{
			name: "invalid signature",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgProposal{
				From:      []byte("bad from"),
				Signature: []byte("bad signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgProposal) {
				assert.ErrorIs(t, store.AddMessage(msg), ErrInvalidSignature)
			},
		},

		{
			name: "msg added",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgProposal) {
				assert.NoError(t, store.AddMessage(msg))
				assert.Len(t, store.GetProposalMessages(msg.View), 1)
			},
		},

		{
			name: "msg removed",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgProposal) {
				require.Len(t, store.GetProposalMessages(msg.View), 0)
				require.NoError(t, store.AddMessage(msg))

				store.RemoveProposalMessages(&types.View{Sequence: msg.View.Sequence + 1})
				require.Len(t, store.GetProposalMessages(msg.View), 1)

				store.RemoveProposalMessages(&types.View{Sequence: msg.View.Sequence, Round: msg.View.Round + 1})
				require.Len(t, store.GetProposalMessages(msg.View), 1)

				store.RemoveProposalMessages(msg.View)
				assert.Len(t, store.GetProposalMessages(msg.View), 0)
			},
		},

		{
			name: "no duplicate msg when added twice",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgProposal) {
				require.NoError(t, store.AddMessage(msg))
				require.NoError(t, store.AddMessage(msg))

				assert.Len(t, store.GetProposalMessages(msg.View), 1)
			},
		},

		{
			name: "2 messages with different round",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				if bytes.Equal(sig, []byte("other signature")) {
					return []byte("other from")
				}
				return nil
			}),
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *Store, msg *types.MsgProposal) {
				msg2 := &types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}

				require.NoError(t, store.AddMessage(msg))
				require.NoError(t, store.AddMessage(msg2))

				assert.Len(t, store.GetProposalMessages(msg.View), 1)
				assert.Len(t, store.GetProposalMessages(msg2.View), 1)
			},
		},

		{
			name: "2 messages with different sequence",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				if bytes.Equal(sig, []byte("other signature")) {
					return []byte("other from")
				}
				return nil
			}),
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *Store, msg *types.MsgProposal) {
				msg2 := &types.MsgProposal{
					View:      &types.View{Sequence: 102, Round: 0},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}

				require.NoError(t, store.AddMessage(msg))
				require.NoError(t, store.AddMessage(msg2))

				assert.Len(t, store.GetProposalMessages(msg.View), 1)
				assert.Len(t, store.GetProposalMessages(msg2.View), 1)
			},
		},

		{
			name: "2 unique messages with same view",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				if bytes.Equal(sig, []byte("other signature")) {
					return []byte("other from")
				}
				return nil
			}),
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgProposal) {
				require.NoError(t, store.AddMessage(msg))
				require.NoError(t, store.AddMessage(&types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}))

				assert.Len(t, store.GetProposalMessages(msg.View), 2)
			},
		},

		{
			name: "no message for given round",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *Store, msg *types.MsgProposal) {
				require.NoError(t, store.AddMessage(msg))

				msgs := store.GetProposalMessages(&types.View{
					Sequence: msg.View.Sequence,
					Round:    msg.View.Round + 1,
				})

				assert.Len(t, msgs, 0)
			},
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.runTestFn(New(tt.sigRecover), tt.msg)
		})
	}
}

func TestStore_MsgPrepare(t *testing.T) {
	t.Parallel()

	testTable := []testTable[types.MsgPrepare]{
		{
			name: "invalid signature",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgPrepare{
				From:      []byte("bad from"),
				Signature: []byte("bad signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgPrepare) {
				assert.ErrorIs(t, store.AddMessage(msg), ErrInvalidSignature)
			},
		},

		{
			name: "msg added",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgPrepare{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgPrepare) {
				assert.NoError(t, store.AddMessage(msg))
				assert.Len(t, store.GetPrepareMessages(msg.View), 1)
			},
		},

		{
			name: "msg removed",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgPrepare{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgPrepare) {
				require.Len(t, store.GetPrepareMessages(msg.View), 0)
				require.NoError(t, store.AddMessage(msg))

				store.RemovePrepareMessages(&types.View{Sequence: msg.View.Sequence + 1})
				require.Len(t, store.GetPrepareMessages(msg.View), 1)

				store.RemovePrepareMessages(&types.View{Sequence: msg.View.Sequence, Round: msg.View.Round + 1})
				require.Len(t, store.GetPrepareMessages(msg.View), 1)

				store.RemovePrepareMessages(msg.View)
				assert.Len(t, store.GetPrepareMessages(msg.View), 0)
			},
		},

		{
			name: "no duplicate msg when added twice",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgPrepare{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgPrepare) {
				require.NoError(t, store.AddMessage(msg))
				require.NoError(t, store.AddMessage(msg))

				assert.Len(t, store.GetPrepareMessages(msg.View), 1)
			},
		},

		{
			name: "2 messages with different round",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				if bytes.Equal(sig, []byte("other signature")) {
					return []byte("other from")
				}
				return nil
			}),
			msg: &types.MsgPrepare{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *Store, msg *types.MsgPrepare) {
				msg2 := &types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}

				require.NoError(t, store.AddMessage(msg))
				require.NoError(t, store.AddMessage(msg2))

				assert.Len(t, store.GetPrepareMessages(msg.View), 1)
				assert.Len(t, store.GetPrepareMessages(msg2.View), 1)
			},
		},

		{
			name: "2 messages with different sequence",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				if bytes.Equal(sig, []byte("other signature")) {
					return []byte("other from")
				}
				return nil
			}),
			msg: &types.MsgPrepare{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *Store, msg *types.MsgPrepare) {
				msg2 := &types.MsgPrepare{
					View:      &types.View{Sequence: 102, Round: 0},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}

				require.NoError(t, store.AddMessage(msg))
				require.NoError(t, store.AddMessage(msg2))

				assert.Len(t, store.GetPrepareMessages(msg.View), 1)
				assert.Len(t, store.GetPrepareMessages(msg2.View), 1)
			},
		},

		{
			name: "2 unique messages with same view",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				if bytes.Equal(sig, []byte("other signature")) {
					return []byte("other from")
				}
				return nil
			}),
			msg: &types.MsgPrepare{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgPrepare) {
				require.NoError(t, store.AddMessage(msg))
				require.NoError(t, store.AddMessage(&types.MsgPrepare{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}))

				assert.Len(t, store.GetPrepareMessages(msg.View), 2)
			},
		},

		{
			name: "no message for given round",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgPrepare{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *Store, msg *types.MsgPrepare) {
				require.NoError(t, store.AddMessage(msg))

				msgs := store.GetPrepareMessages(&types.View{
					Sequence: msg.View.Sequence,
					Round:    msg.View.Round + 1,
				})

				assert.Len(t, msgs, 0)
			},
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.runTestFn(New(tt.sigRecover), tt.msg)
		})
	}
}

func TestStore_MsgCommit(t *testing.T) {
	t.Parallel()

	testTable := []testTable[types.MsgCommit]{
		{
			name: "invalid signature",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgCommit{
				From:      []byte("bad from"),
				Signature: []byte("bad signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgCommit) {
				assert.ErrorIs(t, store.AddMessage(msg), ErrInvalidSignature)
			},
		},

		{
			name: "msg added",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgCommit{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgCommit) {
				assert.NoError(t, store.AddMessage(msg))
				assert.Len(t, store.GetCommitMessages(msg.View), 1)
			},
		},

		{
			name: "msg removed",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgCommit{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgCommit) {
				require.Len(t, store.GetCommitMessages(msg.View), 0)
				require.NoError(t, store.AddMessage(msg))

				store.RemoveCommitMessages(&types.View{Sequence: msg.View.Sequence + 1})
				require.Len(t, store.GetCommitMessages(msg.View), 1)

				store.RemoveCommitMessages(&types.View{Sequence: msg.View.Sequence, Round: msg.View.Round + 1})
				require.Len(t, store.GetCommitMessages(msg.View), 1)

				store.RemoveCommitMessages(msg.View)
				assert.Len(t, store.GetCommitMessages(msg.View), 0)
			},
		},

		{
			name: "no duplicate msg when added twice",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgCommit{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgCommit) {
				require.NoError(t, store.AddMessage(msg))
				require.NoError(t, store.AddMessage(msg))

				assert.Len(t, store.GetCommitMessages(msg.View), 1)
			},
		},

		{
			name: "2 messages with different round",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				if bytes.Equal(sig, []byte("other signature")) {
					return []byte("other from")
				}
				return nil
			}),
			msg: &types.MsgCommit{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *Store, msg *types.MsgCommit) {
				msg2 := &types.MsgCommit{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}

				require.NoError(t, store.AddMessage(msg))
				require.NoError(t, store.AddMessage(msg2))

				assert.Len(t, store.GetCommitMessages(msg.View), 1)
				assert.Len(t, store.GetCommitMessages(msg2.View), 1)
			},
		},

		{
			name: "2 messages with different sequence",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				if bytes.Equal(sig, []byte("other signature")) {
					return []byte("other from")
				}
				return nil
			}),
			msg: &types.MsgCommit{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *Store, msg *types.MsgCommit) {
				msg2 := &types.MsgCommit{
					View:      &types.View{Sequence: 102, Round: 0},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}

				require.NoError(t, store.AddMessage(msg))
				require.NoError(t, store.AddMessage(msg2))

				assert.Len(t, store.GetCommitMessages(msg.View), 1)
				assert.Len(t, store.GetCommitMessages(msg2.View), 1)
			},
		},

		{
			name: "2 unique messages with same view",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				if bytes.Equal(sig, []byte("other signature")) {
					return []byte("other from")
				}
				return nil
			}),
			msg: &types.MsgCommit{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgCommit) {
				require.NoError(t, store.AddMessage(msg))
				require.NoError(t, store.AddMessage(&types.MsgCommit{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}))

				assert.Len(t, store.GetCommitMessages(msg.View), 2)
			},
		},

		{
			name: "no message for given round",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgCommit{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *Store, msg *types.MsgCommit) {
				require.NoError(t, store.AddMessage(msg))

				msgs := store.GetCommitMessages(&types.View{
					Sequence: msg.View.Sequence,
					Round:    msg.View.Round + 1,
				})

				assert.Len(t, msgs, 0)
			},
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.runTestFn(New(tt.sigRecover), tt.msg)
		})
	}
}

func TestStore_MsgRoundChange(t *testing.T) {
	t.Parallel()

	testTable := []testTable[types.MsgRoundChange]{
		{
			name: "invalid signature",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgRoundChange{
				From:      []byte("bad from"),
				Signature: []byte("bad signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgRoundChange) {
				assert.ErrorIs(t, store.AddMessage(msg), ErrInvalidSignature)
			},
		},

		{
			name: "msg added",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgRoundChange{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgRoundChange) {
				assert.NoError(t, store.AddMessage(msg))
				assert.Len(t, store.GetRoundChangeMessages(msg.View), 1)
			},
		},

		{
			name: "msg removed",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgRoundChange{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgRoundChange) {
				require.Len(t, store.GetRoundChangeMessages(msg.View), 0)
				require.NoError(t, store.AddMessage(msg))

				store.RemoveRoundChangeMessages(&types.View{Sequence: msg.View.Sequence + 1})
				require.Len(t, store.GetRoundChangeMessages(msg.View), 1)

				store.RemoveRoundChangeMessages(&types.View{Sequence: msg.View.Sequence, Round: msg.View.Round + 1})
				require.Len(t, store.GetRoundChangeMessages(msg.View), 1)

				store.RemoveRoundChangeMessages(msg.View)
				assert.Len(t, store.GetRoundChangeMessages(msg.View), 0)
			},
		},

		{
			name: "no duplicate msg when added twice",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgRoundChange{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgRoundChange) {
				require.NoError(t, store.AddMessage(msg))
				require.NoError(t, store.AddMessage(msg))

				assert.Len(t, store.GetRoundChangeMessages(msg.View), 1)
			},
		},

		{
			name: "2 messages with different round",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				if bytes.Equal(sig, []byte("other signature")) {
					return []byte("other from")
				}
				return nil
			}),
			msg: &types.MsgRoundChange{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *Store, msg *types.MsgRoundChange) {
				msg2 := &types.MsgRoundChange{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}

				require.NoError(t, store.AddMessage(msg))
				require.NoError(t, store.AddMessage(msg2))

				assert.Len(t, store.GetRoundChangeMessages(msg.View), 1)
				assert.Len(t, store.GetRoundChangeMessages(msg2.View), 1)
			},
		},

		{
			name: "2 messages with different sequence",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				if bytes.Equal(sig, []byte("other signature")) {
					return []byte("other from")
				}
				return nil
			}),
			msg: &types.MsgRoundChange{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *Store, msg *types.MsgRoundChange) {
				msg2 := &types.MsgRoundChange{
					View:      &types.View{Sequence: 102, Round: 0},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}

				require.NoError(t, store.AddMessage(msg))
				require.NoError(t, store.AddMessage(msg2))

				assert.Len(t, store.GetRoundChangeMessages(msg.View), 1)
				assert.Len(t, store.GetRoundChangeMessages(msg2.View), 1)
			},
		},

		{
			name: "2 unique messages with same view",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				if bytes.Equal(sig, []byte("other signature")) {
					return []byte("other from")
				}
				return nil
			}),
			msg: &types.MsgRoundChange{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgRoundChange) {
				require.NoError(t, store.AddMessage(msg))
				require.NoError(t, store.AddMessage(&types.MsgRoundChange{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}))

				assert.Len(t, store.GetRoundChangeMessages(msg.View), 2)
			},
		},

		{
			name: "no message for given round",
			sigRecover: sigRecoverFn(func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}
				return nil
			}),
			msg: &types.MsgRoundChange{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *Store, msg *types.MsgRoundChange) {
				require.NoError(t, store.AddMessage(msg))

				msgs := store.GetRoundChangeMessages(&types.View{
					Sequence: msg.View.Sequence,
					Round:    msg.View.Round + 1,
				})

				assert.Len(t, msgs, 0)
			},
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.runTestFn(New(tt.sigRecover), tt.msg)
		})
	}
}
