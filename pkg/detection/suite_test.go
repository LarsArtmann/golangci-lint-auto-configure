package detection_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDetection(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Detection Suite")
}
