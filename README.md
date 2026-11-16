# Traveloke

A Go-based microservices project for searching flights and hotels, managing user accounts, creating bookings, and processing payments.

## Services

| Service | Default port | Responsibility |
|---|---:|---|
| `user-service` | `8081` | Registration, login, JWT authentication, and user profiles |
| `booking-service` | `8082` | Flight and hotel search, inventory, and bookings |
| `payment-service` | `8083` | Payment creation and asynchronous payment processing |
| `admin-service` | `8084` | Administrative flight, hotel, and analytics operations |

Each service is an independent Go module.

## Architecture

```text
Client
  ├── user-service       :8081
  ├── booking-service    :8082
  ├── payment-service    :8083
  └── admin-service      :8084

booking-service  <── HTTP/internal API ──  payment-service
booking-service  <── RabbitMQ payment result events
payment-service  <── RabbitMQ payment requests/results
```

The services use PostgreSQL, RabbitMQ, and Redis. User, booking, and payment services share the same JWT secret so that authenticated requests can be verified consistently. Internal service requests are protected with `INTERNAL_API_KEY`.

## Requirements

- Go 1.27.1 or later
- PostgreSQL
- RabbitMQ
- Redis
- Docker Desktop with Docker Compose (optional, for containerized setup)

When running locally, create these PostgreSQL databases before starting the services:

- `traveloka_users`
- `traveloka_bookings`
- `traveloka_payments`

Tables are created automatically with GORM migrations when each service starts. The booking service also seeds sample flights and hotels when its database is empty.

When using Docker Compose, `init-db.sql` creates the three databases automatically on the first PostgreSQL initialization.

## Configuration

### Docker Compose configuration

The root `.env.example` contains the variables used by Docker Compose:

```text
.env.example
```

Create the root `.env` file before starting the containers:

```powershell
Copy-Item .\.env.example .\.env
```

Update the placeholder values in `.env`, especially `POSTGRES_PASSWORD`, `JWT_SECRET`, and `INTERNAL_API_KEY`. The root `.env` is ignored by Git and must not contain credentials that are committed to the repository.

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
Copy-Item .\admin-service\.env.example .\admin-service\.env
```

The admin service requires `JWT_SECRET`, `INTERNAL_API_KEY`, and `BOOKING_SERVICE_URL`. Its admin endpoints require a valid JWT with the `is_admin` claim.

Use the same values for `JWT_SECRET` and `INTERNAL_API_KEY` in the user, booking, payment, and admin service configuration files. For local execution, use `localhost` service addresses. Inside Docker Compose, services use container names such as `postgres`, `rabbitmq`, and `redis`.

## Running the Services

### Docker Compose

From the repository root, after creating `.env` from `.env.example`:

```powershell
docker compose up --build -d
docker compose ps
```

The services are available on ports `8081` through `8084`. PostgreSQL, RabbitMQ, and Redis are also started automatically. RabbitMQ management UI is available at `http://localhost:15672`, and Redis is available at `localhost:6379` from the host.

To stop the containers:

```powershell
docker compose down
```

### Running locally with Go

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

### Admin service

```powershell
Set-Location .\admin-service
go mod download
go run .
```

Runs on `http://localhost:8084`.

## API Overview

All protected endpoints require:

```http
Authorization: Bearer <jwt-token>
```

Every service exposes a public Prometheus metrics endpoint:

| Service | Metrics endpoint |
|---|---|
| User service | `GET http://localhost:8081/metrics` |
| Booking service | `GET http://localhost:8082/metrics` |
| Payment service | `GET http://localhost:8083/metrics` |
| Admin service | `GET http://localhost:8084/metrics` |

The metrics include Go runtime metrics and HTTP request counters and duration histograms, including `http_requests_total` and `http_request_duration_seconds`.

To test metrics, send a `GET` request to any service's `/metrics` endpoint using Postman or a browser. The response uses Prometheus text format:

```text
http_requests_total{method="GET",path="/health",status="200"} 1
http_request_duration_seconds_count{method="GET",path="/health"} 1
```

When running the services with Docker Compose after changing the source code, rebuild the affected containers:

```powershell
docker compose up -d --build --force-recreate
```

### User service

| Method | Endpoint | Authentication |
|---|---|---|
| `POST` | `/auth/register` | Public |
| `POST` | `/auth/login` | Public |
| `GET` | `/users/me` | JWT required |
| `PATCH` | `/users/me` | JWT required |

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

### Wallet

| Method | Endpoint | Authentication |
|---|---|---|
| `GET` | `/wallets/me` | JWT required |
| `POST` | `/wallets/topup` | JWT required |

Top up request:

```json
{
  "amount": 5000000
}
```

The amount must be greater than `0` and no more than `100000000`. Payment processing deducts the booking amount from the user's wallet. Payments fail when the wallet balance is insufficient.

### Admin service

| Method | Endpoint | Authentication |
|---|---|---|
| `GET` | `/health` | Public |
| `POST` | `/admin/flights` | Admin JWT required |
| `PATCH` | `/admin/flights/:id` | Admin JWT required |
| `DELETE` | `/admin/flights/:id` | Admin JWT required |
| `POST` | `/admin/hotels` | Admin JWT required |
| `PATCH` | `/admin/hotels/:id` | Admin JWT required |
| `DELETE` | `/admin/hotels/:id` | Admin JWT required |
| `GET` | `/admin/analytics` | Admin JWT required |

All admin endpoints require a valid JWT signed with `JWT_SECRET` and an `is_admin: true` claim. The admin service forwards authorized requests to the booking service's internal API using `INTERNAL_API_KEY`.

Create a flight:

```json
{
  "airline": "Example Air",
  "flight_number": "EA123",
  "origin": "CGK",
  "destination": "DPS",
  "departure_time": "2026-12-01T08:00:00Z",
  "arrival_time": "2026-12-01T10:00:00Z",
  "price": 1250000,
  "total_seats": 180
}
```

Update a flight with any subset of the same fields. `arrival_time` must be later than `departure_time`, and `total_seats` cannot be reduced below the number of seats already booked.

Create a hotel:

```json
{
  "name": "Example Hotel",
  "city": "Bali",
  "stars": 4,
  "price_per_night": 850000,
  "total_rooms": 60
}
```

Update a hotel with any subset of these fields. `stars` must be between 1 and 5, and `total_rooms` cannot be reduced below the number of rooms already booked.

Flight or hotel deletion returns a conflict when the item still has active `held` or `paid` bookings. The analytics endpoint returns total bookings, bookings created today, booking counts by status, total paid revenue, and the five most-booked flight routes.

## Example Workflow

1. Register a user through `user-service`.
2. Log in and save the returned JWT.
3. Search flights or hotels through `booking-service`.
4. Create a booking using the selected item ID.
5. Create a payment using the booking ID.
6. Top up the wallet if necessary using `/wallets/topup`.
7. Check the payment and booking status through their respective endpoints.
8. Use an admin JWT to manage flight and hotel inventory or view booking analytics.

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
- Keep Redis and RabbitMQ accessible only to trusted services in production.
- Use HTTPS and a secure secret-management solution in production.
