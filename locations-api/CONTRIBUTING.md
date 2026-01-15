# Contributing to Locations API

First off, thank you for considering contributing to this project! 🎉

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples** (code snippets, API calls, etc.)
- **Describe the behavior you observed and what you expected**
- **Include your environment details** (OS, Go version, etc.)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **A clear and descriptive title**
- **A detailed description of the proposed functionality**
- **Explain why this enhancement would be useful**
- **List any similar features in other projects** (if applicable)

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** with clear, logical commits
3. **Follow Go best practices** and conventions
4. **Add tests** if applicable
5. **Ensure all tests pass** (`go test ./...`)
6. **Format your code** with `go fmt`
7. **Update documentation** as needed (README, comments, etc.)
8. **Write a clear pull request description** explaining your changes

### Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` to format your code
- Write meaningful commit messages
- Add comments for complex logic
- Keep functions focused and concise

### Adding New Locations

If you'd like to add new countries or location data:

1. Update the appropriate SQL files (`countries.sql`, etc.)
2. Test the database migrations
3. Verify the API endpoints return correct data
4. Include documentation for any new fields

### Testing

- Write unit tests for new functionality
- Ensure existing tests pass
- Test API endpoints manually before submitting
- Verify database migrations work correctly

## Development Setup

1. Ensure you have Go 1.24+ installed
2. Clone your fork:
   ```bash
   git clone https://github.com/your-username/locations-api.git
   cd locations-api
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Run database migrations:
   ```bash
   # Follow project-specific migration instructions
   ```
5. Make your changes
6. Run the application:
   ```bash
   go run main.go
   ```

## Commit Message Guidelines

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line

Examples:
```
Add support for city-level location data

Implements new endpoints for querying cities within countries
with population and coordinate information.

Fixes #123
```

## Questions?

Feel free to open an issue with your question or reach out to the maintainers.

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.
