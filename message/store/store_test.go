package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sig-0/go-ibft/message/types"
)

type testTable[M types.IBFTMessage] struct {
	msg       M
	runTestFn func(*MsgStore, M)
	name      string
}

func Test_Collection_Clear(t *testing.T) {
	c := NewMsgCollection[*types.MsgProposal]()
	view := &types.View{Sequence: 101, Round: 5}
	msg := &types.MsgProposal{
		Metadata: &types.MsgMetadata{
			View:      view,
			Sender:    []byte("someone"),
			Signature: []byte("signature"),
		},
	}

	c.Add(msg)
	require.Equal(t, msg, c.Get(view)[0])

	c.Clear()
	assert.Len(t, c.Get(view), 0)
}

func TestStore_MsgProposal(t *testing.T) {
	t.Parallel()

	table := []testTable[*types.MsgProposal]{
		{
			name: "message added",
			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					View:      &types.View{Sequence: 101, Round: 0},
					Sender:    []byte("from"),
					Signature: []byte("signature"),
				},
			},
			runTestFn: func(store *MsgStore, msg *types.MsgProposal) {
				view := &types.View{Sequence: msg.Sequence(), Round: msg.Round()}
				store.ProposalMessages.Add(msg)
				assert.Len(t, store.ProposalMessages.Get(view), 1)
			},
		},

		{
			name: "message removed",
			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					View:      &types.View{Sequence: 101, Round: 0},
					Sender:    []byte("from"),
					Signature: []byte("signature"),
				},
			},
			runTestFn: func(store *MsgStore, msg *types.MsgProposal) {
				view := &types.View{
					Sequence: msg.Sequence(),
					Round:    msg.Round(),
				}
				require.Len(t, store.ProposalMessages.Get(view), 0)
				store.ProposalMessages.Add(msg)

				vv := &types.View{Sequence: msg.Sequence() + 1}
				store.ProposalMessages.Remove(vv)
				require.Len(t, store.ProposalMessages.Get(view), 1)

				vvv := &types.View{Sequence: msg.Sequence(), Round: view.Round + 1}
				store.ProposalMessages.Remove(vvv)
				require.Len(t, store.ProposalMessages.Get(view), 1)

				store.ProposalMessages.Remove(view)
				require.Len(t, store.ProposalMessages.Get(view), 0)
			},
		},

		{
			name: "no duplicate message when added twice",
			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					View:      &types.View{Sequence: 101, Round: 0},
					Sender:    []byte("from"),
					Signature: []byte("signature"),
				},
			},
			runTestFn: func(store *MsgStore, msg *types.MsgProposal) {
				store.ProposalMessages.Add(msg)
				store.ProposalMessages.Add(msg)

				view := &types.View{Sequence: msg.Sequence(), Round: msg.Round()}
				assert.Len(t, store.ProposalMessages.Get(view), 1)
			},
		},

		{
			name: "2 messages with different round",
			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					View:      &types.View{Sequence: 101, Round: 0},
					Sender:    []byte("from"),
					Signature: []byte("signature"),
				},
			},

			runTestFn: func(store *MsgStore, msg *types.MsgProposal) {
				msg2 := &types.MsgProposal{
					Metadata: &types.MsgMetadata{
						View:      &types.View{Sequence: 101, Round: 1},
						Sender:    []byte("other from"),
						Signature: []byte("other signature"),
					},
				}

				store.ProposalMessages.Add(msg)
				store.ProposalMessages.Add(msg2)

				view1 := &types.View{
					Sequence: msg.Sequence(),
					Round:    msg.Round(),
				}

				view2 := &types.View{
					Sequence: msg2.Sequence(),
					Round:    msg2.Round(),
				}
				assert.Len(t, store.ProposalMessages.Get(view1), 1)
				assert.Len(t, store.ProposalMessages.Get(view2), 1)
			},
		},

		{
			name: "2 messages with different sequence",
			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					View:      &types.View{Sequence: 101, Round: 0},
					Sender:    []byte("from"),
					Signature: []byte("signature"),
				},
			},

			runTestFn: func(store *MsgStore, msg *types.MsgProposal) {
				msg2 := &types.MsgProposal{
					Metadata: &types.MsgMetadata{
						View:      &types.View{Sequence: 102, Round: 0},
						Sender:    []byte("other from"),
						Signature: []byte("other signature"),
					},
				}

				store.ProposalMessages.Add(msg)
				store.ProposalMessages.Add(msg2)

				assert.Len(t, store.ProposalMessages.Get(&types.View{Sequence: msg.Sequence(), Round: msg.Round()}), 1)
				assert.Len(t, store.ProposalMessages.Get(&types.View{Sequence: msg2.Sequence(), Round: msg2.Round()}), 1)
			},
		},

		{
			name: "2 unique messages with same view",
			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					View:      &types.View{Sequence: 101, Round: 0},
					Sender:    []byte("from"),
					Signature: []byte("signature"),
				},
			},
			runTestFn: func(store *MsgStore, msg *types.MsgProposal) {
				store.ProposalMessages.Add(msg)
				store.ProposalMessages.Add(&types.MsgProposal{
					Metadata: &types.MsgMetadata{
						View:      &types.View{Sequence: 101, Round: 0},
						Sender:    []byte("other from"),
						Signature: []byte("other signature"),
					},
				})

				view := &types.View{Sequence: msg.Sequence(), Round: msg.Round()}
				assert.Len(t, store.ProposalMessages.Get(view), 2)
				assert.Len(t, store.ProposalMessages.Get(view), 2)
			},
		},

		{
			name: "no message for given round",
			msg: &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					View:      &types.View{Sequence: 101, Round: 0},
					Sender:    []byte("from"),
					Signature: []byte("signature"),
				},
			},

			runTestFn: func(store *MsgStore, msg *types.MsgProposal) {
				store.ProposalMessages.Add(msg)

				view := &types.View{Sequence: msg.Sequence(), Round: msg.Round() + 1}
				assert.Len(t, store.ProposalMessages.Get(view), 0)
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
				Metadata: &types.MsgMetadata{
					View:      view,
					Signature: []byte("sig"),
				},
			}
		)

		store := NewMsgStore()
		feed := store.Feed()

		store.ProposalMessages.Add(msg)

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
				Metadata: &types.MsgMetadata{
					View:      view1,
					Signature: []byte("sig"),
				},
			}

			view2 = &types.View{Sequence: 101, Round: 5}
			msg2  = &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					View:      view2,
					Signature: []byte("sig"),
				},
			}
		)

		store := NewMsgStore()
		feed := store.Feed()

		store.ProposalMessages.Add(msg1)
		store.ProposalMessages.Add(msg2)

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
				Metadata: &types.MsgMetadata{
					View:      view,
					Signature: []byte("signature 2"),
				},
			}
		)

		store := NewMsgStore()
		feed := store.Feed()

		store.ProposalMessages.Add(msg)
		require.Len(t, store.ProposalMessages.Get(view), 1)

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
				Metadata: &types.MsgMetadata{
					View:      view1,
					Signature: []byte("signature"),
				},
			}

			view3 = &types.View{Sequence: 101, Round: 10}
			msg3  = &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					View:      view3,
					Signature: []byte("signature"),
				},
			}
		)

		store.ProposalMessages.Add(msg3)
		store.ProposalMessages.Add(msg1)

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
			Metadata: &types.MsgMetadata{
				View:      view2,
				Signature: []byte("signature"),
			},
		}

		store.ProposalMessages.Add(msg)

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
				Metadata: &types.MsgMetadata{
					View:      view1,
					Signature: []byte("signature"),
				},
			}

			msg2 = &types.MsgProposal{
				Metadata: &types.MsgMetadata{
					View:      view2,
					Signature: []byte("signature"),
				},
			}
		)

		store.ProposalMessages.Add(msg1)
		store.ProposalMessages.Add(msg2)

		r := <-sub
		messages := r.Unwrap()

		require.Len(t, messages, 1)
		assert.Equal(t, msg2, messages[0])
	})
}
