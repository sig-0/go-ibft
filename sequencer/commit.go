package sequencer

import (
	"bytes"

	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) sendMsgCommit(ctx Context) {
	msg := &types.MsgCommit{
		From:       s.ID(),
		View:       s.state.getView(),
		BlockHash:  s.state.getProposalBlockHash(),
		CommitSeal: s.Sign(s.state.getProposalBlockHash()),
	}

	msg.Signature = s.Sign(ctx.Keccak().Hash(msg.Payload()))

	ctx.Transport().MulticastCommit(msg)
}

func (s *Sequencer) awaitCommit(ctx Context) error {
	commits, err := s.awaitQuorumCommits(ctx)
	if err != nil {
		return err
	}

	for _, commit := range commits {
		s.state.acceptSeal(commit.From, commit.CommitSeal)
	}

	return nil
}

func (s *Sequencer) awaitQuorumCommits(ctx Context) ([]*types.MsgCommit, error) {
	sub, cancelSub := ctx.MessageFeed().CommitMessages(s.state.getView(), false)
	defer cancelSub()

	cache := newMsgCache(func(msg *types.MsgCommit) bool {
		return s.isValidMsgCommit(msg)
	})

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			cache = cache.add(notification.Unwrap())

			commits := cache.get()
			if len(commits) == 0 || !ctx.Quorum().HasQuorum(types.WrapMessages(commits...)) {
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

	if !bytes.Equal(msg.BlockHash, s.state.getProposalBlockHash()) {
		return false
	}

	if !s.IsValidSignature(msg.GetSender(), msg.BlockHash, msg.CommitSeal) {
		return false
	}

	return true
}
