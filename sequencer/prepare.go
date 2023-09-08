package sequencer

import (
	"bytes"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) multicastPrepare(ctx ibft.Context) {
	msg := &types.MsgPrepare{
		From:      s.ID(),
		View:      s.state.currentView,
		BlockHash: s.state.AcceptedBlockHash(),
	}

	msg.Signature = s.Sign(msg.Payload())

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
	cache := newMsgCache(s.isValidMsgPrepare)

	sub, cancelSub := ctx.Feed().Prepare(s.state.currentView, false)
	defer cancelSub()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrapMessages := <-sub:
			cache = cache.Add(unwrapMessages())

			validPrepares := cache.Messages()
			if !ctx.Quorum().HasQuorum(s.state.CurrentSequence(), types.ToMsg(validPrepares)) {
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
