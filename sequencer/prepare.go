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
		BlockHash: s.state.acceptedBlockHash(),
	}

	s.transport.MulticastPrepare(message.SignMsg(msg, s.validator))
}

func (s *Sequencer) awaitPrepareQuorum(ctx context.Context) ([]*message.MsgPrepare, error) {
	sub, cancelSub := s.feed.SubscribePrepare(s.state.sequence, s.state.round, false)
	defer cancelSub()

	cache := store.NewMsgCache(s.isValidMsgPrepare)

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
	// sender is part of the validator set
	if !s.validatorSet.IsValidator(msg.Info.Sender, msg.Info.Sequence) {
		return false
	}

	// block hash and accepted block hash match
	if !bytes.Equal(msg.BlockHash, s.state.acceptedBlockHash()) {
		return false
	}

	return true
}
