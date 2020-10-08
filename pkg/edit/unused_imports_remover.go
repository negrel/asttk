package edit

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/negrel/asttk/pkg/inspector"
)

type unusedImportsRemover struct {
	allImports      []*ast.ImportSpec
	requiredImports []ast.Spec
}

// RemoveUnusedImports return an inspector.Inspector that analyze the required package
// and a function to remove package that are not required.
func RemoveUnusedImports() (inspector.Inspector, func(file *ast.File)) {
	uir := new(unusedImportsRemover)

	return uir.inspect, uir.removeImports
}

func (uir *unusedImportsRemover) inspect(node ast.Node) (recursive bool) {
	recursive = true

	if file, isFile := node.(*ast.File); isFile {
		uir.allImports = file.Imports
	}

	ident, ok := node.(*ast.Ident)
	if !ok || ident.Obj != nil {
		return
	}

	for _, _import := range uir.allImports {
		// Storing package identifier (name or last folder name in path)
		var name string
		if identifier := _import.Name; identifier != nil {
			name = identifier.Name
		} else {
			slice := strings.Split(_import.Path.Value, string(filepath.Separator))
			name = slice[len(slice)-1]
			name = strings.Trim(name, "\"")
		}

		if name == ident.Name {
			uir.requiredImports = append(uir.requiredImports, _import)
		}
	}

	return
}

func (uir *unusedImportsRemover) removeImports(file *ast.File) {
	for _, d := range file.Decls {
		decl, ok := d.(*ast.GenDecl)
		if !ok {
			continue
		}

		if decl.Tok != token.IMPORT {
			continue
		}

		decl.Specs = uir.requiredImports
		break
	}
}
