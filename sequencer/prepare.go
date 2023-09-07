package sequencer

import (
	"bytes"
	"context"

	"github.com/madz-lab/go-ibft/message/types"
)

func (s *Sequencer) awaitPrepare(ctx context.Context, feed MessageFeed) error {
	messages, err := s.awaitQuorumPrepareMessages(ctx, feed)
	if err != nil {
		return err
	}

	s.state.PrepareCertificate(messages)
	s.transport.Multicast(s.buildMsgCommit())

	return nil
}

func (s *Sequencer) awaitQuorumPrepareMessages(ctx context.Context, feed MessageFeed) ([]*types.MsgPrepare, error) {
	sub, cancelSub := feed.SubscribeToPrepareMessages(s.state.currentView, false)
	defer cancelSub()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrapMessages := <-sub:
			validPrepares := types.Filter(unwrapMessages(), s.isValidMsgPrepare)
			if !s.quorum.HasQuorum(types.ToMsg(validPrepares)) {
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

func (s *Sequencer) buildMsgPrepare() *types.MsgPrepare {
	msg := &types.MsgPrepare{
		View:      s.state.currentView,
		From:      s.ID(),
		BlockHash: s.state.AcceptedBlockHash(),
	}

	msg.Signature = s.Sign(msg.Payload())

	return msg
}
