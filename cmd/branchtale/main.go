package main

import (
	"context"
	"fmt"
	"os"

	"github.com/deck/branchtale/internal/config"
	"github.com/deck/branchtale/internal/pr"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	branchPrefix      string
	verbose           bool
	dryRun            bool
	contentGeneration string
)

var rootCmd = &cobra.Command{
	Use:   "branchtale",
	Short: "AI-powered pull request creator",
	Long:  "Branchtale creates GitHub pull requests with AI-generated titles and descriptions based on your code changes.",
	RunE:  runRoot,
}

func main() {
	ctx := context.Background()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		color.New(color.FgRed).Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&branchPrefix, "prefix", "p", "", "Branch name prefix (e.g., 'feature/xyz-123-')")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().StringVarP(&contentGeneration, "content-generation", "c", "local", "Content generation mode (e.g., 'local', 'yandex')")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Enable dry run mode (no changes will be pushed or PR created)")
}

func runRoot(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg.BranchPrefix = branchPrefix
	cfg.Verbose = verbose
	cfg.ContentGeneration = contentGeneration
	cfg.DryRun = dryRun

	if cfg.Verbose {
		color.Green("✓ Configuration loaded successfully")
	}

	service := pr.NewService(cfg)
	return service.Run(ctx)
}
