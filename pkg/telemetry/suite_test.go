package telemetry_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTelemetry(t *testing.T) {
	RegisterFailHandler(Fail) //art-dupl:accept Ginkgo per-package bootstrap
	RunSpecs(t, "Telemetry Suite")
}
