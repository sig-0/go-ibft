package sequencer

import (
	"bytes"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) awaitCommit(ctx ibft.Context) error {
	commits, err := s.awaitQuorumCommits(ctx)
	if err != nil {
		return err
	}

	for _, commit := range commits {
		s.state.AcceptSeal(commit.From, commit.CommitSeal)
	}

	return nil
}

func (s *Sequencer) awaitQuorumCommits(ctx ibft.Context) ([]*types.MsgCommit, error) {
	sub, cancelSub := ctx.Feed().Commit(s.state.currentView, false)
	defer cancelSub()

	isValid := func(msg *types.MsgCommit) bool {
		return s.isValidCommit(msg, ctx.SigRecover())
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrapMessages := <-sub:
			validCommits := types.Filter(unwrapMessages(), isValid)
			if len(validCommits) == 0 {
				continue
			}

			if !ctx.Quorum().HasQuorum(s.state.CurrentSequence(), types.ToMsg(validCommits)) {
				continue
			}

			return validCommits, nil
		}
	}
}

func (s *Sequencer) isValidCommit(msg *types.MsgCommit, sigRecover ibft.SigRecover) bool {
	acceptedBlockHash := s.state.AcceptedBlockHash()
	if !bytes.Equal(msg.BlockHash, acceptedBlockHash) {
		return false
	}

	if !s.IsValidator(msg.From, msg.View.Sequence) {
		return false
	}

	if !bytes.Equal(msg.From, sigRecover.From(acceptedBlockHash, msg.CommitSeal)) {
		return false
	}

	return true
}
