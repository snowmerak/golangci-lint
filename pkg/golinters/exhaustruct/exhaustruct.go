package exhaustruct

import (
	"github.com/GaijinEntertainment/go-exhaustruct/v3/analyzer"
	"golang.org/x/tools/go/analysis"

	"github.com/snowmerak/golangci-lint/pkg/config"
	"github.com/snowmerak/golangci-lint/pkg/goanalysis"
	"github.com/snowmerak/golangci-lint/pkg/golinters/internal"
)

func New(settings *config.ExhaustructSettings) *goanalysis.Linter {
	var include, exclude []string
	if settings != nil {
		include = settings.Include
		exclude = settings.Exclude
	}

	a, err := analyzer.NewAnalyzer(include, exclude)
	if err != nil {
		internal.LinterLogger.Fatalf("exhaustruct configuration: %v", err)
	}

	return goanalysis.NewLinter(
		a.Name,
		a.Doc,
		[]*analysis.Analyzer{a},
		nil,
	).WithLoadMode(goanalysis.LoadModeTypesInfo)
}
