package message

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSignatureVerifier bool

func (v mockSignatureVerifier) Verify(_, _, _ []byte) error {
	if v {
		return nil
	} else {
		return errors.New("signature verification failed")
	}
}

func Test_Store_Add(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		msg              Message
		expectedErrStr   string
		isValidSignature bool
	}{
		{
			expectedErrStr: "missing info",
			msg:            &MsgProposal{},
		},

		{
			expectedErrStr: "missing sender",
			msg:            &MsgProposal{Info: &MsgInfo{}},
		},

		{
			expectedErrStr: "missing signature",
			msg: &MsgProposal{Info: &MsgInfo{
				Sender: []byte("sender"),
			}},
		},

		{
			expectedErrStr: "missing block_hash",
			msg: &MsgProposal{Info: &MsgInfo{
				Sender:    []byte("sender"),
				Signature: []byte("signature"),
			}},
		},

		{
			expectedErrStr: "missing proposed_block",
			msg: &MsgProposal{
				Info: &MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
				BlockHash: []byte("block_hash"),
			},
		},

		{
			expectedErrStr:   "ok",
			isValidSignature: true,
			msg: &MsgProposal{
				Info: &MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
				BlockHash:     []byte("block_hash"),
				ProposedBlock: &ProposedBlock{},
			},
		},

		{
			expectedErrStr: "missing block_hash",
			msg: &MsgPrepare{
				Info: &MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
			},
		},

		{
			expectedErrStr: "missing block_hash",
			msg: &MsgPrepare{
				Info: &MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
			},
		},

		{
			expectedErrStr:   "signature verification failed",
			isValidSignature: false,
			msg: &MsgPrepare{
				Info: &MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
				BlockHash: []byte("block_hash"),
			},
		},

		{
			expectedErrStr:   "ok",
			isValidSignature: true,
			msg: &MsgPrepare{
				Info: &MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
				BlockHash: []byte("block_hash"),
			},
		},

		{
			expectedErrStr: "missing block_hash",
			msg: &MsgCommit{
				Info: &MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
			},
		},

		{
			expectedErrStr: "missing commit_seal",
			msg: &MsgCommit{
				Info: &MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
				BlockHash: []byte("block_hash"),
			},
		},

		{
			expectedErrStr:   "ok",
			isValidSignature: true,
			msg: &MsgCommit{
				Info: &MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
				BlockHash:  []byte("block_hash"),
				CommitSeal: []byte("commit_seal"),
			},
		},

		{
			expectedErrStr:   "ok",
			isValidSignature: true,
			msg: &MsgRoundChange{
				Info: &MsgInfo{
					Sender:    []byte("sender"),
					Signature: []byte("signature"),
				},
				LatestPreparedProposedBlock: nil,
				LatestPreparedCertificate:   nil,
			},
		},
	}

	for _, tt := range testTable {
		t.Run(tt.expectedErrStr, func(t *testing.T) {
			t.Parallel()

			s := NewMsgStore(mockSignatureVerifier(tt.isValidSignature))
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

	msg := &MsgPrepare{
		Info: &MsgInfo{
			Sequence:  0,
			Round:     0,
			Sender:    []byte("sender"),
			Signature: []byte("signature"),
		},
		BlockHash: []byte("block_hash"),
	}

	s := NewMsgStore(mockSignatureVerifier(true))
	require.NoError(t, s.Add(msg))

	s.Clear()
	assert.Len(t, s.PrepareMessages.Get(0, 0), 0)
}
