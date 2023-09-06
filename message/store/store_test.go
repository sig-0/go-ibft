package store

import (
	"bytes"
	"github.com/madz-lab/go-ibft/message/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type mockCodec struct {
	recoverFn func([]byte, []byte) []byte
}

func (c mockCodec) RecoverFrom(data []byte, sig []byte) []byte {
	return c.recoverFn(data, sig)
}

func TestAddMessage(t *testing.T) {
	t.Parallel()

	t.Run("invalid MsgProposal", func(t *testing.T) {
		t.Parallel()

		msg := &types.MsgProposal{
			From:                   []byte("from"),
			Signature:              nil,
			ProposedBlock:          nil,
			BlockHash:              nil,
			RoundChangeCertificate: nil,
		}

		store := New(mockCodec{recoverFn: func(_ []byte, _ []byte) []byte {
			return []byte("not sender")
		}})

		assert.Error(t, store.AddMsgProposal(msg))
	})

	t.Run("add MsgProposal", func(t *testing.T) {
		t.Parallel()

		msg := &types.MsgProposal{
			View:      &types.View{Sequence: 101, Round: 1},
			From:      []byte("from"),
			Signature: nil,
		}

		store := New(mockCodec{recoverFn: func(_ []byte, _ []byte) []byte {
			return []byte("from")
		}})

		assert.NoError(t, store.AddMsgProposal(msg))
		msgs := store.GetProposalMessages(&types.View{Sequence: 101, Round: 1})
		require.Equal(t, msg, msgs[0])
	})

	t.Run("add 2 MsgProposal", func(t *testing.T) {
		t.Parallel()

		msg1 := &types.MsgProposal{
			View:      &types.View{Sequence: 101, Round: 1},
			From:      []byte("from1"),
			Signature: []byte("sig1"),
		}

		msg2 := &types.MsgProposal{
			View:      &types.View{Sequence: 101, Round: 1},
			From:      []byte("from2"),
			Signature: []byte("sig2"),
		}

		store := New(mockCodec{recoverFn: func(_ []byte, sig []byte) []byte {
			if bytes.Equal(sig, []byte("sig1")) {
				return []byte("from1")
			}

			if bytes.Equal(sig, []byte("sig2")) {
				return []byte("from2")
			}

			return nil
		}})

		assert.NoError(t, store.AddMsgProposal(msg1))
		assert.NoError(t, store.AddMsgProposal(msg2))

		msgs := store.GetProposalMessages(&types.View{Sequence: 101, Round: 1})
		require.Len(t, msgs, 2)
		assert.Equal(t, msg1, msgs[0])
		assert.Equal(t, msg2, msgs[1])
	})

	t.Run("add 2 identical MsgProposal", func(t *testing.T) {
		t.Parallel()

		msg1 := &types.MsgProposal{
			View:      &types.View{Sequence: 101, Round: 1},
			From:      []byte("from1"),
			Signature: []byte("sig1"),
		}

		msg2 := &types.MsgProposal{
			View:      &types.View{Sequence: 101, Round: 1},
			From:      []byte("from1"),
			Signature: []byte("sig1"),
		}

		store := New(mockCodec{recoverFn: func(_ []byte, sig []byte) []byte {
			if bytes.Equal(sig, []byte("sig1")) {
				return []byte("from1")
			}

			return nil
		}})

		assert.NoError(t, store.AddMsgProposal(msg1))
		assert.NoError(t, store.AddMsgProposal(msg2))

		msgs := store.GetProposalMessages(&types.View{Sequence: 101, Round: 1})
		require.Len(t, msgs, 1)
		assert.Equal(t, msg1, msgs[0])
	})
}
