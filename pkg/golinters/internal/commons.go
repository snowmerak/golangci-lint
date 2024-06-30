package internal

import "github.com/snowmerak/golangci-lint/pkg/logutils"

// LinterLogger must be use only when the context logger is not available.
var LinterLogger = logutils.NewStderrLog(logutils.DebugKeyLinter)
