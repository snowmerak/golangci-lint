package wsl

import (
	"github.com/bombsimon/wsl/v4"
	"golang.org/x/tools/go/analysis"

	"github.com/snowmerak/golangci-lint/pkg/config"
	"github.com/snowmerak/golangci-lint/pkg/goanalysis"
)

func New(settings *config.WSLSettings) *goanalysis.Linter {
	var conf *wsl.Configuration
	if settings != nil {
		conf = &wsl.Configuration{
			StrictAppend:                     settings.StrictAppend,
			AllowAssignAndCallCuddle:         settings.AllowAssignAndCallCuddle,
			AllowAssignAndAnythingCuddle:     settings.AllowAssignAndAnythingCuddle,
			AllowMultiLineAssignCuddle:       settings.AllowMultiLineAssignCuddle,
			ForceCaseTrailingWhitespaceLimit: settings.ForceCaseTrailingWhitespaceLimit,
			AllowTrailingComment:             settings.AllowTrailingComment,
			AllowSeparatedLeadingComment:     settings.AllowSeparatedLeadingComment,
			AllowCuddleDeclaration:           settings.AllowCuddleDeclaration,
			AllowCuddleWithCalls:             settings.AllowCuddleWithCalls,
			AllowCuddleWithRHS:               settings.AllowCuddleWithRHS,
			ForceCuddleErrCheckAndAssign:     settings.ForceCuddleErrCheckAndAssign,
			ErrorVariableNames:               settings.ErrorVariableNames,
			ForceExclusiveShortDeclarations:  settings.ForceExclusiveShortDeclarations,
			IncludeGenerated:                 true, // force to true because golangci-lint already have a way to filter generated files.
		}
	}

	a := wsl.NewAnalyzer(conf)

	return goanalysis.NewLinter(
		a.Name,
		a.Doc,
		[]*analysis.Analyzer{a},
		nil,
	).WithLoadMode(goanalysis.LoadModeSyntax)
}
