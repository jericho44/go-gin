# Gin Golang Application

A well-structured Gin Golang project with proper organization, following Go best practices and conventions. This structure provides a solid foundation for building REST APIs with clear separation of concerns, proper dependency management, and scalable architecture.

## Project Structure

```
gin-golang-app/
├── main.go                 # Application entry point
├── go.mod                  # Go module file
├── .env                    # Environment variables (development)
├── .gitignore             # Git ignore file
├── README.md              # Project documentation
├── cmd/                   # Main applications
├── internal/              # Private application code
├── pkg/                   # Public library code
└── tests/                 # Test files
```

## Features

- Clean project structure following Go conventions
- Gin web framework for HTTP routing
- Environment-based configuration management
- Middleware support for cross-cutting concerns
- Repository pattern for data access
- Comprehensive error handling
- Unit and integration testing setup

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git

### Installation

1. Clone the repository:

   ```bash
   git clone <repository-url>
   cd gin-golang-app
   ```

2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Set up environment variables:

   ```bash
   cp .env.example .env
   ```

   Edit `.env` file with your configuration values.

4. Run the application:
   ```bash
   go run main.go
   ```

The server will start on `http://localhost:8080`

### Health Check

Visit `http://localhost:8080/health` to verify the server is running.

## Configuration

The application uses environment-based configuration management with support for multiple environments:

### Environment Files

- `.env` - Development environment (not committed to git)
- `.env.example` - Template for environment variables
- `.env.staging` - Staging environment configuration
- `.env.production` - Production environment configuration

### Configuration Variables

| Variable               | Description                                              | Default                       |
| ---------------------- | -------------------------------------------------------- | ----------------------------- |
| `APP_ENV`              | Application environment (development/staging/production) | `development`                 |
| `SERVER_HOST`          | Server host address                                      | `localhost`                   |
| `SERVER_PORT`          | Server port                                              | `8080`                        |
| `SERVER_READ_TIMEOUT`  | Server read timeout in seconds                           | `30`                          |
| `SERVER_WRITE_TIMEOUT` | Server write timeout in seconds                          | `30`                          |
| `DB_HOST`              | Database host                                            | `localhost`                   |
| `DB_PORT`              | Database port                                            | `5432`                        |
| `DB_USER`              | Database user                                            | `postgres`                    |
| `DB_PASSWORD`          | Database password                                        | ``                            |
| `DB_NAME`              | Database name                                            | `gin_app`                     |
| `DB_SSL_MODE`          | Database SSL mode                                        | `disable`                     |
| `JWT_SECRET`           | JWT secret key                                           | `your-secret-key`             |
| `JWT_EXPIRY_HOURS`     | JWT token expiry in hours                                | `24`                          |
| `CORS_ALLOWED_ORIGINS` | Allowed CORS origins (comma-separated)                   | `*`                           |
| `CORS_ALLOWED_METHODS` | Allowed CORS methods (comma-separated)                   | `GET,POST,PUT,DELETE,OPTIONS` |
| `CORS_ALLOWED_HEADERS` | Allowed CORS headers (comma-separated)                   | `Content-Type,Authorization`  |

### Environment-Specific Defaults

The configuration system automatically applies environment-specific defaults:

- **Development**: Uses localhost settings with SSL disabled
- **Staging**: Uses 0.0.0.0 host with SSL preferred
- **Production**: Uses 0.0.0.0 host with SSL required

## Development

### Project Layout

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout):

- `cmd/`: Main applications for this project
- `internal/`: Private application and library code
- `pkg/`: Library code that's ok to use by external applications
- `tests/`: Additional external test apps and test data

### Building

```bash
go build -o bin/app main.go
```

### Testing

```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
