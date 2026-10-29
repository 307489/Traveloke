# Traveloka Clone

A microservices-based Traveloka clone built with Go. The project is organized into separate services for users, administrators, bookings, and payments.

## Services

| Service | Purpose | Current status |
|---|---|---|
| `user-service` | User registration, login, JWT authentication, and profile management | Implemented |
| `admin-service` | Administrative operations | Module initialized; implementation pending |
| `booking-service` | Booking management | Module initialized; implementation pending |
| `payment-service` | Payment processing | Module initialized; implementation pending |

## Architecture

Each service is an independent Go module:

```text
traveloka-clone/
├── admin-service/
├── booking-service/
├── payment-service/
└── user-service/
```

The current user service uses:

- Go
- Gin for HTTP routing
- PostgreSQL for persistence
- GORM for database access
- JWT for authentication
- bcrypt for password hashing

## Requirements

- Go 1.27.1 or later
- PostgreSQL

## Getting Started

### 1. Clone the repository

```powershell
git clone <repository-url>
cd traveloka-clone
```

### 2. Prepare the user service

Create a PostgreSQL database named `traveloka_users`, then create the user service environment file:

```powershell
Copy-Item .\user-service\.env.example .\user-service\.env
```

Update the database credentials and `JWT_SECRET` in `user-service\.env`.

### 3. Start the user service

```powershell
Set-Location .\user-service
go mod download
go run .
```

The user service runs at `http://localhost:8081` by default.

## User Service API

### Register

```http
POST /auth/register
Content-Type: application/json
```

```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "secure-password",
  "full_name": "John Doe",
  "phone": "+62123456789"
}
```

### Login

```http
POST /auth/login
Content-Type: application/json
```

```json
{
  "email": "john@example.com",
  "password": "secure-password"
}
```

The response contains a JWT token:

```json
{
  "token": "your-jwt-token"
}
```

### Get the current user's profile

```http
GET /users/me
Authorization: Bearer your-jwt-token
```

### Update the current user's profile

```http
PUT /users/me
Authorization: Bearer your-jwt-token
Content-Type: application/json
```

```json
{
  "full_name": "John Smith",
  "phone": "+62987654321"
}
```

## Environment Variables

The user service uses `user-service/.env.example` as its configuration template:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-database-password
DB_NAME=traveloka_users
JWT_SECRET=your-jwt-secret
PORT=8081
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

`RABBITMQ_URL` is reserved for future messaging integration and is not currently used by the user service.

## Development

Run commands from the directory of the service being changed:

```powershell
go mod download
go run .
go test ./...
```

The database schema for the user service is created automatically with GORM migrations when the service starts.

## Security

- Do not commit `.env` files or real credentials.
- Use a strong, unique `JWT_SECRET`.
- Use secure secret management and HTTPS in production.
