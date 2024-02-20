package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madz-lab/go-ibft/message/types"
)

type testTable[M types.IBFTMessage] struct {
	msg       *M
	runTestFn func(*Store, *M)
	name      string
}

func TestStore_MsgProposal(t *testing.T) {
	t.Parallel()

	table := []testTable[types.MsgProposal]{
		{
			name: "msg added",
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgProposal) {
				AddMessage(msg, store)
				assert.Len(t, GetMessages[types.MsgProposal](msg.View, store), 1)
			},
		},

		{
			name: "msg removed",
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgProposal) {
				require.Len(t, GetMessages[types.MsgProposal](msg.View, store), 0)
				AddMessage(msg, store)

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
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgProposal) {
				AddMessage(msg, store)
				AddMessage(msg, store)

				assert.Len(t, GetMessages[types.MsgProposal](msg.View, store), 1)
			},
		},

		{
			name: "2 messages with different round",
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

				AddMessage(msg, store)
				AddMessage(msg2, store)

				assert.Len(t, GetMessages[types.MsgProposal](msg.View, store), 1)
				assert.Len(t, GetMessages[types.MsgProposal](msg2.View, store), 1)
			},
		},

		{
			name: "2 messages with different sequence",
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

				AddMessage(msg, store)
				AddMessage(msg2, store)

				assert.Len(t, GetMessages[types.MsgProposal](msg.View, store), 1)
				assert.Len(t, GetMessages[types.MsgProposal](msg2.View, store), 1)
			},
		},

		{
			name: "2 unique messages with same view",
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgProposal) {
				AddMessage(msg, store)
				AddMessage(&types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}, store)

				assert.Len(t, GetMessages[types.MsgProposal](msg.View, store), 2)
			},
		},

		{
			name: "no message for given round",
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *Store, msg *types.MsgProposal) {
				AddMessage(msg, store)

				view := &types.View{Sequence: msg.View.Sequence, Round: msg.View.Round + 1}
				GetMessages[types.MsgProposal](msg.View, store)
				msgs := GetMessages[types.MsgProposal](view, store)

				assert.Len(t, msgs, 0)
			},
		},
	}

	for _, tt := range table {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.runTestFn(New(), tt.msg)
		})
	}
}
