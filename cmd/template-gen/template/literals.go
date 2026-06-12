package template

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/openmcp-project/platform-service-template/template-gen/logs"
)

// NewLiteralTransformer creates a Transformer for ast.BasicLit ast.Nodes
// by matching find and replacing it with replace.
//
// e.g. find=foo and replace=quota will transform
// Resources: []string{"fooservices"}
// to
// Resources: []string{"quotaservices"}
//
// Use LiteralTransform to replace string values written in source code like imports, value assignment, etc.
func NewLiteralTransformer(list ...FindAndReplace) Transformer {
	return &literalTransformer{
		findAndReplaceList: list,
	}
}

var _ Transformer = &literalTransformer{}

type literalTransformer struct {
	findAndReplaceList []FindAndReplace
}

// Transforms implements [Transformer].
func (t *literalTransformer) Transforms(n ast.Node) bool {
	_, ok := n.(*ast.BasicLit)
	return ok
}

// Execute implements [Transformer].
func (t *literalTransformer) Execute(pkg *packages.Package, n ast.Node) {
	f := n.(*ast.BasicLit)
	position := pkg.Fset.Position(n.Pos())
	for _, fr := range t.findAndReplaceList {
		if strings.Contains(f.Value, fr.Find) {
			original := f.Value
			replaced := strings.Replace(original, fr.Find, fr.Replace, 1)
			logs.Debug(fmt.Sprintf("%s: literal (%s) transformed to %s", position.Filename, original, replaced))
			f.Value = replaced
		}
	}
}
