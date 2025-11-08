# Agent Instructions for rakuda

This file provides instructions and guidelines for AI agents (like Jules/GitHub Copilot) working on the rakuda project.

## Commands

- **test** -- `go test ./...`
- **format** -- `goimports -w .` (formats code and removes unused imports)
- **build** -- `go build -o /tmp/rakuda` (always build to /tmp, not to the repo)

## Project Information

- **repo** -- github.com/podhmo/rakuda
- **description** -- Type-safe HTTP router for Go with compile-time lifecycle enforcement
- **main branch** -- main (note: may need to be created)
- **current status** -- Early development, documentation and design phase

## Tool Stack

- **Go version** -- go1.24+ (required for native path parameter support)
- **Standard library** -- net/http only
- **Testing** -- Standard `testing` package, `github.com/google/go-cmp/cmp` for comparisons
- **Logging** -- log/slog (with context-aware methods)

## Prohibited Tools and Practices

- **testify** -- Prohibited. Use standard `testing` package and `go-cmp` instead
- **log package** -- Prohibited. Use `log/slog` instead
- **Committing binaries** -- Do not commit binaries created by `go build`
- **Committing temp files** -- Do not commit temporary debug files

## Required Practices

### Code Quality
- Format code with `goimports` before committing
- Use `github.com/google/go-cmp/cmp` for test comparisons
- Use `log/slog` for logging with context-aware methods (e.g., `DebugContext()`)
- Run tests after every change: `go test ./...`

### Documentation
- Write all documentation in `docs/*.md` in English
- Write all commit messages in English
- Keep TODO.md updated following the format in sketch/prompts.md
- Reference design documents when making architectural decisions

### Testing
- Use table-driven tests for comprehensive coverage
- Test both success and error cases
- Do not use `testify` assertions
- Build to `/tmp` directory, not in the repository

## Communication Guidelines

When responding to users:
- Think and write code in English
- Write documentation and comments in English
- Write commit messages in English
- Respond to user input in Japanese (ユーザーへの返答は日本語で)

## TODO.md Management

Follow the guidelines in [sketch/prompts.md](./sketch/prompts.md) for:
- Updating TODO.md from plan documents
- Finalizing and refactoring TODO.md
- Creating continuation documents for incomplete work

The TODO.md format follows these rules:
- Do not move individual tasks to "Implemented"
- Move entire feature sections only when all sub-tasks are complete
- Use checkboxes (`[x]` for complete, `[ ]` for incomplete)
- For partially complete features, mark individual sub-tasks with `[x]`

## Development Philosophy

The rakuda project follows these core principles:

1. **Type Safety First**: Use the Go type system to prevent errors at compile-time
2. **Predictable Lifecycle**: Clear separation between configuration and execution
3. **Standard Library First**: Prefer `net/http` over custom implementations
4. **Fail Fast**: Catch errors at compile-time or startup, not at runtime
5. **No Magic**: Explicit, clear APIs with predictable behavior

Refer to [docs/router-design.md](./docs/router-design.md) for detailed architectural decisions.

## Project Structure

```
rakuda/
├── README.md           # Project overview and quick start
├── TODO.md            # Implementation status and roadmap
├── AGENTS.md          # This file - agent instructions
├── LICENSE            # MIT License
├── docs/              # Design documents
│   └── router-design.md  # Core architecture and design decisions
└── sketch/            # AI agent files and planning documents
    ├── README.md      # About sketch directory
    └── prompts.md     # Development prompts for Jules
```
