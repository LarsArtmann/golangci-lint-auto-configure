package main

import (
	"github.com/larsartmann/golangci-lint-auto-configure/internal/cli"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/version"
)

func main() {
	cli.Version = version.Get().String()
	cli.VersionInfo = version.Get()

	cli.Main()
}
