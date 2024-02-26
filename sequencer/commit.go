package sequencer

import (
	"bytes"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) multicastCommit(ctx ibft.Context) {
	msg := &types.MsgCommit{
		From:       s.ID(),
		View:       s.state.CurrentView(),
		BlockHash:  s.state.AcceptedBlockHash(),
		CommitSeal: s.Sign(s.state.AcceptedBlockHash()),
	}

	msg.Signature = s.Sign(ctx.Keccak().Hash(msg.Payload()))

	ctx.Transport().Multicast(msg)
}

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
	cache := newMsgCache(func(msg *types.MsgCommit) bool {
		if !s.IsValidSignature(msg.GetSender(), ctx.Keccak().Hash(msg.Payload()), msg.GetSignature()) {
			return false
		}

		return s.isValidMsgCommit(msg)
	})

	sub, cancelSub := ctx.Feed().CommitMessages(s.state.currentView, false)
	defer cancelSub()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			validCommits := cache.Add(notification.Unwrap()).Messages()
			if len(validCommits) == 0 {
				continue
			}

			if !ctx.Quorum().HasQuorum(ibft.WrapMessages(validCommits...)) {
				continue
			}

			return validCommits, nil
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
