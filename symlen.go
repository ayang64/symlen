package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path"
	"runtime"
	"unicode/utf8"
)

type Accumulator struct {
	Name  string
	Total uint64
	Count uint64
	Min   uint64
	Max   uint64
	MaxID string
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

		// while we're testing if we've exceded our prior max, also stash the name
		// of the longest identifier.
		m := max(l, a.Max)
		if m != a.Max {
			a.MaxID = n.Name
		}
		a.Max = m
	}
	return a
}

func (a *Accumulator) String() string {
	return fmt.Sprintf("%-5.5s: identifiers: %-5d total: %-5d average: %-5.5f min: %-5d max: %-5d (%q)", path.Base(a.Name), a.Count, a.Total, float64(a.Total)/float64(a.Count), a.Min, a.Max, a.MaxID)
}

func count(p string) (*Accumulator, error) {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, p, nil, 0)
	if err != nil {
		return nil, err
	}

	a := &Accumulator{
		Min:  ^uint64(0),
		Name: p,
	}

	for _, t := range pkgs {
		ast.Walk(a, t)
	}

	return a, nil
}

func main() {
	j := flag.Int("j", runtime.NumCPU()*16, "maximum number of concurrent counters")
	flag.Parse()

	ach := make(chan *Accumulator, *j)

	go func() {
		// used to clamp concurrent operations to *j
		sem := make(chan struct{}, *j)

		for _, arg := range flag.Args() {
			arg := arg

			sem <- struct{}{}
			go func() {
				acc, err := count(arg)
				if err != nil {
					log.Print(err)
				}
				ach <- acc

				<-sem
			}()
		}

		// drain worker queue
		for i := 0; i < cap(sem); i++ {
			sem <- struct{}{}
		}

		close(ach)
	}()

	// read results from ach and print each to the terminal
	for a := range ach {
		fmt.Printf("%s\n", a)
	}
}
