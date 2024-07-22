package engine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sig-0/go-ibft"
	"github.com/sig-0/go-ibft/message/store"
	"github.com/sig-0/go-ibft/message/types"
	"github.com/sig-0/go-ibft/sequencer"
)

var (
	ErrInvalidConfig    = errors.New("invalid engine config")
	ErrInvalidSignature = errors.New("invalid message signature")
)

type Config struct {
	MsgTransport   ibft.MsgTransport
	Quorum         ibft.Quorum
	Keccak         ibft.Keccak
	Round0Duration time.Duration
}

func (cfg Config) Validate() error {
	if cfg.MsgTransport.Proposal == nil {
		return fmt.Errorf("%w: missing MsgProposal transport", ErrInvalidConfig)
	}

	if cfg.MsgTransport.Prepare == nil {
		return fmt.Errorf("%w: missing MsgPrepare transport", ErrInvalidConfig)
	}

	if cfg.MsgTransport.Commit == nil {
		return fmt.Errorf("%w: missing MsgCommit transport", ErrInvalidConfig)
	}

	if cfg.MsgTransport.RoundChange == nil {
		return fmt.Errorf("%w: missing MsgRoundChange transport", ErrInvalidConfig)
	}

	if cfg.Quorum == nil {
		return fmt.Errorf("%w: missing Quorum", ErrInvalidConfig)
	}

	if cfg.Keccak == nil {
		return fmt.Errorf("%w: missing Keccak", ErrInvalidConfig)
	}

	if cfg.Round0Duration == 0 {
		return fmt.Errorf("%w: round zero duration cannot be 0", ErrInvalidConfig)
	}

	return nil
}

type Engine struct {
	sequencer *sequencer.Sequencer
	messages  *store.MsgStore
	cfg       Config
}

func NewEngine(validator ibft.Validator, cfg Config) Engine {
	return Engine{
		sequencer: sequencer.NewSequencer(validator, cfg.Round0Duration),
		messages:  store.NewMsgStore(),
		cfg:       cfg,
	}
}

func (e Engine) AddMessage(msg types.Message) error {
	if err := msg.Validate(); err != nil {
		return err
	}

	var (
		sender    = msg.Sender()
		signature = msg.Signature()
		digest    = e.cfg.Keccak.Hash(msg.Payload())
	)

	if !e.sequencer.Validator.IsValidSignature(sender, digest, signature) {
		return ErrInvalidSignature
	}

	e.messages.Add(msg)

	return nil
}

func (e Engine) FinalizeSequence(c context.Context, sequence uint64) *types.FinalizedProposal {
	defer func() {
		e.messages.Clear()
	}()

	ctx := sequencer.NewContext(c)
	ctx = ctx.WithQuorum(e.cfg.Quorum)
	ctx = ctx.WithKeccak(e.cfg.Keccak)
	ctx = ctx.WithTransport(e.cfg.MsgTransport)
	ctx = ctx.WithMsgFeed(e.messages.Feed())

	return e.sequencer.Finalize(ctx, sequence)
}
