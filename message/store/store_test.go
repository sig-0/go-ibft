package store

import (
	"github.com/madz-lab/go-ibft/message/types"
	"github.com/stretchr/testify/assert"
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

	t.Run("MsgProposal", func(t *testing.T) {
		t.Parallel()

		msg := &types.MsgProposal{
			View:                   nil,
			From:                   nil,
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
}
