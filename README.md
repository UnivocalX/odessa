# Odessa
Data Catlog

# Configuration

Odessa can be configured using:

- CLI flags (highest priority)
- Environment variables (ODESSA_*)
- YAML configuration file

## Configuration File Example

Create config.yaml:

```yaml
addr: ":8080"

dsn: "postgres://odessa:change-me@database:5432/odessa?sslmode=disable"

auth:
  jwt_secret: "replace-with-a-long-random-secret"

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

## Environment Variables Example

**Generate a JWT Secret, Linux/macOS: `openssl rand -hex 64`**

```sh
###############################################################################
# PostgreSQL
###############################################################################

POSTGRES_DB=odessa
POSTGRES_USER=odessa
POSTGRES_PASSWORD=change-me

###############################################################################
# MinIO
###############################################################################

MINIO_ROOT_USER=minio
MINIO_ROOT_PASSWORD=change-me

###############################################################################
# PgAdmin
###############################################################################

PGADMIN_DEFAULT_EMAIL=admin@example.com
PGADMIN_DEFAULT_PASSWORD=change-me

###############################################################################
# Odessa
###############################################################################

ODESSA_ADDR=:8080

ODESSA_DSN=postgres://odessa:change-me@database:5432/odessa?sslmode=disable

ODESSA_AUTH_JWT_SECRET=REPLACE_WITH_A_LONG_RANDOM_SECRET

ODESSA_STORAGE_FS_ROOT=/data

ODESSA_STORAGE_S3_REGION=us-east-1
ODESSA_STORAGE_S3_ENDPOINT=http://storage:9000

ODESSA_STORAGE_AZURE_ACCOUNT=
ODESSA_STORAGE_AZURE_CONNECTION_STRING=
ODESSA_STORAGE_AZURE_ACCOUNT_KEY=
ODESSA_STORAGE_AZURE_USE_DEFAULT_CREDENTIAL=false
```

## AWS Credentials File for MinIO

If using the AWS SDK with MinIO, create ~/.aws/credentials:

[minio]
aws_access_key_id = minio
aws_secret_access_key = change-me


and `~/.aws/config` :

```sh
[profile minio]
region = us-east-1
```

Then configure Odessa:

```yaml
storage:
  s3:
    region: us-east-1
    endpoint: http://storage:9000
```

or:

```sh
ODESSA_STORAGE_S3_REGION=us-east-1
ODESSA_STORAGE_S3_ENDPOINT=http://storage:9000
```

The AWS SDK will use the minio profile while communicating with the local MinIO server.


## Start Odessa: `odessa --config config.yaml`