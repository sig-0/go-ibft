package store

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madz-lab/go-ibft/message/types"
)

type mockCodec struct {
	recoverFn func([]byte, []byte) []byte
}

func (c mockCodec) RecoverFrom(data []byte, sig []byte) []byte {
	return c.recoverFn(data, sig)
}

type testTable[M msg] struct {
	name        string
	codec       Codec
	msg         *M
	runTestFn   func(*Store, *M) error
	expectedErr error
}

// TODO: remove cases

func TestStore_MsgProposal(t *testing.T) {
	t.Parallel()

	testTable := []testTable[types.MsgProposal]{
		{
			name: "invalid signature",
			codec: mockCodec{func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}

				return nil
			}},
			msg: &types.MsgProposal{
				From:      []byte("bad from"),
				Signature: []byte("bad signature"),
			},
			expectedErr: ErrInvalidSignature,
			runTestFn: func(store *Store, msg *types.MsgProposal) error {
				return store.AddMsgProposal(msg)
			},
		},

		{
			name: "msg added",
			codec: mockCodec{func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}

				return nil
			}},
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgProposal) error {
				return store.AddMsgProposal(msg)
			},
		},

		{
			name: "no duplicate msg when added twice",
			codec: mockCodec{func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}

				return nil
			}},
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgProposal) error {
				require.NoError(t, store.AddMsgProposal(msg))
				require.NoError(t, store.AddMsgProposal(msg))

				if len(store.GetProposalMessages(msg.View)) != 1 {
					t.Fatal("duplicate msg in collection")
				}

				return nil
			},
		},

		{
			name: "2 messages with different round",
			codec: mockCodec{func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}

				if bytes.Equal(sig, []byte("other signature")) {
					return []byte("other from")
				}

				return nil
			}},
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *Store, msg *types.MsgProposal) error {
				msg2 := &types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 1},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}

				require.NoError(t, store.AddMsgProposal(msg))
				require.NoError(t, store.AddMsgProposal(msg2))

				if len(store.GetProposalMessages(msg.View)) != 1 ||
					len(store.GetProposalMessages(msg2.View)) != 1 {
					return errors.New("messages not found")
				}

				return nil
			},
		},

		{
			name: "2 messages with different sequence",
			codec: mockCodec{func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}

				if bytes.Equal(sig, []byte("other signature")) {
					return []byte("other from")
				}

				return nil
			}},
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},

			runTestFn: func(store *Store, msg *types.MsgProposal) error {
				msg2 := &types.MsgProposal{
					View:      &types.View{Sequence: 102, Round: 0},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}

				require.NoError(t, store.AddMsgProposal(msg))
				require.NoError(t, store.AddMsgProposal(msg2))

				if len(store.GetProposalMessages(msg.View)) != 1 ||
					len(store.GetProposalMessages(msg2.View)) != 1 {
					return errors.New("messages not found")
				}

				return nil
			},
		},

		{
			name: "2 unique messages with same view",
			codec: mockCodec{func(_ []byte, sig []byte) []byte {
				if bytes.Equal(sig, []byte("signature")) {
					return []byte("from")
				}

				if bytes.Equal(sig, []byte("other signature")) {
					return []byte("other from")
				}

				return nil
			}},
			msg: &types.MsgProposal{
				View:      &types.View{Sequence: 101, Round: 0},
				From:      []byte("from"),
				Signature: []byte("signature"),
			},
			runTestFn: func(store *Store, msg *types.MsgProposal) error {
				require.NoError(t, store.AddMsgProposal(msg))
				require.NoError(t, store.AddMsgProposal(&types.MsgProposal{
					View:      &types.View{Sequence: 101, Round: 0},
					From:      []byte("other from"),
					Signature: []byte("other signature"),
				}))

				if len(store.GetProposalMessages(msg.View)) != 2 {
					return errors.New("only 1 message in store")
				}

				return nil
			},
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.ErrorIs(t, tt.runTestFn(New(tt.codec), tt.msg), tt.expectedErr)
		})
	}
}
