package git

import "testing"

const logFixture = "f2b9ade92ead5bd272f9024cd8701b8e18e9cc71\x1ff2b9ade\x1fJose Sojo\x1fjosesojo2828@gmail.com\x1f2026-07-15T06:57:12-04:00\x1fHEAD -> main, origin/main, origin/HEAD\x1ffix: add auto-migration init container, remove deprecated version\x1e\n" +
	"f7ad0dc8be9df1e02bb282c58ea1dba4dbceef5f\x1ff7ad0dc\x1fJose Sojo\x1fjosesojo2828@gmail.com\x1f2026-07-14T19:10:10-04:00\x1f\x1ffix: modify compose\x1e\n"

func TestParseLog(t *testing.T) {
	commits, err := ParseLog(logFixture)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if len(commits) != 2 {
		t.Fatalf("se esperaban 2 commits, se obtuvieron %d", len(commits))
	}

	first := commits[0]
	if first.Hash != "f2b9ade92ead5bd272f9024cd8701b8e18e9cc71" {
		t.Errorf("hash = %q", first.Hash)
	}
	if first.Short != "f2b9ade" {
		t.Errorf("short = %q", first.Short)
	}
	if first.Author != "Jose Sojo" {
		t.Errorf("author = %q", first.Author)
	}
	if first.Refs != "HEAD -> main, origin/main, origin/HEAD" {
		t.Errorf("refs = %q", first.Refs)
	}
	if first.Subject != "fix: add auto-migration init container, remove deprecated version" {
		t.Errorf("subject = %q", first.Subject)
	}
	if first.Date.Year() != 2026 || first.Date.Month() != 7 || first.Date.Day() != 15 {
		t.Errorf("date = %s", first.Date)
	}

	second := commits[1]
	if second.Refs != "" {
		t.Errorf("refs del segundo commit = %q, se esperaba vacío", second.Refs)
	}
	if second.Subject != "fix: modify compose" {
		t.Errorf("subject del segundo commit = %q", second.Subject)
	}
}

func TestParseLogEmpty(t *testing.T) {
	commits, err := ParseLog("")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(commits) != 0 {
		t.Fatalf("se esperaban 0 commits, se obtuvieron %d", len(commits))
	}
}

func TestParseLogInvalidFields(t *testing.T) {
	if _, err := ParseLog("hash\x1fshort\x1fauthor\x1femail\x1e\n"); err == nil {
		t.Fatal("se esperaba error por cantidad de campos")
	}
}

func TestParseLogInvalidDate(t *testing.T) {
	record := "hash\x1fshort\x1fauthor\x1femail\x1fno-es-fecha\x1f\x1fsubject\x1e\n"
	if _, err := ParseLog(record); err == nil {
		t.Fatal("se esperaba error por fecha inválida")
	}
}

const branchFixture = "main\x00origin/main\x00\x00*\x00fix: add auto-migration init container\n" +
	"feature/x\x00origin/feature/x\x00[ahead 2]\x00 \x00wip: something with | pipe\n" +
	"fix/y\x00origin/fix/y\x00[behind 3, ahead 1]\x00 \x00mixed state\n" +
	"old\x00\x00[gone]\x00 \x00obsolete branch\n" +
	"local\x00\x00\x00 \x00no upstream\n"

func TestParseBranches(t *testing.T) {
	branches, err := ParseBranches(branchFixture)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if len(branches) != 5 {
		t.Fatalf("se esperaban 5 ramas, se obtuvieron %d", len(branches))
	}

	main := branches[0]
	if main.Name != "main" || !main.Current || main.Upstream != "origin/main" {
		t.Errorf("rama main mal parseada: %+v", main)
	}
	if main.Ahead != 0 || main.Behind != 0 || main.Gone {
		t.Errorf("track de main mal parseado: %+v", main)
	}

	ahead := branches[1]
	if ahead.Name != "feature/x" || ahead.Ahead != 2 || ahead.Behind != 0 || ahead.Current {
		t.Errorf("rama ahead mal parseada: %+v", ahead)
	}
	if ahead.Subject != "wip: something with | pipe" {
		t.Errorf("subject con pipe mal parseado: %q", ahead.Subject)
	}

	mixed := branches[2]
	if mixed.Ahead != 1 || mixed.Behind != 3 {
		t.Errorf("rama divergida mal parseada: %+v", mixed)
	}

	gone := branches[3]
	if !gone.Gone || gone.Upstream != "" {
		t.Errorf("rama gone mal parseada: %+v", gone)
	}

	local := branches[4]
	if local.Name != "local" || local.Upstream != "" || local.Ahead != 0 || local.Behind != 0 {
		t.Errorf("rama sin upstream mal parseada: %+v", local)
	}
}

func TestParseBranchesInvalidFields(t *testing.T) {
	if _, err := ParseBranches("main\x00origin/main\x00\n"); err == nil {
		t.Fatal("se esperaba error por cantidad de campos")
	}
}

func TestParseTrack(t *testing.T) {
	tests := []struct {
		name       string
		track      string
		wantAhead  int
		wantBehind int
		wantGone   bool
	}{
		{"vacío", "", 0, 0, false},
		{"gone", "[gone]", 0, 0, true},
		{"ahead", "[ahead 2]", 2, 0, false},
		{"behind", "[behind 3]", 0, 3, false},
		{"ambos", "[ahead 2, behind 1]", 2, 1, false},
		{"orden invertido", "[behind 10, ahead 4]", 4, 10, false},
		{"con espacios", "  [ahead 1]  ", 1, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ahead, behind, gone := ParseTrack(tt.track)
			if ahead != tt.wantAhead || behind != tt.wantBehind || gone != tt.wantGone {
				t.Errorf("ParseTrack(%q) = (%d, %d, %t), se esperaba (%d, %d, %t)",
					tt.track, ahead, behind, gone, tt.wantAhead, tt.wantBehind, tt.wantGone)
			}
		})
	}
}
