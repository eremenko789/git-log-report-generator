package git

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/p-eremenko/git-log-report-generator/internal/model"
)

var (
	insertionRe = regexp.MustCompile(`(\d+) insertions?\(\+\)`)
	deletionRe  = regexp.MustCompile(`(\d+) deletions?\(-\)`)
)

const prettyFormat = "%H%x1f%h%x1f%an%x1f%ae%x1f%ad%x1f%ar%x1f%s%x1f%b%x1f%D%x1e"

// GetCommits retrieves commits from git and parses them into model objects.
func GetCommits(ctx context.Context, repo, fromRef, toRef, authorFilter string, includeFiles bool) ([]model.Commit, error) {
	args := []string{
		"-C", repo,
		"log",
		"--date=iso-strict",
		fmt.Sprintf("--pretty=format:%s", prettyFormat),
		"--shortstat",
		fmt.Sprintf("%s..%s", fromRef, toRef),
	}
	if authorFilter != "" {
		args = append(args, "--author", authorFilter)
	}

	out, err := runGit(ctx, args...)
	if err != nil {
		return nil, err
	}

	commits, err := parseLogOutput(out)
	if err != nil {
		return nil, err
	}

	if !includeFiles {
		return commits, nil
	}

	for i := range commits {
		files, filesErr := getFileStats(ctx, repo, commits[i].Hash)
		if filesErr != nil {
			return nil, filesErr
		}
		commits[i].Files = files
	}

	return commits, nil
}

func runGit(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git command failed (%s): %w\n%s", strings.Join(args, " "), err, string(out))
	}
	return string(out), nil
}

func parseLogOutput(out string) ([]model.Commit, error) {
	segments := strings.Split(out, "\x1e")
	commits := make([]model.Commit, 0, len(segments))

	for _, segment := range segments {
		chunk := strings.Trim(segment, "\n")
		if chunk == "" {
			continue
		}
		if !strings.Contains(chunk, "\x1f") {
			if len(commits) == 0 {
				continue
			}
			ins, del := parseShortstat(chunk)
			commits[len(commits)-1].Insertions += ins
			commits[len(commits)-1].Deletions += del
			continue
		}

		parts := strings.Split(chunk, "\x1f")
		if len(parts) < 9 {
			return nil, fmt.Errorf("unexpected git log output segment: %q", chunk)
		}
		if idx := strings.LastIndex(parts[0], "\n"); idx >= 0 {
			prefix := strings.TrimSpace(parts[0][:idx])
			if prefix != "" && len(commits) > 0 {
				ins, del := parseShortstat(prefix)
				commits[len(commits)-1].Insertions += ins
				commits[len(commits)-1].Deletions += del
			}
			parts[0] = strings.TrimSpace(parts[0][idx+1:])
		}

		date, err := time.Parse(time.RFC3339, strings.TrimSpace(parts[4]))
		if err != nil {
			return nil, fmt.Errorf("parse commit date: %w", err)
		}

		refs, statsBlob := splitRefsAndStats(parts[8])
		ins, del := parseShortstat(statsBlob)

		commit := model.Commit{
			Hash:        strings.TrimSpace(parts[0]),
			ShortHash:   strings.TrimSpace(parts[1]),
			AuthorName:  strings.TrimSpace(parts[2]),
			AuthorEmail: strings.TrimSpace(parts[3]),
			Date:        date,
			RelDate:     strings.TrimSpace(parts[5]),
			Subject:     strings.TrimSpace(parts[6]),
			Body:        strings.TrimSpace(parts[7]),
			Refs:        strings.TrimSpace(refs),
			Insertions:  ins,
			Deletions:   del,
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

func splitRefsAndStats(raw string) (string, string) {
	lines := strings.Split(raw, "\n")
	if len(lines) == 0 {
		return "", ""
	}

	refs := strings.TrimSpace(lines[0])
	if len(lines) == 1 {
		return refs, ""
	}

	stats := strings.Join(lines[1:], "\n")
	return refs, stats
}

func parseShortstat(stats string) (int, int) {
	ins := extractFirstInt(insertionRe, stats)
	del := extractFirstInt(deletionRe, stats)
	return ins, del
}

func extractFirstInt(re *regexp.Regexp, value string) int {
	match := re.FindStringSubmatch(value)
	if len(match) != 2 {
		return 0
	}
	parsed, err := strconv.Atoi(match[1])
	if err != nil {
		return 0
	}
	return parsed
}

func getFileStats(ctx context.Context, repo, hash string) ([]model.FileStat, error) {
	out, err := runGit(ctx, "-C", repo, "show", "--format=", "--name-status", hash)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	result := make([]model.FileStat, 0, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			continue
		}

		status := string(fields[0][0])
		path := fields[1]
		if (status == "R" || status == "C") && len(fields) >= 3 {
			path = fmt.Sprintf("%s -> %s", fields[1], fields[2])
		}

		result = append(result, model.FileStat{Status: status, Path: path})
	}

	return result, nil
}
