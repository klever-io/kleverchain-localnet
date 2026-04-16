package domain

import (
	"fmt"
	"time"

	"github.com/klever-io/kleverchain-localnet/internal/errs"
)

const StateSchemaVersion = 1

type LocalnetState struct {
	Version            int       `yaml:"version"`
	CreatedAt          time.Time `yaml:"createdAt"`
	Validators         int       `yaml:"validators"`
	MaxSupply          uint64    `yaml:"maxSupply"`
	ConsensusGroupSize int       `yaml:"consensusGroupSize"`
	MinNodes           int       `yaml:"minNodes,omitempty"`
	StartTime          int64     `yaml:"startTime"`
	ChainID            string    `yaml:"chainID"`
	KleverImage        string    `yaml:"kleverImage"`
	SeednodePeerID     string    `yaml:"seednodePeerID,omitempty"`
}

func NewLocalnetState(validators int, maxSupply uint64, consensusGroupSize int) LocalnetState {
	if consensusGroupSize <= 0 {
		consensusGroupSize = validators
	}
	return LocalnetState{
		Version:            StateSchemaVersion,
		CreatedAt:          time.Now().UTC(),
		Validators:         validators,
		MaxSupply:          maxSupply,
		ConsensusGroupSize: consensusGroupSize,
		KleverImage:        KleverImage,
	}
}

func (s LocalnetState) Validate() error {
	if s.Validators <= 0 {
		return fmt.Errorf("%w: validators must be >= 1 (got %d)", errs.ErrInvalidInput, s.Validators)
	}
	if s.ConsensusGroupSize <= 0 {
		return fmt.Errorf("%w: consensusGroupSize must be >= 1 (got %d)", errs.ErrInvalidInput, s.ConsensusGroupSize)
	}
	if s.ConsensusGroupSize > s.Validators {
		return fmt.Errorf("%w: consensusGroupSize=%d exceeds validators=%d", errs.ErrConsensusGroupTooLarge, s.ConsensusGroupSize, s.Validators)
	}
	if s.MaxSupply == 0 {
		return fmt.Errorf("%w: maxSupply must be > 0", errs.ErrInvalidInput)
	}
	totalStaking := KLVDelegation * uint64(s.Validators)
	if s.MaxSupply <= totalStaking+RootKLV {
		return fmt.Errorf("%w: maxSupply=%d must exceed totalStaking+rootKLV=%d", errs.ErrInvalidInput, s.MaxSupply, totalStaking+RootKLV)
	}
	if s.KleverImage == "" {
		return fmt.Errorf("%w: kleverImage is required", errs.ErrInvalidInput)
	}
	if s.Version == 0 {
		return fmt.Errorf("%w: version is required", errs.ErrInvalidInput)
	}
	return nil
}

func (s LocalnetState) EffectiveMinNodes() int {
	m := s.ConsensusGroupSize
	if s.MinNodes > m {
		m = s.MinNodes
	}
	if m < 1 {
		m = 1
	}
	return m
}
