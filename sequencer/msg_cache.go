package sequencer

import (
	"github.com/madz-lab/go-ibft/message/types"
)

type msg interface {
	types.IBFTMessage

	GetSender() []byte
}

type msgCache[M msg] struct {
	filterFn func(M) bool
	seen     map[string]struct{}
	filtered []M
}

func newMsgCache[M msg](filterFn func(M) bool) msgCache[M] {
	return msgCache[M]{
		filterFn: filterFn,
		filtered: make([]M, 0),
		seen:     make(map[string]struct{}),
	}
}

func (c msgCache[M]) Add(messages []M) msgCache[M] {
	for _, msg := range messages {
		sender := msg.GetSender()
		if _, ok := c.seen[string(sender)]; ok {
			continue
		}

		c.seen[string(sender)] = struct{}{}

		if !c.filterFn(msg) {
			continue
		}

		c.filtered = append(c.filtered, msg)
	}

	return c
}

func (c msgCache[M]) Get() []M {
	return c.filtered
}
