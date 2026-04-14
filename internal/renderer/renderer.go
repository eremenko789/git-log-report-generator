package renderer

import (
	"embed"
	"fmt"
	"html"
	"html/template"
	"io"
	"sort"
	"time"

	"github.com/p-eremenko/git-log-report-generator/internal/model"
)

//go:embed templates/report.html.tmpl
var templateFS embed.FS

type reportData struct {
	Title       string
	RepoName    string
	FromRef     string
	ToRef       string
	GeneratedAt string
	Summary     summaryData
	Authors     []authorStat
	Commits     []commitView
}

type summaryData struct {
	Commits    int
	Authors    int
	Insertions int
	Deletions  int
}

type authorStat struct {
	Name  string
	Email string
	Count int
}

type commitView struct {
	model.Commit
	SafeBody template.HTML
	Files    []fileView
}

type fileView struct {
	Status string
	Path   template.HTML
}

// Render writes a static HTML report to a writer.
func Render(w io.Writer, title, repoName, fromRef, toRef string, generatedAt time.Time, commits []model.Commit) error {
	tmpl, err := template.ParseFS(templateFS, "templates/report.html.tmpl")
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	data := reportData{
		Title:       title,
		RepoName:    repoName,
		FromRef:     fromRef,
		ToRef:       toRef,
		GeneratedAt: generatedAt.Format(time.RFC3339),
		Summary:     buildSummary(commits),
		Authors:     buildAuthorStats(commits),
		Commits:     buildCommitViews(commits),
	}

	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("render report: %w", err)
	}
	return nil
}

func buildSummary(commits []model.Commit) summaryData {
	authors := make(map[string]struct{})
	summary := summaryData{Commits: len(commits)}
	for _, c := range commits {
		authors[c.AuthorEmail] = struct{}{}
		summary.Insertions += c.Insertions
		summary.Deletions += c.Deletions
	}
	summary.Authors = len(authors)
	return summary
}

func buildAuthorStats(commits []model.Commit) []authorStat {
	counter := make(map[string]*authorStat)
	for _, c := range commits {
		key := c.AuthorName + "|" + c.AuthorEmail
		if _, ok := counter[key]; !ok {
			counter[key] = &authorStat{Name: c.AuthorName, Email: c.AuthorEmail}
		}
		counter[key].Count++
	}

	result := make([]authorStat, 0, len(counter))
	for _, v := range counter {
		result = append(result, *v)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].Name < result[j].Name
		}
		return result[i].Count > result[j].Count
	})
	return result
}

func buildCommitViews(commits []model.Commit) []commitView {
	result := make([]commitView, 0, len(commits))
	for _, c := range commits {
		files := make([]fileView, 0, len(c.Files))
		for _, file := range c.Files {
			files = append(files, fileView{Status: file.Status, Path: template.HTML(html.EscapeString(file.Path))})
		}
		result = append(result, commitView{Commit: c, SafeBody: template.HTML(html.EscapeString(c.Body)), Files: files})
	}
	return result
}
