package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	version = "1.0.0"

	// Color functions
	successColor = color.New(color.FgGreen).SprintFunc()
	errorColor   = color.New(color.FgRed).SprintFunc()
	warnColor    = color.New(color.FgYellow).SprintFunc()
	infoColor    = color.New(color.FgCyan).SprintFunc()
)

var rootCmd = &cobra.Command{
	Use:   "codepack-cli",
	Short: "A CLI tool to extract and categorize project files into markdown",
	Long: `CodePack CLI is a powerful tool that scans your project directory,
automatically categorizing files into code, data, and configuration files.
It then generates well-structured markdown files that are easy to share
with AI assistants like Claude, GPT-4, and other LLMs.`,
	Example: `  codepack-cli init flutter -o flutter-config.yaml
  codepack-cli generate -c flutter-config.yaml
  codepack-cli generate --verbose`,
	Version: "1.2.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(completionCmd)
}

// Logging helpers
func logInfo(msg string) {
	if verbose {
		fmt.Fprintf(os.Stderr, "%s %s\n", infoColor("[INFO]"), msg)
	}
}

func logSuccess(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", successColor("[SUCCESS]"), msg)
}

func logWarn(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", warnColor("[WARN]"), msg)
}

func logError(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", errorColor("[ERROR]"), msg)
}
