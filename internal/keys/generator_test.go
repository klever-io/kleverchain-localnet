package keys_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/klever-io/kleverchain-localnet/internal/dockercli"
	"github.com/klever-io/kleverchain-localnet/internal/domain"
	"github.com/klever-io/kleverchain-localnet/internal/errs"
	"github.com/klever-io/kleverchain-localnet/internal/keys"
)

type fakeKeygenRunner struct {
	calls [][]string
}

func (f *fakeKeygenRunner) Run(_ context.Context, _ string, args []string, _ dockercli.RunOptions) (dockercli.RunResult, error) {
	f.calls = append(f.calls, append([]string(nil), args...))

	var mount string
	keyType := ""
	for i, a := range args {
		if a == "-v" && i+1 < len(args) {
			mount = args[i+1]
		}
		if a == "--key-type" && i+1 < len(args) {
			keyType = args[i+1]
		}
	}
	if mount == "" {
		return dockercli.RunResult{}, fmt.Errorf("no -v mount specified")
	}

	sep := strings.LastIndex(mount, ":")
	if sep <= 0 {
		return dockercli.RunResult{}, fmt.Errorf("invalid -v mount %q", mount)
	}
	hostPath := mount[:sep]
	if err := os.MkdirAll(hostPath, 0o700); err != nil {
		return dockercli.RunResult{}, err
	}

	suffix := filepath.Base(hostPath)
	switch keyType {
	case "both":
		if err := writeFakePEM(filepath.Join(hostPath, "validatorKey.pem"), "bls_"+suffix); err != nil {
			return dockercli.RunResult{}, err
		}
		if err := writeFakePEM(filepath.Join(hostPath, "walletKey.pem"), "klv1_"+suffix); err != nil {
			return dockercli.RunResult{}, err
		}
	case "wallet":
		if err := writeFakePEM(filepath.Join(hostPath, "walletKey.pem"), "klv1_root"); err != nil {
			return dockercli.RunResult{}, err
		}
	}
	return dockercli.RunResult{}, nil
}

func writeFakePEM(path, label string) error {
	body := fmt.Sprintf("-----BEGIN PRIVATE KEY for %s-----\nAAA\n-----END PRIVATE KEY for %s-----\n", label, label)
	return os.WriteFile(path, []byte(body), 0o600)
}

func TestGenerator_Generate_HappyPath(t *testing.T) {
	dir := t.TempDir()
	state := domain.NewLocalnetState(3, domain.DefaultMaxSupply, 3)

	runner := &fakeKeygenRunner{}
	g := keys.NewGenerator(runner, "image:tag")

	res, err := g.Generate(context.Background(), state, dir, false)
	require.NoError(t, err)
	require.Len(t, res.Validators, 3)
	require.Len(t, res.Wallets, 3)
	require.NotNil(t, res.Root)

	for i, v := range res.Validators {
		require.Equal(t, i, v.Index)
		require.Equal(t, fmt.Sprintf("bls_node-%d", i), v.PubKey)
	}
	for i, w := range res.Wallets {
		require.Equal(t, fmt.Sprintf("klv1_node-%d", i), w.Address)
	}
	require.Equal(t, "klv1_root", res.Root.Address)

	require.Len(t, runner.calls, 4)
}

func TestGenerator_Idempotent(t *testing.T) {
	dir := t.TempDir()
	state := domain.NewLocalnetState(2, domain.DefaultMaxSupply, 2)

	first := &fakeKeygenRunner{}
	_, err := keys.NewGenerator(first, "image:tag").Generate(context.Background(), state, dir, false)
	require.NoError(t, err)
	require.Len(t, first.calls, 3)

	second := &fakeKeygenRunner{}
	_, err = keys.NewGenerator(second, "image:tag").Generate(context.Background(), state, dir, false)
	require.NoError(t, err)
	require.Empty(t, second.calls, "second run without --force must skip docker")
}

func TestGenerator_Force_Regenerates(t *testing.T) {
	dir := t.TempDir()
	state := domain.NewLocalnetState(2, domain.DefaultMaxSupply, 2)

	_, err := keys.NewGenerator(&fakeKeygenRunner{}, "image:tag").Generate(context.Background(), state, dir, false)
	require.NoError(t, err)

	forced := &fakeKeygenRunner{}
	_, err = keys.NewGenerator(forced, "image:tag").Generate(context.Background(), state, dir, true)
	require.NoError(t, err)
	require.Len(t, forced.calls, 3)
}

func TestGenerator_IncompleteStateRequiresForce(t *testing.T) {
	dir := t.TempDir()
	state := domain.NewLocalnetState(2, domain.DefaultMaxSupply, 2)

	runner := &fakeKeygenRunner{}
	_, err := keys.NewGenerator(runner, "image:tag").Generate(context.Background(), state, dir, false)
	require.NoError(t, err)

	require.NoError(t, os.Remove(filepath.Join(dir, "node-1", "walletKey.pem")))

	_, err = keys.NewGenerator(&fakeKeygenRunner{}, "image:tag").Generate(context.Background(), state, dir, false)
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrKeysIncomplete))
}

func TestGenerator_LinuxUserArgs(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("user args only added on Linux")
	}

	dir := t.TempDir()
	state := domain.NewLocalnetState(1, domain.DefaultMaxSupply, 1)

	runner := &fakeKeygenRunner{}
	_, err := keys.NewGenerator(runner, "image:tag").Generate(context.Background(), state, dir, false)
	require.NoError(t, err)

	foundUser := false
	for _, call := range runner.calls {
		for i, a := range call {
			if a == "--user" && i+1 < len(call) {
				foundUser = true
			}
		}
	}
	require.True(t, foundUser, "--user flag must be present on Linux")
}

func TestGenerator_PEMPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod semantics differ on Windows")
	}

	dir := t.TempDir()
	state := domain.NewLocalnetState(1, domain.DefaultMaxSupply, 1)

	runner := &fakeKeygenRunner{}
	_, err := keys.NewGenerator(runner, "image:tag").Generate(context.Background(), state, dir, false)
	require.NoError(t, err)

	pem := filepath.Join(dir, "node-0", "validatorKey.pem")
	info, err := os.Stat(pem)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	nodeDir := filepath.Join(dir, "node-0")
	info, err = os.Stat(nodeDir)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o700), info.Mode().Perm())
}
