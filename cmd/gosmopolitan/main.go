package main

import (
	"github.com/xen0n/gosmopolitan"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(gosmopolitan.Analyzer)
}
