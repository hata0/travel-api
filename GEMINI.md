# Project Goal
This is a Web API for managing travel plans.

# Project Structure
This project follows the layered architecture of a typical Go language project.

# Development Stack
- Language: Go
- Framework: Gin
- Database: PostgreSQL
- Others: sqlc

# Coding Rules

## General

- Comments in the code should be written in Japanese.

# Directory Patterns
```
cmd/               # Entry point of the application.
internal/          # Internal application logic.
  domain/          # Core business logic and entities.
  infrastructure/  # Implementation of external dependencies such as databases.
  interface/.      # Handlers for requests and responses.
  usecase/         # Application-specific business rules.
```
