package store

import (
	"github.com/sig-0/go-ibft/message"
	"sync"
)

type MsgCollection[M message.IBFTMessage] interface {
	Add(msg M)
	Get(sequence uint64, round uint64) []M
	Remove(sequence uint64, round uint64)
	Subscribe(sequence, round uint64, higherRounds bool) (message.Subscription[M], func())
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

func (c *syncCollection[M]) Subscribe(sequence, round uint64, higherRounds bool) (message.Subscription[M], func()) {
	sub := newSubscription[M](sequence, round, higherRounds)
	unregister := c.registerSubscription(sub)

	sub.Notify(c.getNotificationFn(sequence, round, higherRounds))

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
	seq, round := info.Sequence, info.Round

	c.subscriptions.Notify(func(sub subscription[M]) {
		// match the sequence
		if seq != sub.Sequence {
			return
		}

		// exclude lower rounds
		if round < sub.Round {
			return
		}

		sub.Notify(c.getNotificationFn(sub.Sequence, sub.Round, sub.HigherRounds))
	})
}

func (c *syncCollection[M]) Get(sequence, round uint64) []M {
	c.collectionMux.RLock()
	defer c.collectionMux.RUnlock()

	return c.msgCollection.loadSet(sequence, round).Messages()
}

func (c *syncCollection[M]) getNotificationFn(sequence, round uint64, higherRounds bool) message.MsgNotificationFn[M] {
	return func() []M {
		c.collectionMux.RLock()
		defer c.collectionMux.RUnlock()

		if !higherRounds {
			return c.msgCollection.get(sequence, round)
		}

		return c.msgCollection.getMessagesWithHighestRoundNumber(sequence, round)
	}
}

func (c *syncCollection[M]) Remove(sequence, round uint64) {
	c.collectionMux.Lock()
	defer c.collectionMux.Unlock()

	c.msgCollection.remove(sequence, round)
}

type msgCollection[M message.IBFTMessage] map[uint64]map[uint64]msgSet[M]

func (c *msgCollection[M]) add(msg M) {
	var (
		info     = message.Message(msg).GetInfo()
		sequence = info.Sequence
		round    = info.Round
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

func (c *msgCollection[M]) get(sequence, round uint64) []M {
	return c.loadSet(sequence, round).Messages()
}

func (c *msgCollection[M]) loadSet(sequence, round uint64) msgSet[M] {
	sameSequenceMessages, ok := (*c)[sequence]
	if !ok {
		return nil
	}

	set, ok := sameSequenceMessages[round]
	if !ok {
		return nil
	}

	return set
}

func (c *msgCollection[M]) getMessagesWithHighestRoundNumber(sequence, round uint64) []M {
	maxRound := round
	for round := range (*c)[sequence] {
		if maxRound >= round {
			continue
		}

		maxRound = round
	}

	return c.get(sequence, maxRound)
}

func (c *msgCollection[M]) remove(sequence, round uint64) {
	_, ok := (*c)[sequence]
	if !ok {
		return
	}

	_, ok = (*c)[sequence][round]
	if !ok {
		return
	}

	delete((*c)[sequence], round)
}

type msgSet[M message.IBFTMessage] map[string]M

func (s msgSet[M]) Messages() []M {
	messages := make([]M, 0, len(s))
	for _, msg := range s {
		messages = append(messages, msg)
	}

	return messages
}
