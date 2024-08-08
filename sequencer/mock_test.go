package sequencer

import (
	"github.com/sig-0/go-ibft/message"
	"github.com/sig-0/go-ibft/message/transport"
)

type MockValidator struct {
	address           []byte
	signFn            func([]byte) []byte
	buildProposalFn   func(uint64) []byte
	isValidProposalFn func(uint64, []byte) bool
}

func (v MockValidator) Address() []byte {
	return v.address
}

func (v MockValidator) Sign(digest []byte) []byte {
	return v.signFn(digest)
}

func (v MockValidator) BuildProposal(sequence uint64) []byte {
	return v.buildProposalFn(sequence)
}

func (v MockValidator) IsValidProposal(proposal []byte, sequence uint64) bool {
	return v.isValidProposalFn(sequence, proposal)
}

type MockValidatorSet struct {
	isValidatorFn func([]byte, uint64) bool
	isProposerFn  func([]byte, uint64, uint64) bool
	hasQuorumFn   func([]message.Message) bool
}

func (vs MockValidatorSet) IsValidator(addr []byte, sequence uint64) bool {
	return vs.isValidatorFn(addr, sequence)
}

func (vs MockValidatorSet) IsProposer(addr []byte, sequence, round uint64) bool {
	return vs.isProposerFn(addr, sequence, round)
}

func (vs MockValidatorSet) HasQuorum(messages []message.Message) bool {
	return vs.hasQuorumFn(messages)
}

type MockKeccak func([]byte) []byte

func (k MockKeccak) Hash(digest []byte) []byte {
	return k(digest)
}

type MockSignatureVerifier func([]byte, []byte, []byte) error

func (s MockSignatureVerifier) Verify(signature, digest []byte, msg []byte) error {
	return s(signature, digest, msg)
}

type MockFeed struct {
	proposal    map[uint64]map[uint64][]*message.MsgProposal
	prepare     map[uint64]map[uint64][]*message.MsgPrepare
	commit      map[uint64]map[uint64][]*message.MsgCommit
	roundChange map[uint64]map[uint64][]*message.MsgRoundChange
}

func NewMockFeed(messages []message.Message) MockFeed {
	f := MockFeed{
		proposal:    map[uint64]map[uint64][]*message.MsgProposal{},
		prepare:     map[uint64]map[uint64][]*message.MsgPrepare{},
		commit:      map[uint64]map[uint64][]*message.MsgCommit{},
		roundChange: map[uint64]map[uint64][]*message.MsgRoundChange{},
	}

	for _, msg := range messages {
		switch m := msg.(type) {
		case *message.MsgProposal:
			proposalsInSequence, ok := f.proposal[m.Info.Sequence]
			if !ok {
				f.proposal[m.Info.Sequence] = map[uint64][]*message.MsgProposal{}
				proposalsInSequence = f.proposal[m.Info.Sequence]
			}

			if _, ok := proposalsInSequence[m.Info.Round]; !ok {
				proposalsInSequence[m.Info.Round] = []*message.MsgProposal{}
			}

			proposalsInSequence[m.Info.Round] = append(proposalsInSequence[m.Info.Round], m)

		case *message.MsgPrepare:
			preparesInSequence, ok := f.prepare[m.Info.Sequence]
			if !ok {
				f.prepare[m.Info.Sequence] = map[uint64][]*message.MsgPrepare{}
				preparesInSequence = f.prepare[m.Info.Sequence]
			}

			if _, ok := preparesInSequence[m.Info.Round]; !ok {
				preparesInSequence[m.Info.Round] = []*message.MsgPrepare{}
			}

			preparesInSequence[m.Info.Round] = append(preparesInSequence[m.Info.Round], m)

		case *message.MsgCommit:
			commitsInSequence, ok := f.commit[m.Info.Sequence]
			if !ok {
				f.commit[m.Info.Sequence] = map[uint64][]*message.MsgCommit{}
				commitsInSequence = f.commit[m.Info.Sequence]
			}

			if _, ok := commitsInSequence[m.Info.Round]; !ok {
				commitsInSequence[m.Info.Round] = []*message.MsgCommit{}
			}

			commitsInSequence[m.Info.Round] = append(commitsInSequence[m.Info.Round], m)

		case *message.MsgRoundChange:
			roundChangesInSequence, ok := f.roundChange[m.Info.Sequence]
			if !ok {
				f.roundChange[m.Info.Sequence] = map[uint64][]*message.MsgRoundChange{}
				roundChangesInSequence = f.roundChange[m.Info.Sequence]
			}

			if _, ok := roundChangesInSequence[m.Info.Round]; !ok {
				roundChangesInSequence[m.Info.Round] = []*message.MsgRoundChange{}
			}

			roundChangesInSequence[m.Info.Round] = append(roundChangesInSequence[m.Info.Round], m)
		}
	}

	return f
}

func (f MockFeed) SubscribeProposal(sequence, round uint64, higherRounds bool) (message.Subscription[*message.MsgProposal], func()) {
	sub := make(message.Subscription[*message.MsgProposal], 1)
	if !higherRounds {
		notification := func() []*message.MsgProposal {
			return f.proposal[sequence][round]
		}

		sub <- message.MsgNotificationFn[*message.MsgProposal](notification)

		return sub, func() { close(sub) }
	}

	var highestRound uint64
	for round, _ := range f.proposal[sequence] {
		if round >= highestRound {
			highestRound = round
		}
	}

	if highestRound < round {
		sub <- message.MsgNotificationFn[*message.MsgProposal](func() []*message.MsgProposal {
			return nil
		})

		return sub, func() { close(sub) }
	}

	sub <- message.MsgNotificationFn[*message.MsgProposal](func() []*message.MsgProposal {
		return f.proposal[sequence][highestRound]
	})

	return sub, func() { close(sub) }
}

func (f MockFeed) SubscribePrepare(sequence, round uint64, higherRounds bool) (message.Subscription[*message.MsgPrepare], func()) {
	sub := make(message.Subscription[*message.MsgPrepare], 1)
	if !higherRounds {
		notification := func() []*message.MsgPrepare {
			return f.prepare[sequence][round]
		}

		sub <- message.MsgNotificationFn[*message.MsgPrepare](notification)

		return sub, func() { close(sub) }
	}

	var highestRound uint64
	for round, _ := range f.prepare[sequence] {
		if round >= highestRound {
			highestRound = round
		}
	}

	if highestRound < round {
		sub <- message.MsgNotificationFn[*message.MsgPrepare](func() []*message.MsgPrepare {
			return nil
		})

		return sub, func() { close(sub) }
	}

	sub <- message.MsgNotificationFn[*message.MsgPrepare](func() []*message.MsgPrepare {
		return f.prepare[sequence][highestRound]
	})

	return sub, func() { close(sub) }
}

func (f MockFeed) SubscribeCommit(sequence, round uint64, higherRounds bool) (message.Subscription[*message.MsgCommit], func()) {
	sub := make(message.Subscription[*message.MsgCommit], 1)
	if !higherRounds {
		notification := func() []*message.MsgCommit {
			return f.commit[sequence][round]
		}

		sub <- message.MsgNotificationFn[*message.MsgCommit](notification)

		return sub, func() { close(sub) }
	}

	var highestRound uint64
	for round, _ := range f.commit[sequence] {
		if round >= highestRound {
			highestRound = round
		}
	}

	if highestRound < round {
		sub <- message.MsgNotificationFn[*message.MsgCommit](func() []*message.MsgCommit {
			return nil
		})

		return sub, func() { close(sub) }
	}

	sub <- message.MsgNotificationFn[*message.MsgCommit](func() []*message.MsgCommit {
		return f.commit[sequence][highestRound]
	})

	return sub, func() { close(sub) }
}

func (f MockFeed) SubscribeRoundChange(sequence, round uint64, higherRounds bool) (message.Subscription[*message.MsgRoundChange], func()) {
	sub := make(message.Subscription[*message.MsgRoundChange], 1)
	if !higherRounds {
		notification := func() []*message.MsgRoundChange {
			return f.roundChange[sequence][round]
		}

		sub <- message.MsgNotificationFn[*message.MsgRoundChange](notification)

		return sub, func() { close(sub) }
	}

	var highestRound uint64
	for round, _ := range f.roundChange[sequence] {
		if round >= highestRound {
			highestRound = round
		}
	}

	if highestRound < round {
		sub <- message.MsgNotificationFn[*message.MsgRoundChange](func() []*message.MsgRoundChange {
			return nil
		})

		return sub, func() { close(sub) }
	}

	sub <- message.MsgNotificationFn[*message.MsgRoundChange](func() []*message.MsgRoundChange {
		return f.roundChange[sequence][highestRound]
	})

	return sub, func() { close(sub) }
}

type SingeRoundMockFeed MockFeed

func NewSingleRoundMockFeed(messages []message.Message) SingeRoundMockFeed {
	return SingeRoundMockFeed(NewMockFeed(messages))
}

func (f SingeRoundMockFeed) SubscribeProposal(sequence, round uint64, higherRounds bool) (message.Subscription[*message.MsgProposal], func()) {
	sub := make(message.Subscription[*message.MsgProposal], 1)
	if !higherRounds {
		notification := func() []*message.MsgProposal {
			return f.proposal[sequence][round]
		}

		sub <- message.MsgNotificationFn[*message.MsgProposal](notification)

		return sub, func() { close(sub) }
	}

	sub <- message.MsgNotificationFn[*message.MsgProposal](func() []*message.MsgProposal {
		return nil
	})

	return sub, func() { close(sub) }
}

func (f SingeRoundMockFeed) SubscribePrepare(sequence, round uint64, higherRounds bool) (message.Subscription[*message.MsgPrepare], func()) {
	sub := make(message.Subscription[*message.MsgPrepare], 1)
	if !higherRounds {
		notification := func() []*message.MsgPrepare {
			return f.prepare[sequence][round]
		}

		sub <- message.MsgNotificationFn[*message.MsgPrepare](notification)

		return sub, func() { close(sub) }
	}

	sub <- message.MsgNotificationFn[*message.MsgPrepare](func() []*message.MsgPrepare {
		return nil
	})

	return sub, func() { close(sub) }
}

func (f SingeRoundMockFeed) SubscribeCommit(sequence, round uint64, higherRounds bool) (message.Subscription[*message.MsgCommit], func()) {
	sub := make(message.Subscription[*message.MsgCommit], 1)
	if !higherRounds {
		notification := func() []*message.MsgCommit {
			return f.commit[sequence][round]
		}

		sub <- message.MsgNotificationFn[*message.MsgCommit](notification)

		return sub, func() { close(sub) }
	}

	sub <- message.MsgNotificationFn[*message.MsgCommit](func() []*message.MsgCommit {
		return nil
	})

	return sub, func() { close(sub) }
}

func (f SingeRoundMockFeed) SubscribeRoundChange(sequence, round uint64, higherRounds bool) (message.Subscription[*message.MsgRoundChange], func()) {
	sub := make(message.Subscription[*message.MsgRoundChange], 1)
	if !higherRounds {
		notification := func() []*message.MsgRoundChange {
			return f.roundChange[sequence][round]
		}

		sub <- message.MsgNotificationFn[*message.MsgRoundChange](notification)

		return sub, func() { close(sub) }
	}

	sub <- message.MsgNotificationFn[*message.MsgRoundChange](func() []*message.MsgRoundChange {
		return nil
	})

	return sub, func() { close(sub) }
}

func DummyTransport() message.Transport {
	return transport.NewTransport(
		func(_ *message.MsgProposal) {},
		func(_ *message.MsgPrepare) {},
		func(_ *message.MsgCommit) {},
		func(_ *message.MsgRoundChange) {},
	)
}
