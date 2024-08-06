package sequencer

import (
	"bytes"
	"context"
	"github.com/sig-0/go-ibft/message"
	"github.com/sig-0/go-ibft/message/store"
)

func (s *Sequencer) sendMsgCommit() {
	msg := &message.MsgCommit{
		Info: &message.MsgInfo{
			View:   s.state.getView(),
			Sender: s.validator.Address(),
		},
		BlockHash:  s.state.getProposedBlockHash(),
		CommitSeal: s.validator.Sign(s.state.getProposedBlockHash()),
	}

	msg = message.SignMsg(msg, s.validator)

	s.transport.MulticastCommit(msg)
}

func (s *Sequencer) awaitCommit(ctx context.Context) error {
	commits, err := s.awaitQuorumCommits(ctx)
	if err != nil {
		return err
	}

	for _, commit := range commits {
		s.state.acceptSeal(commit.Info.Sender, commit.CommitSeal)
	}

	return nil
}

func (s *Sequencer) awaitQuorumCommits(ctx context.Context) ([]*message.MsgCommit, error) {
	sub, cancelSub := s.feed.SubscribeCommit(s.state.getView(), false)
	defer cancelSub()

	cache := store.NewMsgCache(func(msg *message.MsgCommit) bool {
		return s.isValidMsgCommit(msg)
	})

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			cache = cache.Add(notification.Unwrap())

			commits := cache.Get()
			if len(commits) == 0 || !s.validatorSet.HasQuorum(message.WrapMessages(commits...)) {
				continue
			}

			return commits, nil
		}
	}
}

func (s *Sequencer) isValidMsgCommit(msg *message.MsgCommit) bool {
	// sender is in the validator set
	if !s.validatorSet.IsValidator(msg.Info.Sender, msg.Info.View.Sequence) {
		return false
	}

	// block hash is the same as block hash of the accepted proposal
	if !bytes.Equal(msg.BlockHash, s.state.getProposedBlockHash()) {
		return false
	}

	// sender generated commit seal by signing over block hash
	if err := s.sig.Verify(msg.Info.Sender, msg.BlockHash, msg.CommitSeal); err != nil {
		return false
	}

	return true
}
