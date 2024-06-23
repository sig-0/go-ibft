package sequencer

import (
	"bytes"

	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) sendMsgCommit(ctx Context) {
	msg := &types.MsgCommit{
		BlockHash:  s.state.getProposalBlockHash(),
		CommitSeal: s.Sign(s.state.getProposalBlockHash()),
		Metadata: &types.MsgMetadata{
			View:   s.state.getView(),
			Sender: s.ID(),
		},
	}

	msg.Metadata.Signature = s.Sign(ctx.Keccak().Hash(msg.Payload()))

	ctx.Transport().MulticastCommit(msg)
}

func (s *Sequencer) awaitCommit(ctx Context) error {
	commits, err := s.awaitQuorumCommits(ctx)
	if err != nil {
		return err
	}

	for _, commit := range commits {
		s.state.acceptSeal(commit.Sender(), commit.CommitSeal)
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
	if !s.IsValidator(msg.Sender(), msg.Sequence()) {
		return false
	}

	if !bytes.Equal(msg.BlockHash, s.state.getProposalBlockHash()) {
		return false
	}

	if !s.IsValidSignature(msg.Sender(), msg.BlockHash, msg.CommitSeal) {
		return false
	}

	return true
}
