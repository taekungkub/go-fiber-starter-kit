You are a senior backend engineer.

Generate a production-ready Go Fiber starter kit using Clean Architecture.

Tech stack:

- Go
- Fiber
- PostgreSQL
- Redis (for caching and token management)
- JWT Authentication (access token + refresh token)
- sqlx for sql
- Docker support

Project requirements:

1. Project structure

Use a clean architecture structure like:

.
├── cmd
│ └── server
│ └── main.go
├── config
│ ├── config.go
├── internal
│ ├── api # http handlers
│ │ ├── auth
│ │ │ ├── handler.go
│ │ │ └── router.go
│ │ │ └── dto.go
│ │ │ └── usecase.go
│ │ │ └── repository.go
│ │ └── user
│ │ │ ├── handler.go
│ │ │ └── router.go
│ │ │ └── dto.go
│ │ │ └── usecase.go
│ │ │ └── repository.go
│ │
│ ├── database
│ │ ├── postgres.go
│ │ └── redis.go
│ │
│ ├── ingest # รับข้อมูลจาก external เช่น MQTT
│ │ └── mqtt_consumer.go
│ │
│ ├── queue
│ │ └── job_queue.go
│ │
│ ├── worker # background worker
│ │ ├── job.go
│ │ ├── worker.go
│ │ └── batch_worker.go
│ │
│ │
├── middleware
│ ├── auth_middleware.go
│ └── logger.go
├── pkg
│ ├── common
│ └── core
│ └── utils.go
└── go.mod

auth

- handler
- usecase
- repository
- auth (dto, struct)
- router

user

- handler
- usecase
- repository
- user (dto, struct)
- router

2. Authentication system

Implement JWT authentication:

- register
- login
- refresh token
- logout

JWT must include:

- user_id
- role
- exp

Token system:

Access Token

- short lived (15 minutes)

Refresh Token

- long lived (7 days)

Security:

- refresh token stored hashed in database
- Redis used for caching token or blacklist
- logout should invalidate refresh token

3. Middleware

Create middleware for:

JWT middleware

- verify token
- extract user_id
- extract role
- attach to context

Role middleware

Example:

RequireRole("ADMIN")

4. User module

CRUD user endpoints:

POST /users
GET /users
GET /users/:id
PATCH /users/:id
DELETE /users/:id

Soft delete support.

5. Redis usage

Use Redis for:

- caching user data
- refresh token session
- rate limiting (optional)

6. Password

Use bcrypt for password hashing.

7. Database

PostgreSQL schema:

User table:

model User {
id String
email String
username String
password String
firstName String
lastName String
name String
role UserRole
isActive Boolean
refreshTokenHash String?
telephone String?
deletedAt DateTime?
createdAt DateTime
updatedAt DateTime
}

UserRole enum:

ADMIN
STAFF
CUSTOMER

8. API routes

/auth/register
/auth/login
/auth/refresh
/auth/logout

/users
/users/:id

9. Example responses

Login response:

{
"accessToken": "",
"refreshToken": "",
"user": {}
}

10. Implement

- DTO validation
- error handling
- response helper
- config loader
- environment variables

11. Example code

Provide example implementations for:

- JWT middleware
- Role middleware
- Auth service
- User repository
- Redis connection
- Login flow
- Refresh token flow

12. Bonus

Include:

- Dockerfile
- docker-compose
- .env.example
