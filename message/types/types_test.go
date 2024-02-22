package types

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
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

func Test_RCC_Highest_Round_Block(t *testing.T) {
	t.Parallel()

	table := []struct {
		rcc           *RoundChangeCertificate
		expectedBlock []byte
		expectedRound uint64
	}{
		{
			rcc:           &RoundChangeCertificate{Messages: []*MsgRoundChange{}},
			expectedBlock: nil,
			expectedRound: 0,
		},
		{
			rcc: &RoundChangeCertificate{Messages: []*MsgRoundChange{
				{
					LatestPreparedCertificate:   nil,
					LatestPreparedProposedBlock: nil,
				},
			}},
			expectedBlock: nil,
			expectedRound: 0,
		},
		{
			rcc: &RoundChangeCertificate{Messages: []*MsgRoundChange{
				{
					LatestPreparedCertificate: &PreparedCertificate{
						ProposalMessage: &MsgProposal{
							View: &View{
								Round: 0,
							},
						},
					},
					LatestPreparedProposedBlock: &ProposedBlock{
						Block: []byte("block 0"),
					},
				},
			}},
			expectedBlock: []byte("block 0"),
			expectedRound: 0,
		},
		{
			rcc: &RoundChangeCertificate{Messages: []*MsgRoundChange{
				{
					LatestPreparedCertificate: &PreparedCertificate{
						ProposalMessage: &MsgProposal{
							View: &View{
								Round: 0,
							},
						},
					},
					LatestPreparedProposedBlock: &ProposedBlock{
						Block: []byte("block 0"),
					},
				},
				{
					LatestPreparedCertificate: &PreparedCertificate{
						ProposalMessage: &MsgProposal{
							View: &View{
								Round: 1,
							},
						},
					},
					LatestPreparedProposedBlock: &ProposedBlock{
						Block: []byte("block 1"),
					},
				},
			}},
			expectedBlock: []byte("block 1"),
			expectedRound: 1,
		},
	}

	for i, tt := range table {
		tt := tt

		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Parallel()

			block, round := tt.rcc.HighestRoundBlock()

			assert.Equal(t, tt.expectedBlock, block)
			assert.Equal(t, tt.expectedRound, round)
		})
	}
}

func Test_RCC_Highest_Round_Block_Hash(t *testing.T) {
	t.Parallel()

	table := []struct {
		rcc               *RoundChangeCertificate
		expectedBlockHash []byte
		expectedRound     uint64
	}{
		{
			rcc:               &RoundChangeCertificate{Messages: []*MsgRoundChange{}},
			expectedBlockHash: nil,
			expectedRound:     0,
		},
		{
			rcc: &RoundChangeCertificate{Messages: []*MsgRoundChange{
				{
					LatestPreparedCertificate: nil,
				},
			}},
			expectedBlockHash: nil,
			expectedRound:     0,
		},
		{
			rcc: &RoundChangeCertificate{Messages: []*MsgRoundChange{
				{
					LatestPreparedCertificate: &PreparedCertificate{
						ProposalMessage: &MsgProposal{
							View: &View{
								Round: 0,
							},
							BlockHash: []byte("block hash 0"),
						},
					},
				},
			}},
			expectedBlockHash: []byte("block hash 0"),
			expectedRound:     0,
		},
		{
			rcc: &RoundChangeCertificate{Messages: []*MsgRoundChange{
				{
					LatestPreparedCertificate: &PreparedCertificate{
						ProposalMessage: &MsgProposal{
							View: &View{
								Round: 0,
							},
							BlockHash: []byte("block hash 0"),
						},
					},
				},
				{
					LatestPreparedCertificate: &PreparedCertificate{
						ProposalMessage: &MsgProposal{
							View: &View{
								Round: 1,
							},
							BlockHash: []byte("block hash 1"),
						},
					},
				},
			}},
			expectedBlockHash: []byte("block hash 1"),
			expectedRound:     1,
		},
	}

	for i, tt := range table {
		tt := tt

		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Parallel()

			blockHash, round := tt.rcc.HighestRoundBlockHash()

			assert.Equal(t, tt.expectedBlockHash, blockHash)
			assert.Equal(t, tt.expectedRound, round)
		})
	}
}
