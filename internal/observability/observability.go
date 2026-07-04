// Package observability provides structured logging and optional telemetry
// integration for versionconductor using omniobserve.
package observability

import (
	"context"
	"log/slog"
	"os"

	"github.com/plexusone/omniobserve/observops"
	_ "github.com/plexusone/omniobserve/observops/otlp" // OTLP provider
)

// Config holds observability configuration.
type Config struct {
	// ServiceName is the service name for telemetry.
	ServiceName string

	// ServiceVersion is the service version.
	ServiceVersion string

	// LogLevel is the minimum log level.
	LogLevel slog.Level

	// Verbose enables debug-level logging.
	Verbose bool

	// Provider is the observability provider name (otlp, newrelic, datadog).
	Provider string

	// Endpoint is the OTLP endpoint (e.g., localhost:4317).
	Endpoint string

	// APIKey is the API key for cloud providers (New Relic, Datadog).
	APIKey string

	// Disabled disables telemetry collection (local logging only).
	Disabled bool
}

// DefaultConfig returns a default configuration for local-only logging.
func DefaultConfig() Config {
	return Config{
		ServiceName:    "versionconductor",
		ServiceVersion: "dev",
		LogLevel:       slog.LevelInfo,
		Disabled:       true, // Local-only by default
	}
}

// Observability manages the observability stack.
type Observability struct {
	provider observops.Provider
	logger   *slog.Logger
	config   Config
}

// New creates a new Observability instance with the given configuration.
func New(cfg Config) (*Observability, error) {
	obs := &Observability{config: cfg}

	// Determine log level
	level := cfg.LogLevel
	if cfg.Verbose {
		level = slog.LevelDebug
	}

	// Create console handler for local output
	consoleHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})

	// If telemetry is disabled, use local-only logging
	if cfg.Disabled || cfg.Provider == "" {
		obs.logger = slog.New(consoleHandler)
		return obs, nil
	}

	// Create observability provider
	opts := []observops.ClientOption{
		observops.WithServiceName(cfg.ServiceName),
		observops.WithServiceVersion(cfg.ServiceVersion),
	}

	if cfg.Endpoint != "" {
		opts = append(opts, observops.WithEndpoint(cfg.Endpoint))
	}
	if cfg.APIKey != "" {
		opts = append(opts, observops.WithAPIKey(cfg.APIKey))
	}

	provider, err := observops.Open(cfg.Provider, opts...)
	if err != nil {
		// Fall back to local-only logging
		obs.logger = slog.New(consoleHandler)
		obs.logger.Warn("failed to initialize observability provider, using local logging",
			"provider", cfg.Provider,
			"error", err,
		)
		return obs, nil
	}

	obs.provider = provider

	// Create slog handler with both local and remote output
	handler := provider.SlogHandler(
		observops.WithSlogLocalHandler(consoleHandler),
		observops.WithSlogRemoteLevel(int(slog.LevelInfo)),
	)

	obs.logger = slog.New(handler)
	return obs, nil
}

// Logger returns the configured slog.Logger.
func (o *Observability) Logger() *slog.Logger {
	return o.logger
}

// Provider returns the observability provider (may be nil if disabled).
func (o *Observability) Provider() observops.Provider {
	return o.provider
}

// Shutdown gracefully shuts down the observability stack.
func (o *Observability) Shutdown(ctx context.Context) error {
	if o.provider != nil {
		return o.provider.Shutdown(ctx)
	}
	return nil
}

// Context key for logger.
type ctxKey struct{}

// WithLogger adds a logger to the context.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

// LoggerFromContext retrieves the logger from context.
// Returns slog.Default() if no logger is in context.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok && logger != nil {
		return logger
	}
	return slog.Default()
}

// L is a shorthand for LoggerFromContext.
func L(ctx context.Context) *slog.Logger {
	return LoggerFromContext(ctx)
}
