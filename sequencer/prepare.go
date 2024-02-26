package sequencer

import (
	"bytes"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) multicastPrepare(ctx ibft.Context) {
	msg := &types.MsgPrepare{
		From:      s.ID(),
		View:      s.state.CurrentView(),
		BlockHash: s.state.AcceptedBlockHash(),
	}

	msg.Signature = s.Sign(ctx.Keccak().Hash(msg.Payload()))

	ctx.Transport().Multicast(msg)
}

func (s *Sequencer) awaitPrepare(ctx ibft.Context) error {
	messages, err := s.awaitQuorumPrepares(ctx)
	if err != nil {
		return err
	}

	s.state.PrepareCertificate(messages)

	return nil
}

func (s *Sequencer) awaitQuorumPrepares(ctx ibft.Context) ([]*types.MsgPrepare, error) {
	sub, cancelSub := ctx.Feed().PrepareMessages(s.state.currentView, false)
	defer cancelSub()

	cache := newMsgCache(func(msg *types.MsgPrepare) bool {
		if !s.HasValidSignature(msg) {
			return false
		}

		return s.isValidMsgPrepare(msg)
	})

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			messages := notification.Unwrap()

			cache = cache.Add(messages)
			validPrepares := cache.Messages()

			if !ctx.Quorum().HasQuorum(ibft.WrapMessages(validPrepares...)) {
				continue
			}

			return validPrepares, nil
		}
	}
}

func (s *Sequencer) isValidMsgPrepare(msg *types.MsgPrepare) bool {
	if !s.IsValidator(msg.From, msg.View.Sequence) {
		return false
	}

	if !bytes.Equal(msg.BlockHash, s.state.AcceptedBlockHash()) {
		return false
	}

	return true
}
