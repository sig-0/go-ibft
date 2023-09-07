package store

import (
	"github.com/madz-lab/go-ibft/message/types"
)

type msg interface {
	types.MsgProposal | types.MsgPrepare | types.MsgRoundChange | types.MsgCommit
}

func newCollection[M msg]() collection[M] {
	return map[uint64]map[uint64]msgSet[M]{}
}

type collection[M msg] map[uint64]map[uint64]msgSet[M]

func (c *collection[M]) addMessage(msg *M, view *types.View, from []byte) {
	sameSequenceMessages, ok := (*c)[view.Sequence]
	if !ok {
		(*c)[view.Sequence] = map[uint64]msgSet[M]{}
		sameSequenceMessages = (*c)[view.Sequence]
	}

	sameRoundMessages, ok := sameSequenceMessages[view.Round]
	if !ok {
		(*c)[view.Sequence][view.Round] = msgSet[M]{}
		sameRoundMessages = (*c)[view.Sequence][view.Round]
	}

	sameRoundMessages[string(from)] = msg
}

func (c *collection[M]) getMessages(view *types.View) []*M {
	sameSequenceMessages, ok := (*c)[view.Sequence]
	if !ok {
		return nil
	}

	sameRoundMessages, ok := sameSequenceMessages[view.Round]
	if !ok {
		return nil
	}

	return sameRoundMessages.messages()
}

func (c *collection[M]) getMaxRoundMessages(view *types.View) []*M {
	maxRound := view.Round
	for round := range (*c)[view.Sequence] {
		if maxRound >= round {
			continue
		}

		maxRound = round
	}

	return c.getMessages(&types.View{Sequence: view.Sequence, Round: maxRound})
}

func (c *collection[M]) unwrapFn(view *types.View, higherRounds bool) func() []*M {
	return func() []*M {
		if !higherRounds {
			return c.getMessages(view)
		}

		return c.getMaxRoundMessages(view)
	}
}

func (c *collection[M]) remove(view *types.View) {
	_, ok := (*c)[view.Sequence]
	if !ok {
		return
	}

	_, ok = (*c)[view.Sequence][view.Round]
	if !ok {
		return
	}

	delete((*c)[view.Sequence], view.Round)
}

type msgSet[M msg] map[string]*M

func (s msgSet[M]) messages() []*M {
	messages := make([]*M, 0, len(s))
	for _, msg := range s {
		messages = append(messages, msg)
	}

	return messages
}
