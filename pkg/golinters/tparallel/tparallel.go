package tparallel

import (
	"github.com/moricho/tparallel"
	"golang.org/x/tools/go/analysis"

	"github.com/snowmerak/golangci-lint/pkg/goanalysis"
)

func New() *goanalysis.Linter {
	a := tparallel.Analyzer
	return goanalysis.NewLinter(
		a.Name,
		a.Doc,
		[]*analysis.Analyzer{a},
		nil,
	).WithLoadMode(goanalysis.LoadModeTypesInfo)
}
