## Notify Service

gRPC service for user notifications. Stores notifications in PostgreSQL and enforces ownership by `user_id`.

## Responsibility

`notify-service` is responsible for:

- creating notifications for tasks;
- listing notifications for a user;
- marking notifications as read;
- deleting notifications.

It does not publish notifications itself. It only stores and serves them.

## Architecture

The service follows a layered structure:

- `cmd/notify-service`
  entry point;
- `internal/app`
  bootstrap and dependency wiring;
- `internal/domain`
  models and repository contracts;
- `internal/usecase`
  business logic;
- `internal/store/postgre`
  PostgreSQL repository;
- `internal/transport/grpc`
  gRPC handlers;
- `migrations`
  database schema.

Shared infrastructure lives in `common`:

- `common/postgres`
- `common/runtime`
- `common/authsecurity`
- `common/grpcauth`

## API

The protobuf contract is defined in `api-contracts/proto/notify/v1alpha/notify.proto`.

Exposed operations:

| Method | Purpose |
| --- | --- |
| `CreateNotification` | Create notification |
| `ListNotifications` | List notifications |
| `MarkRead` | Mark notification as read |
| `DeleteNotification` | Delete notification |

## Storage Design

PostgreSQL stores notification data.

Why PostgreSQL:

- notification history must be durable;
- indexing by `user_id` is needed for queries;
- explicit SQL keeps behavior transparent.

## Database Schema

Migration: `notify-service/migrations/001_create_notifications.sql`

Table: `notifications`

- `id`
- `user_id`
- `task_id`
- `type`
- `title`
- `body`
- `is_read`
- `created_at`
- `expires_at`

## Configuration

Configuration is provided through environment variables.

Main variables:

- `NOTIFY_GRPC_ADDR`
- `JWT_PUBLIC_KEY_PEM`
- `JWT_ISSUER`
- `JWT_AUDIENCE`
- `DATABASE_URL`

## Local Run

Requirements:

- Go
- Docker / Docker Compose
- `goose`
- RSA public key in PEM format

Start infrastructure:

```bash
cd notify-service
make db-up
make migrate-up
```

Run service:

```bash
cd notify-service
export JWT_PUBLIC_KEY_PEM=/absolute/path/to/public.pem
go run ./cmd/notify-service
```

Defaults:

- PostgreSQL: `localhost:5432`
- gRPC: `:50053`

## Testing

Run:

```bash
GOCACHE=/tmp/go-build go test ./...
```

## Current Limitations

- no push delivery (email/websocket);
- no structured audit logging.

## Summary

Main design choices:

- keep notification logic in use cases;
- enforce ownership in repo operations;
- use explicit SQL with `pgx`;
- keep configuration explicit via env vars.
