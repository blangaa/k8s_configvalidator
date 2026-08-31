package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blangaa/configvalidator/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: configvalidator <path-to-file-or-directory>")
		os.Exit(1)
	}

	target := os.Args[1]

	info, err := os.Stat(target)
	if err != nil {
		fmt.Printf("cannot access %s: %v\n", target, err)
		os.Exit(1)
	}

	var files []string
	if info.IsDir() {
		files, err = findManifests(target)
		if err != nil {
			fmt.Printf("failed to walk directory: %v\n", err)
			os.Exit(1)
		}
	} else {
		files = []string{target}
	}

	validCount := 0
	invalidCount := 0
	skippedCount := 0

	for _, file := range files {
		manifest, err := config.LoadManifest(file)
		if err != nil {
			fmt.Printf("⚠️  %s: could not process (%v)\n", file, err)
			skippedCount++
			continue
		}

		if err := manifest.Validate(); err != nil {
			fmt.Printf("❌ %s: invalid (%v)\n", file, err)
			invalidCount++
			continue
		}

		fmt.Printf("✅ %s: %s\n", file, manifest.Summary())
		validCount++
	}

	fmt.Printf("\nSummary: %d valid, %d invalid, %d skipped (%d total)\n",
		validCount, invalidCount, skippedCount, len(files))

	if invalidCount > 0 {
		os.Exit(1)
	}
}

// findManifests walks a directory and returns paths to all .yaml/.yml files.
func findManifests(dir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".yaml" || ext == ".yml" {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}