package main

import (
	"gometric/internal/analysis/exitchecker"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"

	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"

	goCritic "github.com/go-critic/go-critic/checkers/analyzer"
	gosecAnalyzers "github.com/securego/gosec/v2/analyzers"
)

func main() {
	analyzers := []*analysis.Analyzer{
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sortslice.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
	}

	// staticcheck.io all SA analyzers
	for _, v := range staticcheck.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}

	// staticcheck.io S analyzers
	simplechecks := map[string]bool{
		"S1001": true,
	}
	for _, v := range simple.Analyzers {
		if simplechecks[v.Analyzer.Name] {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// securego.io analyzers
	analyzers = append(analyzers, gosecAnalyzers.BuildDefaultAnalyzers()...)

	// go-critic.com analyzer
	analyzers = append(analyzers, goCritic.Analyzer)

	// internal exitchecker analyzer
	analyzers = append(analyzers, exitchecker.Analyzer)

	multichecker.Main(
		analyzers...,
	)
}
