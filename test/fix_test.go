package test

import (
	"testing"

	"github.com/snowmerak/golangci-lint/test/testshared/integration"
)

func TestFix(t *testing.T) {
	integration.RunFix(t)
}

func TestFix_pathPrefix(t *testing.T) {
	integration.RunFixPathPrefix(t)
}
