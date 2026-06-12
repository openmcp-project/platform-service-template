package template

import (
	"go/ast"

	"golang.org/x/tools/go/packages"
)

// Transformer transform a Node
type Transformer interface {
	// Transform return true if the Transformer does transform the given node type
	Transforms(ast.Node) bool
	// Execute node transformation
	Execute(*packages.Package, ast.Node)
}

type FindAndReplace struct {
	Find    string
	Replace string
}
