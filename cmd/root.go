package cmd

import (
	"opnsense-auto-dns/internal/logger"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var logLevel string

var rootCmd = &cobra.Command{
	Use:   "opnsense-auto-dns",
	Short: "Automatically update DNS records in OPNsense",
	Long: `OPNsense Auto DNS is a CLI tool that automatically updates DNS records in OPNsense unbound.

This tool can detect your current IP address and update DNS records in OPNsense to keep
your hostnames pointing to the correct IP address. It supports multiple configuration
methods including config files, environment variables, and command line flags.

For more information and examples, see the README.md file.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		level, err := log.ParseLevel(logLevel)
		if err != nil {
			level = log.InfoLevel
		}
		logger.Init(level)
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")
}
