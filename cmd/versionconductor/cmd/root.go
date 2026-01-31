package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

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
	return rootCmd.Execute()
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

	// Bind flags to viper
	_ = viper.BindPFlag("orgs", rootCmd.PersistentFlags().Lookup("orgs"))
	_ = viper.BindPFlag("repos", rootCmd.PersistentFlags().Lookup("repos"))
	_ = viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	_ = viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	_ = viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
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
