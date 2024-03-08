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

type EngineConfig struct {
	Validator      ibft.Validator
	Transport      sequencer.MessageTransport
	Quorum         ibft.Quorum
	Keccak         ibft.Keccak
	Round0Duration time.Duration
}

func (cfg EngineConfig) IsValid() error {
	if cfg.Validator == nil {
		return errors.New("nil Validator")
	}

	if !cfg.Transport.IsValid() {
		return errors.New("invalid transport")
	}

	if cfg.Quorum == nil {
		return errors.New("nil Quorum")
	}

	if cfg.Keccak == nil {
		return errors.New("nil Keccak")
	}

	return nil
}

type SequenceResult = types.FinalizedProposal

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
		return fmt.Errorf("invalid msg: %w", err)
	}

	msgDigest := e.cfg.Keccak.Hash(msg.Payload())
	if !e.Validator.IsValidSignature(msg.GetSender(), msgDigest, msg.GetSignature()) {
		return errors.New("invalid signature")
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
