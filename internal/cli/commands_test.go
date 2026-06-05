package cli_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCLICommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Commands Suite")
}
