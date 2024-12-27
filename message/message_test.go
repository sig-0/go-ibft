package message

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockSigner struct {
	addressFn func() []byte
	signFn    func([]byte) []byte
}

func (s mockSigner) Address() []byte {
	return s.addressFn()
}

func (s mockSigner) Sign(digest []byte) []byte {
	return s.signFn(digest)
}

func Test_SignMsg(t *testing.T) {
	t.Parallel()

	t.Run("MsgProposal", func(t *testing.T) {
		t.Parallel()

		signer := mockSigner{
			addressFn: func() []byte { return []byte("cat") },
			signFn:    func(_ []byte) []byte { return []byte("garfield") },
		}

		m := &MsgProposal{Info: &MsgInfo{
			Sequence: 101,
			Round:    3,
			Sender:   signer.Address(),
		}}

		signedMsg := SignMsg(m, signer)
		assert.Equal(t, []byte("garfield"), signedMsg.Info.Signature)
		signedMsg.Info.Signature = nil
		assert.Equal(t, m, signedMsg)
	})

	t.Run("MsgPrepare", func(t *testing.T) {
		t.Parallel()

		signer := mockSigner{
			addressFn: func() []byte { return []byte("cat") },
			signFn:    func(_ []byte) []byte { return []byte("garfield") },
		}

		m := &MsgPrepare{Info: &MsgInfo{
			Sequence: 101,
			Round:    3,
			Sender:   signer.Address(),
		}}

		signedMsg := SignMsg(m, signer)
		assert.Equal(t, []byte("garfield"), signedMsg.Info.Signature)
		signedMsg.Info.Signature = nil
		assert.Equal(t, m, signedMsg)
	})

	t.Run("MsgCommit", func(t *testing.T) {
		t.Parallel()

		signer := mockSigner{
			addressFn: func() []byte { return []byte("cat") },
			signFn:    func(_ []byte) []byte { return []byte("garfield") },
		}

		m := &MsgCommit{Info: &MsgInfo{
			Sequence: 101,
			Round:    3,
			Sender:   signer.Address(),
		}}

		signedMsg := SignMsg(m, signer)
		assert.Equal(t, []byte("garfield"), signedMsg.Info.Signature)
		signedMsg.Info.Signature = nil
		assert.Equal(t, m, signedMsg)
	})

	t.Run("MsgRoundChange", func(t *testing.T) {
		t.Parallel()

		signer := mockSigner{
			addressFn: func() []byte { return []byte("cat") },
			signFn:    func(_ []byte) []byte { return []byte("garfield") },
		}

		m := &MsgRoundChange{Info: &MsgInfo{
			Sequence: 101,
			Round:    3,
			Sender:   signer.Address(),
		}}

		signedMsg := SignMsg(m, signer)
		assert.Equal(t, []byte("garfield"), signedMsg.Info.Signature)
		signedMsg.Info.Signature = nil
		assert.Equal(t, m, signedMsg)
	})
}

func Test_RCC_HighestRoundBlock(t *testing.T) {
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
					LatestPreparedCertificate: &PreparedCertificate{ProposalMessage: &MsgProposal{
						Info: &MsgInfo{Round: 0},
					}},
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
					LatestPreparedCertificate: &PreparedCertificate{ProposalMessage: &MsgProposal{
						Info: &MsgInfo{Round: 0},
					}},
					LatestPreparedProposedBlock: &ProposedBlock{
						Block: []byte("block 0"),
					},
				},
				{
					LatestPreparedCertificate: &PreparedCertificate{ProposalMessage: &MsgProposal{
						Info: &MsgInfo{Round: 1},
					}},
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
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Parallel()

			block, round := tt.rcc.HighestRoundBlock()

			assert.Equal(t, tt.expectedBlock, block)
			assert.Equal(t, tt.expectedRound, round)
		})
	}
}

func Test_RCC_HighestRoundBlockHash(t *testing.T) {
	t.Parallel()

	table := []struct {
		rcc               *RoundChangeCertificate
		expectedBlockHash []byte
		expectedRound     uint64
	}{
		{
			expectedRound: 0,
			rcc:           &RoundChangeCertificate{Messages: []*MsgRoundChange{}},
		},

		{
			expectedRound: 0,
			rcc: &RoundChangeCertificate{Messages: []*MsgRoundChange{
				{
					LatestPreparedCertificate: nil,
				},
			}},
		},

		{
			expectedBlockHash: []byte("block hash 0"),
			expectedRound:     0,
			rcc: &RoundChangeCertificate{Messages: []*MsgRoundChange{
				{
					LatestPreparedCertificate: &PreparedCertificate{ProposalMessage: &MsgProposal{
						Info:      &MsgInfo{Round: 0},
						BlockHash: []byte("block hash 0"),
					}},
				},
			}},
		},

		{
			expectedBlockHash: []byte("block hash 1"),
			expectedRound:     1,
			rcc: &RoundChangeCertificate{Messages: []*MsgRoundChange{
				{
					LatestPreparedCertificate: &PreparedCertificate{ProposalMessage: &MsgProposal{
						Info:      &MsgInfo{Round: 0},
						BlockHash: []byte("block hash 0"),
					}},
				},

				{
					LatestPreparedCertificate: &PreparedCertificate{ProposalMessage: &MsgProposal{
						Info:      &MsgInfo{Round: 1},
						BlockHash: []byte("block hash 1"),
					}},
				},
			}},
		},
	}

	for i, tt := range table {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Parallel()

			blockHash, round := tt.rcc.HighestRoundBlockHash()
			assert.Equal(t, tt.expectedBlockHash, blockHash)
			assert.Equal(t, tt.expectedRound, round)
		})
	}
}
