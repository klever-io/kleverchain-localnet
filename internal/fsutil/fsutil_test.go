package fsutil_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/klever-io/kleverchain-localnet/internal/fsutil"
)

func TestEnsureDir_Creates(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "a", "b", "c")
	require.NoError(t, fsutil.EnsureDir(target, 0o755))
	require.True(t, fsutil.Exists(target))
}

func TestWriteJSONAndReadJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out", "data.json")

	in := map[string]any{"a": 1.0, "b": "hello"}
	require.NoError(t, fsutil.WriteJSON(path, in))
	require.True(t, fsutil.Exists(path))

	var out map[string]any
	require.NoError(t, fsutil.ReadJSON(path, &out))
	require.Equal(t, in, out)

	marshaled, err := json.MarshalIndent(in, "", "    ")
	require.NoError(t, err)
	_ = marshaled
}

func TestRemoveIfExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	require.NoError(t, fsutil.WriteFile(path, []byte("{}"), 0o644))
	require.True(t, fsutil.Exists(path))
	require.NoError(t, fsutil.RemoveIfExists(path))
	require.False(t, fsutil.Exists(path))
	require.NoError(t, fsutil.RemoveIfExists(path))
}

func TestRelativeOrAbs(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "keys", "node-0")
	require.NoError(t, fsutil.EnsureDir(nested, 0o755))
	rel, err := fsutil.RelativeOrAbs(dir, nested)
	require.NoError(t, err)
	require.Equal(t, "keys/node-0", rel)
}
