package store

import (
	"github.com/madz-lab/go-ibft/message/types"
)

type msg interface {
	types.MsgProposal | types.MsgPrepare | types.MsgRoundChange | types.MsgCommit
}

//
//type syncCollection[M msg] struct {
//	collection[M]
//	mux sync.RWMutex
//
//	subs subscriptions[M]
//}
//
//func newSyncCollection[M msg]() *syncCollection[M] {
//	return &syncCollection[M]{
//		collection: newCollection[M](),
//		subs:       newSubscriptions[M](),
//	}
//}

func newCollection[M msg]() collection[M] {
	return map[uint64]map[uint64]msgSet[M]{}
}

type collection[M msg] map[uint64]map[uint64]msgSet[M]

func (c *collection[M]) addMessage(msg *M, view *types.View, from []byte) {
	c.set(view)[string(from)] = msg
}

func (c *collection[M]) getMessages(view *types.View) []*M {
	return c.get(view).messages()
}

func (c *collection[M]) set(view *types.View) msgSet[M] {
	sameSequenceMessages, ok := (*c)[view.Sequence]
	if !ok {
		(*c)[view.Sequence] = map[uint64]msgSet[M]{}
		sameSequenceMessages = (*c)[view.Sequence]
	}

	set, ok := sameSequenceMessages[view.Round]
	if !ok {
		(*c)[view.Sequence][view.Round] = msgSet[M]{}
		set = (*c)[view.Sequence][view.Round]
	}

	return set
}

func (c *collection[M]) get(view *types.View) msgSet[M] {
	sameSequenceMessages, ok := (*c)[view.Sequence]
	if !ok {
		return nil
	}

	set, ok := sameSequenceMessages[view.Round]
	if !ok {
		return nil
	}

	return set
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
