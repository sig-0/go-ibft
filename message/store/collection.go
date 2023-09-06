package store

import "github.com/madz-lab/go-ibft/message/types"

type msg interface {
	types.MsgProposal | types.MsgPrepare | types.MsgRoundChange | types.MsgCommit
}

func newCollection[M msg]() collection[M] {
	return collection[M]{
		collection: map[uint64]map[uint64]*msgSet[M]{},
	}
}

type collection[M msg] struct {
	collection map[uint64]map[uint64]*msgSet[M]
}

func (c *collection[M]) addMessage(msg *M, view *types.View, from []byte) {
	sameSequenceMessages, ok := c.collection[view.Sequence]
	if !ok {
		c.collection[view.Sequence] = map[uint64]*msgSet[M]{}
		sameSequenceMessages = c.collection[view.Sequence]
	}

	sameRoundMessages, ok := sameSequenceMessages[view.Round]
	if !ok {
		sameSequenceMessages[view.Round] = &msgSet[M]{s: map[string]*M{}}
		sameRoundMessages = sameSequenceMessages[view.Round]
	}

	sameRoundMessages.s[string(from)] = msg
}

func (c *collection[M]) getMessages(view *types.View) []*M {
	sameSequenceMessages, ok := c.collection[view.Sequence]
	if !ok {
		return nil
	}

	sameRoundMessages, ok := sameSequenceMessages[view.Round]
	if !ok {
		return nil
	}

	messages := make([]*M, 0, len(sameRoundMessages.s))
	for _, msg := range sameRoundMessages.s {
		messages = append(messages, msg)
	}

	return messages
}

type msgSet[M msg] struct {
	s map[string]*M
}
