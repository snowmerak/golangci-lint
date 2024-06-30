package processors

import (
	"github.com/snowmerak/golangci-lint/pkg/result"
)

const typeCheckName = "typecheck"

type Processor interface {
	Process(issues []result.Issue) ([]result.Issue, error)
	Name() string
	Finish()
}
