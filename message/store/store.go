package store

import (
	"github.com/madz-lab/go-ibft/message/types"
)

//
//// MsgSigRecover extracts the sender associated with data and sig

// Store is a thread-safe storage for consensus messages with a built-in feed mechanism
type Store struct {
	proposal    *syncCollection[types.MsgProposal]
	prepare     *syncCollection[types.MsgPrepare]
	commit      *syncCollection[types.MsgCommit]
	roundChange *syncCollection[types.MsgRoundChange]
}

// New returns a new Store instance. MsgReceiverFn added to this store
// have their signatures verified before being included
func New() *Store {
	s := &Store{
		proposal:    newSyncCollection[types.MsgProposal](),
		prepare:     newSyncCollection[types.MsgPrepare](),
		commit:      newSyncCollection[types.MsgCommit](),
		roundChange: newSyncCollection[types.MsgRoundChange](),
	}

	return s
}

// AddMessage stores the provided msg if its signature is valid
func AddMessage[M types.IBFTMessage](m *M, store *Store) {
	switch msg := any(m).(type) {
	case *types.MsgProposal:
		store.proposal.AddMessage(msg, msg.View, msg.From)
	case *types.MsgPrepare:
		store.prepare.AddMessage(msg, msg.View, msg.From)
	case *types.MsgCommit:
		store.commit.AddMessage(msg, msg.View, msg.From)
	case *types.MsgRoundChange:
		store.roundChange.AddMessage(msg, msg.View, msg.From)
	}
}

func GetMessages[M types.IBFTMessage](view *types.View, store *Store) []*M {
	var (
		m        *M
		messages []*M
	)

	switch any(m).(type) {
	case *types.MsgProposal:
		typedMessages := store.proposal.GetMessages(view)
		messages = make([]*M, 0, len(typedMessages))

		for _, msg := range typedMessages {
			messages = append(messages, any(msg).(*M)) //nolint:forcetypeassert, gocritic
		}
	case *types.MsgPrepare:
		typedMessages := store.prepare.GetMessages(view)
		messages = make([]*M, 0, len(typedMessages))

		for _, msg := range typedMessages {
			messages = append(messages, any(msg).(*M)) //nolint:forcetypeassert, gocritic
		}
	case *types.MsgCommit:
		typedMessages := store.commit.GetMessages(view)
		messages = make([]*M, 0, len(typedMessages))

		for _, msg := range store.commit.GetMessages(view) {
			messages = append(messages, any(msg).(*M)) //nolint:forcetypeassert, gocritic
		}
	case *types.MsgRoundChange:
		typedMessages := store.roundChange.GetMessages(view)
		messages = make([]*M, 0, len(typedMessages))

		for _, msg := range store.roundChange.GetMessages(view) {
			messages = append(messages, any(msg).(*M)) //nolint:forcetypeassert, gocritic
		}
	}

	return messages
}

func RemoveMessages[M types.IBFTMessage](view *types.View, store *Store) {
	var m *M
	switch any(m).(type) {
	case *types.MsgProposal:
		store.proposal.Remove(view)
	case *types.MsgPrepare:
		store.prepare.Remove(view)
	case *types.MsgCommit:
		store.commit.Remove(view)
	case *types.MsgRoundChange:
		store.roundChange.Remove(view)
	}
}
