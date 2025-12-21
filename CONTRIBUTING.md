# Contributing to BLC Perf Analyzer

First off, thank you for considering contributing to BLC Perf Analyzer! 🎉

## Code of Conduct

This project adheres to a code of professional conduct. By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples**
- **Describe the behavior you observed and what you expected**
- **Include your environment details** (OS, Go version, etc.)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

- **Use a clear and descriptive title**
- **Provide a step-by-step description of the suggested enhancement**
- **Explain why this enhancement would be useful**
- **Include examples of how it would work**

### Pull Requests

1. **Fork the repo** and create your branch from `main`
2. **Add tests** for any new functionality
3. **Ensure all tests pass**: `make test`
4. **Run the linter**: `make lint` (or `golangci-lint run`)
5. **Format your code**: `make fmt`
6. **Update documentation** if needed
7. **Write a clear commit message**

## Development Setup

### Prerequisites

- Go 1.19 or higher
- Linux environment (for testing perf integration)
- Git

### Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/blc-perf-analyzer.git
cd blc-perf-analyzer

# Install dependencies
make deps

# Run tests
make test

# Build
make build
```

## Code Standards

### Go Style Guide

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Keep functions focused and small
- Write descriptive variable names

### Code Quality Rules

1. **No hardcoded values** - Use constants or configuration
2. **Always handle errors** - No silent failures
3. **Write tests** - Aim for >80% coverage on new code
4. **Document public APIs** - Use Go doc comments
5. **Enterprise grade** - Production-ready code only

### Example

```go
// Good ✓
func ParseSample(input string) (*Sample, error) {
    if input == "" {
        return nil, fmt.Errorf("input cannot be empty")
    }
    
    // Implementation...
    
    return sample, nil
}

// Bad ✗
func parse(s string) *Sample {
    // No error handling, unclear name
    return &Sample{}
}
```

## Testing

### Writing Tests

- Test files should be named `*_test.go`
- Use table-driven tests where appropriate
- Include both positive and negative test cases
- Test edge cases and error conditions

### Running Tests

```bash
# All tests
make test

# With coverage
make coverage

# Specific package
go test -v ./internal/parser

# Benchmarks
make bench
```

### Test Coverage Goals

- New code: >80% coverage
- Critical paths: >90% coverage
- Overall project: >70% coverage

## Project Structure

```
blcperfanalyzer/
├── cmd/blc-perf-analyzer/     # Main entry point
├── internal/                   # Internal packages
│   ├── analysis/              # Report generation
│   ├── capture/               # Perf execution
│   ├── detector/              # System detection
│   ├── heatmap/               # Heatmap generation
│   ├── parser/                # Perf script parsing
│   └── process/               # Process utilities
├── .github/workflows/         # CI/CD
├── docs/                      # Additional documentation
└── examples/                  # Usage examples
```

## Commit Messages

Follow the Conventional Commits specification:

```
type(scope): subject

body

footer
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Formatting, missing semicolons, etc
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance

**Example:**
```
feat(parser): add support for multi-line stack traces

Implement parsing of complex stack traces that span
multiple lines with proper handling of continuation.

Closes #123
```

## Release Process

1. Update CHANGELOG.md
2. Update version in relevant files
3. Create a git tag: `git tag -a v1.x.x -m "Release v1.x.x"`
4. Push tag: `git push origin v1.x.x`
5. GitHub Actions will handle the rest

## Areas for Contribution

### High Priority
- [ ] Off-CPU analysis support
- [ ] Memory profiling integration
- [ ] Container/Docker awareness
- [ ] Additional database-specific analysis

### Medium Priority
- [ ] Web UI for live viewing
- [ ] Comparative analysis (diff mode)
- [ ] Prometheus/InfluxDB exporters
- [ ] Additional output formats

### Documentation
- [ ] Video tutorials
- [ ] More usage examples
- [ ] Performance tuning guides
- [ ] Troubleshooting cookbook

## Getting Help

- **Documentation**: Check the [README](README.md)
- **Issues**: Search existing [GitHub issues](https://github.com/santiagolertora/blc-perf-analyzer/issues)
- **Discussions**: Start a [GitHub discussion](https://github.com/santiagolertora/blc-perf-analyzer/discussions)

## Recognition

Contributors will be recognized in:
- CHANGELOG.md for their contributions
- README.md contributors section (if significant contributions)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to BLC Perf Analyzer!** 🚀

