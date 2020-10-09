package utils

import (
	"go/ast"

	"github.com/negrel/asttk/pkg/inspector"
)

func RemoveComments(filter func(comment string) bool) inspector.Inspector {
	return func(node ast.Node) (recursive bool) {
		recursive = true

		if node == nil {
			return
		}

		if file, isFile := node.(*ast.File); isFile {
			file.Comments = nil
			file.Doc = nil
		}

		commentGroup, isCommentGroup := node.(*ast.CommentGroup)
		if !isCommentGroup {
			return
		}

		for i, comment := range commentGroup.List {
			if filter(comment.Text) {
				commentGroup.List = append(commentGroup.List[:i], commentGroup.List[i+1:]...)
			}
		}

		return
	}
}

// RemoveAllComments return an Inspector
func RemoveAllComments() inspector.Inspector {
	return RemoveComments(func(_ string) bool { return true })
}
