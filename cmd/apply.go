package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/eji/outenv/internal/config"
	"github.com/eji/outenv/internal/envfile"
	"github.com/eji/outenv/internal/shell"
)

func RunApply(shellName string) error {
	sh, err := shell.Get(shellName)
	if err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Compute new env from ancestor files
	files, err := config.AncestorEnvFiles(wd)
	if err != nil {
		return fmt.Errorf("failed to find env files: %w", err)
	}

	var newEnv map[string]string
	if len(files) > 0 {
		newEnv, err = envfile.Merge(files)
		if err != nil {
			return fmt.Errorf("failed to merge env files: %w", err)
		}
	} else {
		newEnv = make(map[string]string)
	}

	// Read previous state
	prevDir := os.Getenv("OUTENV_DIR")
	prevJSON := os.Getenv("OUTENV_PREV")

	// Parse previous env snapshot
	prev := make(map[string]*string) // nil means was unset
	if prevJSON != "" {
		if err := json.Unmarshal([]byte(prevJSON), &prev); err != nil {
			prev = make(map[string]*string)
		}
	}

	// Check if anything changed
	if prevDir == wd && !envChanged(prev, newEnv) {
		return nil
	}

	var output string

	// Restore previous values (unset or restore old value)
	prevKeys := sortedKeys(prev)
	for _, k := range prevKeys {
		if _, inNew := newEnv[k]; inNew {
			continue // will be overwritten by new export
		}
		oldVal := prev[k]
		if oldVal == nil {
			output += sh.Unset(k) + "\n"
		} else {
			output += sh.Export(k, *oldVal) + "\n"
		}
	}

	// Build new prev snapshot
	newPrev := make(map[string]*string)
	newKeys := sortedStringKeys(newEnv)
	for _, k := range newKeys {
		v := newEnv[k]
		output += sh.Export(k, v) + "\n"
		// Record what was there before
		if oldVal, ok := os.LookupEnv(k); ok {
			s := oldVal
			newPrev[k] = &s
		} else {
			newPrev[k] = nil
		}
	}

	// Update OUTENV_DIR and OUTENV_PREV
	newPrevJSON, err := json.Marshal(newPrev)
	if err != nil {
		return fmt.Errorf("failed to marshal prev: %w", err)
	}

	if len(newEnv) > 0 || len(prev) > 0 {
		output += sh.Export("OUTENV_DIR", wd) + "\n"
		output += sh.Export("OUTENV_PREV", string(newPrevJSON)) + "\n"
	} else {
		// No env to manage; clean up tracking vars
		output += sh.Unset("OUTENV_DIR") + "\n"
		output += sh.Unset("OUTENV_PREV") + "\n"
	}

	fmt.Print(output)
	return nil
}

func envChanged(prev map[string]*string, newEnv map[string]string) bool {
	if len(prev) != len(newEnv) {
		return true
	}
	for k := range newEnv {
		if _, ok := prev[k]; !ok {
			return true
		}
	}
	for k := range prev {
		if _, ok := newEnv[k]; !ok {
			return true
		}
	}
	return false
}

func sortedKeys(m map[string]*string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedStringKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
