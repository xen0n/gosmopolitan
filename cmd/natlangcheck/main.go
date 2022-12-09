package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"regexp"

	"golang.org/x/tools/go/packages"
)

var reHanChars = regexp.MustCompile(`\p{Han}`)

func main() {
	cfg := packages.Config{
		Mode: packages.NeedCompiledGoFiles | // for acting on anything at all
			packages.NeedTypes | // for Fset for printing positions
			packages.NeedSyntax, // for the actual AST nodes
	}
	pkgs, err := packages.Load(&cfg, "./...")
	if err != nil {
		panic(err)
	}

	numErrors := 0
	for _, pkg := range pkgs {
		numErrors += processPkg(pkg)
	}

	if numErrors > 0 {
		os.Exit(1)
	}
}

func processPkg(pkg *packages.Package) int {
	numErrors := 0
	for _, astRoot := range pkg.Syntax {
		numErrors += processFile(astRoot, pkg.Fset)
	}
	return numErrors
}

func processFile(root *ast.File, fset *token.FileSet) int {
	// fmt.Printf("checking %s\n", filename)

	numErrors := 0
	ast.Inspect(root, func(n ast.Node) bool {
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
			numErrors++

			p := fset.Position(lit.Pos())
			fmt.Printf(
				"%s: string literal contains %s script chars: %s\n",
				p.String(),
				"Han",
				lit.Value,
			)
		}

		return true
	})

	return numErrors
}
