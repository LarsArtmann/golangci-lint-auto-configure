package telemetry_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/telemetry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("telemetry enable gate", func() {
	AfterEach(func() {
		GinkgoT().Setenv(telemetry.EnvVarEnable, "")
	})

	It("is disabled by default (no flag, no env)", func() {
		GinkgoT().Setenv(telemetry.EnvVarEnable, "")

		Expect(telemetry.Enabled(false)).To(BeFalse())
	})

	It("is enabled by the flag alone", func() {
		GinkgoT().Setenv(telemetry.EnvVarEnable, "")

		Expect(telemetry.Enabled(true)).To(BeTrue())
	})

	It("is enabled by the environment variable alone", func() {
		GinkgoT().Setenv(telemetry.EnvVarEnable, "1")

		Expect(telemetry.Enabled(false)).To(BeTrue())
	})
})

var _ = Describe("telemetry session", func() {
	It("Finish is a safe no-op on a zero-value session", func() {
		session := &telemetry.Session{}

		Expect(func() {
			session.Finish(context.Background(), nil)
		}).NotTo(Panic())
	})

	It("exports the root span over OTLP/HTTP when enabled", func() {
		var requests atomic.Int64

		var contentType atomic.Value

		receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			contentType.Store(r.Header.Get("Content-Type"))

			w.WriteHeader(http.StatusOK)
		}))
		defer receiver.Close()

		GinkgoT().Setenv(telemetry.EnvVarEnable, "1")
		GinkgoT().Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", receiver.URL)
		GinkgoT().Setenv("OTEL_EXPORTER_OTLP_TIMEOUT", "3")

		logger := log.NewWithOptions(io.Discard, log.Options{Level: log.ErrorLevel})
		session := &telemetry.Session{}
		session.Setup(context.Background(), "configure", telemetry.Options{
			ServiceVersion: "test",
			Logger:         logger,
		})
		session.Finish(context.Background(), nil)

		Eventually(requests.Load, "5s", "100ms").Should(BeNumerically(">=", 1),
			"the OTLP receiver must receive at least one export request")

		Expect(contentType.Load()).To(ContainSubstring("protobuf"),
			"the export must be binary OTLP protobuf")
	})

	It("stays inert (no egress) when setup was never called", func() {
		var requests atomic.Int64

		receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			requests.Add(1)
			w.WriteHeader(http.StatusOK)
		}))
		defer receiver.Close()

		GinkgoT().Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", receiver.URL)

		session := &telemetry.Session{}
		session.Finish(context.Background(), nil)

		time.Sleep(200 * time.Millisecond)
		Expect(requests.Load()).To(BeZero(),
			"a session without Setup must never export anything")
	})
})
