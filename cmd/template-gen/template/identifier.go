package template

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/openmcp-project/platform-service-template/template-gen/logs"
)

// NewIdentifierTransformer creates a Transformer for ast.Ident ast.Nodes
// by matching find and replacing it with replace.
//
// e.g. find=Foo and replace=Quota will transform
// type FooReconciler struct { ... }
// to
// type QuotaReconciler struct { ... }
//
// Use IdentifierTransform to transform identifiers of functions, types, etc.
func NewIdentifierTransformer(list ...FindAndReplace) Transformer {
	return &identifierTransformer{
		findAndReplaceList: list,
	}
}

var _ Transformer = &identifierTransformer{}

type identifierTransformer struct {
	findAndReplaceList []FindAndReplace
}

// Transforms implements [Transformer].
func (t *identifierTransformer) Transforms(n ast.Node) bool {
	_, ok := n.(*ast.Ident)
	return ok
}

// Execute implements [Transformer].
func (t *identifierTransformer) Execute(pkg *packages.Package, n ast.Node) {
	identifier := n.(*ast.Ident)
	if identifier == nil {
		return
	}
	position := pkg.Fset.Position(n.Pos())
	for _, fr := range t.findAndReplaceList {
		if strings.Contains(identifier.Name, fr.Find) {
			renamed := strings.Replace(identifier.Name, fr.Find, fr.Replace, 1)
			logs.Debug(fmt.Sprintf("%s: identifier (%s) renamed to %s", position.Filename, identifier.Name, renamed))
			identifier.Name = renamed
		}
	}
}
