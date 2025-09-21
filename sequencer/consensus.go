package sequencer

import (
	"context"

	"github.com/sig-0/go-ibft/message"
)

type Consensus interface {
	AwaitProposal(ctx context.Context, sequence Sequence, store *message.Store, fromHigherRounds bool) (*message.Proposal, error)
	AwaitRoundChange(ctx context.Context, sequence Sequence, store *message.Store, fromHigherRounds bool) ([]*message.RoundChange, error)
	AwaitPrepare(ctx context.Context, sequence Sequence, store *message.Store) ([]*message.Prepare, error)
	AwaitCommit(ctx context.Context, sequence Sequence, store *message.Store) ([]*message.Commit, error)
}
