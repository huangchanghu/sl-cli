# Gemini Development Context - sl-cli

## 🚀 Project Overview
`sl-cli` (Super Link CLI) is a highly extensible modern command-line tool built with Go. It features a hybrid engine that supports:
- **Native Go Commands**: High-performance core logic.
- **Dynamic YAML Configuration**: Instantly mount HTTP APIs, Shell scripts, or System command aliases.

## 🛠 Tech Stack
- **Language**: Go 1.24.0+
- **Frameworks**: 
  - [Cobra](https://github.com/spf13/cobra): CLI application framework.
  - [Viper](https://github.com/spf13/viper): Configuration management.
  - [Spinner](https://github.com/briandowns/spinner): Interactive loading indicators.

## 📂 Key Directory Structure
- `cmd/sl-cli/main.go`: Application entry point.
- `internal/config/`: Configuration loading logic (supports splitting into multiple files).
- `internal/executor/`: Execution engine for different command types (HTTP, Shell, System, Pipe).
- `pkg/cmd/`: Implementation of native Go commands (e.g., `version`, `calc`, `config`).
- `Makefile`: Standard build/install scripts for macOS/Linux.
- `Makefile-termux`: Optimized build/install scripts for Android/Termux environments.

## 🛠 Standard Commands
- **Build**: `make build`
- **Install (Standard)**: `make install` (Requires sudo)
- **Install (Termux)**: `make -f Makefile-termux install` (No sudo)
- **Generate Man Pages**: `make gen-man`
- **Check Config**: `./sl-cli config check`

## 🏗 Development Guidelines
### 1. Adding Native Go Commands
- Create a new `.go` file in `pkg/cmd/`.
- Define a `cobra.Command` and register it in the `init()` function using `rootCmd.AddCommand()`.
- Run `make build` to verify.

### 2. Extending Dynamic Commands
- Modify `sl-cli.yaml` or add sub-configuration files in `~/.config/sl-cli/`.
- Use Go Template syntax (e.g., `{{index .args 0}}`) for dynamic arguments.
- Use `{{.vars.name}}` to reference global variables.

### 3. Execution Flow
- `rootCmd` initializes the config loader.
- Config loader merges variables and commands from all YAML files.
- Command execution is dispatched to the corresponding executor in `internal/executor/`.

## ✅ Validation Checklist
- [ ] Code follows standard Go idioms and `go fmt`/`go vet`.
- [ ] New features are verified with manual or automated tests.
- [ ] For CLI changes, verify `help` documentation and auto-completion scripts.
- [ ] In Termux, always use `Makefile-termux` for installation.
