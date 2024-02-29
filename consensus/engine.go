package consensus

import (
	"context"
	"errors"
	"time"

	"github.com/madz-lab/go-ibft"
	"github.com/madz-lab/go-ibft/message/store"
	"github.com/madz-lab/go-ibft/message/types"
	"github.com/madz-lab/go-ibft/sequencer"
)

type EngineConfig struct {
	Validator      ibft.Validator
	Verifier       ibft.Verifier
	Transport      ibft.Transport
	Quorum         ibft.Quorum
	Keccak         ibft.Keccak
	Round0Duration time.Duration
}

func (cfg EngineConfig) IsValid() error {
	if cfg.Validator == nil {
		return errors.New("nil Validator")
	}

	if cfg.Verifier == nil {
		return errors.New("nil Verifier")
	}

	if cfg.Transport == nil {
		return errors.New("nil Transport")
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
	cfg EngineConfig

	messages  *store.MsgStore
	sequencer *sequencer.Sequencer
}

func NewEngine(cfg EngineConfig) Engine {
	return Engine{
		cfg:       cfg,
		messages:  store.NewMsgStore(),
		sequencer: sequencer.New(cfg.Validator, cfg.Verifier, cfg.Round0Duration),
	}
}

func (e Engine) AddMessage(msg ibft.Message) error {
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

func (e Engine) Finalize(ctx context.Context, sequence uint64) *SequenceResult {
	defer func() {
		// clean old messages

	}()

	seqCtx := sequencer.NewContext(ctx,
		sequencer.WithQuorum(e.cfg.Quorum),
		sequencer.WithTransport(e.cfg.Transport),
		sequencer.WithKeccak(e.cfg.Keccak),
		sequencer.WithMessageFeed(e.messages.Feed()),
	)

	return e.sequencer.FinalizeSequence(seqCtx, sequence)
}
