package git

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseLog convierte la salida de `git log --pretty=format:...` en commits.
// Es una función pura para poder testearla sin invocar a git.
func ParseLog(data string) ([]Commit, error) {
	records := strings.Split(data, recordSep)
	commits := make([]Commit, 0, len(records))

	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}

		fields := strings.Split(record, fieldSep)
		if len(fields) != 7 {
			return nil, fmt.Errorf("commit con %d campos, se esperaban 7", len(fields))
		}

		date, err := time.Parse(time.RFC3339, fields[4])
		if err != nil {
			return nil, fmt.Errorf("fecha inválida %q: %w", fields[4], err)
		}

		commits = append(commits, Commit{
			Hash:    fields[0],
			Short:   fields[1],
			Author:  fields[2],
			Email:   fields[3],
			Date:    date,
			Refs:    fields[5],
			Subject: fields[6],
		})
	}

	return commits, nil
}

// ParseBranches convierte la salida de `git for-each-ref` en ramas.
func ParseBranches(data string) ([]Branch, error) {
	lines := strings.Split(data, "\n")
	branches := make([]Branch, 0, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Split(line, branchSep)
		if len(fields) != 5 {
			return nil, fmt.Errorf("rama con %d campos, se esperaban 5", len(fields))
		}

		ahead, behind, gone := ParseTrack(fields[2])

		branches = append(branches, Branch{
			Name:     fields[0],
			Upstream: fields[1],
			Ahead:    ahead,
			Behind:   behind,
			Gone:     gone,
			Current:  fields[3] == "*",
			Subject:  fields[4],
		})
	}

	return branches, nil
}

// ParseTrack interpreta el campo %(upstream:track) de git.
// Valores posibles: "", "[gone]", "[ahead 2]", "[behind 3]" o "[ahead 2, behind 3]".
func ParseTrack(track string) (ahead, behind int, gone bool) {
	track = strings.TrimSpace(track)
	if track == "" {
		return 0, 0, false
	}
	if track == "[gone]" {
		return 0, 0, true
	}

	track = strings.TrimSuffix(strings.TrimPrefix(track, "["), "]")

	for _, part := range strings.Split(track, ",") {
		part = strings.TrimSpace(part)

		switch {
		case strings.HasPrefix(part, "ahead "):
			ahead = parseCount(strings.TrimPrefix(part, "ahead "))
		case strings.HasPrefix(part, "behind "):
			behind = parseCount(strings.TrimPrefix(part, "behind "))
		}
	}

	return ahead, behind, false
}

func parseCount(value string) int {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return n
}
