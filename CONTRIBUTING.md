# Contributing to Farshore AI

Thank you for your interest in contributing to Farshore AI! We welcome contributions from the community and are grateful for your help in improving this project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Documentation](#documentation)
- [Questions and Help](#questions-and-help)

## Code of Conduct

By participating in this project, you are expected to uphold our Code of Conduct:
- Be respectful and inclusive
- Exercise consideration and respect in your speech and actions
- Attempt collaboration before conflict
- Refrain from demeaning, discriminatory, or harassing behavior

## Getting Started

### Prerequisites

- Go 1.19+
- Node.js 16+
- Redis 6+
- MongoDB 4.4+

### Setting Up Development Environment

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/memory-remember.git
   cd memory-remember
   ```

3. Set up the development environment:
   ```bash
   # Backend setup
   cd remember
   go mod download
   
   # Frontend setup
   cd ../remember-web
   npm install
   ```

4. Create a configuration file:
   ```bash
   cp remember/config.yaml.example remember/config.yaml
   # Edit the configuration file with your settings
   ```

## Development Workflow

### Branch Naming

Use descriptive branch names:
- `feature/description` - for new features
- `bugfix/description` - for bug fixes
- `docs/description` - for documentation changes
- `refactor/description` - for code refactoring

### Commit Messages

Follow conventional commit format:
```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: new feature
- `fix`: bug fix
- `docs`: documentation
- `style`: formatting changes
- `refactor`: code refactoring
- `test`: adding tests
- `chore`: maintenance tasks

Example:
```
feat(chat): add streaming response support

- Implement WebSocket connection for real-time responses
- Add progress indicators for long responses
- Update frontend to handle streaming data

Closes #123
```

## Pull Request Process

1. **Create a Feature Branch**: Always create a new branch for your changes
2. **Write Tests**: Ensure your code includes appropriate tests
3. **Update Documentation**: Update relevant documentation
4. **Run Tests Locally**: Make sure all tests pass
5. **Submit PR**: Create a pull request with a clear description

### PR Checklist

- [ ] Code follows the project's coding standards
- [ ] Tests have been added/updated
- [ ] Documentation has been updated
- [ ] All CI checks pass
- [ ] Code has been self-reviewed
- [ ] Changes are focused on a single purpose

## Coding Standards

### Go Code

- Use `gofmt` for formatting
- Follow Go naming conventions
- Write comprehensive comments for exported functions
- Keep functions small and focused
- Use meaningful variable names

### JavaScript/React Code

- Follow ESLint configuration
- Use functional components with hooks
- Implement proper TypeScript types
- Write descriptive prop and state names
- Use meaningful component names

### Error Handling

- Handle errors gracefully
- Provide meaningful error messages
- Log errors appropriately
- Return appropriate HTTP status codes

## Testing

### Backend Testing

```bash
cd remember
go test ./... -v
```

### Frontend Testing

```bash
cd remember-web
npm test
```

### Integration Testing

Ensure all services work together:
```bash
./service.sh start all
# Run integration tests
./service.sh stop all
```

## Documentation

### Code Documentation

- Document all exported functions and types
- Include examples for complex functions
- Update README files when adding features
- Keep API documentation current

### API Documentation

When adding new API endpoints:
- Update the main README.md
- Include request/response examples
- Document authentication requirements
- List possible error codes

## Questions and Help

- **Issues**: Use GitHub Issues for bug reports and feature requests
- **Discussions**: Use GitHub Discussions for questions and ideas
- **Email**: contact@farshore.ai for private inquiries

## Recognition

Contributors will be recognized in:
- Project README
- Release notes
- Contributor hall of fame

Thank you for contributing to Farshore AI! 🚀
