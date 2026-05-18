package genesis_test

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/klever-io/kleverchain-localnet/internal/domain"
	"github.com/klever-io/kleverchain-localnet/internal/errs"
	"github.com/klever-io/kleverchain-localnet/internal/genesis"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func goldenPath(name string) string {
	return filepath.Join("..", "..", "testdata", "golden", name)
}

func makeWallets(n int) []domain.Wallet {
	out := make([]domain.Wallet, n)
	for i := 0; i < n; i++ {
		out[i] = domain.Wallet{Address: fmt.Sprintf("klv1wallet%d", i)}
	}
	return out
}

func checkGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := goldenPath(name)
	if *updateGolden {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, got, 0o644))
		return
	}
	want, err := os.ReadFile(path)
	require.NoError(t, err, "missing golden file %s — run with -update to create", path)
	require.Equal(t, string(want), string(got))
}

func TestBuildGenesis_Golden(t *testing.T) {
	cases := []int{1, 3, 5}
	for _, n := range cases {
		t.Run(fmt.Sprintf("validators_%d", n), func(t *testing.T) {
			state := domain.NewLocalnetState(n, domain.DefaultMaxSupply, n)
			wallets := makeWallets(n)
			root := &domain.Wallet{Address: "klv1rootwallet"}
			g, err := genesis.Build(state, wallets, root)
			require.NoError(t, err)

			out, err := json.MarshalIndent(g, "", "    ")
			require.NoError(t, err)
			out = append(out, '\n')

			checkGolden(t, fmt.Sprintf("genesis_%d.json", n), out)
		})
	}
}

func TestBuildGenesis_NoRoot(t *testing.T) {
	state := domain.NewLocalnetState(2, domain.DefaultMaxSupply, 2)
	wallets := makeWallets(2)
	g, err := genesis.Build(state, wallets, nil)
	require.NoError(t, err)

	totalStaking := domain.KLVDelegation * 2
	expectedEachKLV := (domain.DefaultMaxSupply - totalStaking) / 2
	expectedEachKFI := domain.KFISupply / 2

	require.Len(t, g.Entries, 2)
	require.Nil(t, g.Root)
	for _, e := range g.Entries {
		require.Equal(t, expectedEachKLV, e.Balance)
		require.Equal(t, expectedEachKFI, e.KFIBalance)
		require.Equal(t, e.Address, e.Delegation.Address)
		require.Equal(t, domain.KLVDelegation, e.Delegation.Value)
	}
}

func TestBuildGenesis_WalletCountMismatch(t *testing.T) {
	state := domain.NewLocalnetState(3, domain.DefaultMaxSupply, 3)
	_, err := genesis.Build(state, makeWallets(2), nil)
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrInvalidInput))
}

func TestBuildGenesis_EmptyWallets(t *testing.T) {
	state := domain.NewLocalnetState(0, domain.DefaultMaxSupply, 1)
	state.Validators = 0
	_, err := genesis.Build(state, nil, nil)
	require.Error(t, err)
}

func TestBuildGenesis_RejectsInvalidState(t *testing.T) {
	state := domain.NewLocalnetState(3, domain.DefaultMaxSupply, 5)
	_, err := genesis.Build(state, makeWallets(3), nil)
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrConsensusGroupTooLarge))
}
