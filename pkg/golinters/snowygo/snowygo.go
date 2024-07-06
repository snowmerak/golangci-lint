package snowygo

import (
	"github.com/snowmerak/golangci-lint/pkg/config"
	"github.com/snowmerak/golangci-lint/pkg/goanalysis"
	"github.com/snowmerak/snowygo"
	"golang.org/x/tools/go/analysis"
)

func New(settings *config.SnowyGoSettings) *goanalysis.Linter {
	a := snowygo.NewAnalyzerWithConfig(&snowygo.Config{})

	return goanalysis.
		NewLinter(a.Name, a.Doc, []*analysis.Analyzer{a}, nil).
		WithLoadMode(goanalysis.LoadModeTypesInfo)
}
