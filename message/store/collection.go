package store

import (
	"github.com/sig-0/go-ibft/message"
	"sync"
)

type MsgCollection[M message.IBFTMessage] interface {
	Add(msg M)
	Get(view *message.View) []M
	Remove(view *message.View)
	Subscribe(view *message.View, higherRounds bool) (message.Subscription[M], func())
	Clear()
}

type syncCollection[M message.IBFTMessage] struct {
	msgCollection[M]
	subscriptions[M]

	collectionMux, subscriptionMux sync.RWMutex
}

func NewMsgCollection[M message.IBFTMessage]() MsgCollection[M] {
	return &syncCollection[M]{
		msgCollection: msgCollection[M]{},
		subscriptions: subscriptions[M]{},
	}
}

func (c *syncCollection[M]) Clear() {
	c.collectionMux.Lock()
	defer c.collectionMux.Unlock()

	clear(c.msgCollection)
}

func (c *syncCollection[M]) Subscribe(view *message.View, higherRounds bool) (message.Subscription[M], func()) {
	sub := newSubscription[M](view, higherRounds)
	unregister := c.registerSubscription(sub)

	sub.Notify(c.getNotificationFn(view, higherRounds))

	return sub.Channel, unregister
}

func (c *syncCollection[M]) registerSubscription(sub subscription[M]) func() {
	c.subscriptionMux.Lock()
	defer c.subscriptionMux.Unlock()

	id := c.subscriptions.Add(sub)

	return func() {
		c.subscriptionMux.Lock()
		defer c.subscriptionMux.Unlock()

		c.subscriptions.Remove(id)
	}
}

func (c *syncCollection[M]) Add(msg M) {
	c.collectionMux.Lock()
	defer c.collectionMux.Unlock()

	c.msgCollection.add(msg)

	c.subscriptionMux.RLock()
	defer c.subscriptionMux.RUnlock()

	info := message.Message(msg).GetInfo()
	seq, round := info.View.Sequence, info.View.Round

	c.subscriptions.Notify(func(sub subscription[M]) {
		// match the sequence
		if seq != sub.View.Sequence {
			return
		}

		// exclude lower rounds
		if round < sub.View.Round {
			return
		}

		sub.Notify(c.getNotificationFn(sub.View, sub.HigherRounds))
	})
}

func (c *syncCollection[M]) Get(view *message.View) []M {
	c.collectionMux.RLock()
	defer c.collectionMux.RUnlock()

	return c.msgCollection.loadSet(view).Messages()
}

func (c *syncCollection[M]) getNotificationFn(view *message.View, higherRounds bool) message.MsgNotificationFn[M] {
	return func() []M {
		c.collectionMux.RLock()
		defer c.collectionMux.RUnlock()

		if !higherRounds {
			return c.msgCollection.get(view)
		}

		return c.msgCollection.getMessagesWithHighestRoundNumber(view)
	}
}

func (c *syncCollection[M]) Remove(view *message.View) {
	c.collectionMux.Lock()
	defer c.collectionMux.Unlock()

	c.msgCollection.remove(view)
}

type msgCollection[M message.IBFTMessage] map[uint64]map[uint64]msgSet[M]

func (c *msgCollection[M]) add(msg M) {
	var (
		info     = message.Message(msg).GetInfo()
		sequence = info.View.Sequence
		round    = info.View.Round
		sender   = string(info.Sender)
	)

	c.loadOrStoreSet(sequence, round)[sender] = msg
}

func (c *msgCollection[M]) loadOrStoreSet(sequence, round uint64) msgSet[M] {
	sameSequenceMessages, ok := (*c)[sequence]
	if !ok {
		(*c)[sequence] = map[uint64]msgSet[M]{}
		sameSequenceMessages = (*c)[sequence]
	}

	set, ok := sameSequenceMessages[round]
	if !ok {
		(*c)[sequence][round] = msgSet[M]{}
		set = (*c)[sequence][round]
	}

	return set
}

func (c *msgCollection[M]) get(view *message.View) []M {
	return c.loadSet(view).Messages()
}

func (c *msgCollection[M]) loadSet(view *message.View) msgSet[M] {
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

func (c *msgCollection[M]) getMessagesWithHighestRoundNumber(view *message.View) []M {
	maxRound := view.Round
	for round := range (*c)[view.Sequence] {
		if maxRound >= round {
			continue
		}

		maxRound = round
	}

	return c.get(&message.View{Sequence: view.Sequence, Round: maxRound})
}

func (c *msgCollection[M]) remove(view *message.View) {
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

type msgSet[M message.IBFTMessage] map[string]M

func (s msgSet[M]) Messages() []M {
	messages := make([]M, 0, len(s))
	for _, msg := range s {
		messages = append(messages, msg)
	}

	return messages
}
