package testableexamples

import (
	"github.com/maratori/testableexamples/pkg/testableexamples"
	"golang.org/x/tools/go/analysis"

	"github.com/snowmerak/golangci-lint/pkg/goanalysis"
)

func New() *goanalysis.Linter {
	a := testableexamples.NewAnalyzer()

	return goanalysis.NewLinter(
		a.Name,
		a.Doc,
		[]*analysis.Analyzer{a},
		nil,
	).WithLoadMode(goanalysis.LoadModeSyntax)
}
