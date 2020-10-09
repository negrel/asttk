package parse

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// GoFile define a parsed go file.
type GoFile struct {
	path string
	ast  *ast.File
}

// NewGoFile return a GoFile object using the given parameters.
func NewGoFile(path string, ast *ast.File) *GoFile {
	return &GoFile{path: path, ast: ast}
}

// File parse the file at the given path and return a new *GoFile.
// Test file (*_test.go) are not supported.
func File(filePath string) (*GoFile, error) {
	filePath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}
	dir, filename := filepath.Split(filePath)

	if filepath.Ext(filename) != ".go" {
		return nil, fmt.Errorf("the given file path should end with a \".go\" extension")
	}

	// Loading packages
	config := Config
	config.Dir = dir

	pkgs, err := packages.Load(&config)
	if err != nil {
		return nil, err
	}

	var astFile *ast.File = nil
	for _, pkg := range pkgs {
		for i, goFile := range pkg.GoFiles {
			if goFile == filePath {
				astFile = pkg.Syntax[i]
			}
		}
	}
	if astFile == nil {
		return nil, fmt.Errorf("file not found")
	}

	return &GoFile{
		path: filePath,
		ast:  astFile,
	}, nil
}

// Path return the go file absolute path.
func (f *GoFile) Path() string {
	return f.path
}

// Dir return the parent directory of the go file
func (f *GoFile) Dir() string {
	return filepath.Dir(f.path)
}

// Name return the name of the file
func (f *GoFile) Name() string {
	return filepath.Base(f.path)
}

// AST return the *ast.File object of the file.
func (f *GoFile) AST() *ast.File {
	return f.ast
}

// Fprint "pretty-print" the AST of the file to output.
func (f *GoFile) Fprint(output io.Writer) error {
	return format.Node(output, token.NewFileSet(), f.ast)
}

// Bytes convert the AST of the file as an array of byte.
func (f *GoFile) Bytes() ([]byte, error) {
	buf := &bytes.Buffer{}

	err := f.Fprint(buf)

	return buf.Bytes(), err
}

// WriteFile method write the GoFile source code in the file
// at the given path.
func (f *GoFile) WriteFile(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}

	return f.Fprint(file)
}
