package template

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"log"
	"os"

	"golang.org/x/tools/go/packages"
)

// Runner allows to configure and run the template code transformation
type Runner struct {
	module                 string
	kind                   string
	debug                  bool
	dryRun                 bool
	transformers           []Transformer
	findAndReplaceComments []FindAndReplace
	directories            []string
}

// NewRunner returns a Runner to execute the template code transformation
func NewRunner(module, kind string, debug, dryrun bool) *Runner {
	return &Runner{
		module: module,
		kind:   kind,
		debug:  debug,
		dryRun: dryrun,
	}
}

func (r *Runner) WithTransformers(transformers ...Transformer) *Runner {
	r.transformers = transformers
	return r
}

func (r *Runner) WithCommentReplacements(comments ...FindAndReplace) *Runner {
	r.findAndReplaceComments = comments
	return r
}

func (r *Runner) WithPackageDirectorties(dirs ...string) *Runner {
	r.directories = dirs
	return r
}

func (r *Runner) Execute() {
	for _, dir := range r.directories {
		r.TransformPackages(dir)
	}
}

// TransformPackages traveres each package in a root directory and executes each code Transformer of the Runner
func (r *Runner) TransformPackages(directory string) {
	cfg := &packages.Config{
		Mode:  packages.NeedSyntax | packages.NeedFiles | packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:   directory,
		Tests: true,
	}
	commentTransformer := NewCommentTransformer(r.findAndReplaceComments...)
	// load every package in the directory
	p, err := packages.Load(cfg, "./...")
	if err != nil {
		log.Fatalf("load package failed: %v", err)
		return
	}
	for _, pkg := range p {
		for _, file := range pkg.Syntax {
			if file == nil {
				continue
			}
			// traverse AST and execute code transformations
			ast.Inspect(file, func(n ast.Node) bool {
				if n == nil {
					return true
				}
				for _, t := range r.transformers {
					if t.Transforms(n) {
						t.Execute(pkg, n)
					}
				}
				return true
			})
			// transform comments
			commentTransformer.Transform(file)
			// prepare to write file after code tranformation
			var buffer bytes.Buffer
			if err := format.Node(&buffer, pkg.Fset, file); err != nil {
				log.Fatalf("format node failed: %v", err)
				return
			}
			// write to stdout when dry-run and no debug
			if r.dryRun && !r.debug {
				os.Stdout.Write(buffer.Bytes())
			}
			// write to fs when no dry-run
			if !r.dryRun {
				position := pkg.Fset.Position(file.Pos())
				if err := os.WriteFile(position.Filename, buffer.Bytes(), 0644); err != nil {
					log.Fatalf("write file failed: %v", err)
					return
				}
			}
		}
	}
}

func ParseExpr(expr string) {
	e, err := parser.ParseExpr(expr)
	_ = e
	_ = err
}
