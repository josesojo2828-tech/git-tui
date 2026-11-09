package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Separadores usados para parsear la salida de git de forma inequívoca.
const (
	fieldSep  = "\x1f" // unit separator (git log)
	recordSep = "\x1e" // record separator (git log)
	branchSep = "\x00" // NUL (git for-each-ref)
)

// Commit representa un commit del historial.
type Commit struct {
	Hash    string
	Short   string
	Author  string
	Email   string
	Date    time.Time
	Refs    string
	Subject string
}

// Branch representa una rama local con su estado respecto del upstream.
type Branch struct {
	Name     string
	Upstream string
	Ahead    int
	Behind   int
	Gone     bool
	Current  bool
	Subject  string
}

// Runner ejecuta comandos git dentro de un repositorio.
type Runner struct {
	dir string
}

// NewRunner crea un Runner que opera en el directorio indicado.
func NewRunner(dir string) *Runner {
	return &Runner{dir: dir}
}

// Dir devuelve el directorio de trabajo del runner.
func (r *Runner) Dir() string {
	return r.dir
}

func (r *Runner) run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.dir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}

	return string(out), nil
}

// TopLevel devuelve la raíz del repositorio que contiene al directorio del runner.
func (r *Runner) TopLevel(ctx context.Context) (string, error) {
	out, err := r.run(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// CurrentBranch devuelve el nombre de la rama actual.
func (r *Runner) CurrentBranch(ctx context.Context) (string, error) {
	out, err := r.run(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Log devuelve los últimos commits del historial.
func (r *Runner) Log(ctx context.Context, limit int) ([]Commit, error) {
	if limit <= 0 {
		limit = 20
	}

	format := strings.Join([]string{
		"%H", "%h", "%an", "%ae", "%aI", "%D", "%s",
	}, fieldSep) + recordSep

	out, err := r.run(ctx, "log", "--pretty=format:"+format, "-n", strconv.Itoa(limit))
	if err != nil {
		return nil, err
	}

	return ParseLog(out)
}

// Branches devuelve las ramas locales ordenadas por fecha del último commit.
func (r *Runner) Branches(ctx context.Context) ([]Branch, error) {
	format := strings.Join([]string{
		"%(refname:short)", "%(upstream:short)", "%(upstream:track)", "%(HEAD)", "%(contents:subject)",
	}, "%00")

	out, err := r.run(ctx, "for-each-ref", "--sort=-committerdate",
		"--format="+format, "refs/heads")
	if err != nil {
		return nil, err
	}

	return ParseBranches(out)
}

// Diff devuelve el patch completo de una revisión.
func (r *Runner) Diff(ctx context.Context, rev string) (string, error) {
	if strings.TrimSpace(rev) == "" {
		return "", fmt.Errorf("revisión vacía")
	}
	return r.run(ctx, "show", "--no-color", "--format=medium", rev)
}
