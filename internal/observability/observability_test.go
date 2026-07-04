package observability

import (
	"context"
	"log/slog"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ServiceName != "versionconductor" {
		t.Errorf("ServiceName = %q, want %q", cfg.ServiceName, "versionconductor")
	}
	if cfg.ServiceVersion != "dev" {
		t.Errorf("ServiceVersion = %q, want %q", cfg.ServiceVersion, "dev")
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, slog.LevelInfo)
	}
	if !cfg.Disabled {
		t.Error("Disabled = false, want true")
	}
}

func TestNew_LocalOnly(t *testing.T) {
	cfg := DefaultConfig()
	obs, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() { _ = obs.Shutdown(context.Background()) }()

	if obs.Logger() == nil {
		t.Error("Logger() = nil, want non-nil")
	}
	if obs.Provider() != nil {
		t.Error("Provider() = non-nil, want nil for disabled telemetry")
	}
}

func TestNew_Verbose(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Verbose = true

	obs, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() { _ = obs.Shutdown(context.Background()) }()

	if obs.Logger() == nil {
		t.Error("Logger() = nil, want non-nil")
	}
}

func TestNew_InvalidProvider(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Disabled = false
	cfg.Provider = "invalid-provider"

	obs, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v, want nil (fallback to local)", err)
	}
	defer func() { _ = obs.Shutdown(context.Background()) }()

	// Should fall back to local logging
	if obs.Logger() == nil {
		t.Error("Logger() = nil, want non-nil")
	}
}

func TestWithLogger(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	ctxWithLogger := WithLogger(ctx, logger)

	got := LoggerFromContext(ctxWithLogger)
	if got != logger {
		t.Error("LoggerFromContext() did not return the expected logger")
	}
}

func TestLoggerFromContext_NoLogger(t *testing.T) {
	ctx := context.Background()

	got := LoggerFromContext(ctx)
	if got == nil {
		t.Error("LoggerFromContext() = nil, want slog.Default()")
	}
}

func TestL(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	ctxWithLogger := WithLogger(ctx, logger)

	got := L(ctxWithLogger)
	if got != logger {
		t.Error("L() did not return the expected logger")
	}
}

func TestShutdown_NilProvider(t *testing.T) {
	cfg := DefaultConfig()
	obs, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Shutdown should not error when provider is nil
	if err := obs.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown() error = %v, want nil", err)
	}
}
