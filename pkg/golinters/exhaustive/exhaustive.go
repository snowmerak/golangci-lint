package exhaustive

import (
	"github.com/nishanths/exhaustive"
	"golang.org/x/tools/go/analysis"

	"github.com/snowmerak/golangci-lint/pkg/config"
	"github.com/snowmerak/golangci-lint/pkg/goanalysis"
)

func New(settings *config.ExhaustiveSettings) *goanalysis.Linter {
	a := exhaustive.Analyzer

	var cfg map[string]map[string]any
	if settings != nil {
		cfg = map[string]map[string]any{
			a.Name: {
				exhaustive.CheckFlag:                      settings.Check,
				exhaustive.CheckGeneratedFlag:             settings.CheckGenerated,
				exhaustive.DefaultSignifiesExhaustiveFlag: settings.DefaultSignifiesExhaustive,
				exhaustive.IgnoreEnumMembersFlag:          settings.IgnoreEnumMembers,
				exhaustive.IgnoreEnumTypesFlag:            settings.IgnoreEnumTypes,
				exhaustive.PackageScopeOnlyFlag:           settings.PackageScopeOnly,
				exhaustive.ExplicitExhaustiveMapFlag:      settings.ExplicitExhaustiveMap,
				exhaustive.ExplicitExhaustiveSwitchFlag:   settings.ExplicitExhaustiveSwitch,
				exhaustive.DefaultCaseRequiredFlag:        settings.DefaultCaseRequired,
			},
		}
	}

	return goanalysis.NewLinter(
		a.Name,
		a.Doc,
		[]*analysis.Analyzer{a},
		cfg,
	).WithLoadMode(goanalysis.LoadModeTypesInfo)
}
