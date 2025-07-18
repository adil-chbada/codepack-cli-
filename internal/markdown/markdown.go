package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adil-chbada/codepack-cli/internal/config"
)

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

// getFileSize returns the size of a file, or -1 if error
func getFileSize(filePath string) int64 {
	info, err := os.Stat(filePath)
	if err != nil {
		return -1
	}
	return info.Size()
}

// WriteMarkdown writes a list of files to a markdown file
func WriteMarkdown(outputPath, title string, files []string, cfg *config.Config) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create markdown file: %w", err)
	}
	defer file.Close()

	// Sort files for consistent output
	sort.Strings(files)

	// Write header
	fmt.Fprintf(file, "# %s\n\n", title)

	// Calculate total size
	totalSize := int64(0)
	for _, filePath := range files {
		fileSize := getFileSize(filepath.Join(cfg.ProjectPath, filePath))
		if fileSize >= 0 {
			totalSize += fileSize
		}
	}

	// Write metadata
	fmt.Fprintf(file, "**Project:** %s  \n", getProjectName(cfg))
	fmt.Fprintf(file, "**Generated:** %s  \n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "**Total Files:** %d  \n", len(files))
	fmt.Fprintf(file, "**Total Size:** %s  \n\n", formatFileSize(totalSize))

	if len(files) == 0 {
		fmt.Fprintf(file, "*No files found matching the criteria.*\n")
		return nil
	}

	// Write table of contents
	writeFileTOC(file, files, cfg.ProjectPath)

	// Write file content
	for _, filePath := range files {
		filename := filepath.Base(filePath)
		ext := filepath.Ext(filename)
		content, err := os.ReadFile(filepath.Join(cfg.ProjectPath, filePath))
		if err == nil {
			fmt.Fprintf(file, "### `%s`\n\n", filePath)
			fmt.Fprintf(file, "```%s\n%s\n```\n\n", strings.TrimPrefix(ext, "."), content)
		}
	}

	// Write summary footer
	fmt.Fprintf(file, "---\n\n")
	fmt.Fprintf(file, "**Summary:**\n")
	fmt.Fprintf(file, "- Total files listed: %d\n", len(files))
	fmt.Fprintf(file, "- Total size: %s\n", formatFileSize(totalSize))

	fmt.Fprintf(file, "- Generated by codepack-cli\n")

	return nil
}

// writeFileTOC writes a table of contents with file paths and sizes
func writeFileTOC(file *os.File, files []string, projectPath string) {
	fmt.Fprintf(file, "## Table of Contents\n\n")
	fmt.Fprintf(file, "| File | Size |\n")
	fmt.Fprintf(file, "| --- | --- |\n")

	for _, filePath := range files {
		fileSize := getFileSize(filepath.Join(projectPath, filePath))
		sizeStr := "unknown"
		if fileSize >= 0 {
			sizeStr = formatFileSize(fileSize)
		}
		fmt.Fprintf(file, "| `%s` | %s |\n", filePath, sizeStr)
	}

	fmt.Fprintf(file, "\n")
}

// getProjectName extracts project name from config or path

func getProjectName(cfg *config.Config) string {
	if cfg.ProjectName != "" {
		return cfg.ProjectName
	}

	// Extract from project path
	return filepath.Base(cfg.ProjectPath)
}
