package sequencer

import (
	"bytes"
	"context"
	"github.com/sig-0/go-ibft/message"
	"github.com/sig-0/go-ibft/message/store"
)

func (s *Sequencer) sendMsgPrepare() {
	msg := &message.MsgPrepare{
		Info: &message.MsgInfo{
			Sequence: s.state.sequence,
			Round:    s.state.round,
			Sender:   s.validator.Address(),
		},
		BlockHash: s.state.getProposedBlockHash(),
	}

	msg = message.SignMsg(msg, s.validator)
	s.transport.MulticastPrepare(msg)
}

func (s *Sequencer) awaitPrepare(ctx context.Context) error {
	messages, err := s.awaitQuorumPrepares(ctx)
	if err != nil {
		return err
	}

	s.state.prepareCertificate(messages)

	return nil
}

func (s *Sequencer) awaitQuorumPrepares(ctx context.Context) ([]*message.MsgPrepare, error) {
	sub, cancelSub := s.feed.SubscribePrepare(0, 0, false)
	defer cancelSub()

	cache := store.NewMsgCache(func(msg *message.MsgPrepare) bool {
		return s.isValidMsgPrepare(msg)
	})

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			cache = cache.Add(notification.Unwrap())

			prepares := cache.Get()
			if !s.validatorSet.HasQuorum(message.WrapMessages(prepares...)) {
				continue
			}

			return prepares, nil
		}
	}
}

func (s *Sequencer) isValidMsgPrepare(msg *message.MsgPrepare) bool {
	if !s.validatorSet.IsValidator(msg.Info.Sender, msg.Info.Sequence) {
		return false
	}

	if !bytes.Equal(msg.BlockHash, s.state.getProposedBlockHash()) {
		return false
	}

	return true
}
