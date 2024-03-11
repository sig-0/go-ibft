package sequencer

import (
	"bytes"

	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) sendMsgCommit(ctx Context) {
	msg := &types.MsgCommit{
		From:       s.ID(),
		View:       s.state.View(),
		BlockHash:  s.state.AcceptedBlockHash(),
		CommitSeal: s.Sign(s.state.AcceptedBlockHash()),
	}

	msg.Signature = s.Sign(ctx.Keccak().Hash(msg.Payload()))

	ctx.MessageTransport().Commit.Multicast(msg)
}

func (s *Sequencer) awaitCommit(ctx Context) error {
	commits, err := s.awaitQuorumCommits(ctx)
	if err != nil {
		return err
	}

	for _, commit := range commits {
		s.state.AcceptSeal(commit.From, commit.CommitSeal)
	}

	return nil
}

func (s *Sequencer) awaitQuorumCommits(ctx Context) ([]*types.MsgCommit, error) {
	sub, cancelSub := ctx.MessageFeed().CommitMessages(s.state.View(), false)
	defer cancelSub()

	cache := newMsgCache(func(msg *types.MsgCommit) bool {
		return s.isValidMsgCommit(msg)
	})

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			cache = cache.Add(notification.Unwrap())

			commits := cache.Get()
			if len(commits) == 0 {
				continue
			}

			if !ctx.Quorum().HasQuorum(types.WrapMessages(commits...)) {
				continue
			}

			return commits, nil
		}
	}
}

func (s *Sequencer) isValidMsgCommit(msg *types.MsgCommit) bool {
	if !s.IsValidator(msg.From, msg.View.Sequence) {
		return false
	}

	if !bytes.Equal(msg.BlockHash, s.state.AcceptedBlockHash()) {
		return false
	}

	if !s.IsValidSignature(msg.GetSender(), msg.BlockHash, msg.CommitSeal) {
		return false
	}

	return true
}
