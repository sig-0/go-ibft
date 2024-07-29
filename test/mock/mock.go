package mock

import (
	"bytes"

	"github.com/sig-0/go-ibft"
	"github.com/sig-0/go-ibft/message/types"
)

var (
	NoQuorum      = ibft.QuorumFn(func(_ []types.Message) bool { return false })
	NonZeroQuorum = ibft.QuorumFn(func(messages []types.Message) bool { return len(messages) > 0 })

	OkBlock     = func(_ []byte, _ uint64) bool { return true }
	OkSignature = func(_, _, _ []byte) bool { return true }
)

func DummyKeccak(digest string) ibft.KeccakFn {
	return func(_ []byte) []byte {
		return []byte(digest)
	}
}

func QuorumOf(n int) ibft.QuorumFn {
	return func(messages []types.Message) bool {
		return len(messages) >= n
	}
}

func DummyTransport() ibft.MsgTransport {
	return ibft.MsgTransport{
		Proposal:    ibft.TransportFn[*types.MsgProposal](func(_ *types.MsgProposal) {}),
		Prepare:     ibft.TransportFn[*types.MsgPrepare](func(_ *types.MsgPrepare) {}),
		Commit:      ibft.TransportFn[*types.MsgCommit](func(_ *types.MsgCommit) {}),
		RoundChange: ibft.TransportFn[*types.MsgRoundChange](func(_ *types.MsgRoundChange) {}),
	}
}

type ValidatorID []byte

func NewValidatorID(id string) ValidatorID {
	return []byte(id)
}

func (id ValidatorID) ID() []byte {
	return id
}

func (id ValidatorID) Signer() ibft.SignerFn {
	return func(_ []byte) []byte {
		return []byte("signature")
	}
}

type ValidatorSet struct {
	IsValidatorFn func([]byte, uint64) bool
	IsProposerFn  func([]byte, uint64, uint64) bool
}

func (vs ValidatorSet) IsProposer(id []byte, sequence, round uint64) bool {
	return vs.IsProposerFn(id, sequence, round)
}

func (vs ValidatorSet) IsValidator(id []byte, sequence uint64) bool {
	return vs.IsValidatorFn(id, sequence)
}

type Proposer struct {
	ID    []byte
	Round uint64
}

func ProposersInRounds(proposers ...Proposer) func([]byte, uint64, uint64) bool {
	pp := make(map[uint64][]byte)
	for _, p := range proposers {
		pp[p.Round] = p.ID
	}

	return func(id []byte, _, round uint64) bool {
		return bytes.Equal(id, pp[round])
	}
}

type Validator struct {
	IDFn              func() []byte
	SignFn            func([]byte) []byte
	BuildProposalFn   func(uint64) []byte
	IsValidProposalFn func([]byte, uint64) bool
}

func (v Validator) ID() []byte {
	return v.IDFn()
}

func (v Validator) Sign(digest []byte) []byte {
	return v.SignFn(digest)
}

func (v Validator) BuildProposal(sequence uint64) []byte {
	return v.BuildProposalFn(sequence)
}

func (v Validator) IsValidProposal(proposal []byte, sequence uint64) bool {
	return v.IsValidProposalFn(proposal, sequence)
}

type SigVerifier struct {
	VerifyFn func([]byte, []byte, []byte) error
}

func (v SigVerifier) Verify(id, digest, sig []byte) error {
	return v.VerifyFn(id, digest, sig)
}

type MessageFeed struct {
	Proposal    messagesByView[*types.MsgProposal]
	Prepare     messagesByView[*types.MsgPrepare]
	Commit      messagesByView[*types.MsgCommit]
	RoundChange messagesByView[*types.MsgRoundChange]
}

func NewMessageFeed(messages []types.Message) MessageFeed {
	f := MessageFeed{
		Proposal:    make(messagesByView[*types.MsgProposal]),
		Prepare:     make(messagesByView[*types.MsgPrepare]),
		Commit:      make(messagesByView[*types.MsgCommit]),
		RoundChange: make(messagesByView[*types.MsgRoundChange]),
	}

	for _, msg := range messages {
		switch msg := msg.(type) {
		case *types.MsgProposal:
			f.Proposal.add(msg)
		case *types.MsgPrepare:
			f.Prepare.add(msg)
		case *types.MsgCommit:
			f.Commit.add(msg)
		case *types.MsgRoundChange:
			f.RoundChange.add(msg)
		}
	}

	return f
}

func newSubscription[M msg](notification types.MsgNotification[M]) (types.Subscription[M], func()) {
	c := make(types.Subscription[M], 1)
	c <- notification

	return c, func() { close(c) }
}

func (f MessageFeed) ProposalMessages(
	view *types.View,
	higherRounds bool,
) (types.Subscription[*types.MsgProposal], func()) {
	return newSubscription(f.Proposal.notification(view, higherRounds))
}

func (f MessageFeed) PrepareMessages(
	view *types.View,
	higherRounds bool,
) (types.Subscription[*types.MsgPrepare], func()) {
	return newSubscription(f.Prepare.notification(view, higherRounds))
}

func (f MessageFeed) CommitMessages(
	view *types.View,
	higherRounds bool,
) (types.Subscription[*types.MsgCommit], func()) {
	return newSubscription(f.Commit.notification(view, higherRounds))
}

func (f MessageFeed) RoundChangeMessages(
	view *types.View,
	higherRounds bool,
) (types.Subscription[*types.MsgRoundChange], func()) {
	return newSubscription(f.RoundChange.notification(view, higherRounds))
}

type SingleRoundFeed MessageFeed

func NewSingleRoundFeed(messages []types.Message) SingleRoundFeed {
	return SingleRoundFeed(NewMessageFeed(messages))
}

func (f SingleRoundFeed) ProposalMessages(
	view *types.View,
	higherRounds bool,
) (types.Subscription[*types.MsgProposal], func()) {
	if higherRounds {
		return nil, func() {}
	}

	return newSubscription(f.Proposal.notification(view, false))
}

func (f SingleRoundFeed) PrepareMessages(
	view *types.View,
	higherRounds bool,
) (types.Subscription[*types.MsgPrepare], func()) {
	if higherRounds {
		return nil, func() {}
	}

	return newSubscription(f.Prepare.notification(view, false))
}

func (f SingleRoundFeed) CommitMessages(
	view *types.View,
	higherRounds bool,
) (types.Subscription[*types.MsgCommit], func()) {
	if higherRounds {
		return nil, func() {}
	}

	return newSubscription(f.Commit.notification(view, false))
}

func (f SingleRoundFeed) RoundChangeMessages(
	view *types.View,
	higherRounds bool,
) (types.Subscription[*types.MsgRoundChange], func()) {
	if higherRounds {
		return nil, func() {}
	}

	return newSubscription(f.RoundChange.notification(view, false))
}

type msg interface {
	types.IBFTMessage

	Sequence() uint64
	Round() uint64
}

type messagesByView[M msg] map[uint64]map[uint64][]M

func (m messagesByView[M]) get(view *types.View) []M {
	return m[view.Sequence][view.Round]
}

func (m messagesByView[M]) rounds(sequence uint64) []uint64 {
	rounds := make([]uint64, 0)

	for round := range m[sequence] {
		rounds = append(rounds, round)
	}

	return rounds
}

func (m messagesByView[M]) add(msg M) {
	sequence, round := msg.Sequence(), msg.Round()

	messagesInSequence, ok := m[sequence]
	if !ok {
		m[sequence] = make(map[uint64][]M)
		messagesInSequence = m[sequence]
	}

	messagesInRound, ok := messagesInSequence[round]
	if !ok {
		messagesInSequence[round] = make([]M, 0)
		messagesInRound = messagesInSequence[round]
	}

	messagesInRound = append(messagesInRound, msg)
	messagesInSequence[round] = messagesInRound
}

func (m messagesByView[M]) notification(view *types.View, higherRounds bool) types.MsgNotification[M] {
	return types.MsgNotificationFn[M](func() []M {
		if !higherRounds {
			return m.get(view)
		}

		var highestRound uint64
		for _, round := range m.rounds(view.Sequence) {
			if round >= highestRound {
				highestRound = round
			}
		}

		if highestRound < view.Round {
			return nil
		}

		view.Round = highestRound

		return m.get(view)
	})
}
