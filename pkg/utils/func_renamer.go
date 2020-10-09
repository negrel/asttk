package utils

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/negrel/asttk/pkg/inspector"
)

type funcRenamer struct {
	filter func(name string) (replaceName string, ok bool)
}

// RenameFunc return two inspector.Inspector, one to rename function declaration and another one
// to rename function call.
func RenameFunc(filter func(name string) (replaceName string, ok bool)) (renameFuncDecl, renameFuncCall inspector.Inspector) {
	f := &funcRenamer{
		filter: filter,
	}

	return f.renameFuncDecl, f.renameFuncCall
}

func (f *funcRenamer) renameFuncDecl(node ast.Node) (recursive bool) {
	recursive = true

	funcDecl, isFuncDecl := node.(*ast.FuncDecl)
	if !isFuncDecl {
		return
	}

	funcName := funcDecl.Name.Name

	newName, ok := f.filter(funcName)
	if !ok || newName == "" {
		return false
	}
	funcDecl.Name.Name = newName

	return
}

func (f *funcRenamer) renameFuncCall(node ast.Node) (recursive bool) {
	recursive = true

	callExpr, isCallExpr := node.(*ast.CallExpr)
	if !isCallExpr {
		return
	}

	switch fun := callExpr.Fun.(type) {
	case *ast.Ident:
		newName, ok := f.filter(fun.Name)
		if !ok {
			return
		}
		f.replaceFuncInCallExpr(callExpr, newName)

	default:
		return
	}

	return
}

func (f *funcRenamer) replaceFuncInCallExpr(callExpr *ast.CallExpr, newName string) {
	if !token.IsIdentifier(newName) {
		panic(fmt.Sprintf("%v is an invalid new name", newName))
	}

	split := strings.Split(newName, ".")

	if length := len(split); length == 2 {
		callExpr.Fun = &ast.SelectorExpr{
			Sel: ast.NewIdent(split[0]),
			X:   ast.NewIdent(split[1]),
		}
	} else if length == 1 {
		callExpr.Fun = ast.NewIdent(newName)
	} else {
		panic(fmt.Sprintf("%v is an invalid new name", newName))
	}
}
