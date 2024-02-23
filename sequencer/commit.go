package sequencer

import (
	"bytes"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) multicastCommit(ctx ibft.Context) {
	msg := &types.MsgCommit{
		From:       s.ID(),
		View:       s.state.currentView,
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
	sigRecover := ctx.SigRecover()
	cache := newMsgCache(func(msg *types.MsgCommit) bool {
		if !s.HasValidSignature(msg) {
			return false
		}

		return s.isValidCommit(msg, sigRecover)
	})

	sub, cancelSub := ctx.Feed().CommitMessages(s.state.currentView, false)
	defer cancelSub()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			messages := notification.Unwrap()

			cache = cache.Add(messages)
			validCommits := cache.Messages()

			if len(validCommits) == 0 {
				continue
			}

			if !ctx.Quorum().HasQuorum(s.state.CurrentSequence(), ibft.WrapMessages(validCommits)) {
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
