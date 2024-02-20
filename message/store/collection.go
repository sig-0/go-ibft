package store

import (
	"sync"

	"github.com/madz-lab/go-ibft/message/types"
)

type syncCollection[M types.IBFTMessage] struct {
	collection[M]
	subscriptions[M]

	collectionMux,
	subscriptionMux sync.RWMutex
}

func newSyncCollection[M types.IBFTMessage]() *syncCollection[M] {
	return &syncCollection[M]{
		collection:    collection[M]{},
		subscriptions: subscriptions[M]{},
	}
}

func (c *syncCollection[M]) Subscribe(view *types.View, higherRounds bool) (Subscription[M], func()) {
	sub := newSubscription[M](view, higherRounds)
	unregister := c.RegisterSubscription(sub)

	sub.Notify(c.unwrapMessagesFn(view, higherRounds))

	return sub.Channel, unregister
}

func (c *syncCollection[M]) RegisterSubscription(sub subscription[M]) func() {
	c.subscriptionMux.Lock()
	defer c.subscriptionMux.Unlock()

	id := c.subscriptions.Add(sub)

	return func() {
		c.subscriptionMux.Lock()
		defer c.subscriptionMux.Unlock()

		c.subscriptions.Remove(id)
	}
}

func (c *syncCollection[M]) AddMessage(msg *M, view *types.View, from []byte) {
	c.collectionMux.Lock()
	c.collection.AddMessage(msg, view, from)
	c.collectionMux.Unlock()

	c.subscriptionMux.RLock()
	c.subscriptions.Notify(func(sub subscription[M]) {
		if view.Sequence != sub.View.Sequence {
			return
		}

		if view.Round < sub.View.Round {
			return
		}

		sub.Notify(c.unwrapMessagesFn(sub.View, sub.HigherRounds))
	})
	c.subscriptionMux.RUnlock()
}

func (c *syncCollection[M]) GetMessages(view *types.View) []*M {
	c.collectionMux.RLock()
	defer c.collectionMux.RUnlock()

	return c.collection.Get(view).Messages()
}

func (c *syncCollection[M]) unwrapMessagesFn(view *types.View, higherRounds bool) MsgReceiverFn[M] {
	return func() []*M {
		c.collectionMux.RLock()
		defer c.collectionMux.RUnlock()

		if !higherRounds {
			return c.collection.GetMessages(view)
		}

		return c.collection.GetHighestRoundMessages(view)
	}
}

func (c *syncCollection[M]) Remove(view *types.View) {
	c.collectionMux.Lock()
	defer c.collectionMux.Unlock()

	c.collection.RemoveMessagesWithView(view)
}

type collection[M types.IBFTMessage] map[uint64]map[uint64]msgSet[M]

func (c *collection[M]) AddMessage(msg *M, view *types.View, from []byte) {
	c.Set(view)[string(from)] = msg
}

func (c *collection[M]) GetMessages(view *types.View) []*M {
	return c.Get(view).Messages()
}

func (c *collection[M]) Set(view *types.View) msgSet[M] {
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

func (c *collection[M]) Get(view *types.View) msgSet[M] {
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

func (c *collection[M]) GetHighestRoundMessages(view *types.View) []*M {
	maxRound := view.Round
	for round := range (*c)[view.Sequence] {
		if maxRound >= round {
			continue
		}

		maxRound = round
	}

	return c.GetMessages(&types.View{Sequence: view.Sequence, Round: maxRound})
}

func (c *collection[M]) RemoveMessagesWithView(view *types.View) {
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

type msgSet[M types.IBFTMessage] map[string]*M

func (s msgSet[M]) Messages() []*M {
	messages := make([]*M, 0, len(s))
	for _, msg := range s {
		messages = append(messages, msg)
	}

	return messages
}
