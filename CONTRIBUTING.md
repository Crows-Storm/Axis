# Contributing to Axis

First off, thank you for considering contributing to Axis! It's people like you that make Axis such a great tool.

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

* **Use a clear and descriptive title**
* **Describe the exact steps which reproduce the problem**
* **Provide specific examples to demonstrate the steps**
* **Describe the behavior you observed after following the steps**
* **Explain which behavior you expected to see instead and why**
* **Include screenshots and animated GIFs if possible**

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

* **Use a clear and descriptive title**
* **Provide a step-by-step description of the suggested enhancement**
* **Provide specific examples to demonstrate the steps**
* **Describe the current behavior and explain which behavior you expected to see instead**
* **Explain why this enhancement would be useful**

### Pull Requests

* Fill in the required template
* Do not include issue numbers in the PR title
* Follow the Go coding style
* Include thoughtfully-worded, well-structured tests
* Document new code
* End all files with a newline

## Development Process

### 1. Fork and Clone

```bash
# Fork the repository on GitHub
# Then clone your fork
git clone https://github.com/YOUR_USERNAME/Axis.git
cd Axis
```

### 2. Create a Branch

```bash
git checkout -b feature/my-new-feature
# or
git checkout -b fix/my-bug-fix
```

Branch naming convention:
- `feature/` - for new features
- `fix/` - for bug fixes
- `refactor/` - for code refactoring
- `docs/` - for documentation changes
- `test/` - for test additions/changes

### 3. Set Up Development Environment

```bash
# Install dependencies
make deps

# Start infrastructure services
docker-compose up -d

# Run migrations
make migrate-up
```

### 4. Make Your Changes

* Write clear, commented code
* Follow the existing code style
* Add or update tests as needed
* Update documentation as needed

### 5. Test Your Changes

```bash
# Run unit tests
make test

# Run linter
make lint

# Run integration tests
make test-integration

# Check test coverage
make coverage
```

### 6. Commit Your Changes

We follow [Conventional Commits](https://www.conventionalcommits.org/) specification:

```bash
git commit -m "feat: add new feature"
git commit -m "fix: resolve bug in user service"
git commit -m "docs: update README"
git commit -m "refactor: simplify event dispatcher"
git commit -m "test: add tests for wallet service"
```

Commit types:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools

### 7. Push and Create PR

```bash
git push origin feature/my-new-feature
```

Then go to GitHub and create a Pull Request.

## Coding Standards

### Go Style Guide

* Follow [Effective Go](https://golang.org/doc/effective_go.html)
* Use `gofmt` to format your code
* Run `golangci-lint` before committing
* Keep functions small and focused
* Write meaningful variable and function names
* Add comments for exported functions

### Project-Specific Conventions

#### 1. Package Naming

```go
// ✅ Good
package user
package repository
package http

// ❌ Bad
package userPackage
package user_service
package HTTPHandler
```

#### 2. Interface Naming

```go
// ✅ Good
type Repository interface { ... }
type Handler interface { ... }

// ❌ Bad
type IRepository interface { ... }
type UserRepositoryInterface interface { ... }
```

#### 3. Error Handling

```go
// ✅ Good
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// ❌ Bad
if err != nil {
    return errors.New("error")
}
```

#### 4. Context Usage

```go
// ✅ Good - context as first parameter
func (s *Service) Handle(ctx context.Context, cmd Command) error {
    ...
}

// ❌ Bad
func (s *Service) Handle(cmd Command, ctx context.Context) error {
    ...
}
```

#### 5. DDD Patterns

Follow the project's DDD structure:

```go
// Domain Layer - Pure business logic
type User struct {
    domain.AggregateRoot
    // fields...
}

func (u *User) Activate() error {
    // Business rule
    u.IncrementVersion()
    u.AddEvent(NewUserActivatedEvent(u))
    return nil
}

// Application Layer - Use case orchestration
type CreateUserHandler struct {
    userRepo user.Repository
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd CreateUserCommand) error {
    // Orchestrate domain objects
}
```

#### 6. CQRS Pattern

Use the decorator pattern for handlers:

```go
// Command Handler
type CreateUserCommandHandler decorator.CommandHandler[CreateUserCommand, *User]

func NewCreateUserCommandHandler(
    repo Repository,
    metricsClient decorator.MetricsClient,
) CreateUserCommandHandler {
    return decorator.ApplyCommandDecorators[CreateUserCommand, *User](
        &createUserCommandHandler{repo: repo},
        metricsClient,
    )
}
```

### Testing Standards

#### Unit Tests

```go
func TestUser_Activate(t *testing.T) {
    // Arrange
    user := &User{Status: StatusInactive}
    
    // Act
    err := user.Activate()
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, StatusActive, user.Status)
    assert.True(t, user.HasEvents())
}
```

#### Integration Tests

```go
func TestUserRepository_Save(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()
    
    repo := NewUserRepository(db)
    user := &User{LoginId: "test"}
    
    // Test
    err := repo.Save(context.Background(), user)
    assert.NoError(t, err)
}
```

### Documentation

* Add godoc comments for all exported types and functions
* Update README.md if adding new features
* Add examples in comments when helpful
* Update API documentation if changing endpoints

```go
// CreateUser creates a new user with the given credentials.
// It validates the input, hashes the password, and persists the user.
//
// Example:
//   user, err := service.CreateUser(ctx, CreateUserCommand{
//       LoginId: "john",
//       Email: "john@example.com",
//   })
func (s *Service) CreateUser(ctx context.Context, cmd CreateUserCommand) (*User, error) {
    ...
}
```

## Project Structure Guidelines

When adding new features, follow the established structure:

### Adding a New Service

```
services/your-service/
├── cmd/api/
│   ├── main.go
│   └── wire.go
├── internal/
│   ├── domain/
│   ├── application/
│   ├── infrastructure/
│   └── ports/
├── configs/
├── go.mod
└── Makefile
```

### Adding a New Aggregate

```
internal/domain/aggregate-name/
├── aggregate.go      # Aggregate root
├── repository.go     # Repository interface
├── events.go         # Domain events
├── value_objects.go  # Value objects
└── aggregate_test.go # Tests
```

### Adding a New Command/Query

```
internal/application/command/
├── create_entity.go
├── update_entity.go
└── delete_entity.go

internal/application/query/
├── get_entity.go
└── list_entities.go
```

## Review Process

After you submit a pull request:

1. **Automated Checks**: CI will run tests and linters
2. **Code Review**: Maintainers will review your code
3. **Discussion**: Be prepared to discuss your changes
4. **Revisions**: Make any requested changes
5. **Approval**: Once approved, your PR will be merged

### Review Checklist

- [ ] Code follows project style guidelines
- [ ] Tests pass locally
- [ ] New tests added for new functionality
- [ ] Documentation updated
- [ ] No merge conflicts
- [ ] Commit messages follow convention
- [ ] PR description is clear

## Getting Help

* **Discord**: Join our [Discord server](https://discord.gg/axis)
* **GitHub Discussions**: Ask questions in [Discussions](https://github.com/Crows-Storm/Axis/discussions)
* **Email**: Contact maintainers at crow@axis-platform.com

## Recognition

Contributors will be recognized in:
* README.md contributors section
* Release notes
* Project documentation

Thank you for contributing to Axis! 🎉
