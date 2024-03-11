package types

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/require"
)

func Test_Msg_Bytes_And_Payload(t *testing.T) {
	t.Parallel()

	t.Run("MsgProposal", func(t *testing.T) {
		t.Parallel()

		m := &MsgProposal{
			View:          &View{Sequence: 101, Round: 1},
			From:          []byte("from"),
			Signature:     []byte("signature"),
			ProposedBlock: &ProposedBlock{Block: []byte("block"), Round: 1},
			BlockHash:     []byte("block hash"),
			RoundChangeCertificate: &RoundChangeCertificate{
				Messages: []*MsgRoundChange{
					{
						View:                        &View{Sequence: 101, Round: 1},
						From:                        []byte("from"),
						Signature:                   []byte("signature"),
						LatestPreparedProposedBlock: &ProposedBlock{Block: []byte("block"), Round: 1},
						LatestPreparedCertificate: &PreparedCertificate{
							ProposalMessage: &MsgProposal{
								View:          &View{Sequence: 101, Round: 1},
								From:          []byte("from"),
								Signature:     []byte("signature"),
								ProposedBlock: &ProposedBlock{Block: []byte("block"), Round: 1},
								BlockHash:     []byte("block hash"),
							},
							PrepareMessages: []*MsgPrepare{
								{
									View:      &View{Sequence: 101, Round: 1},
									From:      []byte("from"),
									Signature: []byte("signature"),
									BlockHash: []byte("block hash"),
								},
							},
						},
					},
				},
			},
		}

		mm := &MsgProposal{}
		require.NoError(t, proto.Unmarshal(m.Bytes(), mm))
		require.True(t, reflect.DeepEqual(m.Bytes(), mm.Bytes()))
		require.True(t, reflect.DeepEqual(m.Payload(), mm.Payload()))
	})

	t.Run("MsgPrepare", func(t *testing.T) {
		t.Parallel()

		m := &MsgPrepare{
			View:      &View{Sequence: 101, Round: 0},
			From:      []byte("from"),
			Signature: []byte("signature"),
			BlockHash: []byte("block hash"),
		}

		mm := &MsgPrepare{}
		require.NoError(t, proto.Unmarshal(m.Bytes(), mm))
		require.True(t, reflect.DeepEqual(m.Bytes(), mm.Bytes()))
		require.True(t, reflect.DeepEqual(m.Payload(), mm.Payload()))
	})

	t.Run("MsgCommit", func(t *testing.T) {
		t.Parallel()

		m := &MsgCommit{
			View:       &View{Sequence: 101, Round: 0},
			From:       []byte("from"),
			Signature:  []byte("signature"),
			BlockHash:  []byte("block hash"),
			CommitSeal: []byte("commit seal"),
		}

		mm := &MsgCommit{}
		require.NoError(t, proto.Unmarshal(m.Bytes(), mm))
		require.True(t, reflect.DeepEqual(m.Bytes(), mm.Bytes()))
		require.True(t, reflect.DeepEqual(m.Payload(), mm.Payload()))
	})

	t.Run("MsgRoundChange", func(t *testing.T) {
		t.Parallel()

		m := &MsgRoundChange{
			View:                        &View{Sequence: 101, Round: 0},
			From:                        []byte("from"),
			Signature:                   []byte("signature"),
			LatestPreparedProposedBlock: &ProposedBlock{Block: []byte("block"), Round: 0},
			LatestPreparedCertificate: &PreparedCertificate{
				ProposalMessage: &MsgProposal{
					View:          &View{Sequence: 101, Round: 0},
					From:          []byte("from"),
					Signature:     []byte("signature"),
					ProposedBlock: &ProposedBlock{Block: []byte("block"), Round: 0},
					BlockHash:     []byte("block hash"),
				},
				PrepareMessages: []*MsgPrepare{
					{
						View:      &View{Sequence: 101, Round: 0},
						From:      []byte("from"),
						Signature: []byte("signature"),
						BlockHash: []byte("block hash"),
					},
				},
			},
		}

		mm := &MsgRoundChange{}
		require.NoError(t, proto.Unmarshal(m.Bytes(), mm))
		require.True(t, reflect.DeepEqual(m.Bytes(), mm.Bytes()))
		require.True(t, reflect.DeepEqual(m.Payload(), mm.Payload()))
	})
}
