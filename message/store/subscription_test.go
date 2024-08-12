package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sig-0/go-ibft/message"
)

func Test_Subscribe_Proposal(t *testing.T) {
	t.Parallel()

	t.Run("exact sequence and round", func(t *testing.T) {
		t.Parallel()

		s := NewMsgStore()
		f := s.Feed()

		sub, cancelSub := f.SubscribeProposal(101, 0, false)
		defer cancelSub()

		m := &message.MsgProposal{
			Info: &message.MsgInfo{
				Sequence:  101,
				Round:     0,
				Sender:    []byte("sender"),
				Signature: []byte("signature"),
			},
			ProposedBlock: &message.ProposedBlock{
				Block: []byte("block"),
				Round: 0,
			},
			BlockHash: []byte("block hash"),
		}
		require.NoError(t, s.Add(m))

		notification := <-sub
		mm := notification.Unwrap()[0]

		assert.Equal(t, mm, m)
	})

	t.Run("highest available round", func(t *testing.T) {
		t.Parallel()

		s := NewMsgStore()
		f := s.Feed()

		sub, cancelSub := f.SubscribeProposal(101, 0, true)
		defer cancelSub()

		m := &message.MsgProposal{
			Info: &message.MsgInfo{
				Sequence:  101,
				Round:     0,
				Sender:    []byte("sender"),
				Signature: []byte("signature"),
			},
			ProposedBlock: &message.ProposedBlock{
				Block: []byte("block"),
				Round: 0,
			},
			BlockHash: []byte("block hash"),
		}
		require.NoError(t, s.Add(m))

		m = &message.MsgProposal{
			Info: &message.MsgInfo{
				Sequence:  101,
				Round:     1,
				Sender:    []byte("sender"),
				Signature: []byte("signature"),
			},
			ProposedBlock: &message.ProposedBlock{
				Block: []byte("block"),
				Round: 0,
			},
			BlockHash: []byte("block hash"),
		}
		require.NoError(t, s.Add(m))

		m = &message.MsgProposal{
			Info: &message.MsgInfo{
				Sequence:  101,
				Round:     2,
				Sender:    []byte("sender"),
				Signature: []byte("signature"),
			},
			ProposedBlock: &message.ProposedBlock{
				Block: []byte("block"),
				Round: 0,
			},
			BlockHash: []byte("block hash"),
		}
		require.NoError(t, s.Add(m))

		notification := <-sub
		mm := notification.Unwrap()[0]

		assert.Equal(t, mm, m)
		assert.Equal(t, uint64(2), mm.Info.Round)
	})
}
