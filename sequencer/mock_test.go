//nolint:dupl //because
package sequencer

import (
	"github.com/sig-0/go-ibft/message"
)

var (
	Alice = []byte("Alice")
	Bob   = []byte("Bob")
	Chris = []byte("Chris")
	Nina  = []byte("Nina")

	AlwaysAValidator     = func([]byte, uint64) bool { return true }
	AlwaysValidSignature = mockSignatureVerifier(func(_, _, _ []byte) error { return nil })
	AlwaysValidProposal  = func(_ uint64, _ []byte) bool { return true }
	DummyKeccak          = KeccakFn(func(_ []byte) []byte { return DummyKeccakValue })
	DummyKeccakValue     = []byte("keccak")
	DummySignFn          = func(_ []byte) []byte { return nil }
)

type dummyTransport struct{}

func (t dummyTransport) MulticastProposal(_ *message.MsgProposal) {}

func (t dummyTransport) MulticastPrepare(_ *message.MsgPrepare) {}

func (t dummyTransport) MulticastCommit(_ *message.MsgCommit) {}

func (t dummyTransport) MulticastRoundChange(_ *message.MsgRoundChange) {}

type mockValidator struct {
	signFn            func([]byte) []byte
	buildProposalFn   func(uint64) []byte
	isValidProposalFn func(uint64, []byte) bool
	address           []byte
}

func (v mockValidator) Address() []byte {
	return v.address
}

func (v mockValidator) Sign(digest []byte) []byte {
	return v.signFn(digest)
}

func (v mockValidator) BuildProposal(sequence uint64) []byte {
	return v.buildProposalFn(sequence)
}

func (v mockValidator) IsValidProposal(proposal []byte, sequence uint64) bool {
	return v.isValidProposalFn(sequence, proposal)
}

type mockValidatorSet struct {
	isValidatorFn func([]byte, uint64) bool
	isProposerFn  func([]byte, uint64, uint64) bool
	hasQuorumFn   func([]message.Message) bool
}

func (vs mockValidatorSet) IsValidator(addr []byte, sequence uint64) bool {
	return vs.isValidatorFn(addr, sequence)
}

func (vs mockValidatorSet) IsProposer(addr []byte, sequence, round uint64) bool {
	return vs.isProposerFn(addr, sequence, round)
}

func (vs mockValidatorSet) HasQuorum(messages []message.Message) bool {
	return vs.hasQuorumFn(messages)
}

type mockSignatureVerifier func([]byte, []byte, []byte) error

func (s mockSignatureVerifier) Verify(signature, digest, msg []byte) error {
	return s(signature, digest, msg)
}

type mockFeed struct {
	proposal    map[uint64]map[uint64][]*message.MsgProposal
	prepare     map[uint64]map[uint64][]*message.MsgPrepare
	commit      map[uint64]map[uint64][]*message.MsgCommit
	roundChange map[uint64]map[uint64][]*message.MsgRoundChange
}

func newMockFeed(messages []message.Message) mockFeed {
	f := mockFeed{
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

func (f mockFeed) SubscribeProposal(
	sequence,
	round uint64,
	higherRounds bool,
) (chan func() []*message.MsgProposal, func()) {
	sub := make(chan func() []*message.MsgProposal, 1)
	var notification func() []*message.MsgProposal
	if !higherRounds {
		notification = func() []*message.MsgProposal {
			return f.proposal[sequence][round]
		}
	} else {
		var highestRound uint64
		for round := range f.proposal[sequence] {
			if round >= highestRound {
				highestRound = round
			}
		}

		if highestRound < round {
			notification = func() []*message.MsgProposal { return nil }
		} else {
			notification = func() []*message.MsgProposal {
				return f.proposal[sequence][highestRound]
			}
		}
	}

	sub <- notification

	return sub, func() { close(sub) }
}

func (f mockFeed) SubscribePrepare(
	sequence,
	round uint64,
	higherRounds bool,
) (chan func() []*message.MsgPrepare, func()) {
	sub := make(chan func() []*message.MsgPrepare, 1)
	var notification func() []*message.MsgPrepare
	if !higherRounds {
		notification = func() []*message.MsgPrepare {
			return f.prepare[sequence][round]
		}
	} else {
		var highestRound uint64
		for round := range f.prepare[sequence] {
			if round >= highestRound {
				highestRound = round
			}
		}

		if highestRound < round {
			notification = func() []*message.MsgPrepare { return nil }
		} else {
			notification = func() []*message.MsgPrepare {
				return f.prepare[sequence][highestRound]
			}
		}
	}

	sub <- notification

	return sub, func() { close(sub) }
}

func (f mockFeed) SubscribeCommit(
	sequence,
	round uint64,
	higherRounds bool,
) (chan func() []*message.MsgCommit, func()) {
	sub := make(chan func() []*message.MsgCommit, 1)
	var notification func() []*message.MsgCommit
	if !higherRounds {
		notification = func() []*message.MsgCommit {
			return f.commit[sequence][round]
		}
	} else {
		var highestRound uint64
		for round := range f.commit[sequence] {
			if round >= highestRound {
				highestRound = round
			}
		}

		if highestRound < round {
			notification = func() []*message.MsgCommit { return nil }
		} else {
			notification = func() []*message.MsgCommit {
				return f.commit[sequence][highestRound]
			}
		}
	}

	sub <- notification
	return sub, func() { close(sub) }
}

func (f mockFeed) SubscribeRoundChange(
	sequence,
	round uint64,
	higherRounds bool,
) (chan func() []*message.MsgRoundChange, func()) {
	sub := make(chan func() []*message.MsgRoundChange, 1)
	var notification func() []*message.MsgRoundChange
	if !higherRounds {
		notification = func() []*message.MsgRoundChange {
			return f.roundChange[sequence][round]
		}
	} else {
		var highestRound uint64
		for round := range f.roundChange[sequence] {
			if round >= highestRound {
				highestRound = round
			}
		}

		if highestRound < round {
			notification = func() []*message.MsgRoundChange { return nil }
		} else {
			notification = func() []*message.MsgRoundChange {
				return f.roundChange[sequence][highestRound]
			}
		}
	}

	sub <- notification

	return sub, func() { close(sub) }
}

type SingeRoundMockFeed mockFeed

func NewSingleRoundMockFeed(messages []message.Message) SingeRoundMockFeed {
	return SingeRoundMockFeed(newMockFeed(messages))
}

func (f SingeRoundMockFeed) SubscribeProposal(
	sequence,
	round uint64,
	higherRounds bool,
) (chan func() []*message.MsgProposal, func()) {
	sub := make(chan func() []*message.MsgProposal, 1)
	if !higherRounds {
		notification := func() []*message.MsgProposal {
			return f.proposal[sequence][round]
		}

		sub <- notification

		return sub, func() { close(sub) }
	}

	sub <- func() []*message.MsgProposal {
		return nil
	}

	return sub, func() { close(sub) }
}

func (f SingeRoundMockFeed) SubscribePrepare(
	sequence,
	round uint64,
	higherRounds bool,
) (chan func() []*message.MsgPrepare, func()) {
	sub := make(chan func() []*message.MsgPrepare, 1)
	if !higherRounds {
		notification := func() []*message.MsgPrepare {
			return f.prepare[sequence][round]
		}

		sub <- notification

		return sub, func() { close(sub) }
	}

	sub <- func() []*message.MsgPrepare {
		return nil
	}

	return sub, func() { close(sub) }
}

func (f SingeRoundMockFeed) SubscribeCommit(
	sequence,
	round uint64,
	higherRounds bool,
) (chan func() []*message.MsgCommit, func()) {
	sub := make(chan func() []*message.MsgCommit, 1)
	if !higherRounds {
		notification := func() []*message.MsgCommit {
			return f.commit[sequence][round]
		}

		sub <- notification

		return sub, func() { close(sub) }
	}

	sub <- func() []*message.MsgCommit {
		return nil
	}

	return sub, func() { close(sub) }
}

func (f SingeRoundMockFeed) SubscribeRoundChange(
	sequence,
	round uint64,
	higherRounds bool,
) (chan func() []*message.MsgRoundChange, func()) {
	sub := make(chan func() []*message.MsgRoundChange, 1)
	if !higherRounds {
		notification := func() []*message.MsgRoundChange {
			return f.roundChange[sequence][round]
		}

		sub <- notification

		return sub, func() { close(sub) }
	}

	sub <- func() []*message.MsgRoundChange {
		return nil
	}

	return sub, func() { close(sub) }
}
