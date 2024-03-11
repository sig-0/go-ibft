package engine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/store"
	"github.com/madz-lab/go-ibft/message/types"
	"github.com/madz-lab/go-ibft/sequencer"
)

var (
	ErrInvalidConfig    = errors.New("invalid engine config")
	ErrInvalidSignature = errors.New("recovered address does not match sender")
)

type Config struct {
	TransportMsgProposal    ibft.Transport[*types.MsgProposal]
	TransportMsgPrepare     ibft.Transport[*types.MsgPrepare]
	TransportMsgCommit      ibft.Transport[*types.MsgCommit]
	TransportMsgRoundChange ibft.Transport[*types.MsgRoundChange]
	Quorum                  ibft.Quorum
	Keccak                  ibft.Keccak
	Round0Duration          time.Duration
}

func (cfg Config) Validate() error {
	if cfg.TransportMsgProposal == nil {
		return fmt.Errorf("%w: missing *MsgProposal transport", ErrInvalidConfig)
	}

	if cfg.TransportMsgPrepare == nil {
		return fmt.Errorf("%w: missing *MsgPrepare transport", ErrInvalidConfig)
	}

	if cfg.TransportMsgCommit == nil {
		return fmt.Errorf("%w: missing *MsgCommit transport", ErrInvalidConfig)
	}

	if cfg.TransportMsgRoundChange == nil {
		return fmt.Errorf("%w: missing *MsgRoundChange transport", ErrInvalidConfig)
	}

	if cfg.Quorum == nil {
		return fmt.Errorf("%w: missing quorum", ErrInvalidConfig)
	}

	if cfg.Keccak == nil {
		return fmt.Errorf("%w: missing keccak", ErrInvalidConfig)
	}

	if cfg.Round0Duration == 0 {
		return fmt.Errorf("%w: round zero duration cannot be 0", ErrInvalidConfig)
	}

	return nil
}

type Engine struct {
	sequencer *sequencer.Sequencer

	messages *store.MessageStore
	cfg      Config
}

func NewEngine(validator ibft.Validator, cfg Config) Engine {
	return Engine{
		sequencer: sequencer.New(validator, cfg.Round0Duration),
		messages:  store.NewMessageStore(),
		cfg:       cfg,
	}
}

func (e Engine) AddMessage(msg types.Message) error {
	if err := msg.Validate(); err != nil {
		return fmt.Errorf("malformed message: %w", err)
	}

	var (
		sender    = msg.GetSender()
		signature = msg.GetSignature()
		digest    = e.cfg.Keccak.Hash(msg.Payload())
	)

	if !e.sequencer.Validator.IsValidSignature(sender, digest, signature) {
		return fmt.Errorf("signature verification failed: %w", ErrInvalidSignature)
	}

	switch msg := msg.(type) {
	case *types.MsgProposal:
		e.messages.ProposalMessages.AddMessage(msg)
	case *types.MsgPrepare:
		e.messages.PrepareMessages.AddMessage(msg)
	case *types.MsgCommit:
		e.messages.CommitMessages.AddMessage(msg)
	case *types.MsgRoundChange:
		e.messages.RoundChangeMessages.AddMessage(msg)
	}

	return nil
}

type SequenceResult struct {
	SequenceProposal *types.FinalizedProposal
	Sequence         uint64
}

func (e Engine) FinalizeSequence(ctx context.Context, sequence uint64) SequenceResult {
	defer func() {
		e.messages.ProposalMessages.Clear()
		e.messages.PrepareMessages.Clear()
		e.messages.CommitMessages.Clear()
		e.messages.RoundChangeMessages.Clear()
	}()

	c := sequencer.NewContext(ctx,
		sequencer.WithQuorum(e.cfg.Quorum),
		sequencer.WithKeccak(e.cfg.Keccak),
		sequencer.WithMessageFeed(e.messages.Feed()),
		sequencer.WithMessageTransport(sequencer.MessageTransport{
			Proposal:    e.cfg.TransportMsgProposal,
			Prepare:     e.cfg.TransportMsgPrepare,
			Commit:      e.cfg.TransportMsgCommit,
			RoundChange: e.cfg.TransportMsgRoundChange,
		}),
	)

	return SequenceResult{
		Sequence:         sequence,
		SequenceProposal: e.sequencer.Finalize(c, sequence),
	}
}
