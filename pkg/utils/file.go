package utils

import (
	"go/ast"

	"github.com/negrel/asttk/pkg/inspector"
)

// ChangePackage return an Inspector function that will change the package
// of a file.
func ChangePackage(name string) inspector.Inspector {
	return func(node ast.Node) bool {
		pkg, isPackage := node.(*ast.Package)
		if !isPackage {
			return false
		}

		pkg.Name = name

		return false
	}
}

// ApplyOnTopDecl wraps the given Inspectors and call them on
// every ast.File.Decls node. Inspectors wrapped by this helper
// will be called before other inspectors of your inspector.Lead.
func ApplyOnTopDecl(inspectors ...inspector.Inspector) inspector.Inspector {
	wrappers := make([]inspector.Inspector, len(inspectors))

	for i, isp := range inspectors {
		wrappers[i] = func(node ast.Node) bool {
			file, isFile := node.(*ast.File)
			if !isFile {
				return false
			}

			for _, decl := range file.Decls {
				isp(decl)
			}

			return false
		}
	}

	return inspector.Lieutenant(wrappers...)
}
