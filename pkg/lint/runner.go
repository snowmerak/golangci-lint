package lint

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/snowmerak/golangci-lint/internal/errorutil"
	"github.com/snowmerak/golangci-lint/pkg/config"
	"github.com/snowmerak/golangci-lint/pkg/fsutils"
	"github.com/snowmerak/golangci-lint/pkg/goutil"
	"github.com/snowmerak/golangci-lint/pkg/lint/linter"
	"github.com/snowmerak/golangci-lint/pkg/lint/lintersdb"
	"github.com/snowmerak/golangci-lint/pkg/logutils"
	"github.com/snowmerak/golangci-lint/pkg/result"
	"github.com/snowmerak/golangci-lint/pkg/result/processors"
	"github.com/snowmerak/golangci-lint/pkg/timeutils"
)

type processorStat struct {
	inCount  int
	outCount int
}

type Runner struct {
	Log logutils.Log

	lintCtx    *linter.Context
	Processors []processors.Processor
}

func NewRunner(log logutils.Log, cfg *config.Config, args []string, goenv *goutil.Env,
	lineCache *fsutils.LineCache, fileCache *fsutils.FileCache,
	dbManager *lintersdb.Manager, lintCtx *linter.Context,
) (*Runner, error) {
	// Beware that some processors need to add the path prefix when working with paths
	// because they get invoked before the path prefixer (exclude and severity rules)
	// or process other paths (skip files).
	files := fsutils.NewFiles(lineCache, cfg.Output.PathPrefix)

	skipFilesProcessor, err := processors.NewSkipFiles(cfg.Issues.ExcludeFiles, cfg.Output.PathPrefix)
	if err != nil {
		return nil, err
	}

	skipDirs := cfg.Issues.ExcludeDirs
	if cfg.Issues.UseDefaultExcludeDirs {
		skipDirs = append(skipDirs, processors.StdExcludeDirRegexps...)
	}

	skipDirsProcessor, err := processors.NewSkipDirs(log.Child(logutils.DebugKeySkipDirs), skipDirs, args, cfg.Output.PathPrefix)
	if err != nil {
		return nil, err
	}

	enabledLinters, err := dbManager.GetEnabledLintersMap()
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled linters: %w", err)
	}

	return &Runner{
		Processors: []processors.Processor{
			processors.NewCgo(goenv),

			// Must go after Cgo.
			processors.NewFilenameUnadjuster(lintCtx.Packages, log.Child(logutils.DebugKeyFilenameUnadjuster)),

			// Must go after FilenameUnadjuster.
			processors.NewInvalidIssue(log.Child(logutils.DebugKeyInvalidIssue)),

			// Must be before diff, nolint and exclude autogenerated processor at least.
			processors.NewPathPrettifier(),
			skipFilesProcessor,
			skipDirsProcessor, // must be after path prettifier

			processors.NewAutogeneratedExclude(cfg.Issues.ExcludeGenerated),

			// Must be before exclude because users see already marked output and configure excluding by it.
			processors.NewIdentifierMarker(),

			processors.NewExclude(&cfg.Issues),
			processors.NewExcludeRules(log.Child(logutils.DebugKeyExcludeRules), files, &cfg.Issues),
			processors.NewNolint(log.Child(logutils.DebugKeyNolint), dbManager, enabledLinters),

			processors.NewUniqByLine(cfg),
			processors.NewDiff(&cfg.Issues),
			processors.NewMaxPerFileFromLinter(cfg),
			processors.NewMaxSameIssues(cfg.Issues.MaxSameIssues, log.Child(logutils.DebugKeyMaxSameIssues), cfg),
			processors.NewMaxFromLinter(cfg.Issues.MaxIssuesPerLinter, log.Child(logutils.DebugKeyMaxFromLinter), cfg),
			processors.NewSourceCode(lineCache, log.Child(logutils.DebugKeySourceCode)),
			processors.NewPathShortener(),
			processors.NewSeverity(log.Child(logutils.DebugKeySeverityRules), files, &cfg.Severity),

			// The fixer still needs to see paths for the issues that are relative to the current directory.
			processors.NewFixer(cfg, log, fileCache),

			// Now we can modify the issues for output.
			processors.NewPathPrefixer(cfg.Output.PathPrefix),
			processors.NewSortResults(cfg),
		},
		lintCtx: lintCtx,
		Log:     log,
	}, nil
}

func (r *Runner) Run(ctx context.Context, linters []*linter.Config) ([]result.Issue, error) {
	sw := timeutils.NewStopwatch("linters", r.Log)
	defer sw.Print()

	var (
		lintErrors error
		issues     []result.Issue
	)

	for _, lc := range linters {
		lc := lc
		sw.TrackStage(lc.Name(), func() {
			linterIssues, err := r.runLinterSafe(ctx, r.lintCtx, lc)
			if err != nil {
				lintErrors = errors.Join(lintErrors, fmt.Errorf("can't run linter %s", lc.Linter.Name()), err)
				r.Log.Warnf("Can't run linter %s: %v", lc.Linter.Name(), err)

				return
			}

			issues = append(issues, linterIssues...)
		})
	}

	return r.processLintResults(issues), lintErrors
}

func (r *Runner) runLinterSafe(ctx context.Context, lintCtx *linter.Context,
	lc *linter.Config,
) (ret []result.Issue, err error) {
	defer func() {
		if panicData := recover(); panicData != nil {
			if pe, ok := panicData.(*errorutil.PanicError); ok {
				err = fmt.Errorf("%s: %w", lc.Name(), pe)

				// Don't print stacktrace from goroutines twice
				r.Log.Errorf("Panic: %s: %s", pe, pe.Stack())
			} else {
				err = fmt.Errorf("panic occurred: %s", panicData)
				r.Log.Errorf("Panic stack trace: %s", debug.Stack())
			}
		}
	}()

	issues, err := lc.Linter.Run(ctx, lintCtx)

	if lc.DoesChangeTypes {
		// Packages in lintCtx might be dirty due to the last analysis,
		// which affects to the next analysis.
		// To avoid this issue, we clear type information from the packages.
		// See https://github.com/golangci/golangci-lint/pull/944.
		// Currently, DoesChangeTypes is true only for `unused`.
		lintCtx.ClearTypesInPackages()
	}

	if err != nil {
		return nil, err
	}

	for i := range issues {
		if issues[i].FromLinter == "" {
			issues[i].FromLinter = lc.Name()
		}
	}

	return issues, nil
}

func (r *Runner) processLintResults(inIssues []result.Issue) []result.Issue {
	sw := timeutils.NewStopwatch("processing", r.Log)

	var issuesBefore, issuesAfter int
	statPerProcessor := map[string]processorStat{}

	var outIssues []result.Issue
	if len(inIssues) != 0 {
		issuesBefore += len(inIssues)
		outIssues = r.processIssues(inIssues, sw, statPerProcessor)
		issuesAfter += len(outIssues)
	}

	// finalize processors: logging, clearing, no heavy work here

	for _, p := range r.Processors {
		p := p
		sw.TrackStage(p.Name(), func() {
			p.Finish()
		})
	}

	if issuesBefore != issuesAfter {
		r.Log.Infof("Issues before processing: %d, after processing: %d", issuesBefore, issuesAfter)
	}
	r.printPerProcessorStat(statPerProcessor)
	sw.PrintStages()

	return outIssues
}

func (r *Runner) printPerProcessorStat(stat map[string]processorStat) {
	parts := make([]string, 0, len(stat))
	for name, ps := range stat {
		if ps.inCount != 0 {
			parts = append(parts, fmt.Sprintf("%s: %d/%d", name, ps.outCount, ps.inCount))
		}
	}
	if len(parts) != 0 {
		r.Log.Infof("Processors filtering stat (out/in): %s", strings.Join(parts, ", "))
	}
}

func (r *Runner) processIssues(issues []result.Issue, sw *timeutils.Stopwatch, statPerProcessor map[string]processorStat) []result.Issue {
	for _, p := range r.Processors {
		var newIssues []result.Issue
		var err error
		p := p
		sw.TrackStage(p.Name(), func() {
			newIssues, err = p.Process(issues)
		})

		if err != nil {
			r.Log.Warnf("Can't process result by %s processor: %s", p.Name(), err)
		} else {
			stat := statPerProcessor[p.Name()]
			stat.inCount += len(issues)
			stat.outCount += len(newIssues)
			statPerProcessor[p.Name()] = stat
			issues = newIssues
		}

		if issues == nil {
			issues = []result.Issue{}
		}
	}

	return issues
}
