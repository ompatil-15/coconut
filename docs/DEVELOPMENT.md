# Development Guide

Guide for contributing to Coconut and setting up a development environment.

## Getting Started

### Prerequisites

- Go 1.25.3 or higher
- Git
- Make (optional, but recommended)

### Clone and Build

```bash
# Clone repository
git clone https://github.com/ompatil-15/coconut.git
cd coconut

# Install dependencies
go mod tidy

# Build and install
make build

# Verify installation
coconut --help
```

## Development Workflow

### Building

```bash
# Build and install to $GOPATH/bin
make build

# Or manually
go install
```

### Running

```bash
# Run directly
coconut <command>

# Or without installing
go run main.go <command>
```

### Debugging

#### View Logs

```bash
# Real-time log viewing
make logs

# Or manually
tail -f ~/.coconut/logs/coconut.log
```

#### Database Inspection

```bash
# Dump database contents
make db_dump

# Or manually
go run scripts/db_dump.go
```

#### Clear Database

```bash
# WARNING: Deletes all data
make clear_db

# Or manually
rm ~/.coconut/coconut.db
```



## Adding Features

### Adding a New Command

1. Create new command file

```go
package cmd

import (
    "github.com/ompatil-15/coconut/internal/factory"
    "github.com/spf13/cobra"
)

func NewExportCmd(f *factory.Factory) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "export",
        Short: "Export vault to file",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation
            return nil
        },
    }
    return cmd
}
```

2. Register in root command

```go
func NewRootCmd(f *factory.Factory) *cobra.Command {
    // ... existing code ...
    cmd.AddCommand(NewExportCmd(f))
    return cmd
}
```

### Adding a New Crypto Strategy

1. Implement `CryptoStrategy` interface

```go
package crypto

type ChaCha20Poly1305 struct{}

func NewChaCha20Poly1305() *ChaCha20Poly1305 {
    return &ChaCha20Poly1305{}
}

func (c *ChaCha20Poly1305) Encrypt(key []byte, plaintext string) (string, error) {
    // Implementation
}

func (c *ChaCha20Poly1305) Decrypt(key []byte, ciphertext string) (string, error) {
    // Implementation
}
```

2. Update factory to support selection

```go
func (f *Factory) getCryptoStrategy() crypto.CryptoStrategy {
    // Read from config
    switch config.CryptoAlgorithm {
    case "chacha20poly1305":
        return crypto.NewChaCha20Poly1305()
    default:
        return crypto.NewAESGCM()
    }
}
```

### Adding a New Storage Backend

1. Implement `Repository` interface

```go
package db

type SQLiteRepository struct {
    conn *sql.DB
}

func NewSQLiteRepository(path string) (*SQLiteRepository, error) {
    // Implementation
}

func (s *SQLiteRepository) Get(key string) ([]byte, error) {
    // Implementation
}

func (s *SQLiteRepository) Put(key string, value []byte) error {
    // Implementation
}

// ... other methods
```

2. Update factory initialization

### Adding Configuration Options

1. Update config struct

```go
type Config struct {
    AutoLockSecs    int
    NewOption       int  // New option
}
```

2. Update default values

```go
func Default() *Config {
    return &Config{
        AutoLockSecs: 300,
        NewOption:    15,  // New default
    }
}
```

3. Update config store to persist new option
4. Update config command to handle new option



## Release Process

### Version Numbering

Follow [Semantic Versioning](https://semver.org/):
- MAJOR: Breaking changes
- MINOR: New features (backward compatible)
- PATCH: Bug fixes

### Creating a Release

1. Update version in code
2. Update CHANGELOG.md
3. Create git tag
4. Build binaries
5. Create GitHub release

```bash
# Tag release
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o coconut-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o coconut-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o coconut-windows-amd64.exe
```

## Contributing

### Pull Request Process

1. Fork the repository
2. Create feature branch: `git checkout -b feature/my-feature`
3. Make changes and commit: `git commit -am 'Add feature'`
4. Push to branch: `git push origin feature/my-feature`
5. Open Pull Request

### PR Guidelines

- Clear description of changes
- Include tests for new features
- Update documentation
- Follow existing code style
- One feature per PR

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add export command
fix: resolve memory leak in vault
docs: update security documentation
refactor: simplify crypto strategy selection
test: add integration tests for unlock
```



## Getting Help

- Open an issue on GitHub
- Email: patilom001@gmail.com
- Check existing issues and discussions

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
