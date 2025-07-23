# Project Goal
This is a Web API for managing travel plans.

# Project Structure
This project follows the layered architecture of a typical Go language project.

# Development Stack
- Language: Go
- Framework: Gin
- Database: PostgreSQL, sqlc, golang-migrate/migrate, pgx

# Testing Tools
- testify: Assertion library
- testcontainers-go: Docker container management for testing
- uber-go/mock: Automatic mock generation

# Coding Rules

## Markdown

- Markdown must be written in English.

## Go

- Comments in the go code should be written in Japanese.

## Git

- Commit messages should follow the Conventional Commits rule, but the description should be written in Japanese.

# Directory Patterns
```
cmd/               # Entry point of the application.
internal/          # Internal application logic.
  domain/          # Core business logic and entities.
  infrastructure/  # Implementation of external dependencies such as databases.
  interface/.      # Handlers for requests and responses.
  usecase/         # Application-specific business rules.
```

# Notes

- Please respond to users in Japanese.
- Choose appropriate methods in practice.
- If a fix takes more than 3 minutes, skip it. Finally, summarize and explain the failed fixes.
