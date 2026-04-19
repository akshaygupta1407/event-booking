# Event Booking System Backend

This is a simple backend for an Event Booking System built in Go. It supports two user roles:

- Organizers, who can create and manage their own events
- Customers, who can browse events and book tickets

The project also includes background job processing. When a booking is created, the system queues a booking confirmation task. When an event is updated, it queues notification tasks for customers who have already booked that event. For this assignment, those jobs are simulated with console logs.

## Tech stack

- Go
- Gin
- GORM
- PostgreSQL
- Redis
- Asynq
- JWT authentication

## Features

- User registration and login
- Role-based access control for organizers and customers
- Event creation and update APIs for organizers
- Event browsing APIs for customers
- Ticket booking with capacity checks
- Background jobs for booking confirmations
- Background jobs for event update notifications

## Project structure

- `cmd/api` contains the application entrypoint
- `internal/config` handles environment-based configuration
- `internal/database` contains database setup and migrations
- `internal/handlers` contains HTTP handlers
- `internal/middleware` contains authentication and authorization middleware
- `internal/repositories` contains database access logic
- `internal/services` contains business logic
- `internal/jobs` contains job queue and worker handlers
- `tests` contains API-level test coverage

## Running the project

The easiest way to run everything is with Docker Compose.

```bash
docker compose up --build
```

This starts:

- the API service
- the background worker
- PostgreSQL
- Redis

Once the services are up, the API will be available at:

```text
http://localhost:8080
```

If you prefer testing with Postman, this repository also includes a Postman collection export in [`event-booking-postman-collection.json`](event-booking-postman-collection.json). You can import it directly into Postman and run the requests from there.

## Environment variables

Default values are provided in `.env.example`.

- `APP_MODE` chooses whether the container runs the API or the worker
- `HTTP_PORT` sets the API port
- `DATABASE_URL` is the PostgreSQL connection string
- `REDIS_ADDR` is the Redis address
- `JWT_SECRET` is used to sign tokens
- `JWT_EXPIRY_HOURS` controls token expiry time

## API overview

### Authentication

- `POST /auth/register`
- `POST /auth/login`

### Health

- `GET /health`

### Events

- `GET /events`
- `GET /events/:id`
- `POST /events` organizer only
- `PUT /events/:id` organizer only
- `GET /organizer/events` organizer only

### Bookings

- `POST /events/:id/bookings` customer only
- `GET /bookings/me` customer only

## Background jobs

Two background tasks are implemented using Redis and Asynq:

1. Booking confirmation
   Triggered when a customer books tickets.

2. Event update notification
   Triggered when an organizer updates an event.

For this assignment, both jobs simply log messages to the console. You can verify them by checking the worker logs:

```bash
docker compose logs -f worker
```

## Testing

To run the automated tests:

```bash
go test ./...
```

The tests cover:

- registration and login
- role-based route protection
- event creation and update rules
- booking flow and overbooking checks
- background job enqueue behavior

## API testing with Postman

If you want to test the API manually, you can import the included Postman collection from [`event-booking-postman-collection.json`](event-booking-postman-collection.json).

After importing it:

1. Start the project with Docker Compose
2. Run the auth requests first to generate tokens
3. Copy the returned JWT tokens into the collection variables or request headers
4. Create an event as an organizer
5. Book tickets as a customer
6. Update the event and check the worker logs

The collection covers the main happy-path flow for registration, login, event creation, booking, and event updates.

## Notes

- Organizers can only update events they own
- Customers cannot create or edit events
- Organizers cannot book tickets
- Ticket availability is enforced to prevent overbooking
- Event updates notify each booked customer once per update

## Quick verification flow

After starting the stack, a simple manual flow is:

1. Register an organizer
2. Register a customer
3. Create an event as the organizer
4. Book tickets as the customer
5. Update the event as the organizer
6. Check the worker logs to confirm both background jobs ran

This project was designed to keep the code modular and easy to follow while still feeling close to a real backend service.
