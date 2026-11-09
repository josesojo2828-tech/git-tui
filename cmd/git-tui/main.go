package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/josesojo2828-tech/git-tui/internal/git"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "git-tui: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	command := "help"
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	switch command {
	case "log":
		return runLog(ctx, args)
	case "branches":
		return runBranches(ctx, args)
	case "diff":
		return runDiff(ctx, args)
	case "help", "-h", "--help":
		usage()
		return nil
	default:
		usage()
		return fmt.Errorf("comando desconocido: %q", command)
	}
}

func newRunner(ctx context.Context) (*git.Runner, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	runner := git.NewRunner(dir)
	if _, err := runner.TopLevel(ctx); err != nil {
		return nil, fmt.Errorf("no estás dentro de un repositorio git: %w", err)
	}

	return runner, nil
}

func runLog(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("log", flag.ContinueOnError)
	limit := fs.Int("n", 20, "cantidad de commits a mostrar")
	if err := fs.Parse(args); err != nil {
		return err
	}

	runner, err := newRunner(ctx)
	if err != nil {
		return err
	}

	commits, err := runner.Log(ctx, *limit)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "HASH\tFECHA\tAUTOR\tASUNTO")
	for _, c := range commits {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			c.Short, c.Date.Format("2006-01-02 15:04"), c.Author, c.Subject)
	}

	return w.Flush()
}

func runBranches(ctx context.Context, args []string) error {
	runner, err := newRunner(ctx)
	if err != nil {
		return err
	}

	branches, err := runner.Branches(ctx)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "\tRAMA\tUPSTREAM\tESTADO\tASUNTO")
	for _, b := range branches {
		marker := " "
		if b.Current {
			marker = "*"
		}

		state := "sincronizada"
		switch {
		case b.Gone:
			state = "upstream eliminado"
		case b.Ahead > 0 && b.Behind > 0:
			state = fmt.Sprintf("+%d / -%d", b.Ahead, b.Behind)
		case b.Ahead > 0:
			state = fmt.Sprintf("+%d", b.Ahead)
		case b.Behind > 0:
			state = fmt.Sprintf("-%d", b.Behind)
		case b.Upstream == "":
			state = "sin upstream"
		}

		upstream := b.Upstream
		if upstream == "" {
			upstream = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", marker, b.Name, upstream, state, b.Subject)
	}

	return w.Flush()
}

func runDiff(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("falta la revisión: git-tui diff <hash>")
	}

	runner, err := newRunner(ctx)
	if err != nil {
		return err
	}

	patch, err := runner.Diff(ctx, args[0])
	if err != nil {
		return err
	}

	fmt.Print(patch)
	return nil
}

func usage() {
	fmt.Print(`git-tui — historial de git desde la terminal

Uso:
  git-tui <comando> [opciones]

Comandos:
  log [-n N]     Últimos commits (por defecto 20)
  branches       Ramas locales con estado respecto del upstream
  diff <rev>     Patch completo de un commit
  help           Esta ayuda

Ejemplos:
  git-tui log -n 50
  git-tui branches
  git-tui diff HEAD~1
`)
}
