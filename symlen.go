package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"unicode/utf8"
)

type Accumulator struct {
	Total uint64
	Count uint64
	Min   uint64
	Max   uint64
}

func (a *Accumulator) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	max := func(x uint64, y uint64) uint64 {
		if x > y {
			return x
		}
		return y
	}

	min := func(x uint64, y uint64) uint64 {
		if x < y {
			return x
		}
		return y
	}

	switch n := node.(type) {
	case *ast.Ident:
		l := uint64(utf8.RuneCountInString(n.Name))
		a.Count++
		a.Total += l
		a.Min = min(l, a.Min)
		a.Max = max(l, a.Max)
	}
	return a
}

func count(p string) error {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, p, nil, 0)
	if err != nil {
		return err
	}

	a := &Accumulator{
		Min: ^uint64(0),
	}

	for _, t := range pkgs {
		ast.Walk(a, t)
	}

	fmt.Printf("identifiers: %d, total: %d, min: %d, max: %d,  average: %f\n", a.Count, a.Total, a.Min, a.Max, float64(a.Total)/float64(a.Count))
	return nil
}

func main() {
	// yes, i know -- there are no flags yet.
	flag.Parse()

	for _, arg := range flag.Args() {
		if err := count(arg); err != nil {
			log.Fatal(err)
		}
	}
}
