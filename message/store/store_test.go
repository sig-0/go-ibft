package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madz-lab/go-ibft/message/types"
)

type testTable[M types.IBFTMessage] struct {
	msg       M
	runTestFn func(*MsgStore, M)
	name      string
}

func TestStore_MsgProposal(t *testing.T) {
	t.Parallel()

	table := []testTable[*types.MsgProposal]{
		{
			name: "message added",
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *MsgStore, msg *types.MsgProposal) {
				store.ProposalMessages.AddMessage(msg)
				assert.Len(t, store.ProposalMessages.GetMessages(msg.View), 1)
			},
		},

		{
			name: "message removed",
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *MsgStore, msg *types.MsgProposal) {
				require.Len(t, store.ProposalMessages.GetMessages(msg.View), 0)
				store.ProposalMessages.AddMessage(msg)

				view := &types.View{Sequence: msg.View.Sequence + 1}
				store.ProposalMessages.RemoveMessages(view)
				require.Len(t, store.ProposalMessages.GetMessages(msg.View), 1)

				view = &types.View{Sequence: msg.View.Sequence, Round: msg.View.Round + 1}
				store.ProposalMessages.RemoveMessages(view)
				require.Len(t, store.ProposalMessages.GetMessages(msg.View), 1)

				store.ProposalMessages.RemoveMessages(msg.View)
				require.Len(t, store.ProposalMessages.GetMessages(msg.View), 0)
			},
		},

		{
			name: "no duplicate message when added twice",
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *MsgStore, msg *types.MsgProposal) {
				store.ProposalMessages.AddMessage(msg)
				store.ProposalMessages.AddMessage(msg)

				assert.Len(t, store.ProposalMessages.GetMessages(msg.View), 1)
			},
		},

		{
			name: "2 messages with different round",
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *MsgStore, msg *types.MsgProposal) {
				msg2 := &types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}

				store.ProposalMessages.AddMessage(msg)
				store.ProposalMessages.AddMessage(msg2)

				assert.Len(t, store.ProposalMessages.GetMessages(msg.View), 1)
				assert.Len(t, store.ProposalMessages.GetMessages(msg2.View), 1)
			},
		},

		{
			name: "2 messages with different sequence",
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *MsgStore, msg *types.MsgProposal) {
				msg2 := &types.MsgProposal{
					View:      &types.View{Sequence: 102, Round: 0},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}

				store.ProposalMessages.AddMessage(msg)
				store.ProposalMessages.AddMessage(msg2)

				assert.Len(t, store.ProposalMessages.GetMessages(msg.View), 1)
				assert.Len(t, store.ProposalMessages.GetMessages(msg2.View), 1)
			},
		},

		{
			name: "2 unique messages with same view",
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *MsgStore, msg *types.MsgProposal) {
				store.ProposalMessages.AddMessage(msg)
				store.ProposalMessages.AddMessage(&types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				})

				assert.Len(t, store.ProposalMessages.GetMessages(msg.View), 2)
				assert.Len(t, store.ProposalMessages.GetMessages(msg.View), 2)
			},
		},

		{
			name: "no message for given round",
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *MsgStore, msg *types.MsgProposal) {
				store.ProposalMessages.AddMessage(msg)

				view := &types.View{Sequence: msg.View.Sequence, Round: msg.View.Round + 1}
				assert.Len(t, store.ProposalMessages.GetMessages(view), 0)
			},
		},
	}

	for _, tt := range table {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.runTestFn(NewMsgStore(), tt.msg)
		})
	}
}

func TestFeed_MsgProposal(t *testing.T) {
	t.Parallel()

	t.Run("message received", func(t *testing.T) {
		t.Parallel()

		var (
			view = &types.View{Sequence: 101, Round: 0}
			msg  = &types.MsgProposal{
				View:      view,
				Signature: []byte("sig"),
			}
		)

		store := NewMsgStore()
		feed := store.Feed()

		store.ProposalMessages.AddMessage(msg)

		sub, cancelSub := feed.ProposalMessages(view, false)
		defer cancelSub()

		r := <-sub
		messages := r.Unwrap()

		assert.Equal(t, msg, messages[0])
	})

	t.Run("message received with exact view", func(t *testing.T) {
		t.Parallel()

		var (
			view1 = &types.View{Sequence: 101, Round: 0}
			msg1  = &types.MsgProposal{
				View:      view1,
				Signature: []byte("sig"),
			}

			view2 = &types.View{Sequence: 101, Round: 5}
			msg2  = &types.MsgProposal{
				View:      view2,
				Signature: []byte("sig"),
			}
		)

		store := NewMsgStore()
		feed := store.Feed()

		store.ProposalMessages.AddMessage(msg1)
		store.ProposalMessages.AddMessage(msg2)

		sub, cancelSub := feed.ProposalMessages(view1, false)
		defer cancelSub()

		r := <-sub
		messages := r.Unwrap()

		assert.Equal(t, msg1, messages[0])
	})

	t.Run("future round message received", func(t *testing.T) {
		t.Parallel()

		var (
			view = &types.View{Sequence: 101, Round: 1}
			msg  = &types.MsgProposal{
				View:      view,
				Signature: []byte("signature 2"),
			}
		)

		store := NewMsgStore()
		feed := store.Feed()

		store.ProposalMessages.AddMessage(msg)
		require.Len(t, store.ProposalMessages.GetMessages(view), 1)

		previousView := &types.View{Sequence: view.Sequence, Round: view.Round - 1}

		sub, cancelSub := feed.ProposalMessages(previousView, true)
		defer cancelSub()

		r := <-sub
		messages := r.Unwrap()

		assert.Equal(t, msg, messages[0])
	})

	t.Run("highest round message received", func(t *testing.T) {
		t.Parallel()

		store := NewMsgStore()
		feed := store.Feed()

		sub, cancelSub := feed.ProposalMessages(&types.View{
			Sequence: 101,
			Round:    6,
		},
			true,
		)
		defer cancelSub()

		r := <-sub
		assert.Len(t, r.Unwrap(), 0)

		var (
			view1 = &types.View{Sequence: 101, Round: 1}
			msg1  = &types.MsgProposal{
				View:      view1,
				Signature: []byte("signature"),
			}

			view3 = &types.View{Sequence: 101, Round: 10}
			msg3  = &types.MsgProposal{
				View:      view3,
				Signature: []byte("signature"),
			}
		)

		store.ProposalMessages.AddMessage(msg3)
		store.ProposalMessages.AddMessage(msg1)

		r = <-sub
		messages := r.Unwrap()

		require.Len(t, messages, 1)
		assert.Equal(t, msg3, messages[0])
	})

	t.Run("subscription not notified", func(t *testing.T) {
		t.Parallel()

		store := NewMsgStore()
		feed := store.Feed()

		view1 := &types.View{Sequence: 101, Round: 1}
		view2 := &types.View{Sequence: 102, Round: 1}

		// two subscriptions, same view
		sub, cancelSub := feed.ProposalMessages(view1, true)

		r := <-sub
		require.Len(t, r.Unwrap(), 0)

		msg := &types.MsgProposal{
			View:      view2,
			Signature: []byte("signature"),
		}

		store.ProposalMessages.AddMessage(msg)

		cancelSub() // close the sub so the channel can be read

		_, ok := <-sub
		assert.False(t, ok)
	})

	t.Run("subscription gets latest notification", func(t *testing.T) {
		t.Parallel()

		store := NewMsgStore()
		feed := store.Feed()

		view1 := &types.View{Sequence: 101, Round: 1}
		view2 := &types.View{Sequence: 101, Round: 2}

		// two subscriptions, same view
		sub, cancelSub := feed.ProposalMessages(view1, true)
		defer cancelSub()

		var (
			msg1 = &types.MsgProposal{
				View:      view1,
				Signature: []byte("signature"),
			}

			msg2 = &types.MsgProposal{
				View:      view2,
				Signature: []byte("signature"),
			}
		)

		store.ProposalMessages.AddMessage(msg1)
		store.ProposalMessages.AddMessage(msg2)

		r := <-sub
		messages := r.Unwrap()

		require.Len(t, messages, 1)
		assert.Equal(t, msg2, messages[0])
	})
}
