package sequencer

import (
	"bytes"
	"github.com/sig-0/go-ibft"

	"github.com/sig-0/go-ibft/message/types"
)

func (s *Sequencer) sendMsgCommit(ctx Context) {
	msg := &types.MsgCommit{
		BlockHash:  s.state.getProposedBlockHash(),
		CommitSeal: s.validator.Sign(s.state.getProposedBlockHash()),
		Metadata: &types.MsgMetadata{
			View:   s.state.getView(),
			Sender: s.validator.ID(),
		},
	}

	msg.Metadata.Signature = s.validator.Sign(ctx.Keccak().Hash(msg.Payload()))

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
		return s.isValidMsgCommit(msg, ctx.SigVerifier())
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

func (s *Sequencer) isValidMsgCommit(msg *types.MsgCommit, sigVerifier ibft.SigVerifier) bool {
	// sender is in the validator set
	if !s.validatorSet.IsValidator(msg.Sender(), msg.Sequence()) {
		return false
	}

	// block hash is the same as block hash of the accepted proposal
	if !bytes.Equal(msg.BlockHash, s.state.getProposedBlockHash()) {
		return false
	}

	// sender generated commit seal by signing over block hash
	if err := sigVerifier.Verify(msg.Sender(), msg.BlockHash, msg.CommitSeal); err != nil {
		return false
	}

	return true
}
