package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/p-eremenko/git-log-report-generator/internal/git"
	"github.com/p-eremenko/git-log-report-generator/internal/renderer"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type config struct {
	repo    string
	output  string
	title   string
	author  string
	noFiles bool
	showVer bool
	fromRef string
	toRef   string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	cfg, err := parseArgs(args, stdout)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	if cfg.showVer {
		_, _ = fmt.Fprintf(stdout, "git-html-report version=%s commit=%s date=%s\n", version, commit, date)
		return 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	commits, err := git.GetCommits(ctx, cfg.repo, cfg.fromRef, cfg.toRef, cfg.author, !cfg.noFiles)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error: fetch commits: %v\n", err)
		return 1
	}

	outFile, err := os.Create(cfg.output)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error: create output file: %v\n", err)
		return 1
	}
	defer func() {
		_ = outFile.Close()
	}()

	repoName := filepath.Base(cfg.repo)
	if err := renderer.Render(outFile, cfg.title, repoName, cfg.fromRef, cfg.toRef, time.Now(), commits); err != nil {
		_, _ = fmt.Fprintf(stderr, "error: render report: %v\n", err)
		return 1
	}

	_, _ = fmt.Fprintf(stdout, "report generated: %s\n", cfg.output)
	return 0
}

func parseArgs(args []string, stdout io.Writer) (config, error) {
	cfg := config{}
	fs := flag.NewFlagSet("git-html-report", flag.ContinueOnError)
	fs.SetOutput(stdout)

	fs.StringVar(&cfg.repo, "repo", ".", "Path to git repository")
	fs.StringVar(&cfg.repo, "r", ".", "Path to git repository (shorthand)")
	fs.StringVar(&cfg.output, "output", "git-report.html", "Output HTML file")
	fs.StringVar(&cfg.output, "o", "git-report.html", "Output HTML file (shorthand)")
	fs.StringVar(&cfg.title, "title", "Git Commits Report", "Page title")
	fs.StringVar(&cfg.title, "t", "Git Commits Report", "Page title (shorthand)")
	fs.StringVar(&cfg.author, "author", "", "Author name/email filter")
	fs.BoolVar(&cfg.noFiles, "no-files", false, "Do not include file list")
	fs.BoolVar(&cfg.showVer, "version", false, "Print version and exit")
	fs.BoolVar(&cfg.showVer, "v", false, "Print version and exit (shorthand)")

	flagArgs, positionals, err := splitArgs(args)
	if err != nil {
		return cfg, err
	}

	if err := fs.Parse(flagArgs); err != nil {
		return cfg, err
	}

	if cfg.showVer {
		return cfg, nil
	}

	if len(positionals) != 2 {
		return cfg, fmt.Errorf("expected <from-ref> and <to-ref> positional arguments")
	}
	cfg.fromRef = positionals[0]
	cfg.toRef = positionals[1]

	return cfg, nil
}

func splitArgs(args []string) ([]string, []string, error) {
	withValue := map[string]bool{
		"--repo":   true,
		"-r":       true,
		"--output": true,
		"-o":       true,
		"--title":  true,
		"-t":       true,
		"--author": true,
	}

	var flagArgs []string
	var positionals []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if len(arg) > 0 && arg[0] == '-' {
			flagArgs = append(flagArgs, arg)
			if withValue[arg] && !hasEquals(arg) {
				if i+1 >= len(args) {
					return nil, nil, fmt.Errorf("flag %s requires value", arg)
				}
				i++
				flagArgs = append(flagArgs, args[i])
			}
			continue
		}
		positionals = append(positionals, arg)
	}

	return flagArgs, positionals, nil
}

func hasEquals(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return true
		}
	}
	return false
}
