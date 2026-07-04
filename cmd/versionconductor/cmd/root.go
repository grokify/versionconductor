package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/plexusone/versionconductor/internal/observability"
)

var (
	cfgFile string
	obs     *observability.Observability
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "versionconductor",
	Short: "Automated dependency PR management and maintenance releases",
	Long: `VersionConductor is a CLI tool for automated dependency PR management
and maintenance releases across multiple GitHub repositories.

Features:
  - Scan for Renovate/Dependabot PRs across organizations
  - Auto-review and merge dependency PRs based on Cedar policies
  - Create maintenance releases when dependencies are updated

Part of the DevOpsOrchestra suite alongside PipelineConductor.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	// Set up signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize observability after config is loaded
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return initObservability()
	}

	// Ensure observability is shut down
	rootCmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		if obs != nil {
			return obs.Shutdown(ctx)
		}
		return nil
	}

	return rootCmd.ExecuteContext(ctx)
}

// initObservability initializes the observability stack.
func initObservability() error {
	cfg := observability.Config{
		ServiceName:    "versionconductor",
		ServiceVersion: Version,
		Verbose:        viper.GetBool("verbose"),
		Provider:       viper.GetString("otel-provider"),
		Endpoint:       viper.GetString("otel-endpoint"),
		APIKey:         viper.GetString("otel-api-key"),
		Disabled:       !viper.GetBool("otel-enabled"),
	}

	if cfg.Verbose {
		cfg.LogLevel = slog.LevelDebug
	}

	var err error
	obs, err = observability.New(cfg)
	if err != nil {
		return fmt.Errorf("initializing observability: %w", err)
	}

	// Set as default logger
	slog.SetDefault(obs.Logger())

	return nil
}

// GetLogger returns the configured logger.
func GetLogger() *slog.Logger {
	if obs != nil {
		return obs.Logger()
	}
	return slog.Default()
}

// GetContext returns a context with the logger attached.
func GetContext(ctx context.Context) context.Context {
	return observability.WithLogger(ctx, GetLogger())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.versionconductor.yaml)")
	rootCmd.PersistentFlags().StringSlice("orgs", nil, "GitHub organizations to scan")
	rootCmd.PersistentFlags().StringSlice("repos", nil, "Specific repositories (owner/repo format)")
	rootCmd.PersistentFlags().String("token", "", "GitHub token (or set GITHUB_TOKEN env var)")
	rootCmd.PersistentFlags().String("format", "table", "Output format: table, json, markdown, csv")
	rootCmd.PersistentFlags().Bool("dry-run", false, "Show what would happen without making changes")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")

	// Observability flags
	rootCmd.PersistentFlags().Bool("otel-enabled", false, "Enable OpenTelemetry observability")
	rootCmd.PersistentFlags().String("otel-provider", "otlp", "Observability provider: otlp, newrelic, datadog")
	rootCmd.PersistentFlags().String("otel-endpoint", "", "OTLP endpoint (e.g., localhost:4317)")
	rootCmd.PersistentFlags().String("otel-api-key", "", "API key for cloud providers")

	// Bind flags to viper
	_ = viper.BindPFlag("orgs", rootCmd.PersistentFlags().Lookup("orgs"))
	_ = viper.BindPFlag("repos", rootCmd.PersistentFlags().Lookup("repos"))
	_ = viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	_ = viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	_ = viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("otel-enabled", rootCmd.PersistentFlags().Lookup("otel-enabled"))
	_ = viper.BindPFlag("otel-provider", rootCmd.PersistentFlags().Lookup("otel-provider"))
	_ = viper.BindPFlag("otel-endpoint", rootCmd.PersistentFlags().Lookup("otel-endpoint"))
	_ = viper.BindPFlag("otel-api-key", rootCmd.PersistentFlags().Lookup("otel-api-key"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".versionconductor" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".versionconductor")
	}

	// Environment variables
	viper.SetEnvPrefix("VERSIONCONDUCTOR")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// Also check GITHUB_TOKEN directly
	if viper.GetString("token") == "" {
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			viper.Set("token", token)
		}
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}
