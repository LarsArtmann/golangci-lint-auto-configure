package main

import (
	"github.com/larsartmann/golangci-lint-auto-configure/internal/cli"
)

var version = "dev"

func main() {
	cli.Version = version

	cli.Main()
}
