package sequencer

import (
	"bytes"
	"context"
	"github.com/madz-lab/go-ibft/message/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHappyFlow(t *testing.T) {
	t.Parallel()

	t.Run("validator is not the proposer", func(t *testing.T) {
		t.Parallel()

		var (
			id            = []byte("validator id")
			proposer      = []byte("proposer")
			someValidator = []byte("some validator")
			proposalHash  = []byte("proposal hash")
			sequence      = uint64(101)
		)

		var (
			validator Validator
			transport Transport
			quorum    Quorum
			feed      MessageFeed
		)

		validator = mockValidator{
			recoverFromFn:  func(_ []byte, _ []byte) []byte { return someValidator },
			hashFn:         func(_ []byte) []byte { return proposalHash },
			isValidBlockFn: func(_ []byte) bool { return true },
			idFn:           func() []byte { return id },
			isProposerFn:   func(_ *types.View, from []byte) bool { return !bytes.Equal(from, id) },
			signFn:         func(_ []byte) []byte { return nil },
		}

		feed = mockMessageeFeed{
			subProposalFn: func() []*types.MsgProposal {
				return []*types.MsgProposal{
					{
						View: &types.View{
							Sequence: sequence,
							Round:    0,
						},
						From: proposer,
						ProposedBlock: &types.ProposedBlock{
							Data:  []byte("block data"),
							Round: 0,
						},
						ProposalHash: proposalHash,
					},
				}
			},
			subPrepareFn: func() []*types.MsgPrepare {
				return []*types.MsgPrepare{
					{
						View: &types.View{
							Sequence: sequence,
							Round:    0,
						},
						From:         someValidator,
						ProposalHash: proposalHash,
					},
				}
			},
			subCommitFn: func() []*types.MsgCommit {
				return []*types.MsgCommit{
					{
						View: &types.View{
							Sequence: sequence,
							Round:    0,
						},
						From:         someValidator,
						ProposalHash: proposalHash,
						CommitSeal:   []byte("commit seal"),
					},
				}
			},
		}

		transport = DummyTransport

		quorum = mockQuorum{
			quorumPrepare: func(_ ...*types.MsgPrepare) bool {
				return true
			},
			quorumCommit: func(_ ...*types.MsgCommit) bool {
				return true
			},
		}

		s := New(validator, 1)
		s.WithTransport(transport)
		s.WithQuorum(quorum)

		fb := s.FinalizeSequence(context.Background(), sequence, feed)
		assert.NotNil(t, fb)
		assert.Equal(t, uint64(0), fb.Round)
		assert.Equal(t, []byte("block data"), fb.Block)
		assert.Equal(t, [][]byte{[]byte("commit seal")}, fb.CommitSeals)
	})
}
