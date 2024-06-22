package sequencer

import (
	"bytes"

	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) sendMsgPrepare(ctx Context) {
	msg := &types.MsgPrepare{
		From:      s.ID(),
		View:      s.state.getView(),
		BlockHash: s.state.getProposalBlockHash(),
	}

	msg.Signature = s.Sign(ctx.Keccak().Hash(msg.Payload()))

	ctx.Transport().MulticastPrepare(msg)
}

func (s *Sequencer) awaitPrepare(ctx Context) error {
	messages, err := s.awaitQuorumPrepares(ctx)
	if err != nil {
		return err
	}

	s.state.prepareCertificate(messages)

	return nil
}

func (s *Sequencer) awaitQuorumPrepares(ctx Context) ([]*types.MsgPrepare, error) {
	sub, cancelSub := ctx.MessageFeed().PrepareMessages(s.state.getView(), false)
	defer cancelSub()

	cache := newMsgCache(func(msg *types.MsgPrepare) bool {
		return s.isValidMsgPrepare(msg)
	})

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			cache = cache.add(notification.Unwrap())

			prepares := cache.get()
			if !ctx.Quorum().HasQuorum(types.WrapMessages(prepares...)) {
				continue
			}

			return prepares, nil
		}
	}
}

func (s *Sequencer) isValidMsgPrepare(msg *types.MsgPrepare) bool {
	if !s.IsValidator(msg.From, msg.View.Sequence) {
		return false
	}

	if !bytes.Equal(msg.BlockHash, s.state.getProposalBlockHash()) {
		return false
	}

	return true
}
