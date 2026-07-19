// Package goengine provides Go-specific intelligence: package loading,
// symbol indexing, and reference finding. It uses golang.org/x/tools/go/packages
// to inspect Go source. This is a domain analysis service — no mutations.
package goengine

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Index holds the loaded Go project model.
type Index struct {
	RootPath  string
	Packages  []*PackageInfo
	symbolMap map[string][]*Symbol // name → symbols (across packages)
}

// PackageInfo describes a single loaded Go package.
type PackageInfo struct {
	ImportPath string
	Name       string // package name (last segment)
	Dir        string
	Files      []FileInfo
	Symbols    []*Symbol
	Errors     []packages.Error
}

// FileInfo describes a single Go file.
type FileInfo struct {
	Path     string
	Size     int64
	IsTest   bool
}

// Symbol describes a declared identifier in a package.
type Symbol struct {
	Name     string
	Kind     SymbolKind // func, type, var, const
	Package  string
	Exported bool
	Pos      string // file:line
	File     string
	Line     int
}

// SymbolKind classifies a Go symbol.
type SymbolKind string

const (
	SymbolFunc   SymbolKind = "func"
	SymbolType   SymbolKind = "type"
	SymbolVar    SymbolKind = "var"
	SymbolConst  SymbolKind = "const"
	SymbolMethod SymbolKind = "method"
)

// Load reads all Go packages under dir and builds an index.
// Uses the standard Go toolchain (needs go list / type info).
func Load(dir string) (*Index, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedSyntax | packages.NeedDeps,
		Dir:  dir,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	idx := &Index{
		RootPath:  dir,
		symbolMap: make(map[string][]*Symbol),
	}

	for _, pkg := range pkgs {
		info := &PackageInfo{
			ImportPath: pkg.PkgPath,
			Name:       pkg.Name,
			Dir:        dirOf(pkg),
		}

		for _, f := range pkg.GoFiles {
			info.Files = append(info.Files, FileInfo{Path: f, IsTest: false})
		}
		for _, e := range pkg.Errors {
			info.Errors = append(info.Errors, e)
		}

		// Extract symbols from syntax
		if pkg.Syntax != nil && pkg.Fset != nil {
			for _, file := range pkg.Syntax {
				info.Symbols = append(info.Symbols, extractSymbols(file, pkg.Fset, pkg.PkgPath)...)
			}
		}

		for _, s := range info.Symbols {
			idx.symbolMap[s.Name] = append(idx.symbolMap[s.Name], s)
		}

		idx.Packages = append(idx.Packages, info)
	}

	return idx, nil
}

// extractSymbols walks an AST file and collects top-level declarations.
func extractSymbols(file *ast.File, fset *token.FileSet, pkgPath string) []*Symbol {
	var symbols []*Symbol

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			kind := SymbolFunc
			name := d.Name.Name
			if d.Recv != nil && len(d.Recv.List) > 0 {
				kind = SymbolMethod
				recv := typeString(d.Recv.List[0].Type)
				name = recv + "." + d.Name.Name
			}
			pos := fset.Position(d.Pos())
			symbols = append(symbols, &Symbol{
				Name:     name,
				Kind:     kind,
				Package:  pkgPath,
				Exported: d.Name.IsExported(),
				File:     pos.Filename,
				Line:     pos.Line,
				Pos:      fmt.Sprintf("%s:%d", pos.Filename, pos.Line),
			})

		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					pos := fset.Position(s.Pos())
					symbols = append(symbols, &Symbol{
						Name:     s.Name.Name,
						Kind:     SymbolType,
						Package:  pkgPath,
						Exported: s.Name.IsExported(),
						File:     pos.Filename,
						Line:     pos.Line,
						Pos:      fmt.Sprintf("%s:%d", pos.Filename, pos.Line),
					})
				case *ast.ValueSpec:
					kind := SymbolVar
					if d.Tok == token.CONST {
						kind = SymbolConst
					}
					for _, name := range s.Names {
						pos := fset.Position(name.Pos())
						symbols = append(symbols, &Symbol{
							Name:     name.Name,
							Kind:     kind,
							Package:  pkgPath,
							Exported: name.IsExported(),
							File:     pos.Filename,
							Line:     pos.Line,
							Pos:      fmt.Sprintf("%s:%d", pos.Filename, pos.Line),
						})
					}
				}
			}
		}
	}

	return symbols
}

// FindSymbol returns all symbols matching the name.
func (idx *Index) FindSymbol(name string) []*Symbol {
	return idx.symbolMap[name]
}

// ExportedSymbols returns all exported symbols across all packages.
func (idx *Index) ExportedSymbols() []*Symbol {
	var result []*Symbol
	for _, syms := range idx.symbolMap {
		for _, s := range syms {
			if s.Exported {
				result = append(result, s)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Package != result[j].Package {
			return result[i].Package < result[j].Package
		}
		return result[i].Name < result[j].Name
	})
	return result
}

// PackageByImport returns a package by its import path.
func (idx *Index) PackageByImport(path string) *PackageInfo {
	for _, p := range idx.Packages {
		if p.ImportPath == path {
			return p
		}
	}
	return nil
}

// Summary returns a compact overview of the index.
func (idx *Index) Summary() string {
	pkgCount := len(idx.Packages)
	symbolCount := 0
	fileCount := 0
	for _, p := range idx.Packages {
		symbolCount += len(p.Symbols)
		fileCount += len(p.Files)
	}
	return fmt.Sprintf("%d packages · %d files · %d symbols", pkgCount, fileCount, symbolCount)
}

// PrintPackages returns a formatted listing of all packages.
func (idx *Index) PrintPackages() string {
	var b strings.Builder
	for _, p := range idx.Packages {
		fmt.Fprintf(&b, "  %-50s %d symbols\n", p.ImportPath, len(p.Symbols))
	}
	return b.String()
}

// typeString renders an AST type expression as a string.
func typeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeString(t.X)
	default:
		return "?"
	}
}

func dirOf(pkg *packages.Package) string {
	if len(pkg.GoFiles) > 0 {
		return dirOfPath(pkg.GoFiles[0])
	}
	return ""
}

func dirOfPath(p string) string {
	idx := strings.LastIndex(p, "/")
	if idx < 0 {
		return "."
	}
	return p[:idx]
}

// unused import guard (types used implicitly via packages)
var _ = types.Universe
