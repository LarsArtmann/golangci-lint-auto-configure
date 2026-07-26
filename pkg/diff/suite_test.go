package diff_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDiff(t *testing.T) {
	RegisterFailHandler(Fail) //art-dupl:accept Ginkgo per-package bootstrap
	RunSpecs(t, "Diff Suite")
}
