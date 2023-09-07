package store

import (
	"github.com/madz-lab/go-ibft/message/types"
	"sync"
)

type msg interface {
	types.MsgProposal | types.MsgPrepare | types.MsgRoundChange | types.MsgCommit
}

type syncCollection[M msg] struct {
	collection[M]
	subscriptions[M]

	collectionMux,
	subscriptionMux sync.RWMutex
}

func newSyncCollection[M msg]() *syncCollection[M] {
	return &syncCollection[M]{
		collection:    newCollection[M](),
		subscriptions: newSubscriptions[M](),
	}
}

func (c *syncCollection[M]) subscribe(view *types.View, higherRounds bool) (subscription[M], func()) {
	sub := newSubscription[M](view, higherRounds)
	unregister := c.registerSub(sub)

	sub.notify(c.unwrapMessagesFn(view, higherRounds))

	return sub, unregister
}

func (c *syncCollection[M]) registerSub(sub subscription[M]) func() {
	c.subscriptionMux.Lock()
	defer c.subscriptionMux.Unlock()

	id := c.subscriptions.add(sub)

	return func() {
		c.subscriptionMux.Lock()
		defer c.subscriptionMux.Unlock()

		c.subscriptions.remove(id)
	}
}

func (c *syncCollection[M]) addMessage(msg *M, view *types.View, from []byte) {
	c.collectionMux.Lock()
	c.collection.addMessage(msg, view, from)
	c.collectionMux.Unlock()

	c.subscriptionMux.RLock()
	c.subscriptions.notify(func(sub subscription[M]) {
		if view.Sequence != sub.View.Sequence {
			return
		}

		if view.Round < sub.View.Round {
			return
		}

		sub.notify(c.unwrapMessagesFn(sub.View, sub.HigherRounds))
	})
	c.subscriptionMux.RUnlock()
}

func (c *syncCollection[M]) getMessages(view *types.View) []*M {
	c.collectionMux.RLock()
	defer c.collectionMux.RUnlock()

	return c.collection.get(view).messages()
}

func (c *syncCollection[M]) unwrapMessagesFn(view *types.View, higherRounds bool) func() []*M {
	return func() []*M {
		c.collectionMux.RLock()
		defer c.collectionMux.RUnlock()

		if !higherRounds {
			return c.collection.getMessages(view)
		}

		return c.collection.getHighestRoundMessages(view)
	}
}

func (c *syncCollection[M]) remove(view *types.View) {
	c.collectionMux.Lock()
	defer c.collectionMux.Unlock()

	c.collection.remove(view)
}

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

func (c *collection[M]) getHighestRoundMessages(view *types.View) []*M {
	maxRound := view.Round
	for round := range (*c)[view.Sequence] {
		if maxRound >= round {
			continue
		}

		maxRound = round
	}

	return c.getMessages(&types.View{Sequence: view.Sequence, Round: maxRound})
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
