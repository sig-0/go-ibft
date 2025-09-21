package sequencer

import (
	"github.com/sig-0/go-ibft/message"
)

func (s *Sequencer) buildPrepareMessage(sequence *Sequence) *message.Prepare {
	msg := &message.Prepare{
		Sequence:  sequence.sequence,
		Round:     sequence.round,
		Sender:    s.validator.Address(),
		BlockHash: sequence.proposal.BlockHash,
	}

	// todo: keccak
	msg.Signature = s.validator.Sign(msg.Payload())

	return msg
}

//
//func (s *Sequencer) awaitPrepare(
//	ctx context.Context,
//	sequence *Sequence,
//	store *message.Store,
//) ([]*message.Prepare, error) {
//	sub, cancelSub := store.PrepareMessages.Subscribe(sequence.sequence, sequence.round, false)
//	defer cancelSub()
//
//	//cache := message.NewCache(s.isValidMsgPrepare)
//
//	for {
//		select {
//		case <-ctx.Done():
//			return nil, ctx.Err()
//		case notification := <-sub:
//			//cache.Add(notification()...)
//
//			messages, err := s.consensus.AwaitPrepare(ctx, *sequence, notification())
//			if err != nil {
//				// todo: log
//				continue
//			}
//
//			return messages, nil
//
//			//prepares := cache.Get()
//			//addresses := make([][]byte, 0, len(prepares))
//			//for _, commit := range prepares {
//			//	addresses = append(addresses, commit.GetSender())
//			//}
//			//
//			//if !s.verifier.HasQuorum(addresses, s.state.sequence) {
//			//	continue
//			//}
//			//
//			//return prepares, nil
//		}
//	}
//}

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
