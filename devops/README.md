# BadmintonHub DevOps

Local development uses the all-in-one Compose stack below. Dev and Production release instructions are in [RELEASE.md](RELEASE.md).

Run the complete local stack from this directory.

## Start

```bash
cd devops
docker compose -f docker/compose.yml up --build -d
docker compose -f docker/compose.yml ps
```

The compose file has safe local defaults. Copy `.env.example` to `.env` only when you want to override ports or credentials.

Open `http://localhost:3000`.

Host ports use a project-specific range to avoid collisions with services already
running on your machine: API `18080`, Identity PostgreSQL `15432`, Identity
Redis `16379`, Gateway Redis `16380`, and MongoDB `27018`.

The startup order is the service-owned PostgreSQL and Redis dependencies,
MongoDB, Identity migrations, development seed, Identity, Gateway API, and
web. Nginx serves the React application and proxies `/api`, `/health`, and
`/ready` to the Gateway API, so local CORS configuration is not required. The
Gateway forwards the current Auth routes to the private Identity service.

## Demo account

```text
Email: user@badmintonhub.local
Password: Password@123
```

The seed is idempotent and intended for local development only.

## Common commands

```bash
# Follow application logs
docker compose -f docker/compose.yml logs -f identity api web

# Stop containers but keep data
docker compose -f docker/compose.yml down

# Reset all local data and start again
docker compose -f docker/compose.yml down -v
docker compose -f docker/compose.yml up --build -d

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
- `scripts/ensure-service-databases.sh`: idempotently create service databases
  on an existing DB VM volume.
