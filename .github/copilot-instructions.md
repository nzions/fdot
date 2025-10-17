# Copilot Instructions for Go Development

## Project Context
This is a Go project using Go 1.24.4 with standard project structure:
- `cmd/` - Main applications (executables)
- `pkg/` - Library code for this project
- `go.mod` - Go module definition

## Go Development Guidelines

### Code Style & Standards
- Follow official Go formatting with `gofmt`
- Use `golint` and `go vet` for code quality
- Prefer short, descriptive variable names
- Use camelCase for exported functions/types, lowercase for unexported
- Write self-documenting code with clear function/type names
- Add godoc comments for all exported functions, types, and packages

### Project Structure Best Practices
- Place main packages in `cmd/` subdirectories (e.g., `cmd/netcrawl/`)
- Put reusable library code in `pkg/`
- Use internal packages (`internal/`) for code that should not be imported externally
- Group related functionality in packages with clear, single responsibilities

### Error Handling
- Always handle errors explicitly - never ignore them
- Use the standard `error` interface
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Return errors as the last return value
- Use early returns to reduce nesting

### Testing
- Write tests for all exported functions
- Place tests in `*_test.go` files in the same package
- Use table-driven tests for multiple test cases
- Mock external dependencies using interfaces
- Aim for high test coverage but focus on meaningful tests

### Dependencies & Modules
- Use `go mod tidy` to maintain clean dependencies
- Prefer standard library over external dependencies when possible
- Keep dependencies minimal and well-maintained

### Semantic Versioning Requirements
- ALWAYS follow strict semantic versioning (SemVer 2.0.0)
- Version format: MAJOR.MINOR.PATCH (e.g., v1.2.3)
- MAJOR: Increment for incompatible API changes (breaking changes)
- MINOR: Increment for backward-compatible functionality additions
- PATCH: Increment for backward-compatible bug fixes

#### Version Increment Rules:
- Breaking changes (API removal, signature changes): MAJOR bump
- New features (new functions, methods, types): MINOR bump  
- Bug fixes, security patches, refactoring: PATCH bump
- Documentation-only changes: PATCH bump
- Internal changes with no API impact: PATCH bump


### Performance & Concurrency
- Use goroutines for concurrent operations
- Use channels for communication between goroutines
- Prefer sync.WaitGroup or context.Context for coordination
- Be mindful of memory allocations in hot paths
- Use profiling tools (`go tool pprof`) for performance optimization

### Security
- Validate all inputs, especially from external sources
- Use crypto/rand for cryptographic randomness
- Be careful with file permissions and paths
- Sanitize user inputs to prevent injection attacks

### Documentation
- Write clear package documentation
- Use examples in godoc comments when helpful
- Keep README.md updated with build/run instructions
- Document complex algorithms or business logic

### Build & Deployment
- ALWAYS compile binaries to `<project root>/bin/` directory using `-o bin/<binary-name>`
- Use `go build` for production builds
- Set appropriate build flags for releases: `-ldflags="-s -w"`
- Use `go generate` for code generation when needed
- Consider using Docker for containerized deployments

### Code Generation Preferences
- Generate idiomatic Go code following these standards
- Include appropriate error handling in all generated code
- Add TODO comments for areas needing human review
- Use meaningful variable names even in short code snippets
- Include necessary imports and package declarations

### File Creation & Validation (CRITICAL)
When creating new Go files:
- NEVER duplicate the `package` declaration - it must appear EXACTLY ONCE at the top
- File structure must be: package declaration → imports → code
- IMMEDIATELY validate with `go build` or check for errors after file creation
- Use `gofmt` to verify syntax correctness
- Example of CORRECT file structure:
  ```go
  package mypackage
  
  import (
      "fmt"
  )
  
  func MyFunc() { }
  ```
- WRONG (duplicate package):
  ```go
  package mypackage  // WRONG - duplicate!
  package mypackage
  
  import (
  ```

### Common Patterns to Use
- Interface-based design for testability
- Constructor functions for complex types
- Options pattern for configuration
- Context passing for cancellation and timeouts
- Structured logging with levels

### Avoid These Patterns
- Global variables (except for configuration)
- Panic in library code (use errors instead)
- Empty catch blocks
- Magic numbers (use named constants)
- Deep nesting (prefer early returns)

# CRITICAL section
- ALWAYS DRY, KISS, YAGNI
- ALWAYS prioritize readability and maintainability
- ALWAYS use latest go version features
- ALWAYS use go 1.22+ web mux
- ALWAYS use any instead of interface{}
- ALWAYS compile binaries to `<project root>/bin/` directory
- ALWAYS follow strict semantic versioning (SemVer 2.0.0)
- ALWAYS validate new Go files with `go build` or `get_errors` tool immediately after creation
- NEVER duplicate the `package` declaration in Go files (it must appear EXACTLY ONCE)
- NEVER use deprecated or outdated libraries