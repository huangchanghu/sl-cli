# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Install

- `make build` — produces `./sl-cli` in repo root (`go build -o sl-cli cmd/sl-cli/main.go`).
- `make install` — builds, generates man pages, copies binary to `/usr/local/bin`, installs shell completion for the detected `$SHELL` (zsh or bash), and provisions default config files (`sl-cli.yaml` + `sl-cli.example.yaml`) into `$HOME/.config/sl-cli/`. Existing config files are **not** overwritten — the target prints a skip message instead. Uses `sudo` for binary/man/completion writes; the config step runs as the invoking user against `$HOME`.
- `make -f Makefile-termux install` — Android/Termux variant; same flow but targets `/data/data/com.termux/files/...` and runs without `sudo`. **Use this on Termux instead of the plain `Makefile`.**
- `make install-config` — config-only provisioning step (can be run standalone if you want to drop in updated config templates without rebuilding the binary).
- `make uninstall` — removes binary, completions, and all `sl-cli*.1` man pages. **Does not touch `$HOME/.config/sl-cli/`** — user config is preserved across reinstalls.
- `make clean` — removes built binary and `./man1/`.
- `go build ./...` / `go vet ./...` — for quick correctness checks during development. There is no test suite yet.
- `./sl-cli config check` — validates the active config file (syntax + per-type field requirements). Run this after editing any YAML.

The `.go-version` file pins Go 1.24; `go.mod` requires `go 1.24.0` with `toolchain go1.24.4`.

## Architecture

`sl-cli` is a Cobra-based CLI with a **hybrid command engine**: some commands are compiled-in Go code, others are loaded dynamically from YAML at startup. Both kinds register against the same `rootCmd` and are indistinguishable to the end user.

### Startup flow (`pkg/cmd/root.go`)

`Execute()` runs three phases in order — each later phase depends on the previous one:
1. `preParseConfigFlag()` — manually scans `os.Args` for `--config <path>` / `--config=<path>` **before** Cobra parses anything. This is needed because dynamic commands have to be registered before `rootCmd.Execute()` runs, but Cobra's normal flag parsing happens inside `Execute()`.
2. `initConfig()` — picks the config file:
   1. `--config <path>` if explicitly passed
   2. Otherwise the canonical `$HOME/.config/sl-cli/sl-cli.yaml`

   There is no fallback to `./sl-cli.yaml` or `$HOME/.sl-cli.yaml` — those legacy locations were removed in favor of a single source of truth provisioned by `make install-config`. If the canonical file is missing, `viper.ConfigFileUsed()` returns empty and `loadDynamicCommands()` silently skips (only built-in Go commands remain).
3. `loadDynamicCommands()` — parses the YAML via `internal/config.LoadConfig`, then builds Cobra commands via `buildCommand`. **Duplicate-name handling:** if a dynamic command's name already exists on `rootCmd` (e.g. the built-in `config`), its subcommands are merged into the existing command rather than overwriting it. This is how the example `sl-cli.yaml` adds a `config show` subcommand alongside the built-in `config init`/`config check`.

### Config loader (`internal/config/loader.go`)

Configs support **`imports:`** — a list of additional YAML files (relative paths resolved against the importing file's directory). Imports are loaded recursively with circular-import detection; later definitions override earlier ones for `vars`, and `commands` from all files are concatenated. Top-level `vars:` is a global string→string map exposed to templates as `{{.vars.KEY}}`; values are themselves expanded via `os.ExpandEnv` and templated with the CLI args (but **not** with other vars, to avoid recursion — see `resolveVars` in the executor).

### Executor (`internal/executor/executor.go`)

`Run()` dispatches on `cfg.Type` to one of three handlers — these are the three valid values; anything else errors out:

- **`http`** — Go templates the URL, headers, and body (`{{index .args 0}}`, `{{.vars.token}}`, plus `os.ExpandEnv` for `$VAR` / `${VAR}`). Sends the request with a spinner. On 2xx, optionally pipes the response body through a chain of `pipes:` (each entry is `command` + `args`); pipes are wired so all commands `Start()` before any `Wait()`, giving true streaming. On non-2xx, prints the body to stdout and returns an error — **the pipe chain is intentionally skipped on errors** so `jq` doesn't choke on HTML error pages.
- **`shell`** — templates `script` and runs it via `/bin/sh -c`. Stdin/stdout/stderr are wired to the terminal, so scripts can prompt interactively.
- **`system`** — runs `command` with `args + os.Args-tail`. Config-provided args come first, user-supplied args are appended.

### Cobra flag parsing quirk

In `buildCommand`, **`DisableFlagParsing` is set to `true` for `shell` and `system` types**. Without this, invoking `sl-cli ls -la` makes Cobra reject `-la` as an unknown shorthand flag. HTTP commands keep flag parsing on (they take positional template args, not flags). When adding new dynamic command types, decide deliberately whether to disable flag parsing.

### Templating contract

Templates receive a single map: `{"args": []string, "vars": map[string]string}`. The canonical positional-arg access is `{{index .args 0}}` (not `{{.args.0}}` — that doesn't work in `text/template`). Env expansion runs **after** template rendering, so a template can output `${VAR}` for later substitution.

## Adding commands

- **Native Go command**: add a file under `pkg/cmd/`, define a `cobra.Command`, register it in `init()` with `rootCmd.AddCommand(...)`. The package shares `rootCmd` across all files. See `pkg/cmd/version.go` for the minimal pattern. Then `make install`.
- **Dynamic command**: edit `~/.config/sl-cli/sl-cli.yaml` (or a file `imports:`'d from it). Run `sl-cli config check` to validate. No rebuild needed.

## Hidden / build-tooling commands

- `sl-cli gen-man [dir]` is `Hidden: true` (not in help output) — it's only used by the Makefile's `gen-man` target to populate `./man1/` before install.
