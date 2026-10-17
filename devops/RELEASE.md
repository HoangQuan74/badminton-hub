# Release Runbook

## Topology

```text
Internet
   |
Web VM: Nginx/React + public Gateway + private Identity
   |
private network only
   |
DB VM: PostgreSQL + Redis + MongoDB
```

The Web VM is the only public application host. Bind database ports to the DB VM private IP and allow inbound traffic only from the Web VM private IP. Do not deploy the local seed job outside local development.

## Branch and release policy

- Pull requests run validation only.
- `develop` builds immutable images, pushes them to GHCR, and deploys Dev automatically.
- A Git tag matching `v*` builds the same immutable images and pauses for production approval.
- `main` validates and publishes images but does not deploy by itself.
- Every deployment uses the Git SHA as `IMAGE_TAG`; do not deploy `latest`.

## Bootstrap the DB VM

Install Docker Engine and the Compose plugin, then provision:

```bash
sudo mkdir -p /opt/badminton-hub/db/devops/environments/db
sudo chown -R "$USER":"$USER" /opt/badminton-hub
```

Copy `devops/environments/db/compose.yml` and create `/opt/badminton-hub/db/.env` from the matching example. Set `DB_BIND_ADDRESS` to the DB VM private IP.

```bash
cd /opt/badminton-hub/db
docker compose --project-name badminton-hub-db --env-file .env -f devops/environments/db/compose.yml up -d
docker compose --project-name badminton-hub-db --env-file .env -f devops/environments/db/compose.yml ps
devops/scripts/ensure-service-databases.sh
```

Firewall rules must allow TCP `5432`, `6379`, and `27017` from the Web VM private IP only. Keep SSH restricted to the administrator or bastion address.

## Bootstrap each Web VM environment

Install Docker Engine, the Compose plugin, `curl`, and the Jenkins deployment public key.

```bash
sudo mkdir -p /opt/badminton-hub/dev /opt/badminton-hub/prod
sudo chown -R "$USER":"$USER" /opt/badminton-hub
```

Create `/opt/badminton-hub/dev/.env` from `.env.example` and `/opt/badminton-hub/prod/.env` from `.env.prod.example`. The examples isolate each environment's PostgreSQL database, service-owned schemas, Redis namespace, MongoDB database, JWT secrets, and web ports. Jenkins intentionally does not overwrite these secret files.

Log in to GHCR once on the Web VM if the images are private:

```bash
docker login ghcr.io
```

## Jenkins configuration

Required plugins: Pipeline, Docker Pipeline, SSH Agent, and Workspace Cleanup.

Configure these credentials:

- `ghcr-credentials`: GHCR username and token with package write/read access.
- `badminton-hub-dev-ssh`: SSH private key for the Dev Web VM.
- `badminton-hub-prod-ssh`: SSH private key for the Production Web VM.

Configure global environment variables:

- `DEV_DEPLOY_HOST`, for example `deploy@10.0.0.20`.
- `PROD_DEPLOY_HOST`, for example `deploy@10.0.0.21`.

Create a Multibranch Pipeline, set **Script Path** to `devops/jenkins/Jenkinsfile`, and set `REGISTRY_NAMESPACE` to the lowercase GHCR owner.

## Deployment behavior

Jenkins uploads only deployment manifests, scripts, and the migrations owned by
the service being deployed. On the Web VM, `deploy.sh`:

1. Pulls the exact Identity, API, and Web image tags.
2. Runs forward migrations against the service-scoped database URL and schema as a
   one-shot container.
3. Starts Identity, API, and Web services in dependency order.
4. Waits for `/ready` and checks the web root.
5. Records the current and previous application image tags.

Rollback on the Web VM:

```bash
cd /opt/badminton-hub/prod
devops/scripts/rollback.sh prod
```

Rollback changes application images only. It never runs migration `down`; schema changes must remain backward-compatible with the previous application release.

## Backup

Run on the DB VM before production schema changes and on a schedule. Pass the
environment database containing the service schema being changed:

```bash
cd /opt/badminton-hub/db
devops/scripts/backup-db.sh badminton_identity_prod
```

When moving an existing environment from former shared tables, take a verified
backup and perform an explicit Identity data migration into the `identity`
schema before enforcing the new boundary. The deployment does not copy data
automatically.

Copy backups to storage outside the DB VM and periodically test restoration. A backup that has never been restored is not a verified recovery mechanism.

## Kubernetes decision

Kubernetes can run on two Linux VMs, but the current `DB VM + Web VM` topology is not a meaningful highly available cluster:

- A single-node K3s installation on the Web VM gives Kubernetes deployment practice but no control-plane or workload HA.
- Making the DB VM a K3s worker mixes stateful database and application workloads, weakens isolation, and still leaves a single control plane.
- Running PostgreSQL, Redis, and MongoDB inside that two-node cluster adds storage and recovery complexity without improving availability.

Recommended learning path:

1. Release the two-VM Compose design first and prove CI/CD, migration, backup, smoke test, and rollback.
2. Install single-node K3s on the Web VM and move only stateless API/Web workloads into it.
3. Keep databases on the DB VM as external services.
4. Add a third VM before claiming Kubernetes HA; use three server/control-plane nodes or a managed control plane for a production design.

The Docker images and environment contracts in this repository are reusable when moving from Compose to K3s. Keep Identity private and expose only the Gateway through the Web VM.
