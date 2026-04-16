package compose_test

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/klever-io/kleverchain-localnet/internal/compose"
	"github.com/klever-io/kleverchain-localnet/internal/domain"
	"github.com/klever-io/kleverchain-localnet/internal/errs"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func goldenPath(name string) string {
	return filepath.Join("..", "..", "testdata", "golden", name)
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

func TestBuildCompose_Golden(t *testing.T) {
	cases := []int{1, 3, 5}
	for _, n := range cases {
		t.Run(fmt.Sprintf("validators_%d", n), func(t *testing.T) {
			state := domain.NewLocalnetState(n, domain.DefaultMaxSupply, n)
			services, err := compose.BuildValidatorServices(state, "keys")
			require.NoError(t, err)

			out, err := compose.Build(state, services, compose.DefaultResources())
			require.NoError(t, err)

			checkGolden(t, fmt.Sprintf("docker-compose_%d.yaml", n), []byte(out))
		})
	}
}

func TestBuildCompose_ValidYAML(t *testing.T) {
	state := domain.NewLocalnetState(3, domain.DefaultMaxSupply, 3)
	services, err := compose.BuildValidatorServices(state, "keys")
	require.NoError(t, err)
	out, err := compose.Build(state, services, compose.DefaultResources())
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(out), &parsed))
	svc, ok := parsed["services"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, svc, "seednode")
	require.Contains(t, svc, "node0")
	require.Contains(t, svc, "node1")
	require.Contains(t, svc, "node2")
}

func TestBuildCompose_BindMountPrefix(t *testing.T) {
	state := domain.NewLocalnetState(1, domain.DefaultMaxSupply, 1)
	services, err := compose.BuildValidatorServices(state, "keys")
	require.NoError(t, err)
	out, err := compose.Build(state, services, compose.DefaultResources())
	require.NoError(t, err)
	require.Contains(t, out, "./keys/node-0/validatorKey.pem:/opt/klever-blockchain/config/node/validatorKey.pem")
	require.Contains(t, out, "- ./config/seednode:/opt/klever-blockchain/config/seednode")
	require.Contains(t, out, "- ./config/node:/opt/klever-blockchain/config/node")
	require.Contains(t, out, "- ./dbs/node-0/:/opt/klever-blockchain/db")
	require.Contains(t, out, "- ./logs/node-0:/opt/klever-blockchain/logs")
}

func TestBuildCompose_PinnedImage(t *testing.T) {
	state := domain.NewLocalnetState(2, domain.DefaultMaxSupply, 2)
	services, err := compose.BuildValidatorServices(state, "keys")
	require.NoError(t, err)
	out, err := compose.Build(state, services, compose.DefaultResources())
	require.NoError(t, err)
	require.NotContains(t, out, ":latest")
	require.Equal(t, 3, strings.Count(out, domain.KleverImage))
}

func TestBuildCompose_Healthchecks(t *testing.T) {
	state := domain.NewLocalnetState(2, domain.DefaultMaxSupply, 2)
	services, err := compose.BuildValidatorServices(state, "keys")
	require.NoError(t, err)
	out, err := compose.Build(state, services, compose.DefaultResources())
	require.NoError(t, err)
	require.Equal(t, 2, strings.Count(out, "healthcheck:"), "only validators should have healthchecks — seednode has no HTTP endpoint")
	require.NotContains(t, out, "condition: service_healthy", "seednode has no healthcheck so validators cannot wait on service_healthy")
}

func TestBuildCompose_RejectsBareRelativePath(t *testing.T) {
	state := domain.NewLocalnetState(1, domain.DefaultMaxSupply, 1)
	bad := []compose.ValidatorService{{Index: 0, RESTPort: 8800, ValidatorKeyPath: "keys/node-0/validatorKey.pem"}}
	_, err := compose.Build(state, bad, compose.DefaultResources())
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrInvalidInput))
}
