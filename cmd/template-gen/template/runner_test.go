package template_test

import (
	"testing"

	"github.com/openmcp-project/platform-service-template/template-gen/template"
)

func TestParseExpr(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		expr string
	}{
		{
			name: "&Foo{}",
			expr: "&Foo{}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template.ParseExpr(tt.expr)
		})
	}
}
