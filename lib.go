package gosmopolitan

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"
	"strings"

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

	// ignore test files, because test files could be full of i18n and l10n
	// fixtures, and we want to focus on the actual run-time logic
	//
	// TODO: is there a way to both ignore test files earlier, and make use of
	// inspect.Analyzer's cached results? currently Inspector doesn't provide
	// a way to selectively travese some files' AST but not others.
	isBelongingToTestFiles := func(n ast.Node) bool {
		return strings.HasSuffix(c.p.Fset.File(n.Pos()).Name(), "_test.go")
	}

	insp.Nodes(nil, func(n ast.Node, push bool) bool {
		// we only need to look at each node once
		if !push {
			return false
		}

		if isBelongingToTestFiles(n) {
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
			c.p.Report(analysis.Diagnostic{
				Pos:     lit.Pos(),
				End:     lit.End(),
				Message: fmt.Sprintf("string literal contains %s script char(s)", "Han"),
			})
		}

		return true
	})

	// check time.Local usages
	insp.Nodes([]ast.Node{(*ast.Ident)(nil)}, func(n ast.Node, push bool) bool {
		// we only need to look at each node once
		if !push {
			return false
		}

		if isBelongingToTestFiles(n) {
			return false
		}

		ident := n.(*ast.Ident)

		d := c.p.TypesInfo.ObjectOf(ident)
		if d == nil || d.Pkg() == nil {
			return true
		}

		if d.Pkg().Path() == "time" && d.Name() == "Local" {
			c.p.Report(analysis.Diagnostic{
				Pos:     n.Pos(),
				End:     n.End(),
				Message: "usage of time.Local",
			})
		}

		return true
	})

	return nil, nil
}
