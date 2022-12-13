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

type AnalyzerConfig struct {
	// EscapeHatches is optionally a list of fully qualified names, in the
	// `(full/pkg/path).name` form, to act as "i18n escape hatches". Inside
	// call-like expressions to those names, the string literal script check
	// is ignored.
	//
	// With this functionality in place, you can use type aliases like
	// `type R = string` as markers, or have explicitly i18n-aware functions
	// exempt from the checks.
	EscapeHatches []string
}

func NewAnalyzer() *analysis.Analyzer {
	var escapeHatchesStr string

	a := &analysis.Analyzer{
		Name: "gosmopolitan",
		Doc:  "gosmopolitan checks for possible hurdles to i18n/l10n",
		Requires: []*analysis.Analyzer{
			inspect.Analyzer,
		},
		Run: func(p *analysis.Pass) (any, error) {
			cfg := AnalyzerConfig{
				EscapeHatches: strings.Split(escapeHatchesStr, ","),
			}
			pctx := processCtx{cfg: &cfg, p: p}
			return pctx.run()
		},
		RunDespiteErrors: false,
	}

	a.Flags.StringVar(
		&escapeHatchesStr,
		"escapehatches",
		"",
		"comma-separated list of fully qualified names to act as i18n escape hatches",
	)

	return a
}

func NewAnalyzerWithConfig(cfg *AnalyzerConfig) *analysis.Analyzer {
	a := &analysis.Analyzer{
		Name: "gosmopolitan",
		Doc:  "gosmopolitan checks for possible hurdles to i18n/l10n",
		Requires: []*analysis.Analyzer{
			inspect.Analyzer,
		},
		Run: func(p *analysis.Pass) (any, error) {
			pctx := processCtx{cfg: cfg, p: p}
			return pctx.run()
		},
		RunDespiteErrors: false,
	}

	return a
}

var DefaultAnalyzer = NewAnalyzer()

var reHanChars = regexp.MustCompile(`\p{Han}`)

type processCtx struct {
	cfg *AnalyzerConfig
	p   *analysis.Pass
}

func sliceToSet[T comparable](x []T) map[T]struct{} {
	// lo.SliceToMap(x, func(k T) (T, struct{}) { return k, struct{}{} })
	y := make(map[T]struct{}, len(x))
	for _, k := range x {
		y[k] = struct{}{}
	}
	return y
}

func (c *processCtx) run() (any, error) {
	escapeHatchesSet := sliceToSet(c.cfg.EscapeHatches)

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

		// and don't look inside escape hatches
		if ce, ok := n.(*ast.CallExpr); ok {
			var ident *ast.Ident
			switch x := ce.Fun.(type) {
			case *ast.Ident:
				ident = x
			case *ast.SelectorExpr:
				ident = x.Sel
			}
			referent := c.p.TypesInfo.Uses[ident]
			fullQualifiedName := fmt.Sprintf("(%s).%s", referent.Pkg().Path(), referent.Name())
			_, isEscapeHatch := escapeHatchesSet[fullQualifiedName]
			// if isEscapeHatch: don't recurse (false)
			return !isEscapeHatch
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
