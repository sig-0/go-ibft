package sequencer

import (
	"bytes"

	"github.com/madz-lab/go-ibft"
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
	sub, cancelSub := ctx.MessageFeed().CommitMessages(s.state.view, false)
	defer cancelSub()

	isValidMsg := func(msg *types.MsgCommit) bool {
		if !s.IsValidSignature(msg.GetSender(), ctx.Keccak().Hash(msg.Payload()), msg.GetSignature()) {
			return false
		}

		return s.isValidMsgCommit(msg)
	}
	cache := newMsgCache(isValidMsg)

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

			if !ctx.Quorum().HasQuorum(ibft.WrapMessages(commits...)) {
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

	acceptedBlockHash := s.state.AcceptedBlockHash()
	if !bytes.Equal(msg.BlockHash, acceptedBlockHash) {
		return false
	}

	if !s.IsValidSignature(msg.GetSender(), acceptedBlockHash, msg.CommitSeal) {
		return false
	}

	return true
}
