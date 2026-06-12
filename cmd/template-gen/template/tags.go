package template

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/openmcp-project/platform-service-template/template-gen/logs"
)

// NewTagTransformer creates a Transformer for ast.Field ast.Nodes where field.Tag is set
// by matching find and replacing it with replace.
//
// e.g. find=foo and replace=quota will transform
// Foo *string `json:"foo,omitempty"`
// to
// Foo *string `json:"quota,omitempty"`
//
// Use TagTransformer to transform struct field tags
func NewTagTransformer(list ...FindAndReplace) Transformer {
	return &tagTransformer{
		findAndReplaceList: list,
	}
}

var _ Transformer = &tagTransformer{}

type tagTransformer struct {
	findAndReplaceList []FindAndReplace
}

// Transforms implements [Transformer].
func (t *tagTransformer) Transforms(n ast.Node) bool {
	f, ok := n.(*ast.Field)
	return ok && f != nil && f.Tag != nil
}

// Execute implements [Transformer].
func (t *tagTransformer) Execute(pkg *packages.Package, n ast.Node) {
	f := n.(*ast.Field)
	position := pkg.Fset.Position(n.Pos())
	for _, fr := range t.findAndReplaceList {
		if strings.Contains(f.Tag.Value, fr.Find) {
			original := f.Tag.Value
			replaced := strings.Replace(original, fr.Find, fr.Replace, 1)
			logs.Debug(fmt.Sprintf("%s: tag (%s) transformed to %s", position.Filename, original, replaced))
			f.Tag.Value = replaced
		}
	}
}
