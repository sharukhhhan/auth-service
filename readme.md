# Authentication Service

An authentication service written in Go that provides JWT-based token generation and refresh functionality. The project uses PostgreSQL for storage and includes mocked dependencies for testing purposes.

## How to run

### Clone the Repository
```bash
git clone https://github.com/yourusername/auth-service.git
cd auth-service
```

### Configs
Configure env variables in config.yaml
```yaml
http:
  port: ":8080"
  shutdown_timeout: 5s

log:
  level: "debug"
  log_path: "./logs"

database:
  postgres:
    host: "host.docker.internal"
    port: 5432
    user: "postgres"
    password: "12345"
    name: "medods-tz"
    migration_path: "./migrations"

jwt:
  sign_key: "hello"
  token_ttl: 20m
  refresh_token_ttl: 168h

smtp:
  host: "smtp.example.com"
  port: 587
  user: "your_email@example.com"
  password: "your_password"
```

### Build and Run
#### Without Docker
```bash
go mod download
go build -o auth-service ./cmd/app
./auth-service
```
#### With Docker
```bash
docker build -t auth-service .
docker run -p 8080:8080 auth-service
```

### Usage
#### Endpoints

- POST /api/v1/auth/token: Generate a new access and refresh token pair.
```json
{
  "user_id": "7c452d37-4e83-4f7c-ac41-ab1a6b510c59",
  "client_ip": "192.0.2.1"
}
```

- POST /api/v1/auth/refresh: Refresh tokens using a valid refresh token.
```json
{
  "refresh_token": "your_refresh_token",
  "access_token": "your_expired_access_token"
}
```

### Testing
Run tests using the following command:
```bash
go test ./...
```