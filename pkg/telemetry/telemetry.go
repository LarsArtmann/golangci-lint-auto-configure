// Package telemetry provides strictly opt-in OpenTelemetry tracing for CLI
// runs. Tracing is OFF by default: nothing is initialized, the global tracer
// stays the noop implementation, and no network egress happens unless the
// user enables it via the --otel flag or the EnvVarEnable environment
// variable (SECURITY.md's local-only posture holds by default).
//
// Design split (how-to-golang policy): reusable packages (pkg/linter, the
// migrate command) only call the OpenTelemetry API via otel.Tracer — noop
// unless a host registers a provider. The SDK and the OTLP/HTTP exporter
// live HERE and are wired exclusively at the CLI boundary, so embedded
// consumers (linter-autoconfigure-sdk hosts such as BuildFlow) never pay
// for, or accidentally activate, telemetry.
//
// Like the audit ledger, telemetry is best-effort: a setup or export
// failure is logged as a warning and degrades to noop; a run never fails
// because tracing could not be initialized.
package telemetry

import (
	"context"
	"os"
	"time"

	"charm.land/log/v2"
	errorfamily "github.com/larsartmann/go-error-family"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
)

// EnvVarEnable turns optional telemetry on ("1" or any non-empty value),
// mirroring the audit ledger's EnvVarDisable pattern.
const EnvVarEnable = "GOLANGCI_LINT_AUTO_CONFIGURE_OTEL"

// ServiceName identifies this tool in exported traces.
const ServiceName = "golangci-lint-auto-configure"

// ShutdownTimeout bounds the final flush so a slow collector can never hang
// an interactive CLI run.
const ShutdownTimeout = 5 * time.Second

// Enabled reports whether telemetry should start: the --otel flag value OR
// the EnvVarEnable environment variable.
func Enabled(flagValue bool) bool {
	return flagValue || os.Getenv(EnvVarEnable) != ""
}

// Options carries the immutable setup inputs.
type Options struct {
	// ServiceVersion is stamped on the trace resource (pkg/version.Get().Version).
	ServiceVersion string
	// Logger receives best-effort warnings; never nil when Setup is called by the CLI.
	Logger *log.Logger
}

// Session owns one CLI run's telemetry lifetime: the root span and the
// tracer-provider shutdown. The zero value is ready to use and Finish is a
// no-op until Setup succeeded, so callers never need to check state.
type Session struct {
	span     trace.Span
	shutdown func(context.Context) error
	logger   *log.Logger
}

// Setup initializes the global tracer provider and starts the root span
// ("cli.<command>") when telemetry is enabled, returning the span context so
// the caller can propagate it (cobra: cmd.SetContext) — child spans then
// derive from the root span. Best-effort: on failure it logs a warning,
// leaves the session inert, and returns the incoming context unchanged.
//
//nolint:spancheck // the root span deliberately outlives Setup: it is ended in Session.Finish after the run
func (s *Session) Setup(ctx context.Context, commandName string, opts Options) context.Context {
	logger := fallbackLogger(opts.Logger)

	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		logger.Warnf("Telemetry disabled: OTLP exporter init failed: %v", err)

		return ctx
	}

	res, err := newResource(opts.ServiceVersion)
	if err != nil {
		logger.Warnf("Telemetry disabled: resource init failed: %v", err)

		return ctx
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(provider)

	spanCtx, span := provider.Tracer(ServiceName).Start(ctx, "cli."+commandName)

	s.span = span
	s.shutdown = provider.Shutdown
	s.logger = logger

	return spanCtx
}

// fallbackLogger returns the provided logger or a silent error-level default
// so Setup can always warn on its best-effort paths.
func fallbackLogger(logger *log.Logger) *log.Logger {
	if logger == nil {
		return log.NewWithOptions(nil, log.Options{Level: log.ErrorLevel})
	}

	return logger
}

// newResource builds the trace resource identifying this service and version.
func newResource(serviceVersion string) (*resource.Resource, error) {
	res, err := resource.Merge(resource.Default(), resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(ServiceName),
		semconv.ServiceVersion(serviceVersion),
	))
	if err != nil {
		return nil, errorfamily.Wrap(err, errorfamily.Infrastructure, "telemetry.resource", "merge trace resource")
	}

	return res, nil
}

// Finish records the run outcome on the root span (error family/code on
// failure), ends it, and flushes the provider with ShutdownTimeout. Safe to
// call on an inert session and safe to call exactly once.
func (s *Session) Finish(ctx context.Context, runErr error) {
	if s.span == nil {
		return
	}

	if runErr != nil {
		s.span.SetAttributes(
			attribute.String("error.family", errorfamily.Classify(runErr).String()),
			attribute.String("error.code", errorfamily.Code(runErr)),
		)
		s.span.SetStatus(codes.Error, runErr.Error())
	} else {
		s.span.SetStatus(codes.Ok, "")
	}

	s.span.End()

	if s.shutdown == nil {
		return
	}

	if err := s.shutdown(ctx); err != nil {
		// Best-effort by contract; a lost span batch must never fail the run.
		s.logger.Warnf("Telemetry shutdown: flushing spans failed: %v", err)
	}
}
