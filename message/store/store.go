package store

import (
	"bytes"
	"errors"

	"github.com/madz-lab/go-ibft/message/types"
)

var ErrInvalidSignature = errors.New("invalid signature")

type msgVerifier struct {
	types.SigRecover
}

// Store is a thread-safe storage for consensus messages with a built-in Feed mechanism
type Store struct {
	verifier msgVerifier

	proposal    *syncCollection[types.MsgProposal]
	prepare     *syncCollection[types.MsgPrepare]
	commit      *syncCollection[types.MsgCommit]
	roundChange *syncCollection[types.MsgRoundChange]
}

// New returns a new Store instance. Messages added to this store
// have their signatures verified before being included
func New(sigRecover types.SigRecover) *Store {
	s := &Store{
		verifier:    msgVerifier{sigRecover},
		proposal:    newSyncCollection[types.MsgProposal](),
		prepare:     newSyncCollection[types.MsgPrepare](),
		commit:      newSyncCollection[types.MsgCommit](),
		roundChange: newSyncCollection[types.MsgRoundChange](),
	}

	return s
}

func isValidSignature[M msg](msg *M, sigRecover types.SigRecover) bool {
	msgI := any(msg).(types.Msg) //nolint:forcetypeassert, gocritic, errcheck

	return bytes.Equal(msgI.GetFrom(), sigRecover.From(msgI.Payload(), msgI.GetSignature()))
}

// AddMessage stores the provided msg if its signature is valid
func AddMessage[M msg](m *M, store *Store) error {
	if !isValidSignature(m, store.verifier) {
		return ErrInvalidSignature
	}

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

	return nil
}

func GetMessages[M msg](view *types.View, store *Store) []*M {
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

func RemoveMessages[M msg](view *types.View, store *Store) {
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
