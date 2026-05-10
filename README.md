### Run

```
go run ./cmd
```

### Run with air (auto-reload)

```
air
```

### Build

```
go build -o app ./cmd
```

# Go Fiber Starter Kit

[![Go](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)
[![Fiber](https://img.shields.io/badge/Fiber-v2-green.svg)](https://gofiber.io)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-blue.svg)](https://www.postgresql.org)
[![Redis](https://img.shields.io/badge/Redis-7-red.svg)](https://redis.io)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://www.docker.com)

A production-ready Go Fiber starter kit implementing Clean Architecture with JWT authentication, user management, PostgreSQL database, and Redis caching.

## ✨ Features

- 🏗️ **Clean Architecture** - Separation of concerns with layered architecture
- 🔐 **JWT Authentication** - Access token (15min) + Refresh token (7 days) system
- 👥 **User Management** - Complete CRUD operations with soft delete
- 🛡️ **Role-Based Authorization** - ADMIN, STAFF, CUSTOMER roles
- 🗄️ **PostgreSQL** - Database with migrations and proper indexing
- ⚡ **Redis** - Caching, token management, and rate limiting
- 🐳 **Docker Support** - Ready-to-use Docker Compose setup
- 🚦 **Rate Limiting** - Built-in rate limiting middleware
- 🔒 **Security** - Password hashing, token blacklisting, CORS support

## 🛠 Tech Stack

- **Backend**: Go 1.24+ with Fiber v2
- **Database**: PostgreSQL 17
- **Cache**: Redis 7
- **Authentication**: JWT (access + refresh tokens)
- **ORM**: sqlx for SQL operations
- **Validation**: go-playground/validator
- **Deployment**: Docker & Docker Compose

## 📁 Project Structure

```
go-fiber-starter-kit/
├── cmd/
│   └── main.go              # Application entry point
├── config/                  # Configuration management
├── internal/
│   ├── api/
│   │   ├── auth/           # Authentication module
│   │   │   ├── handler.go
│   │   │   ├── usecase.go
│   │   │   ├── repository.go
│   │   │   └── router.go
│   │   └── user/           # User management module
│   │       ├── handler.go
│   │       ├── usecase.go
│   │       ├── repository.go
│   │       └── router.go
│   └── database/
│       ├── postgres.go     # PostgreSQL connection
│       └── redis.go        # Redis connection
├── middleware/
│   ├── jwt_middleware.go   # JWT validation
│   └── role_middleware.go  # Role-based access
├── migrations/
│   └── 001_init.sql        # Database schema
├── pkg/
│   ├── common/             # Common utilities
│   └── core/               # Core functionality
├── .env.example            # Environment variables template
├── docker-compose.yml      # Docker setup
└── Dockerfile             # Container build
```

## 🚀 Quick Start

### Prerequisites

- Go 1.24 or higher
- Docker and Docker Compose
- PostgreSQL and Redis (if running locally)

### Using Docker Compose (Recommended)

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd go-fiber-starter-kit
   ```

2. **Configure environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start the application**

   ```bash
   docker-compose up -d
   ```

4. **Access the API**
   - API Base URL: `http://localhost:8080`
   - Health Check: `http://localhost:8080/health`

### Local Development

1. **Install dependencies**

   ```bash
   go mod download
   ```

2. **Set up PostgreSQL and Redis**
   - Ensure PostgreSQL is running on localhost:5432
   - Ensure Redis is running on localhost:6379
   - Create database: `go_fiber_starter_db`

3. **Configure environment**

   ```bash
   cp .env.example .env
   # Update .env with your database credentials
   ```

4. **Run the application**
   ```bash
   go run cmd/main.go
   ```

## ⚙️ Environment Variables

Copy `.env.example` to `.env` and configure the following variables:

```bash
# Application
APP_ENV="dev"

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_NAME=go_fiber_starter_db
DB_USER=postgres
DB_PASS=postgres

# JWT Configuration
JWT_SECRET=change-this-to-a-secure-random-string-at-least-32-chars
JWT_EXPIRES_IN_ACCESS_TOKEN=15
JWT_EXPIRES_IN_REFRESH_TOKEN=10080
REFRESH_SECRET=change-this-to-another-secure-random-string-at-least-32-chars

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
```

## 📚 API Endpoints

### Authentication

| Method | Endpoint             | Description                 |
| ------ | -------------------- | --------------------------- |
| POST   | `/api/auth/register` | Register new user           |
| POST   | `/api/auth/login`    | User login                  |
| POST   | `/api/auth/refresh`  | Refresh access token        |
| POST   | `/api/auth/logout`   | User logout (requires auth) |

### User Management

| Method | Endpoint         | Description      | Authentication            |
| ------ | ---------------- | ---------------- | ------------------------- |
| GET    | `/api/users`     | Get all users    | Required (ADMIN/STAFF)    |
| GET    | `/api/users/:id` | Get user by ID   | Required                  |
| PATCH  | `/api/users/:id` | Update user      | Required (owner or ADMIN) |
| DELETE | `/api/users/:id` | Soft delete user | Required (ADMIN)          |

### Example Requests

#### Register User

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "johndoe",
    "password": "securepassword",
    "firstName": "John",
    "lastName": "Doe"
  }'
```

#### Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword"
  }'
```

#### Login Response

```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid-here",
    "email": "user@example.com",
    "username": "johndoe",
    "firstName": "John",
    "lastName": "Doe",
    "role": "CUSTOMER",
    "isActive": true,
    "createdAt": "2024-01-01T00:00:00Z"
  }
}
```

#### Access Protected Route

```bash
curl -X GET http://localhost:8080/api/users \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## 🗄 Database Schema

### Users Table

```sql
CREATE TYPE user_role AS ENUM ('ADMIN', 'STAFF', 'CUSTOMER');

CREATE TABLE users (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email               VARCHAR(255) NOT NULL UNIQUE,
    username            VARCHAR(100) NOT NULL UNIQUE,
    password            VARCHAR(255) NOT NULL,
    first_name          VARCHAR(100) NOT NULL DEFAULT '',
    last_name           VARCHAR(100) NOT NULL DEFAULT '',
    name                VARCHAR(200) GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED,
    role                user_role NOT NULL DEFAULT 'CUSTOMER',
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    refresh_token_hash  VARCHAR(255),
    telephone           VARCHAR(20),
    deleted_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### User Roles

- **ADMIN**: Full access to all resources
- **STAFF**: Limited access, can manage users
- **CUSTOMER**: Default role, can manage own profile

## 🔒 Security Features

### JWT Token System

- **Access Token**: 15 minutes expiry, used for API requests
- **Refresh Token**: 7 days expiry, used to obtain new access tokens
- **Token Blacklisting**: Logout adds tokens to Redis blacklist
- **Secure Storage**: Refresh tokens stored hashed in database

### Password Security

- **Hashing**: bcrypt with automatic salt generation
- **Validation**: Minimum password requirements enforced

### Rate Limiting

- **Global Limit**: 100 requests per minute per IP
- **Redis Storage**: Distributed rate limiting across instances

### CORS Support

- Configurable CORS middleware for cross-origin requests

## 🛠 Development

### Database Migrations

Migrations are automatically applied when starting with Docker Compose. For manual setup:

```bash
# Apply migrations
psql -h localhost -U postgres -d go_fiber_starter_db -f migrations/001_init.sql
```

### Adding New Modules

Follow the existing pattern in `internal/api/`:

1. Create module folder with `handler.go`, `usecase.go`, `repository.go`, `router.go`
2. Define DTOs and models
3. Implement business logic in usecase
4. Add routes in router
5. Register in `main.go`

### Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

## 🚀 Deployment

### Docker Deployment

The application is containerized and ready for production deployment:

```bash
# Build and deploy
docker-compose -f docker-compose.yml up -d

# View logs
docker-compose logs -f app
```

### Environment Configuration

- **Production**: Set `APP_ENV=prod`
- **Development**: Set `APP_ENV=dev`
- Ensure strong JWT secrets in production
- Use environment-specific database credentials

### Health Monitoring

- Health endpoint: `GET /health`
- Application logs include request logging
- Database connection monitoring

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go naming conventions
- Write clean, commented code
- Add tests for new features
- Update documentation as needed
- Ensure all tests pass before submitting

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

If you encounter any issues or have questions:

1. Check the existing issues
2. Create a new issue with detailed information
3. Include environment details and error logs

---

**Built with ❤️ using Go Fiber and Clean Architecture**
'@
