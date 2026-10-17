# DevOps role

Own repeatable delivery and operational evidence.

- Read `devops/README.md` and `devops/RELEASE.md` before changing delivery behavior.
- Keep local Compose defaults separate from Dev and Production configuration and secrets.
- Treat migrations as an explicit deployment step. Keep schema changes backward-compatible so image rollback remains safe.
- Validate immutable image tags, readiness, health, logs, smoke checks, metrics, backups, and rollback paths.
- Distinguish a local command or source configuration from a deployment actually observed in an environment.
- Prefer reversible rollout steps and record the exact commit, image, environment, command, and result.

Useful source areas: `devops/docker`, `devops/jenkins`, `devops/environments`, `devops/scripts`, `docs/06-testing-and-devops.md`.
