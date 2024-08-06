package sequencer

import (
	"time"

	"github.com/sig-0/go-ibft/message"
)

type Config struct {
	Validator         Validator
	ValidatorSet      ValidatorSet
	Transport         message.Transport
	Feed              message.Feed
	Keccak            message.Keccak
	SignatureVerifier message.SignatureVerifier
	Round0Duration    time.Duration
}

func NewConfig(opts ...Option) Config {
	cfg := Config{}
	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

type Option func(*Config)

func WithValidator(v Validator) Option {
	return func(cfg *Config) {
		cfg.Validator = v
	}
}

func WithValidatorSet(vs ValidatorSet) Option {
	return func(cfg *Config) {
		cfg.ValidatorSet = vs
	}
}

func WithTransport(t message.Transport) Option {
	return func(cfg *Config) {
		cfg.Transport = t
	}
}

func WithFeed(f message.Feed) Option {
	return func(cfg *Config) {
		cfg.Feed = f
	}
}

func WithKeccak(k message.Keccak) Option {
	return func(cfg *Config) {
		cfg.Keccak = k
	}
}

func WithSignatureVerifier(v message.SignatureVerifier) Option {
	return func(cfg *Config) {
		cfg.SignatureVerifier = v
	}
}

func WithRound0Duration(d time.Duration) Option {
	return func(cfg *Config) {
		cfg.Round0Duration = d
	}
}
