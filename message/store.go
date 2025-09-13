package message

// Store is a thread-safe storage for consensus messages with a built-in sequencer.Feed mechanism
type Store struct {
	ProposalMessages    *Collection[*Proposal]
	PrepareMessages     *Collection[*Prepare]
	CommitMessages      *Collection[*Commit]
	RoundChangeMessages *Collection[*RoundChange]
}

// NewMsgStore returns a new Store instance
func NewMsgStore() *Store {
	return &Store{
		ProposalMessages:    NewMsgCollection[*Proposal](),
		PrepareMessages:     NewMsgCollection[*Prepare](),
		CommitMessages:      NewMsgCollection[*Commit](),
		RoundChangeMessages: NewMsgCollection[*RoundChange](),
	}
}

// Add includes the message in the store
func (s *Store) Add(msg message) {
	switch msg := msg.(type) {
	case *Proposal:
		s.ProposalMessages.Add(msg)
	case *Prepare:
		s.PrepareMessages.Add(msg)
	case *Commit:
		s.CommitMessages.Add(msg)
	case *RoundChange:
		s.RoundChangeMessages.Add(msg)
	}

	panic("unknown message")
}

// Clear removes all messages from store
func (s *Store) Clear() {
	s.ProposalMessages.Clear()
	s.PrepareMessages.Clear()
	s.CommitMessages.Clear()
	s.RoundChangeMessages.Clear()
}
