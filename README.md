# Task Manager API

A comprehensive REST API service for managing tasks in teams with role-based access control, change history.

## Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for local development)
- MySQL 8.0+
- Redis 7+

### Quick Start with Docker

1. **Start services**:
   ```bash
   docker compose up -d
   ```

2. **Check services are running**:
   ```bash
   docker compose ps
   ```

3. **API is available at**: `http://localhost:8080`

4. **Stop services**:
   ```bash
   docker compose down
   ```

## API Endpoints

### Authentication

- **POST** `/api/v1/register` - Register new user
  ```json
  {
    "email": "user@example.com",
    "username": "username",
    "password": "password"
  }
  ```

- **POST** `/api/v1/login` - Login user
  ```json
  {
    "email": "user@example.com",
    "password": "password"
  }
  ```

### Teams

- **POST** `/api/v1/teams` - Create team
- **GET** `/api/v1/teams` - List user's teams
- **GET** `/api/v1/teams/{id}` - Get team details
- **GET** `/api/v1/teams/{id}/members` - List team members
- **POST** `/api/v1/teams/{id}/invite` - Invite user to team
  ```json
  {
    "user_email": "user@example.com",
    "role": "member"
  }
  ```

### Tasks

- **POST** `/api/v1/tasks` - Create task
  ```json
  {
    "team_id": "team-id",
    "title": "Task title",
    "description": "Task description",
    "priority": "high",
    "assignee_id": "user-id"
  }
  ```

- **GET** `/api/v1/tasks?team_id=X&status=todo&assignee_id=Y&page=1&page_size=20` - List tasks with filters
- **GET** `/api/v1/tasks/{id}` - Get task details
- **PUT** `/api/v1/tasks/{id}` - Update task
- **GET** `/api/v1/tasks/{id}/history` - Get task change history

### Monitoring

- **GET** `/metrics` - Prometheus metrics

## Performance Features

### Caching
- Task lists cached in Redis with 5-minute TTL

### Rate Limiting
- 100 requests per minute per user

## Development

### Build
```bash
make build
```

### Run
```bash
make run
```

### Test
```bash
make test          # Run all tests
make test-cover    # Run with coverage report
```

### Code Quality
```bash
make fmt           # Format code
make lint          # Run linter
make lint-install  # Install golangci-lint v2.4.0+
```
