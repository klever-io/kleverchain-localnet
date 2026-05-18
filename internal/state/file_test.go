package state_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/klever-io/kleverchain-localnet/internal/domain"
	"github.com/klever-io/kleverchain-localnet/internal/errs"
	"github.com/klever-io/kleverchain-localnet/internal/state"
)

func TestState_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, state.DefaultFileName)

	original := domain.NewLocalnetState(3, domain.DefaultMaxSupply, 2)
	original.StartTime = 1_713_276_075
	original.ChainID = "22843681"
	original.CreatedAt = time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	original.SeednodePeerID = "16UiuFake"

	require.NoError(t, state.Save(path, original))
	require.True(t, state.Exists(path))

	loaded, err := state.Load(path)
	require.NoError(t, err)
	require.Equal(t, original.Validators, loaded.Validators)
	require.Equal(t, original.MaxSupply, loaded.MaxSupply)
	require.Equal(t, original.ConsensusGroupSize, loaded.ConsensusGroupSize)
	require.Equal(t, original.StartTime, loaded.StartTime)
	require.Equal(t, original.ChainID, loaded.ChainID)
	require.Equal(t, original.SeednodePeerID, loaded.SeednodePeerID)
	require.Equal(t, domain.StateSchemaVersion, loaded.Version)
}

func TestState_LoadMissing(t *testing.T) {
	_, err := state.Load("/nonexistent/.localnet-state.yaml")
	require.Error(t, err)
	require.True(t, errors.Is(err, os.ErrNotExist))
}

func TestState_LoadMalformed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	require.NoError(t, os.WriteFile(path, []byte("not: valid: yaml: ["), 0o644))

	_, err := state.Load(path)
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrInvalidInput))
}

func TestState_LoadMissingVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "noversion.yaml")
	require.NoError(t, os.WriteFile(path, []byte("validators: 3\n"), 0o644))

	_, err := state.Load(path)
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrInvalidInput))
}

func TestState_LoadFutureVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "future.yaml")
	require.NoError(t, os.WriteFile(path, []byte("version: 99\nvalidators: 3\n"), 0o644))

	_, err := state.Load(path)
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrInvalidInput))
}
