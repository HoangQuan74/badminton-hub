# BadmintonHub DevOps

Local development uses the all-in-one Compose stack below. Dev and Production release instructions are in [RELEASE.md](RELEASE.md).

Run the complete local stack from this directory.

## Start

Run these commands from the `devops` directory. The local env file is ignored
by Git and can be created from the example:

```bash
cd devops
cp .env.example .env
```

### Dependencies

Start PostgreSQL, Redis, MongoDB, and Kafka:

```bash
docker compose --env-file .env -f docker/compose.yml up -d postgres redis mongodb kafka
```

Bootstrap the Kafka topics:

```bash
docker compose --env-file .env -f docker/compose.yml up -d kafka-init
```

The four topics are `booking.events`, `payment.events`,
`notification.commands`, and `payout.events`. `kafka-init` exits after the
topics exist; this is expected.

### Identity data

Apply Identity migrations and local development seed:

```bash
docker compose --env-file .env -f docker/compose.yml run --rm migrate
docker compose --env-file .env -f docker/compose.yml run --rm seed
```

Both commands are one-shot jobs and exit after success.

### Backend services

Start Identity:

```bash
docker compose --env-file .env -f docker/compose.yml up --build -d identity
```

Start Identity and the Socket service together:

```bash
docker compose --env-file .env -f docker/compose.yml --profile socket up --build -d identity socket
```

### Gateways

Start the User Gateway:

```bash
docker compose --env-file .env -f docker/compose.yml up --build -d user-gateway
```

Start the Admin Gateway:

```bash
docker compose --env-file .env -f docker/compose.yml --profile admin up --build -d admin-gateway
```

Start the Socket Gateway and its Socket service:

```bash
docker compose --env-file .env -f docker/compose.yml --profile socket up --build -d socket socket-gateway
```

### Web and complete stack

Start the web container after the User Gateway is healthy:

```bash
docker compose --env-file .env -f docker/compose.yml up --build -d web
```

Start every service, including Admin and Socket profiles, with one command:

```bash
docker compose --env-file .env -f docker/compose.yml --profile admin --profile socket up --build -d
```

Check all running and exited containers:

```bash
docker compose --env-file .env -f docker/compose.yml --profile admin --profile socket ps -a
```

Host processes connect through `localhost:19093`; Compose services use
`kafka:9092`. The four initial topics are `booking.events`, `payment.events`,
`notification.commands`, and `payout.events`.

The compose file has safe local defaults. Copy `.env.example` to `.env` only when you want to override ports or credentials.

Open `http://localhost:3000`.

Host ports use a project-specific range to avoid collisions with services already
running on your machine: User Gateway `18080`, Admin Gateway `18082`, Socket
Gateway `18083`, Identity PostgreSQL `15432`, Identity Redis `16379`, Gateway
Redis `16380`, and MongoDB `27018`.

The startup order is PostgreSQL, Redis, MongoDB, Kafka, Identity migrations,
development seed, Identity, User Gateway, Admin Gateway, Socket service, Socket
Gateway, and web. Nginx serves the React application and proxies `/api`,
`/health`, and `/ready` to the User Gateway, so local CORS configuration is not
required.

Business calls target private gRPC service contracts. The current local HTTP
Identity listener is retained only for health checks and the migration bridge;
it is not a public service API.

## Demo account

```text
Email: user@badmintonhub.local
Password: Password@123
```

The seed is idempotent and intended for local development only.

## Common commands

```bash
# Follow all application logs
docker compose --env-file .env -f docker/compose.yml --profile admin --profile socket logs -f

# Follow one service
docker compose --env-file .env -f docker/compose.yml logs -f identity

# Stop containers but keep data
docker compose --env-file .env -f docker/compose.yml --profile admin --profile socket down

# Reset all local data and start again
docker compose --env-file .env -f docker/compose.yml --profile admin --profile socket down -v
docker compose --env-file .env -f docker/compose.yml --profile admin --profile socket up --build -d

# Check API readiness through the web gateway
curl http://localhost:3000/ready
```

Do not reuse the local credentials or JWT secret in a deployed environment.

## Release assets

- `jenkins/Jenkinsfile`: validation, immutable image publishing, Dev deployment, and approved Production deployment.
- `environments/db`: DB VM services.
- `environments/app`: Web VM application services.
- `scripts/deploy.sh`: migration, deployment, and smoke-test orchestration.
- `scripts/rollback.sh`: application image rollback.
- `scripts/backup-db.sh`: PostgreSQL logical backup on the DB VM.
- `scripts/ensure-service-databases.sh`: idempotently prepare the PostgreSQL
  instance for service-owned schemas on an existing DB VM volume.
