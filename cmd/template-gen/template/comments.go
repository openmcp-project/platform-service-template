package template

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/openmcp-project/platform-service-template/template-gen/logs"
)

type CommentTransformer struct {
	findAndReplaceList []FindAndReplace
}

func NewCommentTransformer(findAndReplaceList ...FindAndReplace) *CommentTransformer {
	return &CommentTransformer{
		findAndReplaceList: findAndReplaceList,
	}
}

func (t *CommentTransformer) Transform(file *ast.File) {
	for _, c := range file.Comments {
		for _, comm := range c.List {
			original := comm.Text
			transformed := false
			for _, fr := range t.findAndReplaceList {
				if strings.Contains(comm.Text, fr.Find) {
					comm.Text = strings.ReplaceAll(comm.Text, fr.Find, fr.Replace)
					transformed = true
				}
			}
			if transformed {
				logs.Debug(fmt.Sprintf("replaced comment (%s) with %s", original, comm.Text))
			}
		}
	}
}
