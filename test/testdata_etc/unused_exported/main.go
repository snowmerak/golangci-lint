package main

import (
	"github.com/snowmerak/golangci-lint/test/testdata_etc/unused_exported/lib"
)

func main() {
	lib.PublicFunc()
}
