package sequencer

import (
	"context"

	"github.com/sig-0/go-ibft/message"
)

func (s *Sequencer) sendMsgPrepare() {
	msg := &message.Prepare{
		Sequence:  s.sequence.sequence,
		Round:     s.sequence.round,
		Sender:    s.validator.Address(),
		BlockHash: s.sequence.acceptedBlockHash(),
	}

	msg.Signature = s.validator.Sign(msg.Payload())

	s.transport.MulticastPrepare(msg)
}

func (s *Sequencer) awaitPrepareQuorum(ctx context.Context) ([]*message.Prepare, error) {
	sub, cancelSub := s.feed.PrepareMessages.Subscribe(s.sequence.sequence, s.sequence.round, false)
	defer cancelSub()

	//cache := message.NewCache(s.isValidMsgPrepare)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case notification := <-sub:
			//cache.Add(notification()...)

			messages, err := s.vrf.CheckPrepare(ctx, &s.sequence, notification())
			if err != nil {
				// todo: log
				continue
			}

			return messages, nil

			//prepares := cache.Get()
			//addresses := make([][]byte, 0, len(prepares))
			//for _, commit := range prepares {
			//	addresses = append(addresses, commit.GetSender())
			//}
			//
			//if !s.verifier.HasQuorum(addresses, s.state.sequence) {
			//	continue
			//}
			//
			//return prepares, nil
		}
	}
}

//
//func (s *Sequencer) isValidMsgPrepare(msg *message.Prepare) bool {
//	// sender is part of the validator set
//	if !s.verifier.IsValidator(msg.Sender, msg.Sequence) {
//		return false
//	}
//
//	// block hash and accepted block hash match
//	if !bytes.Equal(msg.BlockHash, s.state.acceptedBlockHash()) {
//		return false
//	}
//
//	return true
//}
