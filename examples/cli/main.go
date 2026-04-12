// Package main demonstrates a Cobra-based CLI with subcommands, flags,
// Viper config integration, and output formatting via internal/cli helpers.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/example/go-template/internal/cli"
)

var (
	cfgFile  string
	output   string
	noColor  bool
	printer  *cli.Printer
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// rootCmd is the base command. Persistent flags here are inherited by all subcommands.
var rootCmd = &cobra.Command{
	Use:   "myapp",
	Short: "A demo CLI showing Cobra + Viper patterns",
	Long: `myapp demonstrates idiomatic Go CLI development:
  • Cobra subcommands with persistent and local flags
  • Viper config file and environment variable binding
  • Multiple output formats (table, json, yaml)
  • Shell completion (myapp completion bash|zsh|fish|powershell)`,
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		printer = cli.NewPrinter(cli.ParseFormat(output), noColor)
		printer.Out = cmd.OutOrStdout()
	},
}

func init() {
	// Persistent flags — available to all subcommands.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "output format: table, json, yaml")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")

	// Bind persistent flags to Viper so config file/env can set them too.
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output")) // #nosec G104 -- bind failure is non-fatal

	rootCmd.AddCommand(greetCmd, configCmd, listCmd)
}

// --- greet subcommand ---

var (
	greetName     string
	greetGreeting string
	greetCount    int
)

var greetCmd = &cobra.Command{
	Use:   "greet",
	Short: "Print a greeting",
	Long:  "Greet demonstrates local flags and argument validation.",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		for i := 0; i < greetCount; i++ {
			printer.Success(fmt.Sprintf("%s, %s!", greetGreeting, greetName))
		}
		return nil
	},
}

func init() {
	// Local flags — only for this subcommand.
	greetCmd.Flags().StringVarP(&greetName, "name", "n", "World", "who to greet")
	greetCmd.Flags().StringVarP(&greetGreeting, "greeting", "g", "Hello", "greeting word")
	greetCmd.Flags().IntVarP(&greetCount, "count", "c", 1, "number of times to greet")
}

// --- config subcommand ---

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show resolved configuration",
	Long:  "Loads config from file, env vars, and defaults via Viper, then displays it.",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		cfg, err := cli.LoadConfig(cfgFile, "MYAPP")
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		if f := cli.ConfigFileUsed(); f != "" {
			printer.Info("Config file: " + f)
		} else {
			printer.Warn("No config file found — using defaults and env vars")
		}
		return printer.Print(cfg)
	},
}

// --- list subcommand ---

var listCmd = &cobra.Command{
	Use:   "list [filter]",
	Short: "List sample items with optional filter",
	Long:  "Demonstrates table output, JSON/YAML modes, and positional arg validation.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		type item struct {
			Name   string `json:"name" yaml:"name"`
			Status string `json:"status" yaml:"status"`
			Region string `json:"region" yaml:"region"`
		}
		items := []item{
			{"api-gateway", "running", "us-east-1"},
			{"worker", "running", "us-west-2"},
			{"scheduler", "stopped", "eu-west-1"},
			{"cache", "running", "us-east-1"},
		}

		// Apply optional filter argument.
		filter := ""
		if len(args) == 1 {
			filter = strings.ToLower(args[0])
		}

		var filtered []item
		for _, it := range items {
			if filter == "" || strings.Contains(strings.ToLower(it.Name), filter) ||
				strings.Contains(strings.ToLower(it.Status), filter) {
				filtered = append(filtered, it)
			}
		}

		if len(filtered) == 0 {
			printer.Warn("No items match filter: " + filter)
			return nil
		}

		// For structured formats, print the slice directly.
		if printer.Format != cli.FormatTable {
			return printer.Print(filtered)
		}

		// For table format, convert to [][]string.
		rows := [][]string{{"NAME", "STATUS", "REGION"}}
		for _, it := range filtered {
			rows = append(rows, []string{it.Name, it.Status, it.Region})
		}
		return printer.Print(rows)
	},
}
