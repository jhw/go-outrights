# Claude Code Assistant Configuration

This file provides context and instructions for Claude Code when working on this Go project.

## Project Overview

This is a Go-based sports betting simulation engine that calculates outright market probabilities for football leagues. The project uses Poisson-based team ratings with Dixon-Coles adjustments, genetic algorithm optimization, and Monte Carlo simulation.

## Build and Test Commands

When making changes, always run these commands to ensure code quality:

```bash
go build ./...    # Build all packages
go test ./...     # Run all tests (when available)
```

## File Patterns to Ignore

When analyzing or searching the codebase, ignore these files and patterns:

### Go Build Artifacts
- `go-outrights` (main executable binary)
- `*.exe` (Windows executables)
- `*.so` (shared libraries)
- `*.dylib` (macOS dynamic libraries)
- `*.a` (static libraries)

### Go Module Files
- `go.sum` (dependency checksums - auto-generated)
- `vendor/` (vendored dependencies)

### Development and IDE Files
- `.DS_Store` (macOS filesystem metadata)
- `.vscode/` (VS Code settings)
- `.idea/` (IntelliJ/GoLand settings)
- `*.swp`, `*.swo` (Vim swap files)
- `*~` (backup files)

### Log and Output Files
- `*.log`
- `*.out`
- `debug/`
- `tmp/`

### OS and System Files
- `Thumbs.db` (Windows thumbnails)
- `.env` (environment files with secrets)
- `.env.*` (environment file variants)

## Key Architecture

- **pkg/outrights/**: Main package containing all core functionality
  - `api.go`: Main simulation orchestration
  - `solver.go`: Genetic algorithm optimization
  - `simulator.go`: Monte Carlo simulation
  - `matrix.go`: Score matrix and probability calculations
  - `math.go`: Mathematical utility functions (Poisson, statistics)
  - `state.go`: League table management
  - `markets.go`: Betting market handling
  - `types.go`: Data structures and contracts

- **fixtures/**: Test data and configuration files
  - `ENG1-events.json`: Match results and fixture data
  - `ENG1-markets.json`: Betting market definitions

- **demo.go**: Command-line demo application

## Code Conventions

- Use integer types for points, goal difference, and games played
- Only convert to float64 in expected season points calculations
- Mathematical functions are centralized in `math.go`
- Market payoff strings use format like "1|4x0.25|19x0" for flexible payout definitions

## Common Tasks

- To add new betting markets: Update `ENG1-markets.json`
- To modify team ratings: See `solver.go` genetic algorithm
- To adjust simulation parameters: Update constants in `matrix.go`
- To add mathematical functions: Add to `math.go` module