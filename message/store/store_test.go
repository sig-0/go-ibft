package store

import (
	"github.com/sig-0/go-ibft/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Store_Add(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		msg            message.Message
		expectedErrStr string
	}{
		{
			expectedErrStr: "missing info",
			msg:            &message.MsgProposal{},
		},

		{
			expectedErrStr: "missing sender",
			msg:            &message.MsgProposal{Info: &message.MsgInfo{}},
		},

		{
			expectedErrStr: "missing signature",
			msg: &message.MsgProposal{Info: &message.MsgInfo{
				Sender: []byte("sender"),
			}},
		},

		{
			expectedErrStr: "missing block_hash",
			msg: &message.MsgProposal{Info: &message.MsgInfo{
				Sender:    []byte("sender"),
				Signature: []byte("signature"),
			}},
		},

		{
			expectedErrStr: "missing proposed_block",
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
				BlockHash: []byte("block_hash"),
			},
		},

		{
			expectedErrStr: "ok",
			msg: &message.MsgProposal{
				Info: &message.MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
				BlockHash:     []byte("block_hash"),
				ProposedBlock: &message.ProposedBlock{},
			},
		},

		{
			expectedErrStr: "missing block_hash",
			msg: &message.MsgPrepare{
				Info: &message.MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
			},
		},

		{
			expectedErrStr: "missing block_hash",
			msg: &message.MsgPrepare{
				Info: &message.MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
			},
		},
		{
			expectedErrStr: "ok",
			msg: &message.MsgPrepare{
				Info: &message.MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
				BlockHash: []byte("block_hash"),
			},
		},

		{
			expectedErrStr: "missing block_hash",
			msg: &message.MsgCommit{
				Info: &message.MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
			},
		},

		{
			expectedErrStr: "missing commit_seal",
			msg: &message.MsgCommit{
				Info: &message.MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
				BlockHash: []byte("block_hash"),
			},
		},

		{
			expectedErrStr: "ok",
			msg: &message.MsgCommit{
				Info: &message.MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
				BlockHash:  []byte("block_hash"),
				CommitSeal: []byte("commit_seal"),
			},
		},

		{
			expectedErrStr: "ok",
			msg: &message.MsgRoundChange{
				Info: &message.MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
				LatestPreparedProposedBlock: nil,
				LatestPreparedCertificate:   nil,
			},
		},
	}

	for _, tt := range testTable {
		tt := tt
		t.Run(tt.expectedErrStr, func(t *testing.T) {
			t.Parallel()

			s := NewMsgStore()
			if tt.expectedErrStr != "ok" {
				assert.ErrorContains(t, s.Add(tt.msg), tt.expectedErrStr)
			} else {
				assert.NoError(t, s.Add(tt.msg))
			}
		})
	}
}

func Test_Store_Clear(t *testing.T) {
	t.Parallel()

	msg := &message.MsgPrepare{
		Info: &message.MsgInfo{
			Sequence:  0,
			Round:     0,
			Sender:    []byte("sender"),
			Signature: []byte("signature"),
		},
		BlockHash: []byte("block_hash"),
	}

	s := NewMsgStore()
	require.NoError(t, s.Add(msg))

	s.Clear()
	assert.Len(t, s.PrepareMessages.Get(0, 0), 0)
}
