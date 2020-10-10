package parse

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

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

func fmtErrors(errors []packages.Error) error {
	if length := len(errors); length != 0 {
		errors := fmt.Sprintf("%v error(s) found:\n", length)

		for _, err := range errors {
			errors += fmt.Sprint(err) + "\n"
		}

		return fmt.Errorf(errors)
	}

	return nil
}

func extractFile(pkg *packages.Package) []*GoFile {
	goFiles := make([]*GoFile, len(pkg.Syntax))
	for i := 0; i < len(goFiles); i++ {
		goFiles[i] = &GoFile{
			path: pkg.GoFiles[i],
			ast:  pkg.Syntax[i],
			fset: pkg.Fset,
		}
	}

	return goFiles
}
