package sequencer

import (
	"bytes"
	"context"
	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) awaitCommit(ctx context.Context, feed MessageFeed) error {
	commits, err := s.awaitQuorumCommits(ctx, feed)
	if err != nil {
		return err
	}

	for _, commit := range commits {
		s.state.AcceptSeal(commit.From, commit.CommitSeal)
	}

	return nil
}

func (s *Sequencer) awaitQuorumCommits(ctx context.Context, feed MessageFeed) ([]*types.MsgCommit, error) {
	sub, cancelSub := feed.SubscribeToCommitMessages(s.state.currentView, false)
	defer cancelSub()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrapMessages := <-sub:
			validCommits := types.Filter(unwrapMessages(), s.isValidCommit)
			if len(validCommits) == 0 {
				continue
			}

			if !s.quorum.HasQuorum(types.ToMsg(validCommits)) {
				continue
			}

			return validCommits, nil
		}
	}
}

func (s *Sequencer) isValidCommit(msg *types.MsgCommit) bool {
	acceptedBlockHash := s.state.AcceptedBlockHash()
	if !bytes.Equal(msg.BlockHash, acceptedBlockHash) {
		return false
	}

	if !s.verifier.IsValidator(msg.From, msg.View.Sequence) {
		return false
	}

	if !bytes.Equal(msg.From, s.cdc.RecoverFrom(acceptedBlockHash, msg.CommitSeal)) {
		return false
	}

	return true
}

func (s *Sequencer) seal() []byte {
	return s.validator.Sign(s.hash(s.state.AcceptedProposedBlock()))
}

func (s *Sequencer) buildMsgCommit() *types.MsgCommit {
	msg := &types.MsgCommit{
		View:       s.state.currentView,
		From:       s.id,
		BlockHash:  s.state.AcceptedBlockHash(),
		CommitSeal: s.seal(),
	}

	msg.Signature = s.validator.Sign(msg.Payload())

	return msg
}
