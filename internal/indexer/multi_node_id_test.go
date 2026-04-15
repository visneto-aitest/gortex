package indexer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/zzet/gortex/internal/config"
	"github.com/zzet/gortex/internal/graph"
	"github.com/zzet/gortex/internal/search"
)

// setupGoRepoWithTypes creates a Go repo with a Handler type and a method —
// enough to exercise multi-repo indexing without needing the Go toolchain
// for semantic enrichment.
func setupGoRepoWithTypes(t *testing.T, name string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "api"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "go.mod"),
		[]byte("module example.com/"+name+"\n\ngo 1.21\n"),
		0o644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "api", "handler.go"),
		[]byte(`package api

type Handler struct{}

func (h *Handler) CreateTuck() string { return "ok" }
`),
		0o644,
	))
	return dir
}

// TestTrackRepoCtx_FirstOfManyStillGetsPrefix guards against the bug where
// the first repo tracked via TrackRepoCtx at daemon warmup was indexed
// without a RepoPrefix because willBeMultiRepo was decided by counting
// `mi.repos` (which is empty at iteration 0). The symptom was asymmetric
// IDs across repos: one repo's nodes under "internal/api/handler.go::X",
// the rest under "worker/internal/api/handler.go::X". Halved Go edge
// density in multi-repo graphs.
func TestTrackRepoCtx_FirstOfManyStillGetsPrefix(t *testing.T) {
	repoA := setupGoRepoWithTypes(t, "repo-a")
	repoB := setupGoRepoWithTypes(t, "repo-b")

	tmpCfg := filepath.Join(t.TempDir(), "config.yaml")
	gc := &config.GlobalConfig{
		Repos: []config.RepoEntry{
			{Path: repoA, Name: "repo-a"},
			{Path: repoB, Name: "repo-b"},
		},
	}
	gc.SetConfigPath(tmpCfg)
	require.NoError(t, gc.Save())

	cm, err := config.NewConfigManager(tmpCfg)
	require.NoError(t, err)

	g := graph.New()
	mi := NewMultiIndexer(g, newTestRegistry(), search.NewBM25(), cm, zap.NewNop())

	// Simulate warmupDaemonState's loop: TrackRepoCtx each config'd repo
	// in order. The first call is the one that used to skip prefixing.
	for _, entry := range cm.Global().Repos {
		_, err := mi.TrackRepoCtx(context.Background(), entry)
		require.NoError(t, err, "tracking %s", entry.Name)
	}

	require.True(t, mi.IsMultiRepo(), "setup must produce multi-repo mode")

	// Every node must carry a non-empty RepoPrefix and its FilePath must
	// live under that prefix. Any violation means a code path bypassed
	// applyRepoPrefix.
	var missingPrefix, badFilePaths []string
	for _, n := range g.AllNodes() {
		if n.RepoPrefix == "" {
			missingPrefix = append(missingPrefix, n.ID)
			continue
		}
		if n.FilePath != "" && !strings.HasPrefix(n.FilePath, n.RepoPrefix+"/") {
			badFilePaths = append(badFilePaths,
				n.ID+" (FilePath="+n.FilePath+", RepoPrefix="+n.RepoPrefix+")")
		}
	}
	assert.Empty(t, missingPrefix,
		"nodes without RepoPrefix leaked into multi-repo graph (first-repo prefix bug):\n  %s",
		strings.Join(missingPrefix, "\n  "))
	assert.Empty(t, badFilePaths,
		"nodes with FilePath outside their RepoPrefix:\n  %s",
		strings.Join(badFilePaths, "\n  "))

	// No node ID should begin with an absolute filesystem path — that's
	// the shape stale snapshot nodes take, and no current indexing path
	// should produce it.
	for _, n := range g.AllNodes() {
		assert.False(t, strings.HasPrefix(n.ID, "/"),
			"node ID begins with absolute path: %s", n.ID)
	}

	// Both repos must have contributed Handler.CreateTuck, each under its
	// own prefix. This is the positive counterpart to the prefix check.
	want := map[string]bool{
		"repo-a/api/handler.go::Handler.CreateTuck": false,
		"repo-b/api/handler.go::Handler.CreateTuck": false,
	}
	for _, n := range g.AllNodes() {
		if _, ok := want[n.ID]; ok {
			want[n.ID] = true
		}
	}
	for id, found := range want {
		assert.True(t, found, "expected prefixed node %s not found in graph", id)
	}
}
