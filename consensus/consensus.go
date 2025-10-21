package consensus

import (
	"bytes"
	"context"
	"sort"

	"github.com/sig-0/go-ibft/message"
	"github.com/sig-0/go-ibft/sequencer"
)

type ValidatorSet interface {
	sequencer.ProposerAlgo

	GetValidators(ctx context.Context, sequence uint64) ([][]byte, error)
	CheckQuorum(ctx context.Context, sequence uint64, validators [][]byte) (bool, error)
}

type ProposalVerifier interface {
	Verify(ctx context.Context, sequence uint64, proposal []byte) error
}

type Consensus struct {
	vs                ValidatorSet
	proposal          ProposalVerifier
	sig               message.SignatureVerifier
	currentValidators map[string]struct{}
}

func New(
	vs ValidatorSet,
	proposal ProposalVerifier,
	vrf message.SignatureVerifier,
) Consensus {
	return Consensus{
		vs:                vs,
		proposal:          proposal,
		sig:               vrf,
		currentValidators: make(map[string]struct{}),
	}
}

func (c Consensus) InitSequence(ctx context.Context, sequence uint64) error {
	validators, err := c.vs.GetValidators(ctx, sequence)
	if err != nil {
		return err
	}

	clear(c.currentValidators)
	for _, validator := range validators {
		c.currentValidators[string(validator)] = struct{}{}
	}

	return nil
}

func (c Consensus) AwaitProposal(ctx context.Context, sequence sequencer.Sequence, store *message.Store) (*message.Proposal, error) {
	sub, cancel := store.ProposalMessages.Subscribe(sequence.Number)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrap := <-sub:
			messages := unwrap()
			for _, msg := range messages {
				//if _, ok := seen[string(msg.Signature)]; ok {
				//	continue
				//}
				if msg.Round != sequence.Round {
					continue // only interested in this round
				}

				if !c.isValidProposal(ctx, sequence, msg) {
					continue
				}

				// as soon as we hit a valid proposal message, we should return
				// It's unlikely and benign if a proposer decides to gossip 2 valid proposals
				return msg, nil
			}
		}
	}
}

func (c Consensus) AwaitFutureProposal(ctx context.Context, sequence sequencer.Sequence, store *message.Store) (*message.Proposal, error) {
	seen := make(map[string]struct{})
	proposalsInRounds := make(map[uint64]*message.Proposal)

	sub, cancel := store.ProposalMessages.Subscribe(sequence.Number)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrap := <-sub:
			messages := unwrap()
			for _, msg := range messages {
				if _, ok := seen[string(msg.Signature)]; ok {
					continue
				}

				seen[string(msg.Signature)] = struct{}{}

				if msg.Round <= sequence.Round {
					// only interested in higher rounds
					continue
				}

				if !c.isValidProposal(ctx, sequence, msg) {
					continue
				}

				proposalsInRounds[msg.Round] = msg

				// now we check in descending order
				rounds := make([]uint64, 0, len(messages))
				for round := range proposalsInRounds {
					rounds = append(rounds, round)
				}

				sort.SliceStable(rounds, func(i, j int) bool { return rounds[i] > rounds[j] })

				// take the proposal from the highest round
				highestRound := rounds[0]
				return proposalsInRounds[highestRound], nil
			}
		}
	}
}

func (c Consensus) isValidRoundChange(
	ctx context.Context,
	sequence sequencer.Sequence,
	msg *message.RoundChange,
) bool {
	//sender is part of the validator set
	if _, ok := c.currentValidators[string(msg.Sender)]; !ok {
		return false
	}

	var (
		pb = msg.LatestPreparedProposedBlock
		pc = msg.LatestPreparedCertificate
	)

	// if both pb and pc are missing, the message is valid
	if pb == nil && pc == nil {
		return true
	}

	// both pb and pc must be set
	if pb == nil || pc == nil {
		return false
	}

	if !c.isValidPC(ctx, pc, msg) {
		return false
	}

	// block hash in proposal message and keccak hash of proposed block match
	if !bytes.Equal(pc.ProposalMessage.BlockHash, message.GetProposalHash(pb)) {
		return false
	}

	return true
}

func (c Consensus) AwaitRoundChange(ctx context.Context, sequence sequencer.Sequence, store *message.Store) ([]*message.RoundChange, error) {
	sub, cancel := store.RoundChangeMessages.Subscribe(sequence.Number)
	defer cancel()

	seen := make(map[string]struct{})
	valid := make([]*message.RoundChange, 0)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrap := <-sub:
			messages := unwrap()
			for _, msg := range messages {
				if msg.Round != sequence.Round {
					continue // only interested in this particular round
				}

				if _, ok := seen[string(msg.Signature)]; ok {
					continue
				}

				if !c.isValidRoundChange(ctx, sequence, msg) {
					continue
				}

				valid = append(valid, msg)
				senders := make([][]byte, 0, len(valid))
				for _, msg := range valid {
					senders = append(senders, msg.Sender)
				}

				ok, err := c.vs.CheckQuorum(ctx, sequence.Number, senders)
				if err != nil {
					return nil, err // todo
				}

				if !ok {
					continue
				}

				return valid, nil
			}
		}
	}
}

func (c Consensus) AwaitFutureRoundChange(ctx context.Context, seq sequencer.Sequence, store *message.Store) ([]*message.RoundChange, error) {
	sequence := seq.Number
	round := seq.Round
	seen := make(map[string]struct{})
	rccByRounds := make(map[uint64]*message.RoundChangeCertificate)

	sub, cancel := store.RoundChangeMessages.Subscribe(seq.Number)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrap := <-sub:
			messages := unwrap()
			for _, msg := range messages {
				if msg.Round <= round {
					// only interested in higher rounds
					continue
				}

				if _, ok := seen[string(msg.Signature)]; ok {
					continue
				}

				seen[string(msg.Signature)] = struct{}{}

				if !c.isValidRoundChange(ctx, seq, msg) {
					continue
				}

				rcc := rccByRounds[msg.Round]
				if rcc == nil {
					rcc = &message.RoundChangeCertificate{}
					rccByRounds[msg.Round] = rcc
				}

				rcc.Messages = append(rcc.Messages, msg)
			}

			// now we check in descending order
			rounds := make([]uint64, 0, len(rccByRounds))
			for round := range rccByRounds {
				rounds = append(rounds, round)
			}

			sort.SliceStable(rounds, func(i, j int) bool { return rounds[i] > rounds[j] })

			getValidators := func(messages ...*message.RoundChange) [][]byte {
				validators := make([][]byte, 0, len(messages))
				for _, msg := range messages {
					validators = append(validators, msg.Signature)
				}

				return validators
			}

			// take the proposal from the highest round
			for _, round := range rounds {
				messages := rccByRounds[round].Messages
				ok, err := c.vs.CheckQuorum(ctx, sequence, getValidators(messages...))
				if err != nil {
					// todo: log
					continue
				}

				if !ok {
					continue
				}

				return messages, nil
			}
		}
	}
}

func (c Consensus) AwaitPrepare(ctx context.Context, seq sequencer.Sequence, store *message.Store) ([]*message.Prepare, error) {
	seen := make(map[string]struct{})
	valid := make([]*message.Prepare, 0)
	sequence := seq.Number
	round := seq.Round

	sub, cancel := store.PrepareMessages.Subscribe(sequence)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrap := <-sub:
			messages := unwrap()
			for _, msg := range messages {
				if msg.Round != round {
					// only interested in higher rounds
					continue
				}

				if _, ok := seen[string(msg.Signature)]; ok {
					continue
				}

				seen[string(msg.Signature)] = struct{}{}

				if !c.isValidPrepare(ctx, seq, msg) {
					continue
				}

				valid = append(valid, msg)
			}
		}

		getValidators := func(messages ...*message.Prepare) [][]byte {
			validators := make([][]byte, 0, len(messages))
			for _, msg := range messages {
				validators = append(validators, msg.Signature)
			}

			return validators
		}

		ok, err := c.vs.CheckQuorum(ctx, sequence, getValidators(valid...))
		if err != nil {
			// todo: log
			continue
		}

		if !ok {
			continue
		}

		return valid, nil
	}
}

func (c Consensus) isValidPrepare(_ context.Context, seq sequencer.Sequence, msg *message.Prepare) bool {
	if _, ok := c.currentValidators[string(msg.Sender)]; !ok {
		return false
	}

	// block hash and accepted block hash match
	if !bytes.Equal(msg.BlockHash, seq.Proposal.BlockHash) {
		return false
	}

	return true
}

func (c Consensus) AwaitCommit(ctx context.Context, seq sequencer.Sequence, store *message.Store) ([]*message.Commit, error) {
	seen := make(map[string]struct{})
	valid := make([]*message.Commit, 0)
	sequence := seq.Number
	round := seq.Round

	sub, cancel := store.CommitMessages.Subscribe(sequence)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case unwrap := <-sub:
			messages := unwrap()
			for _, msg := range messages {
				if _, ok := seen[string(msg.Signature)]; ok {
					continue
				}

				seen[string(msg.Signature)] = struct{}{}

				if msg.Round != round {
					// only interested in active round
					continue
				}

				if !c.isValidCommit(ctx, seq, msg) {
					continue
				}

				valid = append(valid, msg)
			}
		}

		getValidators := func(messages ...*message.Commit) [][]byte {
			validators := make([][]byte, 0, len(messages))
			for _, msg := range messages {
				validators = append(validators, msg.Signature)
			}

			return validators
		}

		ok, err := c.vs.CheckQuorum(ctx, sequence, getValidators(valid...))
		if err != nil {
			// todo: log
			continue
		}

		if !ok {
			continue
		}

		return valid, nil
	}
}

func (c Consensus) isValidProposal(ctx context.Context, sequence sequencer.Sequence, msg *message.Proposal) bool {
	if msg.ProposedBlock.Round != msg.Round {
		return false
	}

	//if bytes.Equal(msg.Sender, s.validator.Address()) {
	//	return false
	//}

	// sender is the elected proposer
	proposer, err := c.vs.GetProposer(ctx, sequence.Number, sequence.Round)
	if err != nil {
		return false // todo: log?
	}

	if notTheProposer := !bytes.Equal(msg.Sender, proposer); notTheProposer {
		return false
	}

	// block hash and keccak hash of proposed block match
	if !bytes.Equal(msg.BlockHash, message.GetProposalHash(msg.ProposedBlock)) {
		return false
	}

	if msg.Round == 0 {
		return c.proposal.Verify(ctx, msg.Sequence, msg.ProposedBlock.Block) == nil
	}

	/* non zero round proposals */

	rcc := msg.RoundChangeCertificate
	if !c.isValidRCC(ctx, rcc, msg) {
		return false
	}

	trimmedRCC := &message.RoundChangeCertificate{}
	for _, msg := range rcc.Messages {
		pc := msg.LatestPreparedCertificate
		if pc == nil {
			continue
		}

		// any included prepared certificate must be valid
		if c.isValidPC(ctx, pc, msg) {
			trimmedRCC.Messages = append(trimmedRCC.Messages, msg)
		}
	}

	blockHash, round := trimmedRCC.HighestRoundBlockHash()
	if blockHash == nil {
		// there is no previously agreed upon block hash, build a new proposal
		return c.proposal.Verify(ctx, msg.Sequence, msg.ProposedBlock.Block) == nil
	}

	// reuse the proposed block from previous (highest) round
	pb := &message.ProposedBlock{
		Block: msg.ProposedBlock.Block,
		Round: round,
	}

	// block hash and a keccak hash of proposed block match
	return bytes.Equal(blockHash, message.GetProposalHash(pb))
}

func (c Consensus) isValidPC(
	ctx context.Context,
	pc *message.PreparedCertificate,
	msg *message.RoundChange,
) bool {
	// both proposal message and prepare messages must be included
	if pc.ProposalMessage == nil || pc.PrepareMessages == nil {
		return false
	}

	var (
		sequence = pc.ProposalMessage.Sequence
		round    = pc.ProposalMessage.Round
	)

	if sequence != msg.Sequence {
		return false
	}

	if round >= msg.Round {
		return false
	}

	proposer, err := c.vs.GetProposer(ctx, sequence, round)
	if err != nil {
		return false // todo: log
	}

	// proposal sender in pc must be the selected proposer
	if notAProposer := !bytes.Equal(pc.ProposalMessage.Sender, proposer); notAProposer {
		return false
	}

	uniqueSenders := map[string]struct{}{
		string(pc.ProposalMessage.Sender): {}, // proposer
	}

	for _, msg := range pc.PrepareMessages {
		// prepare msg sequence (round) and proposal msg sequence (round) must match
		if msg.Sequence != sequence || msg.Round != round {
			return false
		}

		// prepare msg block hash and proposal msg block hash must match
		if !bytes.Equal(msg.BlockHash, pc.ProposalMessage.BlockHash) {
			return false
		}

		// prepare msg sender must be part of the validator set
		if _, ok := c.currentValidators[string(msg.Sender)]; !ok {
			return false
		}

		uniqueSenders[string(msg.Sender)] = struct{}{}
	}

	// 1 (proposer) + len(prepare) unique validators
	if len(uniqueSenders) != 1+len(pc.PrepareMessages) {
		return false
	}

	senders := make([][]byte, 0, len(uniqueSenders))
	for sender, _ := range uniqueSenders {
		senders = append(senders, []byte(sender))
	}

	// all messages in pc satisfy a quorum
	ok, err := c.vs.CheckQuorum(ctx, sequence, senders)
	if err != nil {
		return false // todo: log
	}

	return ok
}

func (c Consensus) isValidRCC(
	ctx context.Context,
	rcc *message.RoundChangeCertificate,
	proposal *message.Proposal,
) bool {
	// rcc must be included
	if rcc == nil || len(rcc.Messages) == 0 {
		return false
	}

	var (
		sequence      = proposal.Sequence
		round         = proposal.Round
		uniqueSenders = make(map[string]struct{})
	)

	for _, msg := range rcc.Messages {
		// round change msg sequence (round) and proposal msg sequence (round) must match
		if msg.Sequence != sequence || msg.Round != round {
			return false
		}

		// sender must be part of the validator set
		if _, ok := c.currentValidators[string(msg.Sender)]; !ok {
			return false
		}

		uniqueSenders[string(msg.Sender)] = struct{}{}
	}

	// all messages must be unique
	if len(uniqueSenders) != len(rcc.Messages) {
		return false
	}

	senders := make([][]byte, len(uniqueSenders))
	for sender, _ := range uniqueSenders {
		senders = append(senders, []byte(sender))
	}

	ok, err := c.vs.CheckQuorum(ctx, sequence, senders)
	if err != nil {
		return false // todo: log
	}

	return ok
}

func (c Consensus) isValidCommit(ctx context.Context, seq sequencer.Sequence, msg *message.Commit) bool {
	// sender must be part of the validator set
	if _, ok := c.currentValidators[string(msg.Sender)]; !ok {
		return false
	}

	// block hash and accepted block hash must match
	if !bytes.Equal(msg.BlockHash, seq.Proposal.BlockHash) {
		return false
	}

	// commit seal was generated by signing block hash
	if err := c.sig.Verify(msg.Sender, msg.BlockHash, msg.CommitSeal); err != nil {
		return false
	}

	return true
}
