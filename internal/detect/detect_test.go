package detect_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/klever-io/kleverchain-localnet/internal/detect"
)

type fakeProbe struct {
	services []detect.DockerService
}

func (f *fakeProbe) ListServices(_ context.Context) ([]detect.DockerService, error) {
	return f.services, nil
}

func writeFile(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(""), 0o644))
}

func TestDetect_Uninitialized(t *testing.T) {
	dir := t.TempDir()
	d := detect.Detector{WorkDir: dir}
	snap := d.Detect(context.Background())
	require.Equal(t, detect.Uninitialized, snap.State)
}

func TestDetect_KeysOnly(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "keys", "walletKey.pem"))
	writeFile(t, filepath.Join(dir, "keys", "node-0", "validatorKey.pem"))
	writeFile(t, filepath.Join(dir, "keys", "node-0", "walletKey.pem"))

	d := detect.Detector{WorkDir: dir}
	snap := d.Detect(context.Background())
	require.Equal(t, detect.KeysOnly, snap.State)
	require.Equal(t, 1, snap.Validators)
}

func TestDetect_Ready(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "keys", "walletKey.pem"))
	writeFile(t, filepath.Join(dir, "keys", "node-0", "validatorKey.pem"))
	writeFile(t, filepath.Join(dir, "keys", "node-0", "walletKey.pem"))
	writeFile(t, filepath.Join(dir, "config", "node", "genesis.json"))
	writeFile(t, filepath.Join(dir, "config", "node", "nodesSetup.json"))
	writeFile(t, filepath.Join(dir, "docker-compose.yaml"))

	d := detect.Detector{WorkDir: dir}
	snap := d.Detect(context.Background())
	require.Equal(t, detect.Ready, snap.State)
}

func TestDetect_Running(t *testing.T) {
	dir := t.TempDir()
	probe := &fakeProbe{services: []detect.DockerService{
		{Name: "seednode", State: "running", Health: "healthy"},
		{Name: "node0", State: "running", Health: "healthy"},
	}}

	d := detect.Detector{WorkDir: dir, Docker: probe}
	snap := d.Detect(context.Background())
	require.Equal(t, detect.Running, snap.State)
	require.Equal(t, 2, snap.RunningServices)
	require.Equal(t, 2, snap.HealthyServices)
}

func TestDetect_Partial(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "keys", "walletKey.pem"))
	writeFile(t, filepath.Join(dir, "keys", "node-0", "validatorKey.pem"))

	d := detect.Detector{WorkDir: dir}
	snap := d.Detect(context.Background())
	require.Equal(t, detect.Partial, snap.State)
}
