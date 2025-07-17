package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/adil-chbada/codepack-cli/internal/config"
	"github.com/adil-chbada/codepack-cli/internal/markdown"
	"github.com/adil-chbada/codepack-cli/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	configPath string
	outputDir  string
	verbose    bool
)

// Default config file names to search for (in order of preference)
var defaultConfigFiles = []string{
	"config.codepack.yaml",
	".codepack.yaml",
}

// FileCategory represents different types of files
type FileCategory struct {
	Name     string
	Title    string
	Files    []string
	Size     int64
	Filename string
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate markdown files from project based on config",
	Long: `Scan your project directory and generate three markdown files:
- project-code.md: All code files and main local files
- project-data.md: Data files only (*.data.dart, *.json, /data/**, etc.)
- project-locals.md: All other local/configuration files

The tool respects .gitignore patterns and custom exclude patterns from your config.

If no config file is specified, the tool will automatically search for default
config files in the following order: config.codepack.yaml, .codepack.yaml`,
	Example: `  codepack-cli generate
  codepack-cli generate -c config.yaml
  codepack-cli generate -c flutter-config.yaml -o ./output
  codepack-cli generate --config myproject.yaml --output-dir ./docs`,
	RunE: runGenerate,
}

func init() {
	generateCmd.Flags().StringVarP(&configPath, "config", "c", "", "path to config file (if not specified, searches for default config files)")
	generateCmd.Flags().StringVarP(&outputDir, "output-dir", "o", ".", "output directory for markdown files")
	generateCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Resolve config path
	resolvedConfigPath, err := resolveConfigPath()
	if err != nil {
		return fmt.Errorf("config resolution failed: %w", err)
	}

	if verbose {
		logInfo(fmt.Sprintf("Loading config from: %s", resolvedConfigPath))
	}

	// Load and validate config
	cfg, err := config.LoadConfig(resolvedConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config from %s: %w", resolvedConfigPath, err)
	}

	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Ensure output directory exists
	if err := ensureOutputDir(outputDir); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if verbose {
		logInfo(fmt.Sprintf("Scanning project directory: %s", cfg.ProjectPath))
	}

	// Scan project
	result, err := scanner.Scan(cfg.ProjectPath, cfg)
	if err != nil {
		return fmt.Errorf("failed to scan project directory %s: %w", cfg.ProjectPath, err)
	}

	// Prepare file categories
	categories := []FileCategory{
		{
			Name:     "code",
			Title:    "Project Code Files",
			Files:    result.Code,
			Filename: "project-code.md",
		},
		{
			Name:     "data",
			Title:    "Project Data Files",
			Files:    result.Data,
			Filename: "project-data.md",
		},
		{
			Name:     "locals",
			Title:    "Project Local Files",
			Files:    result.Locals,
			Filename: "project-locals.md",
		},
	}

	// Calculate sizes concurrently
	if err := calculateCategorySizes(categories, cfg.ProjectPath); err != nil {
		return fmt.Errorf("failed to calculate file sizes: %w", err)
	}

	// Write markdown files concurrently
	if err := writeMarkdownFiles(categories, cfg); err != nil {
		return fmt.Errorf("failed to write markdown files: %w", err)
	}

	// Print summary
	printSummary(result, categories)

	return nil
}

// resolveConfigPath resolves the config file path
func resolveConfigPath() (string, error) {
	if configPath != "" {
		// Validate explicitly provided config path
		if _, err := os.Stat(configPath); err != nil {
			return "", fmt.Errorf("config file not found: %s", configPath)
		}
		return configPath, nil
	}

	// Search for default config files
	foundConfig, err := findDefaultConfig()
	if err != nil {
		return "", fmt.Errorf("no config file found. Please specify one with -c flag or create one of: %s",
			strings.Join(defaultConfigFiles, ", "))
	}

	return foundConfig, nil
}

// validateConfig performs basic validation on the loaded config
func validateConfig(cfg *config.Config) error {
	if cfg.ProjectPath == "" {
		return fmt.Errorf("project path is required in config")
	}

	// Check if project path exists
	if _, err := os.Stat(cfg.ProjectPath); err != nil {
		return fmt.Errorf("project path does not exist: %s", cfg.ProjectPath)
	}

	return nil
}

// ensureOutputDir creates the output directory if it doesn't exist
func ensureOutputDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	return nil
}

// calculateCategorySizes calculates file sizes for each category concurrently
func calculateCategorySizes(categories []FileCategory, projectPath string) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstError error

	for i := range categories {
		wg.Add(1)
		go func(cat *FileCategory) {
			defer wg.Done()

			size := calculateTotalSize(cat.Files, projectPath)

			mu.Lock()
			cat.Size = size
			mu.Unlock()
		}(&categories[i])
	}

	wg.Wait()
	return firstError
}

// writeMarkdownFiles writes all markdown files concurrently
func writeMarkdownFiles(categories []FileCategory, cfg *config.Config) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstError error

	for _, category := range categories {
		wg.Add(1)
		go func(cat FileCategory) {
			defer wg.Done()

			outputPath := filepath.Join(outputDir, cat.Filename)

			if verbose {
				logInfo(fmt.Sprintf("Writing %s (%d files)", outputPath, len(cat.Files)))
			}

			if err := markdown.WriteMarkdown(outputPath, cat.Title, cat.Files, cfg); err != nil {
				mu.Lock()
				if firstError == nil {
					firstError = fmt.Errorf("failed to write %s: %w", cat.Filename, err)
				}
				mu.Unlock()
			}
		}(category)
	}

	wg.Wait()
	return firstError
}

// printSummary prints a formatted summary of the generation results
func printSummary(result *scanner.ScanResult, categories []FileCategory) {
	// Calculate totals
	var totalSize int64
	for _, cat := range categories {
		totalSize += cat.Size
	}

	// Print header
	fmt.Printf("\n%s\n", successColor("✓ Generation completed successfully!"))
	fmt.Printf("Total files scanned: %d (%s)\n", result.Total, formatFileSize(totalSize))

	// Sort categories by size for consistent output
	sortedCategories := make([]FileCategory, len(categories))
	copy(sortedCategories, categories)
	sort.Slice(sortedCategories, func(i, j int) bool {
		return len(sortedCategories[i].Files) > len(sortedCategories[j].Files)
	})

	// Print category details
	for i, cat := range sortedCategories {
		var prefix string
		if i == len(sortedCategories)-1 {
			prefix = "└─"
		} else {
			prefix = "├─"
		}

		fmt.Printf("%s %s: %d (%s)\n",
			prefix,
			strings.Title(cat.Name)+" files",
			len(cat.Files),
			formatFileSize(cat.Size))
	}

	if result.Excluded > 0 {
		fmt.Printf("├─ Excluded files: %d\n", result.Excluded)
	}

	fmt.Printf("\nMarkdown files written to: %s\n", outputDir)
}

// formatFileSize formats file size in human readable format
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// calculateTotalSize calculates the total size of a list of files
func calculateTotalSize(files []string, projectPath string) int64 {
	var totalSize int64

	for _, filePath := range files {
		fullPath := filepath.Join(projectPath, filePath)
		if info, err := os.Stat(fullPath); err == nil {
			totalSize += info.Size()
		}
		// Silently skip files that can't be stat'd (might have been deleted)
	}

	return totalSize
}

// findDefaultConfig searches for default config files in the current directory
func findDefaultConfig() (string, error) {
	for _, filename := range defaultConfigFiles {
		if _, err := os.Stat(filename); err == nil {
			return filename, nil
		}
	}
	return "", fmt.Errorf("no default config file found")
}
