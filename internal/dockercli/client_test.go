package dockercli_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/klever-io/kleverchain-localnet/internal/dockercli"
	"github.com/klever-io/kleverchain-localnet/internal/errs"
)

func TestVersion_Parses(t *testing.T) {
	f := newFakeRunner()
	f.expect("docker --version", fakeResponse{Stdout: "Docker version 24.0.7, build afdd53b\n"})

	c := dockercli.New(dockercli.WithRunner(f))
	v, err := c.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, 24, v.Major)
	require.Equal(t, 0, v.Minor)
	require.Equal(t, 7, v.Patch)
}

func TestVersion_ErrorPropagated(t *testing.T) {
	f := newFakeRunner()
	f.defaultRe = fakeResponse{Err: errors.New("docker not found")}

	c := dockercli.New(dockercli.WithRunner(f))
	_, err := c.Version(context.Background())
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrPrereq))
}

func TestComposeVersion_Parses(t *testing.T) {
	f := newFakeRunner()
	f.expect("docker compose version", fakeResponse{Stdout: "Docker Compose version v2.27.0\n"})

	c := dockercli.New(dockercli.WithRunner(f))
	ci, err := c.ComposeVersion(context.Background())
	require.NoError(t, err)
	require.Equal(t, 2, ci.Major)
	require.Equal(t, 27, ci.Minor)
}

func TestDaemonReachable(t *testing.T) {
	f := newFakeRunner()
	f.expect("docker info", fakeResponse{Stdout: "Server Version: 24.0.7\n"})

	c := dockercli.New(dockercli.WithRunner(f))
	require.NoError(t, c.DaemonReachable(context.Background()))
}

func TestComposeUp_RequiresComposeFile(t *testing.T) {
	dir := t.TempDir()
	f := newFakeRunner()
	c := dockercli.New(dockercli.WithRunner(f), dockercli.WithWorkDir(dir))

	_, err := c.ComposeUp(context.Background())
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrComposeMissing))
}

func TestComposeUp_WithComposeFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docker-compose.yaml"), []byte(""), 0o644))
	f := newFakeRunner()
	f.expect("docker compose up -d", fakeResponse{Stdout: "creating...\n"})

	c := dockercli.New(dockercli.WithRunner(f), dockercli.WithWorkDir(dir))
	_, err := c.ComposeUp(context.Background())
	require.NoError(t, err)
	require.Len(t, f.calls, 1)
	require.Equal(t, []string{"compose", "up", "-d"}, f.calls[0].Args)
}

func TestComposeDown_AltFileName(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(""), 0o644))
	f := newFakeRunner()
	f.expect("docker compose down", fakeResponse{Stdout: "done\n"})

	c := dockercli.New(dockercli.WithRunner(f), dockercli.WithWorkDir(dir))
	_, err := c.ComposeDown(context.Background())
	require.NoError(t, err)
}

func TestFindComposeFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte("x"), 0o644))
	got, err := dockercli.FindComposeFile(dir)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(dir, "docker-compose.yml"), got)

	other := t.TempDir()
	_, err = dockercli.FindComposeFile(other)
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrComposeMissing))
}
