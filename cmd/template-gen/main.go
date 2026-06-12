package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/openmcp-project/platform-service-template/template-gen/logs"
	"github.com/openmcp-project/platform-service-template/template-gen/template"
)

func main() {
	debug := flag.Bool("debug", false, "print debug logs instead of in-memory result")
	dryrun := flag.Bool("dry-run", true, "print in-memory result to stdout without altering any files")
	module := flag.String("module", "github.com/openmcp-project/platform-service-foo", "the go module name of your platform service")
	kind := flag.String("kind", "Foo", "the name of the API to watch on the onboarding cluster")

	flag.Parse()

	logs.Init(*debug)

	log.Println("code transformation start...")

	template.NewRunner(*module, *kind, *debug, *dryrun).
		WithTransformers(
			template.NewIdentifierTransformer(
				template.FindAndReplace{Find: "Foo", Replace: *kind},
			),
			template.NewTagTransformer(
				template.FindAndReplace{Find: "foo", Replace: strings.ToLower(*kind)},
			),
			template.NewLiteralTransformer(
				template.FindAndReplace{Find: "github.com/openmcp-project/platform-service-template", Replace: *module},
				template.FindAndReplace{Find: "foo", Replace: strings.ToLower(*kind)},
				template.FindAndReplace{Find: "platform-service-template", Replace: "platform-service-" + strings.ToLower(*kind)},
			)).
		WithCommentReplacements(
			template.FindAndReplace{Find: "Foo", Replace: "Quota"},
			template.FindAndReplace{Find: "foo", Replace: "quota"},
		).
		WithPackageDirectorties(
			"../../api",
			"../platform-service-template",
			"../../internal",
			"../..test",
		).
		Execute()

	// rename cmd subfolder
	if !*dryrun {
		platformServiceName := filepath.Base(*module)
		if err := os.Rename("cmd/platform-service-template", "cmd/"+platformServiceName); err != nil {
			log.Fatalf("failed to rename directory: %v", err)
		}
	}
}
