package finding

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFinding(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Finding Suite")
}
