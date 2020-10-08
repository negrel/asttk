package parse

import (
	"fmt"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// GoPackage define a loaded/parsed go package.
type GoPackage struct {
	pkgPath string
	path    string
	subPkgs []*GoPackage
	Files   []*GoFile
	fset    *token.FileSet
}

func findSubPkgs(dir string) (subPkgs []*GoPackage) {
	filesInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	for _, fileInfo := range filesInfo {
		if !fileInfo.IsDir() {
			continue
		}

		filePath := filepath.Join(dir, fileInfo.Name())
		subPkg, err := Package(filePath, true)
		if err != nil {
			continue
		}

		subPkgs = append(subPkgs, subPkg)
	}

	return
}

// Package parse an entire package at the given path and return a new *GoPackage.
func Package(pkgPath string, parseSubPkgs bool) (*GoPackage, error) {
	if pkgPath == "" {
		return nil, fmt.Errorf("the given path is empty")
	}

	pkgPath, err := filepath.Abs(pkgPath)
	if err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(pkgPath)
	if err != nil {
		return nil, err
	}
	if !fileInfo.IsDir() {
		return nil, fmt.Errorf("the given path is not a directory")
	}

	config := Config
	config.Dir = pkgPath
	pkgs, err := packages.Load(&config)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		if len(pkg.GoFiles) == 0 {
			continue
		}

		path := filepath.Dir(pkg.GoFiles[0])
		if pkgPath != path {
			continue
		}

		subPkgs := []*GoPackage{}
		if parseSubPkgs {
			subPkgs = findSubPkgs(pkgPath)
		}

		if length := len(pkg.Errors); length != 0 {
			errors := fmt.Sprintf("%v error(s) found:\n", length)
			for _, err := range pkg.Errors {
				errors += fmt.Sprint(err) + "\n"
			}

			return nil, fmt.Errorf(errors)
		}

		goFiles := make([]*GoFile, len(pkg.Syntax))
		for i := 0; i < len(goFiles); i++ {
			goFiles[i] = &GoFile{
				path: pkg.GoFiles[i],
				ast:  pkg.Syntax[i],
			}
		}

		return &GoPackage{
			pkgPath: pkg.PkgPath,
			path:    path,
			subPkgs: subPkgs,
			Files:   goFiles,
		}, nil
	}

	return nil, fmt.Errorf("package not found")
}

// Path return the package absolute path.
func (p *GoPackage) Path() string {
	return p.path
}

// PkgPath return the package import path.
func (p *GoPackage) PkgPath() string {
	return p.pkgPath
}

// Name return the package name.
func (p *GoPackage) Name() string {
	return filepath.Base(p.path)
}

// FileSet return the fileset of the package.
func (p *GoPackage) FileSet() *token.FileSet {
	return p.fset
}

// SubPkgs return all the subpackages.
func (p *GoPackage) SubPkgs() []*GoPackage {
	return p.subPkgs
}

// WritePkg method write the go file source code in the file at the given
// path.
func (p *GoPackage) WritePkg(path string, writeSubPkgs bool) error {
	for _, file := range p.Files {
		err := file.WriteFile(filepath.Join(path, file.Name()))
		if err != nil {
			return err
		}
	}

	if !writeSubPkgs {
		return nil
	}
	for _, subPkg := range p.subPkgs {
		err := subPkg.WritePkg(filepath.Join(path, subPkg.Name()), true)
		if err != nil {
			return err
		}
	}

	return nil
}
