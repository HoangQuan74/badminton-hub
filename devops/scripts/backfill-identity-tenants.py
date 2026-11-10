#!/usr/bin/env python3
"""Copy a frozen Catalog tenant snapshot into Identity without overwriting data."""

import argparse
import hashlib
import os
from pathlib import Path
import shutil
import subprocess
import sys
from urllib.parse import parse_qsl, unquote, urlsplit

MAX_EXPORT_BYTES = 50 * 1024 * 1024
MAX_ROWS_PER_TABLE = 100_000
SQL_DIR = Path(__file__).resolve().parent.parent / "sql" / "tenant-backfill"
OPTIONS = {"sslmode": "PGSSLMODE", "sslrootcert": "PGSSLROOTCERT", "sslcert": "PGSSLCERT", "sslkey": "PGSSLKEY", "connect_timeout": "PGCONNECT_TIMEOUT"}


def connection_env(variable):
    value = os.environ.get(variable, "")
    parsed = urlsplit(value)
    if parsed.scheme not in ("postgres", "postgresql") or not parsed.hostname or not parsed.username or not parsed.path.strip("/"):
        raise ValueError(f"{variable} must be a PostgreSQL URL with host, user, and database")
    env = {k: v for k, v in os.environ.items() if not k.startswith("PG") and k not in ("SOURCE_CATALOG_DATABASE_URL", "TARGET_IDENTITY_DATABASE_URL")}
    env.update(PGHOST=parsed.hostname, PGPORT=str(parsed.port or 5432), PGDATABASE=unquote(parsed.path.lstrip("/")), PGUSER=unquote(parsed.username), PGCONNECT_TIMEOUT="10")
    if parsed.password is not None:
        env["PGPASSWORD"] = unquote(parsed.password)
    for key, item in parse_qsl(parsed.query, keep_blank_values=True):
        if key not in OPTIONS:
            raise ValueError(f"unsupported PostgreSQL URL option: {key}")
        env[OPTIONS[key]] = item
    return env


def psql(env, sql, transactional=False):
    cmd = ["psql", "-X", "-q", "-A", "-t", "-v", "ON_ERROR_STOP=1"]
    if transactional:
        cmd.append("--single-transaction")
    result = subprocess.run(cmd, input=sql, stdout=subprocess.PIPE, stderr=subprocess.PIPE, env=env)
    if result.returncode:
        raise RuntimeError("psql failed: " + result.stderr.decode(errors="replace").strip())
    return result.stdout


def export_snapshot(env):
    outputs = {}
    for name in ("organizations", "members", "venues"):
        data = psql(env, (SQL_DIR / f"{name}.sql").read_bytes())
        if len(data) > MAX_EXPORT_BYTES or len(data.splitlines()) > MAX_ROWS_PER_TABLE:
            raise RuntimeError(f"{name} exceeds the bounded transfer limit; use a planned batch migration")
        if any(line == b"\\." for line in data.splitlines()):
            raise RuntimeError(f"unsafe COPY terminator found in {name} export")
        outputs[name] = data
    return outputs


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--apply", action="store_true", help="write to Identity and mark Catalog ready")
    parser.add_argument("--confirm-all-tenant-writes-stopped", action="store_true", help="confirm tenant and venue writers in both services are stopped")
    args = parser.parse_args()
    if not args.apply or not args.confirm_all_tenant_writes_stopped:
        parser.error("both --apply and --confirm-all-tenant-writes-stopped are required")
    if not shutil.which("psql"):
        raise RuntimeError("psql is required; install a PostgreSQL client before running this tool")
    source = connection_env("SOURCE_CATALOG_DATABASE_URL")
    target = connection_env("TARGET_IDENTITY_DATABASE_URL")
    state = psql(source, b"SELECT ready FROM catalog.identity_cutover_state WHERE singleton = TRUE;\n").strip()
    if state == b"t":
        print("Catalog cutover is already marked ready; no data was changed.")
        return
    if state != b"f":
        raise RuntimeError("Catalog cutover state is missing or invalid; apply Catalog migration 000005 first")
    snapshot = export_snapshot(source)
    template = (SQL_DIR / "import.sql").read_bytes()
    for name, data in snapshot.items():
        marker = f"__{name.upper()}_CSV__".encode()
        if template.count(marker) != 1:
            raise RuntimeError(f"invalid import template marker for {name}")
        template = template.replace(marker + b"\n", data if not data or data.endswith(b"\n") else data + b"\n")
    psql(target, template, transactional=True)
    current = export_snapshot(source)
    if any(hashlib.sha256(snapshot[name]).digest() != hashlib.sha256(current[name]).digest() for name in snapshot):
        raise RuntimeError("legacy source changed during transfer; Catalog remains blocked and requires reconciliation")
    result = psql(source, b"UPDATE catalog.identity_cutover_state SET ready = TRUE, verified_at = NOW() WHERE singleton = TRUE AND ready = FALSE RETURNING ready;\n").strip()
    if result != b"t":
        raise RuntimeError("transfer verified but Catalog readiness was not updated; retry after investigating")
    print("Identity tenant backfill verified:", ", ".join(f"{name}={len(data.splitlines())}" for name, data in snapshot.items()))


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, ValueError, OSError) as error:
        print(f"tenant backfill failed: {error}", file=sys.stderr)
        sys.exit(1)
