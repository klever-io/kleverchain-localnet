package nodes_test

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
	"github.com/klever-io/kleverchain-localnet/internal/nodes"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func goldenPath(name string) string {
	return filepath.Join("..", "..", "testdata", "golden", name)
}

func makePairs(n int) []nodes.Pair {
	out := make([]nodes.Pair, n)
	for i := 0; i < n; i++ {
		out[i] = nodes.Pair{
			NodeIndex: i,
			PubKey:    fmt.Sprintf("bls_pubkey_%d", i),
			Address:   fmt.Sprintf("klv1wallet%d", i),
		}
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

func TestBuildNodes_Golden(t *testing.T) {
	const fixedStart int64 = 1_713_276_000
	cases := []int{1, 3, 5}
	for _, n := range cases {
		t.Run(fmt.Sprintf("validators_%d", n), func(t *testing.T) {
			state := domain.NewLocalnetState(n, domain.DefaultMaxSupply, n)
			state.StartTime = fixedStart

			ns, err := nodes.Build(state, makePairs(n))
			require.NoError(t, err)

			out, err := json.MarshalIndent(ns, "", "    ")
			require.NoError(t, err)
			out = append(out, '\n')

			checkGolden(t, fmt.Sprintf("nodesSetup_%d.json", n), out)
		})
	}
}

func TestBuildNodes_ConsensusGroupTooLarge(t *testing.T) {
	state := domain.NewLocalnetState(3, domain.DefaultMaxSupply, 5)
	_, err := nodes.Build(state, makePairs(3))
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrConsensusGroupTooLarge))
}

func TestBuildNodes_StartTimeRounding(t *testing.T) {
	got := nodes.RoundUpToSlotBoundary(100)
	require.Equal(t, int64(150), got)
	require.Equal(t, int64(0)%domain.SlotRoundSeconds, got%domain.SlotRoundSeconds)

	got = nodes.RoundUpToSlotBoundary(75)
	require.Equal(t, int64(150), got)

	got = nodes.RoundUpToSlotBoundary(0)
	require.Equal(t, int64(75), got)
}

func TestBuildNodes_ChainIDDerivedFromStartTime(t *testing.T) {
	state := domain.NewLocalnetState(1, domain.DefaultMaxSupply, 1)
	state.StartTime = 75 * 1000
	ns, err := nodes.Build(state, makePairs(1))
	require.NoError(t, err)
	require.Equal(t, int64(75*1001), ns.StartTime)
	require.Equal(t, "1001", ns.ChainID)
}

func TestBuildNodes_DeterministicOrder(t *testing.T) {
	state := domain.NewLocalnetState(3, domain.DefaultMaxSupply, 3)
	state.StartTime = 1_000_000
	pairs := []nodes.Pair{
		{NodeIndex: 2, PubKey: "b", Address: "klv2"},
		{NodeIndex: 0, PubKey: "a", Address: "klv0"},
		{NodeIndex: 1, PubKey: "c", Address: "klv1"},
	}
	ns, err := nodes.Build(state, pairs)
	require.NoError(t, err)
	require.Equal(t, "a", ns.InitialNodes[0].PubKey)
	require.Equal(t, "c", ns.InitialNodes[1].PubKey)
	require.Equal(t, "b", ns.InitialNodes[2].PubKey)
}

func TestBuildNodes_MinNodesFromEffective(t *testing.T) {
	state := domain.NewLocalnetState(3, domain.DefaultMaxSupply, 2)
	state.MinNodes = 5
	state.StartTime = 1_000_000
	ns, err := nodes.Build(state, makePairs(3))
	require.NoError(t, err)
	require.Equal(t, 5, ns.MinNodes)
}
