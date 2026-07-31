# Odessa

Odessa is a data-origin and dataset platform for organizations, data engineering, and machine-learning workflows. It provides a shared catalog of storage origins, blobs, locations, and scan tasks across authenticated users.

## Architecture

- **API:** Go HTTP server using `net/http`
- **Database:** PostgreSQL with embedded, versioned migrations
- **Storage:** filesystem, S3-compatible storage, and Azure Blob Storage
- **Authentication:** bcrypt password hashing and JWT bearer access tokens

## API

Public endpoints:

- `POST /api/signup`
- `POST /api/login`
- `POST /api/refresh`
- `POST /api/password/reset/request`
- `POST /api/password/reset/confirm`
- `GET /api/health`

Protected endpoints require `Authorization: Bearer <access-token>`:

- `GET /api/v1/origins`
- `POST /api/v1/origins`
- `POST /api/v1/origins/{id}/scan`
- `GET /api/v1/origins/{id}/scan`
- `POST /api/logout`
- `POST /api/password/change`
- `POST /api/account/disable`

Origins and related data are shared resources in the current product model. Authentication is not authorization; organization and role support is planned for a future phase.

## Configuration

Configuration precedence is:

1. CLI flags
2. `ODESSA_*` environment variables
3. YAML configuration file
4. Built-in defaults

### Configuration file

Create `config.yaml`:

```yaml
addr: ":8080"

dsn: "postgres://odessa:change-me@database:5432/odessa?sslmode=disable"

http:
  read_timeout: 10s
  write_timeout: 10s
  idle_timeout: 60s
  shutdown_timeout: 10s
  max_header_bytes: 1048576
  max_request_body_bytes: 1048576

auth:
  jwt_secret: "replace-with-a-long-random-secret"
  access_token_lifetime: 15m
  refresh_token_lifetime: 720h
  reset_token_lifetime: 1h
  password_reset_url: "https://app.example.com/reset-password"

email:
  smtp:
    host: "smtp.example.com"
    port: 587
    username: "smtp-user"
    password: "replace-with-smtp-password"
    from: "no-reply@example.com"

storage:
  fs:
    disabled: false
    root: "/data"

  s3:
    disabled: false
    region: "us-east-1"
    endpoint: "http://storage:9000"

  azure:
    disabled: true
    account: ""
    connection_string: ""
    account_key: ""
    use_default_credential: false
```

HTTP limits are application-level defense in depth. Configure TLS, rate limiting, request filtering, WAF rules, bot protection, and connection controls at the API gateway/reverse proxy. The application still enforces body limits so the service is protected if it is reached through another network path.

### Environment variables

Generate a strong JWT secret on Linux/macOS:

```sh
openssl rand -hex 64
```

```sh
ODESSA_ADDR=:8080
ODESSA_DSN=postgres://odessa:change-me@database:5432/odessa?sslmode=disable
ODESSA_AUTH_JWT_SECRET=REPLACE_WITH_A_LONG_RANDOM_SECRET
ODESSA_AUTH_ACCESS_TOKEN_LIFETIME=15m
ODESSA_AUTH_REFRESH_TOKEN_LIFETIME=720h
ODESSA_AUTH_RESET_TOKEN_LIFETIME=1h
ODESSA_AUTH_PASSWORD_RESET_URL=https://app.example.com/reset-password
ODESSA_EMAIL_SMTP_HOST=smtp.example.com
ODESSA_EMAIL_SMTP_PORT=587
ODESSA_EMAIL_SMTP_USERNAME=smtp-user
ODESSA_EMAIL_SMTP_PASSWORD=REPLACE_WITH_SMTP_PASSWORD
ODESSA_EMAIL_SMTP_FROM=no-reply@example.com

ODESSA_HTTP_READ_TIMEOUT=10s
ODESSA_HTTP_WRITE_TIMEOUT=10s
ODESSA_HTTP_IDLE_TIMEOUT=60s
ODESSA_HTTP_SHUTDOWN_TIMEOUT=10s
ODESSA_HTTP_MAX_HEADER_BYTES=1048576
ODESSA_HTTP_MAX_REQUEST_BODY_BYTES=1048576

ODESSA_STORAGE_FS_ROOT=/data
ODESSA_STORAGE_S3_REGION=us-east-1
ODESSA_STORAGE_S3_ENDPOINT=http://storage:9000

ODESSA_STORAGE_AZURE_ACCOUNT=
ODESSA_STORAGE_AZURE_CONNECTION_STRING=
ODESSA_STORAGE_AZURE_ACCOUNT_KEY=
ODESSA_STORAGE_AZURE_USE_DEFAULT_CREDENTIAL=false
```

Do not use placeholder credentials or `sslmode=disable` outside local development. Store secrets in a deployment secret manager, mounted secret files, or protected environment configuration. Avoid passing secrets as CLI flags because command-line arguments may be visible to other processes.

## Local development with Compose

The Compose file starts PostgreSQL, MinIO, and pgAdmin for local development. Infrastructure ports are exposed for convenience and should be kept on a private network in deployed environments. Replace all development credentials before sharing or deploying the stack.

For MinIO, the AWS SDK can use a profile in `~/.aws/credentials`:

```ini
[minio]
aws_access_key_id = minio
aws_secret_access_key = change-me
```

And `~/.aws/config`:

```ini
[profile minio]
region = us-east-1
```

The corresponding local storage configuration is:

```yaml
storage:
  s3:
    region: us-east-1
    endpoint: http://storage:9000
```

## Authentication lifecycle

Login returns a short-lived access token and a rotating refresh token. Refresh tokens are stored only as hashes, are revoked on use, and are revoked when a password changes or an account is disabled. Logout revokes the submitted refresh token. Password reset requests intentionally return the same response whether or not an account exists.

Password-reset delivery uses SMTP when `email.smtp.host` and `email.smtp.from` are configured. The raw reset token is sent only in the email link; only its hash is persisted. If SMTP is not configured, reset requests are accepted but no email is sent, which is suitable only for local development. Use TLS-capable SMTP credentials and a password-reset URL served by your frontend.

The filesystem backend validates its configured root, resolves the root's symlinks, rejects remote file URI hosts, and prevents existing path components from resolving outside the configured root. Keep the service process unprivileged and do not mount sensitive host directories into the container.

## Database migrations

Migrations are embedded and applied at startup. New authentication tables and the user disablement column are added by migration `007_create_auth_sessions`; migration numbers must never be reused after deployment.

## Build and run

```sh
go build -o bin/odessa-server ./cmd/server
./bin/odessa-server --config config.yaml
```