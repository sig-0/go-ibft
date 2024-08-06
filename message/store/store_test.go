package store

import (
	"github.com/sig-0/go-ibft/message"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testTable[M message.IBFTMessage] struct {
	msg       M
	runTestFn func(*MsgStore, M)
	name      string
}

func Test_Collection_Clear(t *testing.T) {
	c := NewMsgCollection[*message.MsgProposal]()
	view := &message.View{Sequence: 101, Round: 5}
	msg := &message.MsgProposal{
		Info: &message.MsgInfo{
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

	table := []testTable[*message.MsgProposal]{
		{
			name: "message added",
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					View:      &message.View{Sequence: 101, Round: 0},
					Sender:    []byte("from"),
					Signature: []byte("signature"),
				},
			},
			runTestFn: func(store *MsgStore, msg *message.MsgProposal) {
				view := &message.View{Sequence: msg.Info.View.Sequence, Round: msg.Info.View.Round}
				store.ProposalMessages.Add(msg)
				assert.Len(t, store.ProposalMessages.Get(view), 1)
			},
		},

		{
			name: "message removed",
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					View:      &message.View{Sequence: 101, Round: 0},
					Sender:    []byte("from"),
					Signature: []byte("signature"),
				},
			},
			runTestFn: func(store *MsgStore, msg *message.MsgProposal) {
				view := &message.View{
					Sequence: msg.Info.View.Sequence,
					Round:    msg.Info.View.Round,
				}
				require.Len(t, store.ProposalMessages.Get(view), 0)
				store.ProposalMessages.Add(msg)

				vv := &message.View{Sequence: msg.Info.View.Sequence + 1}
				store.ProposalMessages.Remove(vv)
				require.Len(t, store.ProposalMessages.Get(view), 1)

				vvv := &message.View{Sequence: msg.Info.View.Sequence, Round: view.Round + 1}
				store.ProposalMessages.Remove(vvv)
				require.Len(t, store.ProposalMessages.Get(view), 1)

				store.ProposalMessages.Remove(view)
				require.Len(t, store.ProposalMessages.Get(view), 0)
			},
		},

		{
			name: "no duplicate message when added twice",
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					View:      &message.View{Sequence: 101, Round: 0},
					Sender:    []byte("from"),
					Signature: []byte("signature"),
				},
			},
			runTestFn: func(store *MsgStore, msg *message.MsgProposal) {
				store.ProposalMessages.Add(msg)
				store.ProposalMessages.Add(msg)

				view := &message.View{Sequence: msg.Info.View.Sequence, Round: msg.Info.View.Round}
				assert.Len(t, store.ProposalMessages.Get(view), 1)
			},
		},

		{
			name: "2 messages with different round",
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					View:      &message.View{Sequence: 101, Round: 0},
					Sender:    []byte("from"),
					Signature: []byte("signature"),
				},
			},

			runTestFn: func(store *MsgStore, msg *message.MsgProposal) {
				msg2 := &message.MsgProposal{
					Info: &message.MsgInfo{
						View:      &message.View{Sequence: 101, Round: 1},
						Sender:    []byte("other from"),
						Signature: []byte("other signature"),
					},
				}

				store.ProposalMessages.Add(msg)
				store.ProposalMessages.Add(msg2)

				view1 := &message.View{
					Sequence: msg.Info.View.Sequence,
					Round:    msg.Info.View.Round,
				}

				view2 := &message.View{
					Sequence: msg2.Info.View.Sequence,
					Round:    msg2.Info.View.Round,
				}
				assert.Len(t, store.ProposalMessages.Get(view1), 1)
				assert.Len(t, store.ProposalMessages.Get(view2), 1)
			},
		},

		{
			name: "2 messages with different sequence",
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					View:      &message.View{Sequence: 101, Round: 0},
					Sender:    []byte("from"),
					Signature: []byte("signature"),
				},
			},

			runTestFn: func(store *MsgStore, msg *message.MsgProposal) {
				msg2 := &message.MsgProposal{
					Info: &message.MsgInfo{
						View:      &message.View{Sequence: 102, Round: 0},
						Sender:    []byte("other from"),
						Signature: []byte("other signature"),
					},
				}

				store.ProposalMessages.Add(msg)
				store.ProposalMessages.Add(msg2)

				assert.Len(t, store.ProposalMessages.Get(&message.View{Sequence: msg.Info.View.Sequence, Round: msg.Info.View.Round}), 1)
				assert.Len(t, store.ProposalMessages.Get(&message.View{Sequence: msg2.Info.View.Sequence, Round: msg2.Info.View.Round}), 1)
			},
		},

		{
			name: "2 unique messages with same view",
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					View:      &message.View{Sequence: 101, Round: 0},
					Sender:    []byte("from"),
					Signature: []byte("signature"),
				},
			},
			runTestFn: func(store *MsgStore, msg *message.MsgProposal) {
				store.ProposalMessages.Add(msg)
				store.ProposalMessages.Add(&message.MsgProposal{
					Info: &message.MsgInfo{
						View:      &message.View{Sequence: 101, Round: 0},
						Sender:    []byte("other from"),
						Signature: []byte("other signature"),
					},
				})

				view := &message.View{Sequence: msg.Info.View.Sequence, Round: msg.Info.View.Round}
				assert.Len(t, store.ProposalMessages.Get(view), 2)
				assert.Len(t, store.ProposalMessages.Get(view), 2)
			},
		},

		{
			name: "no message for given round",
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					View:      &message.View{Sequence: 101, Round: 0},
					Sender:    []byte("from"),
					Signature: []byte("signature"),
				},
			},

			runTestFn: func(store *MsgStore, msg *message.MsgProposal) {
				store.ProposalMessages.Add(msg)

				view := &message.View{Sequence: msg.Info.View.Sequence, Round: msg.Info.View.Round + 1}
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
			view = &message.View{Sequence: 101, Round: 0}
			msg  = &message.MsgProposal{
				Info: &message.MsgInfo{
					View:      view,
					Signature: []byte("sig"),
				},
			}
		)

		store := NewMsgStore()
		feed := store.Feed()

		store.ProposalMessages.Add(msg)

		sub, cancelSub := feed.SubscribeProposal(view, false)
		defer cancelSub()

		r := <-sub
		messages := r.Unwrap()

		assert.Equal(t, msg, messages[0])
	})

	t.Run("message received with exact view", func(t *testing.T) {
		t.Parallel()

		var (
			view1 = &message.View{Sequence: 101, Round: 0}
			msg1  = &message.MsgProposal{
				Info: &message.MsgInfo{
					View:      view1,
					Signature: []byte("sig"),
				},
			}

			view2 = &message.View{Sequence: 101, Round: 5}
			msg2  = &message.MsgProposal{
				Info: &message.MsgInfo{
					View:      view2,
					Signature: []byte("sig"),
				},
			}
		)

		store := NewMsgStore()
		feed := store.Feed()

		store.ProposalMessages.Add(msg1)
		store.ProposalMessages.Add(msg2)

		sub, cancelSub := feed.SubscribeProposal(view1, false)
		defer cancelSub()

		r := <-sub
		messages := r.Unwrap()

		assert.Equal(t, msg1, messages[0])
	})

	t.Run("future round message received", func(t *testing.T) {
		t.Parallel()

		var (
			view = &message.View{Sequence: 101, Round: 1}
			msg  = &message.MsgProposal{
				Info: &message.MsgInfo{
					View:      view,
					Signature: []byte("signature 2"),
				},
			}
		)

		store := NewMsgStore()
		feed := store.Feed()

		store.ProposalMessages.Add(msg)
		require.Len(t, store.ProposalMessages.Get(view), 1)

		previousView := &message.View{Sequence: view.Sequence, Round: view.Round - 1}

		sub, cancelSub := feed.SubscribeProposal(previousView, true)
		defer cancelSub()

		r := <-sub
		messages := r.Unwrap()

		assert.Equal(t, msg, messages[0])
	})

	t.Run("highest round message received", func(t *testing.T) {
		t.Parallel()

		store := NewMsgStore()
		feed := store.Feed()

		sub, cancelSub := feed.SubscribeProposal(&message.View{
			Sequence: 101,
			Round:    6,
		},
			true,
		)
		defer cancelSub()

		r := <-sub
		assert.Len(t, r.Unwrap(), 0)

		var (
			view1 = &message.View{Sequence: 101, Round: 1}
			msg1  = &message.MsgProposal{
				Info: &message.MsgInfo{
					View:      view1,
					Signature: []byte("signature"),
				},
			}

			view3 = &message.View{Sequence: 101, Round: 10}
			msg3  = &message.MsgProposal{
				Info: &message.MsgInfo{
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

		view1 := &message.View{Sequence: 101, Round: 1}
		view2 := &message.View{Sequence: 102, Round: 1}

		// two subscriptions, same view
		sub, cancelSub := feed.SubscribeProposal(view1, true)

		r := <-sub
		require.Len(t, r.Unwrap(), 0)

		msg := &message.MsgProposal{
			Info: &message.MsgInfo{
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

		view1 := &message.View{Sequence: 101, Round: 1}
		view2 := &message.View{Sequence: 101, Round: 2}

		// two subscriptions, same view
		sub, cancelSub := feed.SubscribeProposal(view1, true)
		defer cancelSub()

		var (
			msg1 = &message.MsgProposal{
				Info: &message.MsgInfo{
					View:      view1,
					Signature: []byte("signature"),
				},
			}

			msg2 = &message.MsgProposal{
				Info: &message.MsgInfo{
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
