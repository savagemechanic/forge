package folder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscover_DetectsGoModule(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module github.com/test/project\n\ngo 1.23\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644))

	info, err := Discover(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, "github.com/test/project", info.GoModule)
	assert.True(t, containsType(info.Types, ProjectTypeGoModule))
}

func TestDiscover_DetectsGitRoot(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a git repo
	require.NoError(t, os.Mkdir(filepath.Join(tmpDir, ".git"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example\n\ngo 1.23\n"), 0644))

	// Create a subdirectory and discover from there
	subDir := filepath.Join(tmpDir, "internal", "foo")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "foo.go"), []byte("package foo"), 0644))

	info, err := Discover(subDir)
	require.NoError(t, err)

	assert.Equal(t, tmpDir, info.GitRoot)
	assert.Equal(t, tmpDir, info.RootPath)
	assert.Equal(t, "example", info.GoModule)
}

func TestDiscover_NoProjectType(t *testing.T) {
	tmpDir := t.TempDir()
	info, err := Discover(tmpDir)
	require.NoError(t, err)
	assert.True(t, containsType(info.Types, ProjectTypeUnknown))
}

func TestDiscover_DetectsGoWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.work"), []byte("go 1.23\n"), 0644))

	info, err := Discover(tmpDir)
	require.NoError(t, err)
	assert.True(t, info.HasGoWork)
	assert.True(t, containsType(info.Types, ProjectTypeGoWorkspace))
}

func TestDiscover_FindsSubpackages(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.23\n"), 0644))

	for _, pkg := range []string{"cmd/app", "internal/domain", "internal/service"} {
		dir := filepath.Join(tmpDir, filepath.FromSlash(pkg))
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte("package "+filepath.Base(dir)), 0644))
	}

	info, err := Discover(tmpDir)
	require.NoError(t, err)
	assert.NotEmpty(t, info.Subpackages)
}

func TestProjectInfo_Summary(t *testing.T) {
	info := &ProjectInfo{
		RootPath:  "/tmp/project",
		GoModule:  "github.com/test/project",
		GitRoot:   "/tmp/project",
		Types:     []ProjectType{ProjectTypeGo, ProjectTypeGoModule, ProjectTypeGit},
		Subpackages: []string{"cmd", "internal"},
	}
	s := info.Summary()
	assert.Contains(t, s, "Go module")
	assert.Contains(t, s, "git root")
}
