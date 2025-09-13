package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Cache(t *testing.T) {
	t.Parallel()

	t.Run("valid message added", func(t *testing.T) {
		t.Parallel()

		msg := &MsgProposal{Info: &MsgInfo{Sender: []byte("sender")}}
		cache := NewMsgCache(func(_ *MsgProposal) bool {
			return true
		})

		cache.Add(msg)
		assert.Len(t, cache.Get(), 1)
	})

	t.Run("invalid message skipped", func(t *testing.T) {
		t.Parallel()

		msg := &MsgProposal{Info: &MsgInfo{Sender: []byte("sender")}}
		cache := NewMsgCache(func(_ *MsgProposal) bool {
			return false
		})

		cache.Add(msg)
		assert.Len(t, cache.Get(), 0)
	})

	t.Run("duplicate message skipped", func(t *testing.T) {
		t.Parallel()

		msg := &MsgProposal{Info: &MsgInfo{Sender: []byte("sender")}}
		cache := NewMsgCache(func(_ *MsgProposal) bool {
			return true
		})

		cache.Add(msg)
		cache.Add(msg)
		assert.Len(t, cache.Get(), 1)
	})
}
