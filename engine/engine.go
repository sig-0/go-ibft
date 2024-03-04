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
	Transport      ibft.Transport
	Quorum         ibft.Quorum
	Keccak         ibft.Keccak
	Round0Duration time.Duration
}

func (cfg EngineConfig) IsValid() error {
	if cfg.Validator == nil {
		return errors.New("nil Validator")
	}

	if cfg.Transport == nil {
		return errors.New("nil MessageTransport")
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

	messages   *store.MsgStore
	ctxOptions []sequencer.ContextOption
}

func NewEngine(cfg EngineConfig) Engine {
	msgStore := store.NewMsgStore()
	msgFeed := msgStore.Feed()

	return Engine{
		Sequencer: sequencer.New(cfg.Validator, cfg.Round0Duration),
		messages:  msgStore,
		ctxOptions: []sequencer.ContextOption{
			sequencer.WithQuorum(cfg.Quorum),
			sequencer.WithMessageTransport(cfg.Transport),
			sequencer.WithKeccak(cfg.Keccak),
			sequencer.WithMessageFeed(msgFeed),
		},
	}
}

func (e Engine) AddMessage(msg ibft.Message) error {
	// todo: verify msg

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

	return e.Finalize(sequencer.NewContext(ctx, e.ctxOptions...), sequence)
}
