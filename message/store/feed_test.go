package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madz-lab/go-ibft/message/types"
)

func TestFeed_MsgProposal(t *testing.T) {
	t.Parallel()

	t.Run("msg received", func(t *testing.T) {
		t.Parallel()

		store := New(mockCodec{func(bytes []byte, bytes2 []byte) []byte {
			return nil
		}})

		msg := &types.MsgProposal{
			View:      &types.View{Sequence: 101, Round: 0},
			Signature: []byte("signature"),
		}

		require.NoError(t, store.AddMsgProposal(msg))

		feed := Feed{store}

		sub, cancelSub := feed.SubscribeToProposalMessages(&types.View{
			Sequence: 101,
			Round:    0,
		},
			false,
		)
		defer cancelSub()

		unwrap := <-sub
		messages := unwrap()

		assert.Equal(t, msg, messages[0])
	})

	t.Run("future round msg received", func(t *testing.T) {
		t.Parallel()

		store := New(mockCodec{func(bytes []byte, bytes2 []byte) []byte {
			return nil
		}})

		var (
			view = &types.View{Sequence: 101, Round: 1}
			msg  = &types.MsgProposal{
				View:      view,
				Signature: []byte("signature 2"),
			}
		)

		require.NoError(t, store.AddMsgProposal(msg))
		require.Len(t, store.GetProposalMessages(view), 1)

		feed := Feed{store}

		sub, cancelSub := feed.SubscribeToProposalMessages(&types.View{
			Sequence: 101,
			Round:    0,
		},
			true,
		)
		defer cancelSub()

		unwrap := <-sub
		messages := unwrap()

		assert.Equal(t, msg, messages[0])
	})
}
