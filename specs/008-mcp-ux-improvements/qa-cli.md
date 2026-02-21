# QA Test Suite — CLI Commands
**Branch:** `008-mcp-ux-improvements` | **Date:** 2026-02-20

All tests assume:
- `loko` is on `$PATH` (`task install` was run)
- Working directory is the repo root unless specified otherwise
- The fixture project at `./test` is available

---

## Setup

```bash
loko --version
# Expected: loko vX.Y.Z (commit: ..., built: ...)
```

Create a clean scratch directory for write tests:

```bash
mkdir -p /tmp/loko-qa
```

---

## T01 — version

```bash
loko --version
loko version
```

**Pass criteria:**
- Both print a version string in the format `loko vX.Y.Z (commit: ..., built: ...)`
- Exit code 0

---

## T02 — help

```bash
loko --help
loko new --help
loko build --help
```

**Pass criteria:**
- Each prints a usage summary listing available flags/subcommands
- Exit code 0

---

## T03 — init (new project)

```bash
loko init qa-project \
  --path /tmp/loko-qa/qa-project \
  --description "QA test project"
```

**Pass criteria:**
- Exit code 0
- `/tmp/loko-qa/qa-project/loko.toml` exists
- `/tmp/loko-qa/qa-project/src/` directory exists
- `loko.toml` contains `name = "qa-project"` (or `"QA Project"` with display name)

**Cleanup:** All subsequent tests in this file use `/tmp/loko-qa/qa-project` as `PROJECT`.

```bash
export PROJECT=/tmp/loko-qa/qa-project
```

---

## T04 — new system

```bash
loko new system "Notification Service" \
  --project $PROJECT \
  --description "Sends email and SMS notifications" \
  --technology "Go + AWS Lambda"
```

**Pass criteria:**
- Exit code 0
- `$PROJECT/src/notification-service/system.md` exists
- `system.md` frontmatter contains `name: "Notification Service"`
- Output (stdout) confirms creation with the slug ID `notification-service`

---

## T05 — new system (invalid name — starts with hyphen)

```bash
loko new system "-BadName" --project $PROJECT 2>&1; echo "exit: $?"
```

**Pass criteria:**
- Exit code non-zero
- stderr contains a validation error message

---

## T06 — new container

```bash
loko new container "API Gateway" \
  --project $PROJECT \
  --parent "Notification Service" \
  --description "HTTP entry point" \
  --technology "Go + Fiber"
```

**Pass criteria:**
- Exit code 0
- `$PROJECT/src/notification-service/api-gateway/container.md` exists
- `$PROJECT/src/notification-service/api-gateway/api-gateway.d2` exists (D2 diagram created)
- Output confirms `api-gateway` created under `notification-service`

---

## T07 — new container (parent not found)

```bash
loko new container "Some Container" \
  --project $PROJECT \
  --parent "Nonexistent System" 2>&1; echo "exit: $?"
```

**Pass criteria:**
- Exit code non-zero
- Error mentions `"not found"` or `"nonexistent-system"`

---

## T08 — new component

```bash
loko new component "Email Sender" \
  --project $PROJECT \
  --parent "API Gateway" \
  --description "Dispatches email via SES" \
  --technology "Go"
```

**Pass criteria:**
- Exit code 0
- `$PROJECT/src/notification-service/api-gateway/email-sender/component.md` exists
- Output confirms `email-sender` created

---

## T09 — new component (--preview flag)

```bash
loko new component "SMS Sender" \
  --project $PROJECT \
  --parent "API Gateway" \
  --description "Dispatches SMS via SNS" \
  --preview
```

**Pass criteria:**
- Exit code 0
- Component created on disk
- Output includes a D2 diagram preview snippet (the `--preview` flag generates a diagram preview)

---

## T10 — new container (second container, for relationship tests)

```bash
loko new container "Message Queue" \
  --project $PROJECT \
  --parent "Notification Service" \
  --description "SQS queues for async delivery" \
  --technology "AWS SQS"
```

**Pass criteria:**
- Exit code 0
- `$PROJECT/src/notification-service/message-queue/container.md` exists

---

## T11 — validate (clean project)

```bash
loko validate --project $PROJECT
```

**Pass criteria:**
- Exit code 0
- Output mentions `"valid"` or `"no errors"`
- No crashes

---

## T12 — validate --strict

```bash
loko validate --project $PROJECT --strict
```

**Pass criteria:**
- Runs without crashing
- If warnings exist, exit code is non-zero when `--strict` is set
- If no warnings, exit code 0

---

## T13 — validate --exit-code

```bash
loko validate --project $PROJECT --exit-code; echo "exit: $?"
```

**Pass criteria:**
- The process exits with code 1 if validation found issues, 0 if clean
- Output is human-readable (not a crash/panic)

---

## T14 — build (HTML output)

```bash
loko build \
  --project $PROJECT \
  --output /tmp/loko-qa/dist-html \
  --format html
```

**Pass criteria:**
- Exit code 0
- `/tmp/loko-qa/dist-html/index.html` exists
- `/tmp/loko-qa/dist-html/systems/` directory exists

---

## T15 — build (multiple formats)

```bash
loko build \
  --project $PROJECT \
  --output /tmp/loko-qa/dist-multi \
  --format html,markdown
```

**Pass criteria:**
- Exit code 0
- Both `index.html` and at least one `.md` file exist in `/tmp/loko-qa/dist-multi/`

---

## T16 — build --clean

```bash
loko build \
  --project $PROJECT \
  --output /tmp/loko-qa/dist-html \
  --clean
```

**Pass criteria:**
- Exit code 0
- Rebuilds from scratch without error (previous output directory is cleaned first)

---

## T17 — build (fixture project — full render)

```bash
loko build \
  --project ./test \
  --output /tmp/loko-qa/dist-fixture
```

**Pass criteria:**
- Exit code 0
- `index.html` exists
- At least one SVG or diagram file is present under `dist-fixture/diagrams/`

**Cleanup:** `rm -rf /tmp/loko-qa/dist-fixture`

---

## T18 — watch (start and interrupt)

```bash
timeout 3 loko watch --project $PROJECT --output /tmp/loko-qa/dist-watch 2>&1; echo "exit: $?"
```

**Pass criteria:**
- Process starts and outputs a "watching" or "ready" message before timeout
- Exits cleanly after interrupt (exit code 124 from `timeout` is expected; no panic output)

---

## T19 — export html

```bash
loko export html \
  --project $PROJECT \
  --output /tmp/loko-qa/export-html
```

**Pass criteria:**
- Exit code 0
- `/tmp/loko-qa/export-html/index.html` exists

---

## T20 — export markdown

```bash
loko export markdown \
  --project $PROJECT \
  --output /tmp/loko-qa/export-md
```

**Pass criteria:**
- Exit code 0
- At least one `.md` file exists under `/tmp/loko-qa/export-md/`

---

## T21 — serve (start and interrupt)

```bash
timeout 3 loko serve \
  --project $PROJECT \
  --output /tmp/loko-qa/dist-html \
  --port 18080 2>&1; echo "exit: $?"
```

**Pass criteria:**
- Process starts (listen message visible)
- Exits cleanly on interrupt
- No panic output

---

## T22 — api (start and interrupt)

```bash
timeout 3 loko api \
  --project $PROJECT \
  --port 18081 2>&1; echo "exit: $?"
```

**Pass criteria:**
- Process starts (listen message visible on port 18081)
- Exits cleanly on interrupt
- No panic output

---

## T23 — mcp (start and interrupt)

```bash
timeout 2 loko mcp --project $PROJECT 2>&1; echo "exit: $?"
```

**Pass criteria:**
- A blank line appears on stderr (MCP ready signal)
- No panic output on interrupt

---

## T24 — completion scripts

```bash
loko completion bash 2>&1 | head -5; echo "exit: $?"
loko completion zsh 2>&1 | head -5; echo "exit: $?"
```

**Pass criteria:**
- Each prints shell completion script content (starts with `#` comments or `compdef`)
- Exit code 0 for both

---

## T25 — project flag override (--project vs cwd)

```bash
# From a directory that is NOT the project root:
cd /tmp && loko validate --project $PROJECT
```

**Pass criteria:**
- Exit code 0
- Validates the correct project regardless of `cwd`
- No "project not found" error

---

## T26 — build with D2 theme

```bash
loko build \
  --project $PROJECT \
  --output /tmp/loko-qa/dist-theme \
  --d2-theme dark-mauve
```

**Pass criteria:**
- Exit code 0
- Output produced (theme applied without crash)

---

## T27 — full lifecycle (smoke test)

Run the following in order as a single smoke test:

```bash
export PROJECT=/tmp/loko-qa/smoke-$$

loko init smoke-test --path $PROJECT
loko new system "Backend" --project $PROJECT
loko new container "API" --project $PROJECT --parent "Backend"
loko new component "Handler" --project $PROJECT --parent "API"
loko validate --project $PROJECT
loko build --project $PROJECT --output $PROJECT/dist

echo "Files in dist:"
ls $PROJECT/dist/
```

**Pass criteria:**
- Every command exits 0
- `$PROJECT/dist/index.html` exists at the end
- No panic or stack trace at any step

**Cleanup:** `rm -rf $PROJECT`

---

## Cleanup (all tests)

```bash
rm -rf /tmp/loko-qa
```
