package nosprintfhostport

import (
	"github.com/stbenjam/no-sprintf-host-port/pkg/analyzer"
	"golang.org/x/tools/go/analysis"

	"github.com/snowmerak/golangci-lint/pkg/goanalysis"
)

func New() *goanalysis.Linter {
	a := analyzer.Analyzer

	return goanalysis.NewLinter(
		a.Name,
		a.Doc,
		[]*analysis.Analyzer{a},
		nil,
	).WithLoadMode(goanalysis.LoadModeSyntax)
}
