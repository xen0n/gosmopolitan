package gosmopolitan

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "gosmopolitan",
	Doc:  "gosmopolitan checks for possible hurdles to i18n/l10n",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run: func(p *analysis.Pass) (any, error) {
		pctx := processCtx{p: p}
		return pctx.run()
	},
	RunDespiteErrors: false,
}

var reHanChars = regexp.MustCompile(`\p{Han}`)

type processCtx struct {
	p *analysis.Pass
}

func (c *processCtx) fset() *token.FileSet {
	return c.p.Fset
}

func (c *processCtx) run() (any, error) {
	insp := c.p.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	insp.Nodes(nil, func(n ast.Node, push bool) bool {
		// we only need to look at each node once
		if !push {
			return false
		}

		// skip import blocks that can contain string literals but are not
		// interesting for us
		if _, ok := n.(*ast.ImportSpec); ok {
			return false
		}

		// check only string literals
		lit, ok := n.(*ast.BasicLit)
		if !ok {
			return true
		}
		if lit.Kind != token.STRING {
			return true
		}

		// report string literals containing characters of given script (in
		// the sense of "writing system")
		// for now only the Han script is being checked
		if reHanChars.MatchString(lit.Value) {
			diag := analysis.Diagnostic{
				Pos:     lit.Pos(),
				End:     lit.End(),
				Message: fmt.Sprintf("string literal contains %s script char(s)", "Han"),
			}
			c.p.Report(diag)
		}

		return true
	})

	// check time.Local usages
	insp.Nodes([]ast.Node{(*ast.Ident)(nil)}, func(n ast.Node, push bool) bool {
		// we only need to look at each node once
		if !push {
			return false
		}

		ident := n.(*ast.Ident)

		d := c.p.TypesInfo.ObjectOf(ident)
		if d == nil || d.Pkg() == nil {
			return true
		}

		if d.Pkg().Path() == "time" && d.Name() == "Local" {
			diag := analysis.Diagnostic{
				Pos:     n.Pos(),
				End:     n.End(),
				Message: "usage of time.Local",
			}
			c.p.Report(diag)

		}

		return true
	})

	return nil, nil
}
