package debug

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/chia-network/go-chia-libs/pkg/config"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chia-network/chia-tools/cmd"
	"github.com/chia-network/chia-tools/cmd/network"
)

// Define a fixed column width for size
const sizeColumnWidth = 14

// FileInfo stores file path and size
type FileInfo struct {
	Size int64
	Path string
}

// Exclusions - List of patterns to exclude in the default mode
var exclusions = []string{
	`\.DS_Store$`,
	`data_layer/db/server_files_location.*/.*delta.*`, // Don't show delta files by default
	`wallet/db/temp.*`,
	`run/.*`,
}

// debugCmd represents the config command
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Outputs debugging information about Chia",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("# Version Information")
		fmt.Println(strings.Repeat("-", 60)) // Separator
		ShowVersionInfo()

		fmt.Println("\n# Network Information")
		fmt.Println(strings.Repeat("-", 60)) // Separator
		network.ShowNetworkInfo()

		fmt.Println("\n# File Sizes")
		debugFileSizes()
	},
}

// debugFileSizes retrieves the Chia root path and prints sorted file paths with sizes
func debugFileSizes() {
	chiaroot, err := config.GetChiaRootPath()
	if err != nil {
		fmt.Printf("Could not determine CHIA_ROOT: %s\n", err.Error())
		return
	}

	fmt.Println("Scanning:", chiaroot)
	fmt.Printf("%-*s %s\n", sizeColumnWidth, "Size", "File") // Header
	fmt.Println(strings.Repeat("-", 60))                     // Separator

	// Collect files and sort them by size
	files := collectFiles(chiaroot)
	if viper.GetBool("debug-sort") {
		sort.Slice(files, func(i, j int) bool {
			return files[i].Size > files[j].Size // Sort descending
		})
	}

	// Print sorted files
	for _, file := range files {
		fmt.Printf("%-*s %s\n", sizeColumnWidth, humanReadableSize(file.Size), file.Path)
	}
}

// collectFiles recursively collects file paths and sizes
func collectFiles(root string) []FileInfo {
	var files []FileInfo
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, err := os.Stat(path)
			if err == nil {
				relPath, _ := filepath.Rel(root, path)

				// Apply exclusions
				if !viper.GetBool("debug-all-files") && isExcluded(relPath) {
					return nil // Skip this file
				}

				files = append(files, FileInfo{Size: info.Size(), Path: relPath})
			}
		}
		return nil
	})
	if err != nil {
		slogs.Logr.Fatal("error scanning chia root")
	}
	return files
}

// isExcluded checks if a file path matches any exclusion pattern
func isExcluded(path string) bool {
	for _, pattern := range exclusions {
		match, _ := regexp.MatchString(pattern, path)
		if match {
			return true
		}
	}
	return false
}

// humanReadableSize converts bytes into a human-friendly format (KB, MB, GB, etc.)
func humanReadableSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func init() {
	debugCmd.PersistentFlags().Bool("sort", false, "Sort the files largest first")
	debugCmd.PersistentFlags().Bool("all-files", false, "Show all files. By default, some typically small files are excluded from the output")

	cobra.CheckErr(viper.BindPFlag("debug-sort", debugCmd.PersistentFlags().Lookup("sort")))
	cobra.CheckErr(viper.BindPFlag("debug-all-files", debugCmd.PersistentFlags().Lookup("all-files")))

	cmd.RootCmd.AddCommand(debugCmd)
}
