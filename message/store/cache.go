package store

import "github.com/sig-0/go-ibft/message"

type MsgCache[M message.IBFTMessage] struct {
	filterFn func(M) bool
	seen     map[string]struct{}
	messages []M
}

func NewMsgCache[M message.IBFTMessage](filterFn func(M) bool) MsgCache[M] {
	return MsgCache[M]{
		filterFn: filterFn,
		messages: make([]M, 0),
		seen:     make(map[string]struct{}),
	}
}

func (c MsgCache[M]) Add(messages []M) MsgCache[M] {
	for _, msg := range messages {
		sender := message.Message(msg).GetInfo().Sender
		if _, ok := c.seen[string(sender)]; ok {
			continue
		}

		c.seen[string(sender)] = struct{}{}

		if !c.filterFn(msg) {
			continue
		}

		c.messages = append(c.messages, msg)
	}

	return c
}

func (c MsgCache[M]) Get() []M {
	return c.messages
}
