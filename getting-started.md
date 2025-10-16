# Getting Started with fdot

This guide will help you set up your development environment for the fdot Go project.

## Prerequisites

- Go 1.24.4 or later
- Git with SSH keys configured for GitHub
- Linux/macOS/Windows development environment

## Project Structure

```
fdot/
├── cmd/                    # Main applications (executables)
│   ├── credmgr/           # Credential manager application
│   └── netcrawl/          # Network crawler application
├── pkg/                   # Library code for this project
│   └── fdh/               # Fdot helper packages
│       ├── credmgr/       # Credential management
│       └── user/          # User utilities
├── bin/                   # Compiled binaries (created during build)
├── go.mod                 # Go module definition
└── Makefile              # Build automation
```

## Development Environment Setup

### 1. SSH Configuration for Go Modules

To use SSH keys for Git operations with Go modules (especially for private repositories):

```bash
# Configure Git to use SSH for GitHub URLs
git config --global url."git@github.com:".insteadOf "https://github.com/"

# Set GOPRIVATE for your repositories
export GOPRIVATE="github.com/nzions/*"

# Make it persistent (add to ~/.bashrc or ~/.zshrc)
echo 'export GOPRIVATE="github.com/nzions/*"' >> ~/.bashrc
```

### 2. Verify SSH Configuration

```bash
# Check Git URL rewriting
git config --global --get-regexp url

# Test SSH connection to GitHub
ssh -T git@github.com

# Test Go module fetching
go get -u github.com/nzions/dsjdb
```

### 3. SSH Key Setup (if needed)

If you don't have SSH keys set up:

```bash
# Generate SSH key
ssh-keygen -t ed25519 -C "your.email@example.com"

# Add to SSH agent
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_ed25519

# Copy public key to clipboard (Linux)
cat ~/.ssh/id_ed25519.pub | xclip -selection clipboard

# Add the public key to your GitHub account at:
# https://github.com/settings/ssh/new
```

## Building the Project

### Using Makefile

```bash
# Build all applications
make

# Build specific application
make credmgr
make netcrawl

# Clean build artifacts
make clean
```

### Manual Build Commands

All binaries are compiled to the `bin/` directory:

```bash
# Build credmgr
go build -o bin/credmgr ./cmd/credmgr

# Build netcrawl
go build -o bin/netcrawl ./cmd/netcrawl

# Build with release flags
go build -ldflags="-s -w" -o bin/credmgr ./cmd/credmgr
```

## Running Applications

### Credential Manager

```bash
# Run from source
go run ./cmd/credmgr

# Run compiled binary
./bin/credmgr
```

### Network Crawler

```bash
# Run from source
go run ./cmd/netcrawl

# Run compiled binary
./bin/netcrawl
```

## Development Workflow

### 1. Install Dependencies

```bash
# Download and install dependencies
go mod download

# Clean up unused dependencies
go mod tidy
```

### 2. Code Quality

```bash
# Format code
go fmt ./...

# Run linter
golint ./...

# Vet code
go vet ./...
```

### 3. Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/fdh/credmgr
```

### 4. Git Workflow

```bash
# Check status
git status

# Add changes
git add .

# Commit changes
git commit -m "Your commit message"

# Push changes
git push origin master
```

## Common Issues and Solutions

### SSH Authentication Issues

If you encounter authentication issues:

1. Verify SSH key is added to GitHub account
2. Test SSH connection: `ssh -T git@github.com`
3. Check SSH agent is running: `ssh-add -l`
4. Ensure GOPRIVATE is set correctly

### Module Download Issues

If `go get` fails:

1. Check network connectivity
2. Verify repository exists and you have access
3. Clear module cache: `go clean -modcache`
4. Try with verbose output: `go get -v`

### Build Issues

If builds fail:

1. Ensure Go version is 1.24.4+: `go version`
2. Clean and rebuild: `make clean && make`
3. Check for missing dependencies: `go mod download`

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Git SSH Setup](https://docs.github.com/en/authentication/connecting-to-github-with-ssh)
- [Go Modules Reference](https://golang.org/ref/mod)

## Project-Specific Guidelines

See `.github/copilot-instructions.md` for detailed development guidelines and coding standards specific to this project.