# QA Test Suite — HTTP REST API
**Branch:** `008-mcp-ux-improvements` | **Date:** 2026-02-20

All tests use `curl`. Default API port is `8081`.

---

## Setup

Start the API server against the fixture project:

```bash
loko api --project ./test --port 8081 &
export API_PID=$!
export BASE=http://localhost:8081
```

Wait for the server to be ready:

```bash
sleep 1 && curl -sf $BASE/health | jq .status
# Expected: "ok"
```

Teardown after all tests:

```bash
kill $API_PID
```

---

## T01 — GET /health

```bash
curl -s $BASE/health | jq .
```

**Pass criteria:**
- HTTP 200
- `status` = `"ok"`
- `version` is a non-empty string (e.g., `"v0.2.0"`)
- Response in < 100 ms

---

## T02 — GET /health (no auth required)

If an API key is configured (`LOKO_API_KEY`), the health endpoint must still respond without auth:

```bash
curl -s $BASE/health
# Even with LOKO_API_KEY set, this must return 200
```

**Pass criteria:**
- HTTP 200 regardless of whether auth is configured

---

## T03 — GET /api/v1/project

```bash
curl -s $BASE/api/v1/project | jq .
```

**Pass criteria:**
- HTTP 200
- `success` = `true`
- `name` is a non-empty string
- `system_count` ≥ 1
- `container_count` ≥ 1
- `component_count` ≥ 1

---

## T04 — GET /api/v1/systems

```bash
curl -s $BASE/api/v1/systems | jq .
```

**Pass criteria:**
- HTTP 200
- `success` = `true`
- `systems` is a non-empty array
- `total_count` ≥ 1
- Each system object has `id`, `name`, `container_count`

---

## T05 — GET /api/v1/systems (validate shape)

```bash
curl -s $BASE/api/v1/systems | jq '.systems[0]'
```

**Pass criteria:**
- Object has all required fields: `id`, `name`, `description`, `container_count`, `component_count`
- `id` is a slug (lowercase, hyphens only)

---

## T06 — GET /api/v1/systems/{id} (valid system)

```bash
curl -s $BASE/api/v1/systems/notification-service | jq .
```

**Pass criteria:**
- HTTP 200
- `success` = `true`
- `system.id` = `"notification-service"`
- `system.name` = `"Notification Service"` (or similar display name)
- `containers` is a non-empty array
- Each container has `id`, `name`, `technology`, `component_count`

---

## T07 — GET /api/v1/systems/{id} (not found)

```bash
curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/systems/does-not-exist
```

**Pass criteria:**
- HTTP 404
- Response body has `error` field

```bash
curl -s $BASE/api/v1/systems/does-not-exist | jq '.error'
# Expected: non-null string
```

---

## T08 — GET /api/v1/systems/{id} (case sensitivity)

```bash
curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/systems/Notification-Service
```

**Pass criteria:**
- HTTP 404 (IDs are slugs; uppercase path does not match)
- No server crash

---

## T09 — POST /api/v1/build (trigger build)

```bash
curl -s -X POST $BASE/api/v1/build \
  -H "Content-Type: application/json" \
  -d '{"format": "html", "output_dir": "/tmp/loko-qa-api-dist"}' | jq .
```

**Pass criteria:**
- HTTP 202 Accepted
- `success` = `true`
- `build_id` is a non-empty string
- `status` = `"building"` or `"queued"`

**Save the returned `build_id` for T10.**

```bash
BUILD_ID=$(curl -s -X POST $BASE/api/v1/build \
  -H "Content-Type: application/json" \
  -d '{"format": "html", "output_dir": "/tmp/loko-qa-api-dist"}' | jq -r '.build_id')
echo "Build ID: $BUILD_ID"
```

---

## T10 — GET /api/v1/build/{id} (poll status)

```bash
sleep 5  # give the build time to complete
curl -s $BASE/api/v1/build/$BUILD_ID | jq .
```

**Pass criteria:**
- HTTP 200
- `build_id` matches the submitted ID
- `status` is one of: `"building"`, `"complete"`, `"failed"`
- When `status` = `"complete"`: `files_generated` ≥ 1
- When `status` = `"complete"`: `/tmp/loko-qa-api-dist/index.html` exists on disk

---

## T11 — GET /api/v1/build/{id} (not found)

```bash
curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/build/nonexistent-build-id
```

**Pass criteria:**
- HTTP 404
- Response body has `error` field

---

## T12 — GET /api/v1/validate

```bash
curl -s $BASE/api/v1/validate | jq .
```

**Pass criteria:**
- HTTP 200
- `success` = `true`
- `valid` is a boolean
- `issues` is an array (may be empty)
- Each issue (if any) has `code`, `severity`, `message`

---

## T13 — GET /api/v1/validate (structure check)

```bash
curl -s $BASE/api/v1/validate | jq '{valid, error_count, warning_count}'
```

**Pass criteria:**
- `error_count` is a non-negative integer
- `warning_count` is a non-negative integer
- `error_count + warning_count` equals `length of issues array`

---

## T14 — error response shape (404)

```bash
curl -s $BASE/api/v1/systems/zzz-nonexistent | jq .
```

**Pass criteria:**
- Response has `error` field (string)
- Response optionally has `code` field matching `"NOT_FOUND"`
- No HTML error page (pure JSON response)

---

## T15 — CORS headers

```bash
curl -s -I -X OPTIONS \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET" \
  $BASE/api/v1/project
```

**Pass criteria:**
- Response includes `Access-Control-Allow-Origin` header
- No 5xx response code

---

## T16 — Content-Type header

```bash
curl -s -I $BASE/api/v1/project
```

**Pass criteria:**
- `Content-Type` header includes `application/json`

---

## T17 — POST /api/v1/build (empty body)

```bash
curl -s -X POST $BASE/api/v1/build \
  -H "Content-Type: application/json" \
  -d '{}' | jq .
```

**Pass criteria:**
- HTTP 202 OR 400 (either is acceptable — server may accept defaults or reject missing fields)
- No 5xx response
- No server crash

---

## T18 — POST /api/v1/build (invalid JSON)

```bash
curl -s -o /dev/null -w "%{http_code}" \
  -X POST $BASE/api/v1/build \
  -H "Content-Type: application/json" \
  -d 'not-json'
```

**Pass criteria:**
- HTTP 400 (Bad Request)
- No 5xx response
- Server continues serving subsequent requests (no crash)

Verify server is still up:

```bash
curl -s $BASE/health | jq .status
# Expected: "ok"
```

---

## T19 — unsupported method (PUT /api/v1/project)

```bash
curl -s -o /dev/null -w "%{http_code}" \
  -X PUT $BASE/api/v1/project \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Pass criteria:**
- HTTP 405 Method Not Allowed OR 404
- No 5xx response

---

## T20 — unknown endpoint

```bash
curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/does-not-exist
```

**Pass criteria:**
- HTTP 404
- No 5xx response

---

## T21 — concurrent requests

```bash
for i in $(seq 1 5); do
  curl -s $BASE/api/v1/project > /dev/null &
done
wait
echo "All done"
curl -s $BASE/health | jq .status
```

**Pass criteria:**
- All 5 requests complete without error
- Server is still healthy after concurrent load
- No race condition panics in server logs

---

## T22 — GET /api/v1/systems (response time)

```bash
time curl -s $BASE/api/v1/systems > /dev/null
```

**Pass criteria:**
- Response time < 2 seconds for the fixture project (notification-service with ~20 components)

---

## T23 — API with auth (if LOKO_API_KEY is set)

```bash
export LOKO_API_KEY=test-secret-key-qa
loko api --project ./test --port 18082 &
API2_PID=$!
sleep 1

# Without key — should be rejected:
curl -s -o /dev/null -w "%{http_code}" http://localhost:18082/api/v1/project

# With key — should succeed:
curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer test-secret-key-qa" \
  http://localhost:18082/api/v1/project

# Health endpoint — no key required:
curl -s -o /dev/null -w "%{http_code}" http://localhost:18082/health

kill $API2_PID
unset LOKO_API_KEY
```

**Pass criteria:**
- Without key: HTTP 401
- With key: HTTP 200
- Health: HTTP 200 (no auth required)

> **Skip this test** if `LOKO_API_KEY` is not documented as supported in the current build.

---

## T24 — GET /api/v1/systems/{id} (all fixture systems)

Verify every system from the fixture project is accessible:

```bash
SYSTEMS=$(curl -s $BASE/api/v1/systems | jq -r '.systems[].id')
for SYS in $SYSTEMS; do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/systems/$SYS)
  echo "$SYS: $STATUS"
done
```

**Pass criteria:**
- Every system returns HTTP 200
- No 500 errors

---

## T25 — recovery after panic simulation

```bash
# Send a request that targets an extreme path depth:
curl -s -o /dev/null -w "%{http_code}" \
  "$BASE/api/v1/systems/a/b/c/d/e/f/g/h"

# Server must still respond normally:
curl -s $BASE/health | jq .status
```

**Pass criteria:**
- Deep path: HTTP 404 (not 500)
- Health check still returns `"ok"` (recovery middleware active)

---

## Cleanup

```bash
kill $API_PID 2>/dev/null
rm -rf /tmp/loko-qa-api-dist
```
