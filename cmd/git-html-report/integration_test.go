package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGeneratesHTMLReport(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.name", "Test User")
	runGit(t, repo, "config", "user.email", "test@example.com")

	if err := os.WriteFile(filepath.Join(repo, "seed.txt"), []byte("seed\n"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, repo, "add", "seed.txt")
	runGit(t, repo, "commit", "-m", "chore: seed")

	if err := os.WriteFile(filepath.Join(repo, "a.txt"), []byte("hello\n"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, repo, "add", "a.txt")
	runGit(t, repo, "commit", "-m", "feat: first")

	if err := os.WriteFile(filepath.Join(repo, "a.txt"), []byte("hello\nworld\n"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, repo, "add", "a.txt")
	runGit(t, repo, "commit", "-m", "fix: second")

	output := filepath.Join(repo, "report.html")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := run([]string{"HEAD~2", "HEAD", "--repo", repo, "--output", output}, stdout, stderr)
	if exitCode != 0 {
		t.Fatalf("run() failed, code=%d, stderr=%s", exitCode, stderr.String())
	}

	// #nosec G304 -- output path is created from t.TempDir() in this test.
	content, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	html := string(content)
	if !strings.Contains(html, "feat: first") || !strings.Contains(html, "fix: second") {
		t.Fatalf("report does not contain expected commit subjects")
	}
}

func runGit(t *testing.T, repo string, args ...string) {
	t.Helper()
	// #nosec G204 -- test helper intentionally executes git with controlled arguments.
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v (%s)", args, err, string(out))
	}
}
