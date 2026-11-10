#!/usr/bin/env python3
"""Exercise Catalog HTTP APIs against a disposable, seeded environment only."""

import json
import os
import struct
import urllib.error
import urllib.request
import uuid
import zlib

if os.getenv("CATALOG_E2E_ALLOW_WRITES") != "1":
    raise SystemExit("Set CATALOG_E2E_ALLOW_WRITES=1 only for a disposable environment")
base = os.environ["CATALOG_E2E_BASE_URL"].rstrip("/")
password = os.environ["CATALOG_E2E_PASSWORD"]
owner_email = os.environ.get("CATALOG_E2E_OWNER_EMAIL", "owner@badmintonhub.local")
staff_email = os.environ.get("CATALOG_E2E_STAFF_EMAIL", "staff@badmintonhub.local")
seed_org = "10000000-0000-4000-8000-000000000001"
seed_venue = "20000000-0000-4000-8000-000000000001"
checks = 0


def call(method, path, expected, payload=None, token=None, content_type=None):
    global checks
    headers = {}
    if token:
        headers["Authorization"] = "Bearer " + token
    if payload is not None and not isinstance(payload, bytes):
        payload = json.dumps(payload).encode()
        content_type = "application/json"
    if content_type:
        headers["Content-Type"] = content_type
    req = urllib.request.Request(base + path, data=payload, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=15) as response:
            status, body, response_type = response.status, response.read(), response.headers.get("Content-Type", "")
    except urllib.error.HTTPError as error:
        status, body, response_type = error.code, error.read(), error.headers.get("Content-Type", "")
    if status != expected:
        raise AssertionError(f"{method} {path}: expected HTTP {expected}, got {status}: {body[:400]!r}")
    checks += 1
    if "application/json" in response_type:
        return json.loads(body)
    return body


def require(condition, message):
    global checks
    if not condition:
        raise AssertionError(message)
    checks += 1


def png_image():
    def chunk(kind, content):
        return struct.pack(">I", len(content)) + kind + content + struct.pack(">I", zlib.crc32(kind + content) & 0xFFFFFFFF)
    return b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", struct.pack(">IIBBBBB", 1, 1, 8, 2, 0, 0, 0)) + chunk(b"IDAT", zlib.compress(b"\x00\xff\x00\x00")) + chunk(b"IEND", b"")


def upload(path, token, content, expected):
    boundary = "catalog-" + uuid.uuid4().hex
    body = (f"--{boundary}\r\nContent-Disposition: form-data; name=\"file\"; filename=\"test.png\"\r\nContent-Type: image/png\r\n\r\n".encode() + content + f"\r\n--{boundary}--\r\n".encode())
    return call("POST", path, expected, body, token, "multipart/form-data; boundary=" + boundary)


call("GET", "/api/v1/tenant/organizations", 401)
owner = call("POST", "/api/v1/tenant/auth/login", 200, {"email": owner_email, "password": password})["tokens"]["access_token"]
staff = call("POST", "/api/v1/tenant/auth/login", 200, {"email": staff_email, "password": password})["tokens"]["access_token"]
venue_staff = call("POST", "/api/v1/tenant/auth/login", 200, {"email": "venue-staff@badmintonhub.local", "password": password})["tokens"]["access_token"]
venue_staff_id = "00000000-0000-4000-8000-000000000005"
other_venue = "20000000-0000-4000-8000-000000000002"
scoped = call("GET", "/api/v1/tenant/venues", 200, token=venue_staff)
require([item["id"] for item in scoped.get("items", [])] == [seed_venue], "venue staff must see only assigned venue")
call("GET", "/api/v1/tenant/venues/" + other_venue + "/courts", 403, token=venue_staff)
call("GET", "/api/v1/tenant/venues/" + other_venue + "/courts", 200, token=staff)
assignment = "/api/v1/tenant/venues/" + other_venue + "/staff/" + venue_staff_id
call("PUT", assignment, 403, {"status": "ACTIVE"}, staff)
call("PUT", assignment, 200, {"status": "ACTIVE"}, owner)
scoped = call("GET", "/api/v1/tenant/venues", 200, token=venue_staff)
require({item["id"] for item in scoped.get("items", [])} == {seed_venue, other_venue}, "assigned venue missing")
call("PUT", assignment, 200, {"status": "INACTIVE"}, owner)
call("GET", "/api/v1/tenant/venues/" + other_venue + "/courts", 403, token=venue_staff)
membership = "/api/v1/tenant/organizations/" + seed_org + "/staff/" + venue_staff_id
call("PUT", membership, 200, {"role": "venue_staff", "status": "INACTIVE"}, owner)
scoped = call("GET", "/api/v1/tenant/venues", 200, token=venue_staff)
require(not scoped.get("items", []), "inactive membership must not expose venues")
call("PUT", membership, 200, {"role": "venue_staff", "status": "ACTIVE"}, owner)
public = call("GET", "/api/v1/venues", 200)
require(len(public.get("items", [])) >= 2, "seeded published venues missing")
organizations = call("GET", "/api/v1/tenant/organizations", 200, token=owner)
require(any(x["id"] == seed_org for x in organizations.get("items", [])), "seed organization missing")
call("POST", "/api/v1/tenant/organizations", 403, {"name": "Staff", "slug": "staff-" + uuid.uuid4().hex}, staff)
call("POST", "/api/v1/tenant/organizations", 409, {"name": "Duplicate", "slug": "smash-center"}, owner)
new_org = call("POST", "/api/v1/tenant/organizations", 201, {"name": "E2E organization", "slug": "e2e-org-" + uuid.uuid4().hex}, owner)["organization"]
require(new_org["status"] == "PENDING_VERIFICATION", "organization must start pending verification")
org_path = "/api/v1/tenant/organizations/" + new_org["id"]
call("GET", org_path, 200, token=owner)
call("PATCH", org_path, 200, {"name": "Updated E2E organization"}, owner)
call("POST", org_path + "/venues", 403, {"name": "blocked", "slug": "blocked"}, owner)
call("DELETE", org_path, 204, token=owner)
call("DELETE", "/api/v1/tenant/organizations/" + seed_org, 409, token=owner)
slug = "e2e-venue-" + uuid.uuid4().hex
call("POST", "/api/v1/tenant/organizations/" + seed_org + "/venues", 409, {"name": "Duplicate", "slug": "smash-center"}, owner)
venue = call("POST", "/api/v1/tenant/organizations/" + seed_org + "/venues", 201, {"name": "E2E venue", "slug": slug}, owner)["venue"]
require(venue["status"] == "DRAFT", "venue must start as draft")
venue_id = venue["id"]
venue_path = "/api/v1/tenant/venues/" + venue_id
call("GET", venue_path, 200, token=staff)
call("PATCH", venue_path, 403, {"name": "staff edit"}, staff)
updated = call("PATCH", venue_path, 200, {"name": "Updated E2E venue"}, owner)["venue"]
require(updated["name"] == "Updated E2E venue" and updated["slug"] == slug, "partial venue update changed omitted fields")
call("GET", "/api/v1/venues/" + slug, 404)
court = call("POST", venue_path + "/courts", 201, {"name": "E2E court", "slot_interval_minutes": 30}, owner)["court"]
court_path = "/api/v1/tenant/courts/" + court["id"]
call("GET", court_path, 200, token=staff)
call("PATCH", court_path, 200, {"name": "Updated court", "status": "ACTIVE", "slot_interval_minutes": 60}, owner)
call("DELETE", court_path, 204, token=owner)
png = png_image()
image_path = venue_path + "/images"
upload(image_path, staff, png, 403)
upload(image_path, owner, b"not an image", 400)
upload(image_path, owner, b"x" * (2 * 1024 * 1024 + 1), 400)
image = upload(image_path, owner, png, 201)["image"]
image_id = image["id"]
require(call("GET", image["url"], 200, token=staff) == png, "tenant preview bytes differ")
call("GET", "/api/v1/venue-images/" + image_id, 404)
listed = call("GET", image_path, 200, token=staff)
require(any(item["id"] == image_id for item in listed.get("items", [])), "draft image missing from tenant list")
published = upload("/api/v1/tenant/venues/" + seed_venue + "/images", owner, png, 201)["image"]
require(call("GET", "/api/v1/venue-images/" + published["id"], 200) == png, "public image bytes differ")
call("DELETE", venue_path, 204, token=owner)
call("GET", venue_path, 404, token=owner)
call("GET", image["url"], 404, token=owner)
print(f"Catalog HTTP E2E passed: {checks} assertions")
