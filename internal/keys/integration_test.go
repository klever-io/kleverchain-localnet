//go:build integration

package keys_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/klever-io/kleverchain-localnet/internal/dockercli"
	"github.com/klever-io/kleverchain-localnet/internal/domain"
	"github.com/klever-io/kleverchain-localnet/internal/keys"
)

func TestIntegration_Keygen_Real(t *testing.T) {
	if os.Getenv("CI") == "" && os.Getenv("LOCAL_INTEGRATION") == "" {
		t.Skip("integration test: set LOCAL_INTEGRATION=1 or run in CI")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	dir := t.TempDir()
	state := domain.NewLocalnetState(1, domain.DefaultMaxSupply, 1)

	g := keys.NewGenerator(dockercli.NewExecRunner(), domain.KleverImage)
	res, err := g.Generate(ctx, state, dir, false)
	require.NoError(t, err)
	require.Len(t, res.Validators, 1)
	require.Len(t, res.Wallets, 1)
	require.NotNil(t, res.Root)

	require.True(t, fileExists(filepath.Join(dir, "node-0", "validatorKey.pem")))
	require.True(t, fileExists(filepath.Join(dir, "node-0", "walletKey.pem")))
	require.True(t, fileExists(filepath.Join(dir, "walletKey.pem")))
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
