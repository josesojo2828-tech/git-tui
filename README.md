# git-tui

Historial de git desde la terminal: commits, ramas y diffs, sin salir del shell.

Proyecto de portfolio escrito en Go, sin dependencias externas. La lógica de
parseo está separada de la UI y cubierta por tests, para que la futura TUI se
apoye sobre una base verificable.

## Estado

- [x] Lectura de historial (`log`)
- [x] Ramas locales con estado de upstream (`branches`)
- [x] Diff de un commit (`diff`)
- [ ] TUI interactiva (Bubble Tea)
- [ ] Métricas de salud del repositorio (`/metrics`)

## Requisitos

- Go 1.24+
- git en el `PATH`

## Uso

```bash
# Compilar
make build

# Últimos 20 commits
./bin/git-tui log

# Últimos 50 commits
./bin/git-tui log -n 50

# Ramas con estado respecto del upstream
./bin/git-tui branches

# Patch de un commit
./bin/git-tui diff HEAD~1
```

Salida de `branches`:

```
  RAMA        UPSTREAM            ESTADO        ASUNTO
* main        origin/main         sincronizada  fix: add auto-migration init container
  feature/x   origin/feature/x    +2            wip: something
  fix/y       origin/fix/y        +1 / -3       mixed state
  old         -                   upstream eliminado  obsolete branch
```

## Tests

```bash
make test
```

Los tests cubren los parsers puros (`ParseLog`, `ParseBranches`, `ParseTrack`)
con fixtures de salida real de git, sin invocar al binario.

## Diseño

- **Shell out a `git`** en lugar de una librería: funciona con cualquier
  repositorio y versión de git, y el contrato es la salida del propio comando.
- **Separadores inequívocos** (`\x1f` y `\x1e` en `git log`, `\x00` en
  `for-each-ref`) para que espacios, comas y pipes en los mensajes no rompan el
  parseo.
- **Parsers puros**: reciben `string` y devuelven structs. Se testean sin git.
