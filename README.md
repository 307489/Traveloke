# Traveloka Clone

A Go-based microservices project for searching flights and hotels, managing user accounts, creating bookings, and processing payments.

## Services

| Service | Default port | Responsibility | Status |
|---|---:|---|---|
| `user-service` | `8081` | Registration, login, JWT authentication, and user profiles | Implemented |
| `booking-service` | `8082` | Flight and hotel search, inventory, and bookings | Implemented |
| `payment-service` | `8083` | Payment creation and asynchronous payment processing | Implemented |
| `admin-service` | - | Reserved for administrative features | Not implemented |

Each service is an independent Go module.

## Architecture

```text
Client
  ├── user-service       :8081
  ├── booking-service    :8082
  └── payment-service    :8083

booking-service  <── HTTP/internal API ──  payment-service
booking-service  <── RabbitMQ payment result events
payment-service  ── RabbitMQ payment requests/results
```

The services use PostgreSQL and RabbitMQ. User, booking, and payment services share the same JWT secret so that authenticated requests can be verified consistently. Internal booking requests are protected with `INTERNAL_API_KEY`.

## Requirements

- Go 1.27.1 or later
- PostgreSQL
- RabbitMQ

Create these PostgreSQL databases before starting the services:

- `traveloka_users`
- `traveloka_bookings`
- `traveloka_payments`

Tables are created automatically with GORM migrations when each service starts. The booking service also seeds sample flights and hotels when its database is empty.

## Configuration

Each service has its own environment file:

```text
user-service/.env.example
booking-service/.env.example
payment-service/.env.example
admin-service/.env.example
```

Create the local files from the templates:

```powershell
Copy-Item .\user-service\.env.example .\user-service\.env
Copy-Item .\booking-service\.env.example .\booking-service\.env
Copy-Item .\payment-service\.env.example .\payment-service\.env
```

`admin-service/.env.example` is currently empty because the admin service has no implementation or runtime configuration yet.

Use the same values for `JWT_SECRET` and `INTERNAL_API_KEY` in the user, booking, and payment service configuration files. Update database credentials and the RabbitMQ URL for the local environment.

## Running the Services

Open a separate terminal for each service.

### User service

```powershell
Set-Location .\user-service
go mod download
go run .
```

Runs on `http://localhost:8081`.

### Booking service

```powershell
Set-Location .\booking-service
go mod download
go run .
```

Runs on `http://localhost:8082`.

### Payment service

```powershell
Set-Location .\payment-service
go mod download
go run .
```

Runs on `http://localhost:8083`.

## API Overview

All protected endpoints require:

```http
Authorization: Bearer <jwt-token>
```

### User service

| Method | Endpoint | Authentication |
|---|---|---|
| `POST` | `/auth/register` | Public |
| `POST` | `/auth/login` | Public |
| `GET` | `/users/me` | JWT required |
| `PUT` | `/users/me` | JWT required |

Register:

```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "secure-password",
  "full_name": "John Doe",
  "phone": "+62123456789"
}
```

Login:

```json
{
  "email": "john@example.com",
  "password": "secure-password"
}
```

The login response contains a JWT token for the protected endpoints.

### Booking service

| Method | Endpoint | Authentication |
|---|---|---|
| `GET` | `/health` | Public |
| `GET` | `/flights` | Public |
| `GET` | `/hotels` | Public |
| `POST` | `/bookings` | JWT required |
| `GET` | `/bookings/me` | JWT required |
| `GET` | `/bookings/:id` | JWT required |
| `POST` | `/bookings/:id/cancel` | JWT required |
| `GET` | `/internal/bookings/:id` | Internal API key |

Flight search supports `from`, `to`, `date`, `min_price`, `max_price`, `passengers`, `page`, `limit`, and `sort` query parameters. Hotel search supports `city`, `min_stars`, `min_price`, `max_price`, `rooms`, `page`, `limit`, and `sort`.

Create a booking:

```json
{
  "type": "flight",
  "item_id": "<flight-id>",
  "quantity": 1
}
```

Supported booking types are `flight` and `hotel`. A booking is held for 10 minutes while payment is completed.

### Payment service

| Method | Endpoint | Authentication |
|---|---|---|
| `GET` | `/health` | Public |
| `POST` | `/payments` | JWT required |
| `GET` | `/payments/me` | JWT required |

Create a payment for a held booking:

```json
{
  "booking_id": "<booking-id>"
}
```

Payment requests are published to RabbitMQ and processed asynchronously. The payment service validates the booking through the booking service's internal endpoint before creating a payment.

## Example Workflow

1. Register a user through `user-service`.
2. Log in and save the returned JWT.
3. Search flights or hotels through `booking-service`.
4. Create a booking using the selected item ID.
5. Create a payment using the booking ID.
6. Check the payment and booking status through their respective endpoints.

## Development

Run tests for an individual service from that service's directory:

```powershell
go test ./...
```

The repository does not currently have a root Go module, so commands must be run inside the relevant service directory.

## Security

- Do not commit `.env` files or real credentials.
- Use strong, unique values for `JWT_SECRET` and `INTERNAL_API_KEY`.
- Keep the booking service's internal endpoint accessible only to trusted services.
- Use HTTPS and a secure secret-management solution in production.
