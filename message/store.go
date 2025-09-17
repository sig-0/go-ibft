package message

// Store is a thread-safe storage for consensus messages with a built-in sequencer.Feed mechanism
type Store struct {
	ProposalMessages    *Collection[*Proposal]
	PrepareMessages     *Collection[*Prepare]
	CommitMessages      *Collection[*Commit]
	RoundChangeMessages *Collection[*RoundChange]
}

// NewStore returns a new Store instance
func NewStore() *Store {
	return &Store{
		ProposalMessages:    NewMsgCollection[*Proposal](),
		PrepareMessages:     NewMsgCollection[*Prepare](),
		CommitMessages:      NewMsgCollection[*Commit](),
		RoundChangeMessages: NewMsgCollection[*RoundChange](),
	}
}
