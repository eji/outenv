package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/eji/outenv/internal/config"
	"github.com/eji/outenv/internal/envfile"
)

func RunExport() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	files, err := config.AncestorEnvFiles(wd)
	if err != nil {
		return fmt.Errorf("failed to find env files: %w", err)
	}

	if len(files) == 0 {
		return nil
	}

	merged, err := envfile.Merge(files)
	if err != nil {
		return fmt.Errorf("failed to merge env files: %w", err)
	}

	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Printf("%s=%s\n", k, merged[k])
	}
	return nil
}
