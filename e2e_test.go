package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "outenv-e2e-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	binPath = filepath.Join(tmp, "outenv")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build outenv: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

// runOutenv executes the outenv binary and returns stdout, stderr, and exit code.
func runOutenv(t *testing.T, dir string, env []string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(binPath, args...)
	cmd.Dir = dir
	cmd.Env = append([]string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
	}, env...)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run outenv: %v", err)
		}
	}
	return
}

// newTestDirs creates isolated temp directories for a test, resolving symlinks
// to avoid macOS /tmp → /private/tmp mismatches.
func newTestDirs(t *testing.T) (dataHome, configHome, workDir string) {
	t.Helper()
	tmp := t.TempDir()
	tmp, err := filepath.EvalSymlinks(tmp)
	if err != nil {
		t.Fatal(err)
	}
	dataHome = filepath.Join(tmp, "data")
	configHome = filepath.Join(tmp, "config")
	workDir = filepath.Join(tmp, "work")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return
}

// xdgEnv returns XDG environment variables for the given directories.
func xdgEnv(dataHome, configHome string) []string {
	return []string{
		"XDG_DATA_HOME=" + dataHome,
		"XDG_CONFIG_HOME=" + configHome,
	}
}

// setupEnvFile creates an env file at the expected XDG data path for the given directory.
func setupEnvFile(t *testing.T, dataHome, absDir, content string) {
	t.Helper()
	rel := strings.TrimPrefix(absDir, "/")
	envPath := filepath.Join(dataHome, "outenv", rel, "env")
	if err := os.MkdirAll(filepath.Dir(envPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestNoArgs(t *testing.T) {
	dataHome, configHome, workDir := newTestDirs(t)
	stdout, stderr, exitCode := runOutenv(t, workDir, xdgEnv(dataHome, configHome))

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("expected stderr to contain 'Usage:', got: %s", stderr)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got: %s", stdout)
	}
}

func TestUnknownCommand(t *testing.T) {
	dataHome, configHome, workDir := newTestDirs(t)
	_, stderr, exitCode := runOutenv(t, workDir, xdgEnv(dataHome, configHome), "foo")

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "unknown command") {
		t.Errorf("expected stderr to contain 'unknown command', got: %s", stderr)
	}
}

func TestInit(t *testing.T) {
	dataHome, configHome, workDir := newTestDirs(t)
	env := xdgEnv(dataHome, configHome)

	_, _, exitCode := runOutenv(t, workDir, env, "init")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	rel := strings.TrimPrefix(workDir, "/")
	envPath := filepath.Join(dataHome, "outenv", rel, "env")
	if _, err := os.Stat(envPath); err != nil {
		t.Errorf("expected env file at %s: %v", envPath, err)
	}
}

func TestInitIdempotent(t *testing.T) {
	dataHome, configHome, workDir := newTestDirs(t)
	env := xdgEnv(dataHome, configHome)

	_, _, exitCode := runOutenv(t, workDir, env, "init")
	if exitCode != 0 {
		t.Fatalf("first init: expected exit code 0, got %d", exitCode)
	}

	_, stderr, exitCode := runOutenv(t, workDir, env, "init")
	if exitCode != 0 {
		t.Fatalf("second init: expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stderr, "already exists") {
		t.Errorf("expected stderr to contain 'already exists', got: %s", stderr)
	}
}

func TestExportEmpty(t *testing.T) {
	dataHome, configHome, workDir := newTestDirs(t)
	env := xdgEnv(dataHome, configHome)

	stdout, _, exitCode := runOutenv(t, workDir, env, "export")
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got: %s", stdout)
	}
}

func TestExportSimple(t *testing.T) {
	dataHome, configHome, workDir := newTestDirs(t)
	env := xdgEnv(dataHome, configHome)

	setupEnvFile(t, dataHome, workDir, "FOO=bar\n")

	stdout, _, exitCode := runOutenv(t, workDir, env, "export")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "FOO=bar") {
		t.Errorf("expected stdout to contain 'FOO=bar', got: %s", stdout)
	}
}

func TestExportMerge(t *testing.T) {
	dataHome, configHome, workDir := newTestDirs(t)
	parentDir := filepath.Join(workDir, "parent")
	childDir := filepath.Join(parentDir, "child")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatal(err)
	}

	env := xdgEnv(dataHome, configHome)

	setupEnvFile(t, dataHome, parentDir, "A=1\n")
	setupEnvFile(t, dataHome, childDir, "B=2\n")

	stdout, _, exitCode := runOutenv(t, childDir, env, "export")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "A=1") {
		t.Errorf("expected stdout to contain 'A=1', got: %s", stdout)
	}
	if !strings.Contains(stdout, "B=2") {
		t.Errorf("expected stdout to contain 'B=2', got: %s", stdout)
	}
}

func TestExportOverride(t *testing.T) {
	dataHome, configHome, workDir := newTestDirs(t)
	parentDir := filepath.Join(workDir, "parent")
	childDir := filepath.Join(parentDir, "child")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatal(err)
	}

	env := xdgEnv(dataHome, configHome)

	setupEnvFile(t, dataHome, parentDir, "KEY=parent\n")
	setupEnvFile(t, dataHome, childDir, "KEY=child\n")

	stdout, _, exitCode := runOutenv(t, childDir, env, "export")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "KEY=child") {
		t.Errorf("expected child value to override parent, got: %s", stdout)
	}
	if strings.Contains(stdout, "KEY=parent") {
		t.Errorf("parent value should be overridden, got: %s", stdout)
	}
}

func TestEncrypt(t *testing.T) {
	dataHome, configHome, workDir := newTestDirs(t)
	env := xdgEnv(dataHome, configHome)

	stdout, _, exitCode := runOutenv(t, workDir, env, "encrypt", "hello")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if !strings.HasPrefix(strings.TrimSpace(stdout), "ENC:") {
		t.Errorf("expected stdout to start with 'ENC:', got: %s", stdout)
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	dataHome, configHome, workDir := newTestDirs(t)
	env := xdgEnv(dataHome, configHome)

	// Encrypt a value
	stdout, _, exitCode := runOutenv(t, workDir, env, "encrypt", "secret_value")
	if exitCode != 0 {
		t.Fatalf("encrypt: expected exit code 0, got %d", exitCode)
	}
	encrypted := strings.TrimSpace(stdout)

	// Write encrypted value to env file
	setupEnvFile(t, dataHome, workDir, "SECRET="+encrypted+"\n")

	// Export should decrypt it
	stdout, _, exitCode = runOutenv(t, workDir, env, "export")
	if exitCode != 0 {
		t.Fatalf("export: expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "SECRET=secret_value") {
		t.Errorf("expected decrypted value 'SECRET=secret_value', got: %s", stdout)
	}
}

func TestHookBash(t *testing.T) {
	dataHome, configHome, workDir := newTestDirs(t)
	stdout, _, exitCode := runOutenv(t, workDir, xdgEnv(dataHome, configHome), "hook", "bash")

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "PROMPT_COMMAND") {
		t.Errorf("expected bash hook to contain 'PROMPT_COMMAND', got: %s", stdout)
	}
}

func TestHookZsh(t *testing.T) {
	dataHome, configHome, workDir := newTestDirs(t)
	stdout, _, exitCode := runOutenv(t, workDir, xdgEnv(dataHome, configHome), "hook", "zsh")

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "precmd_functions") {
		t.Errorf("expected zsh hook to contain 'precmd_functions', got: %s", stdout)
	}
}

func TestHookFish(t *testing.T) {
	dataHome, configHome, workDir := newTestDirs(t)
	stdout, _, exitCode := runOutenv(t, workDir, xdgEnv(dataHome, configHome), "hook", "fish")

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "--on-variable PWD") {
		t.Errorf("expected fish hook to contain '--on-variable PWD', got: %s", stdout)
	}
}

func TestHookInvalidShell(t *testing.T) {
	dataHome, configHome, workDir := newTestDirs(t)
	_, _, exitCode := runOutenv(t, workDir, xdgEnv(dataHome, configHome), "hook", "invalid")

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestApply(t *testing.T) {
	dataHome, configHome, workDir := newTestDirs(t)
	env := xdgEnv(dataHome, configHome)

	setupEnvFile(t, dataHome, workDir, "FOO=bar\n")

	stdout, _, exitCode := runOutenv(t, workDir, env, "_apply", "bash")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, `export FOO="bar"`) {
		t.Errorf("expected stdout to contain 'export FOO=\"bar\"', got: %s", stdout)
	}
}
