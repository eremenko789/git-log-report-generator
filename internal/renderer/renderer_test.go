package renderer

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/p-eremenko/git-log-report-generator/internal/model"
)

func TestRenderEscapesBody(t *testing.T) {
	t.Parallel()

	commits := []model.Commit{
		{
			Hash:        "abc",
			ShortHash:   "abc1234",
			AuthorName:  "Alice",
			AuthorEmail: "alice@example.com",
			Date:        time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC),
			RelDate:     "2 days ago",
			Subject:     "feat: xss test",
			Body:        "<script>alert('xss')</script>",
			Files:       []model.FileStat{{Status: "M", Path: "<img src=x onerror=alert(1)>"}},
			Insertions:  1,
			Deletions:   2,
		},
	}

	var out bytes.Buffer
	if err := Render(&out, "Title", "repo", "a", "b", time.Now(), commits); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	html := out.String()
	if strings.Contains(html, "<script>alert('xss')</script>") {
		t.Fatalf("expected escaped body, got raw script tag")
	}
	if !strings.Contains(html, "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;") {
		t.Fatalf("expected escaped script content in output")
	}
}
