package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/klever-io/kleverchain-localnet/internal/domain"
	"github.com/klever-io/kleverchain-localnet/internal/errs"
)

func TestLocalnetState_Validate_Happy(t *testing.T) {
	s := domain.NewLocalnetState(3, domain.DefaultMaxSupply, 3)
	require.NoError(t, s.Validate())
}

func TestLocalnetState_Validate_NonPositiveValidators(t *testing.T) {
	cases := []int{0, -1, -100}
	for _, n := range cases {
		s := domain.NewLocalnetState(n, domain.DefaultMaxSupply, 1)
		s.Validators = n
		err := s.Validate()
		require.Error(t, err)
		require.True(t, errors.Is(err, errs.ErrInvalidInput))
	}
}

func TestLocalnetState_Validate_ConsensusGroupTooLarge(t *testing.T) {
	s := domain.NewLocalnetState(2, domain.DefaultMaxSupply, 5)
	err := s.Validate()
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrConsensusGroupTooLarge))
}

func TestLocalnetState_Validate_MaxSupplyTooSmall(t *testing.T) {
	s := domain.NewLocalnetState(3, domain.KLVDelegation*3+domain.RootKLV, 3)
	err := s.Validate()
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrInvalidInput))
}

func TestLocalnetState_Validate_ZeroMaxSupply(t *testing.T) {
	s := domain.NewLocalnetState(1, 0, 1)
	err := s.Validate()
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrInvalidInput))
}

func TestLocalnetState_Validate_EmptyImage(t *testing.T) {
	s := domain.NewLocalnetState(1, domain.DefaultMaxSupply, 1)
	s.KleverImage = ""
	err := s.Validate()
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrInvalidInput))
}

func TestLocalnetState_Validate_ZeroVersion(t *testing.T) {
	s := domain.NewLocalnetState(1, domain.DefaultMaxSupply, 1)
	s.Version = 0
	err := s.Validate()
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrInvalidInput))
}

func TestLocalnetState_Validate_ConsensusDefaultsToValidators(t *testing.T) {
	s := domain.NewLocalnetState(5, domain.DefaultMaxSupply, 0)
	require.Equal(t, 5, s.ConsensusGroupSize)
	require.NoError(t, s.Validate())
}

func TestLocalnetState_EffectiveMinNodes(t *testing.T) {
	s := domain.NewLocalnetState(3, domain.DefaultMaxSupply, 2)
	require.Equal(t, 2, s.EffectiveMinNodes())

	s.MinNodes = 5
	require.Equal(t, 5, s.EffectiveMinNodes())
}
