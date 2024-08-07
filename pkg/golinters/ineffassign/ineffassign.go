package ineffassign

import (
	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"golang.org/x/tools/go/analysis"

	"github.com/snowmerak/golangci-lint/pkg/goanalysis"
)

func New() *goanalysis.Linter {
	a := ineffassign.Analyzer

	return goanalysis.NewLinter(
		a.Name,
		"Detects when assignments to existing variables are not used",
		[]*analysis.Analyzer{a},
		nil,
	).WithLoadMode(goanalysis.LoadModeSyntax)
}
