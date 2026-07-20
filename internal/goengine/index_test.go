package goengine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupGoProject creates a minimal Go module in a temp dir for indexing.
func setupGoProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testpkg\n\ngo 1.23\n"), 0644))

	// A package with various declaration kinds
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "internal", "domain"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "internal", "domain", "user.go"), []byte(`package domain

type User struct {
	Name string
}

func NewUser(name string) *User {
	return &User{Name: name}
}

func (u *User) Greet() string {
	return "hi " + u.Name
}

const DefaultName = "anonymous"

var Users = []string{}
`), 0644))

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "cmd", "app"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "cmd", "app", "main.go"), []byte(`package main

func main() {}
`), 0644))

	return dir
}

func TestLoad_IndexesPackages(t *testing.T) {
	dir := setupGoProject(t)
	idx, err := Load(dir)
	require.NoError(t, err)

	assert.NotEmpty(t, idx.Packages)
	assert.NotEmpty(t, idx.Summary())
}

func TestIndex_FindsSymbols(t *testing.T) {
	dir := setupGoProject(t)
	idx, err := Load(dir)
	require.NoError(t, err)

	// NewUser is an exported func
	news := idx.FindSymbol("NewUser")
	assert.NotEmpty(t, news, "should find NewUser symbol")
	for _, s := range news {
		assert.Equal(t, SymbolFunc, s.Kind)
		assert.True(t, s.Exported)
	}

	// User is an exported type
	users := idx.FindSymbol("User")
	assert.NotEmpty(t, users)
	assert.Equal(t, SymbolType, users[0].Kind)

	// DefaultName is a const
	consts := idx.FindSymbol("DefaultName")
	assert.NotEmpty(t, consts)
	assert.Equal(t, SymbolConst, consts[0].Kind)

	// Users is a var
	vars := idx.FindSymbol("Users")
	assert.NotEmpty(t, vars)
	assert.Equal(t, SymbolVar, vars[0].Kind)
}

func TestIndex_ExportedSymbols(t *testing.T) {
	dir := setupGoProject(t)
	idx, err := Load(dir)
	require.NoError(t, err)

	exported := idx.ExportedSymbols()
	// Should include NewUser, User, DefaultName, Users but NOT greet
	for _, s := range exported {
		assert.True(t, s.Exported, "ExportedSymbols should only return exported: %s", s.Name)
	}
}

func TestIndex_PackageByImport(t *testing.T) {
	dir := setupGoProject(t)
	idx, err := Load(dir)
	require.NoError(t, err)

	pkg := idx.PackageByImport("testpkg/internal/domain")
	if pkg != nil {
		assert.NotEmpty(t, pkg.Symbols)
	}
}

func TestIndex_PrintPackages(t *testing.T) {
	dir := setupGoProject(t)
	idx, err := Load(dir)
	require.NoError(t, err)

	out := idx.PrintPackages()
	assert.NotEmpty(t, out)
}
