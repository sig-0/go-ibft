package consensus

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
	ErrInvalidConfig  = errors.New("invalid engine config")
	ErrInvalidMessage = errors.New("invalid ibft message")
)

type EngineConfig struct {
	Validator      ibft.Validator
	Transport      sequencer.MessageTransport
	Quorum         ibft.Quorum
	Keccak         ibft.Keccak
	Round0Duration time.Duration
}

func (cfg EngineConfig) IsValid() error {
	if cfg.Validator == nil {
		return fmt.Errorf("%w: missing validator", ErrInvalidConfig)
	}

	if !cfg.Transport.IsValid() {
		return fmt.Errorf("%w: invalid transport", ErrInvalidConfig)
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
	*sequencer.Sequencer

	messages *store.MsgStore
	cfg      EngineConfig
}

func NewEngine(cfg EngineConfig) Engine {
	return Engine{
		Sequencer: sequencer.New(cfg.Validator, cfg.Round0Duration),
		messages:  store.NewMsgStore(),
		cfg:       cfg,
	}
}

func (e Engine) AddMessage(msg types.Message) error {
	if err := msg.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidMessage, err)
	}

	msgDigest := e.cfg.Keccak.Hash(msg.Payload())
	if !e.Validator.IsValidSignature(msg.GetSender(), msgDigest, msg.GetSignature()) {
		return fmt.Errorf("%w: invalid signature", ErrInvalidMessage)
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

type SequenceResult = types.FinalizedProposal

func (e Engine) FinalizeSequence(ctx context.Context, sequence uint64) *SequenceResult {
	defer func() {
		// todo: clean old messages

	}()

	opts := []sequencer.ContextOption{
		sequencer.WithQuorum(e.cfg.Quorum),
		sequencer.WithKeccak(e.cfg.Keccak),
		sequencer.WithMessageTransport(e.cfg.Transport),
		sequencer.WithMessageFeed(e.messages.Feed()),
	}

	return e.Finalize(sequencer.NewContext(ctx, opts...), sequence)
}
