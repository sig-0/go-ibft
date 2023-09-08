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
				assert.ErrorIs(t, AddMessage(msg, store), ErrInvalidSignature)
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
				assert.NoError(t, AddMessage(msg, store))
				assert.Len(t, GetMessages[types.MsgProposal](msg.View, store), 1)
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
				require.Len(t, GetMessages[types.MsgProposal](msg.View, store), 0)
				require.NoError(t, AddMessage(msg, store))

				view := &types.View{Sequence: msg.View.Sequence + 1}
				RemoveMessages[types.MsgProposal](view, store)
				require.Len(t, GetMessages[types.MsgProposal](msg.View, store), 1)

				view = &types.View{Sequence: msg.View.Sequence, Round: msg.View.Round + 1}
				RemoveMessages[types.MsgProposal](view, store)
				require.Len(t, GetMessages[types.MsgProposal](msg.View, store), 1)

				RemoveMessages[types.MsgProposal](msg.View, store)
				assert.Len(t, GetMessages[types.MsgProposal](msg.View, store), 0)
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
				require.NoError(t, AddMessage(msg, store))
				require.NoError(t, AddMessage(msg, store))

				assert.Len(t, GetMessages[types.MsgProposal](msg.View, store), 1)
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

				require.NoError(t, AddMessage(msg, store))
				require.NoError(t, AddMessage(msg2, store))

				assert.Len(t, GetMessages[types.MsgProposal](msg.View, store), 1)
				assert.Len(t, GetMessages[types.MsgProposal](msg2.View, store), 1)
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

				require.NoError(t, AddMessage(msg, store))
				require.NoError(t, AddMessage(msg2, store))

				assert.Len(t, GetMessages[types.MsgProposal](msg.View, store), 1)
				assert.Len(t, GetMessages[types.MsgProposal](msg2.View, store), 1)
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
				require.NoError(t, AddMessage(msg, store))
				require.NoError(t, AddMessage(&types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}, store))

				assert.Len(t, GetMessages[types.MsgProposal](msg.View, store), 2)
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
				require.NoError(t, AddMessage(msg, store))

				view := &types.View{Sequence: msg.View.Sequence, Round: msg.View.Round + 1}
				GetMessages[types.MsgProposal](msg.View, store)
				msgs := GetMessages[types.MsgProposal](view, store)

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
