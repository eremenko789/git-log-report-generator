package model

import "time"

// Commit contains parsed commit metadata from git log.
type Commit struct {
	Hash        string
	ShortHash   string
	AuthorName  string
	AuthorEmail string
	Date        time.Time
	RelDate     string
	Subject     string
	Body        string
	Refs        string
	Files       []FileStat
	Insertions  int
	Deletions   int
}

// FileStat describes a single changed file in a commit.
type FileStat struct {
	Status string
	Path   string
}
