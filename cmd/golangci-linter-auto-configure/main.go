package main

import (
	"github.com/larsartmann/golangcli-linter-auto-configure/internal/cli"
)

var version = "dev"

func main() {
	cli.Version = version

	cli.Main()
}
