package git

import (
	"strings"
	"testing"
	"time"
)

func TestParseLogOutput(t *testing.T) {
	t.Parallel()

	payload := strings.Join([]string{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\x1faaaaaaa\x1fAlice\x1falice@example.com\x1f2026-04-01T10:00:00+00:00\x1f2 weeks ago\x1ffeat: add parser\x1fBody line 1\nBody line 2\x1fHEAD -> main\n 2 files changed, 12 insertions(+), 3 deletions(-)\x1e",
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\x1fbbbbbbb\x1fBob\x1fbob@example.com\x1f2026-04-02T10:00:00+00:00\x1f13 days ago\x1ffix: empty body\x1f\x1f\n 1 file changed, 5 insertions(+)\x1e",
	}, "")

	commits, err := parseLogOutput(payload)
	if err != nil {
		t.Fatalf("parseLogOutput() error = %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}

	if commits[0].Insertions != 12 || commits[0].Deletions != 3 {
		t.Fatalf("unexpected shortstat for commit 0: %+v", commits[0])
	}
	if commits[1].Deletions != 0 {
		t.Fatalf("unexpected deletions for commit 1: %+v", commits[1])
	}
	if commits[1].Body != "" {
		t.Fatalf("expected empty body, got %q", commits[1].Body)
	}
	if commits[0].Date.Format(time.RFC3339) != "2026-04-01T10:00:00Z" {
		t.Fatalf("unexpected date: %s", commits[0].Date.Format(time.RFC3339))
	}
}
