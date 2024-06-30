package exportloopref

import (
	"testing"

	"github.com/snowmerak/golangci-lint/test/testshared/integration"
)

func TestFromTestdata(t *testing.T) {
	integration.RunTestdata(t)
}
