# Quickstart & Test Scenarios: loko v0.1.0

**Created**: 2025-12-17  
**Status**: Ready for Integration Testing

---

## Quickstart Tutorial

This 5-minute walkthrough should be completable by any developer to verify all core functionality works.

### Prerequisites

```bash
# Check Go version
go version  # Should be 1.23+

# Check d2 installed
d2 --version

# Clone loko
git clone https://github.com/madstone-tech/loko
cd loko

# Build loko binary
make build  # Produces ./loko binary
```

### Step 1: Initialize Project (1 minute)

```bash
# Create new architecture project
./loko init my-architecture

# Follow interactive prompts:
# > Project name: my-architecture
# > Description: My sample architecture
# > Author: Your Name

# Verify structure created
ls -la my-architecture/
# Expected: loko.toml, src/ directory
```

### Step 2: Scaffold System (1 minute)

```bash
cd my-architecture

# Create first system
../loko new system PaymentService

# Create second system
../loko new system AuthService

# Verify files
find src/ -name "*.md"
# Expected: src/PaymentService/system.md, src/AuthService/system.md
```

### Step 3: Add Containers (1 minute)

```bash
# Add containers to Payment Service
../loko new container PaymentService API
../loko new container PaymentService Database

# Add containers to Auth Service
../loko new container AuthService APIGateway
../loko new container AuthService TokenStore

# Verify structure
tree src/
```

### Step 4: Build Documentation (1 minute)

```bash
# Build HTML documentation
../loko build

# Verify output
ls -la dist/
# Expected: index.html, diagrams/, css/

# Serve documentation
../loko serve

# Open browser to http://localhost:8080
# Expected: HTML site with both systems visible, sidebar navigation
```

### Step 5: Watch Mode (1 minute)

```bash
# Start watch mode (background)
../loko watch &

# Edit system.md file
vim src/PaymentService/system.md
# Change description, save

# Check browser (should auto-refresh)
# Expected: Updated content visible without manual rebuild
```

---

## Acceptance Test Scenarios

These scenarios verify that each user story works end-to-end.

---

### US-1: LLM-Driven Architecture Design

**Test**: Claude Desktop can design a 3-system architecture via MCP

**Precondition**: loko MCP server running

```bash
# Start MCP server
./loko mcp
# Expected output: MCP server listening
```

**Test Steps**:

1. **Open Claude Desktop**
   - Settings → Developer → Enable Claude for Desktop Dev

2. **Add loko MCP to config**
   ```json
   {
     "mcpServers": {
       "loko": {
         "command": "./loko",
         "args": ["mcp"],
         "cwd": "/path/to/my-architecture"
       }
     }
   }
   ```

3. **Start design conversation**
   ```
   Claude Prompt:
   "I'm building a microservices architecture with 3 main systems:
    - PaymentService (handles payments)
    - UserService (manages user accounts)
    - NotificationService (sends emails/SMS)
    
   Each system has an API, a database, and internal services.
   
   Can you create this structure and show me the diagram?"
   ```

4. **Claude uses loko tools**
   - Calls `create_system` for PaymentService
   - Calls `create_system` for UserService
   - Calls `create_system` for NotificationService
   - Calls `create_container` for each system's containers
   - Calls `update_diagram` to add D2 diagrams
   - Calls `build_docs` to generate HTML

5. **Verify output**
   - All 3 systems created
   - 9 containers created (3 per system)
   - Diagrams rendered and visible in HTML
   - Documentation site navigable

**Expected Result**: ✅ PASS
- 3 systems and their containers exist
- D2 diagrams rendered
- HTML docs show all architecture
- No human intervention needed

---

### US-2: Direct File Editing Workflow

**Test**: Developer can edit files and see live updates

**Setup**:

```bash
cd my-architecture
../loko watch &  # Start watch mode
../loko serve &  # Start web server
```

**Test Steps**:

1. **Open browser to http://localhost:8080**
   - See current PaymentService documentation

2. **Edit system.md**
   ```bash
   # Open in editor
   vim src/PaymentService/system.md
   
   # Change description (e.g., add a sentence)
   # Save file (Ctrl+S or :wq)
   ```

3. **Watch rebuild trigger**
   - Check terminal: "Rebuilding src/PaymentService/..."
   - Measure time from save to refresh: **should be <500ms**

4. **Browser auto-refreshes**
   - No manual refresh needed
   - Updated text visible immediately

5. **Edit D2 diagram**
   ```bash
   vim src/PaymentService/system.d2
   
   # Change: Add new node or connection
   # Example: Add "MessageQueue -> Database"
   # Save
   ```

6. **Verify diagram update**
   - Diagram re-rendered
   - Browser shows updated SVG
   - <500ms latency

**Expected Result**: ✅ PASS
- File changes detected immediately
- HTML updated within 500ms
- Browser auto-refreshes
- No manual commands needed

---

### US-3: Project Scaffolding

**Test**: Quick scaffold with templates

**Setup**:

```bash
rm -rf test-project/
./loko init test-project
```

**Test Steps**:

1. **Run init command**
   ```bash
   ./loko init test-project
   
   # Provide:
   # - Project name: test-project
   # - Description: Test scaffolding
   # - Author: Tester
   ```

2. **Verify project structure**
   ```bash
   ls -la test-project/
   
   # Expected:
   # - loko.toml (with valid config)
   # - src/ (directory)
   # - README.md
   ```

3. **Create system from template**
   ```bash
   cd test-project
   ../loko new system SampleAPI
   ```

4. **Verify template applied**
   ```bash
   cat src/SampleAPI/system.md
   
   # Expected: Valid markdown with frontmatter
   # - Contains: name, description
   # - Contains: H1 heading
   ```

5. **Create container from template**
   ```bash
   ../loko new container SampleAPI Web
   ```

6. **Verify container structure**
   ```bash
   find src/SampleAPI -name "*.md"
   
   # Expected:
   # - src/SampleAPI/system.md
   # - src/SampleAPI/Web/container.md
   ```

7. **Use custom template**
   ```bash
   mkdir -p .loko/templates/custom-system
   cp ../templates/system.md .loko/templates/custom-system/
   # Edit custom template
   ../loko new system CustomSystem --template custom-system
   ```

8. **Verify custom template used**
   ```bash
   cat src/CustomSystem/system.md
   # Should match custom template
   ```

**Expected Result**: ✅ PASS
- Project structure created with loko.toml
- Systems and containers scaffold from templates
- Custom templates discovered and used
- Files follow C4 conventions

---

### US-4: API Integration

**Test**: CI/CD script can trigger builds via HTTP

**Setup**:

```bash
# Start API server
./loko api --port 8081 &

# Wait for server ready
sleep 2
```

**Test Steps**:

1. **Query project via API**
   ```bash
   curl -X GET http://localhost:8081/api/v1/systems
   
   # Expected: JSON listing all systems
   # Response:
   # {
   #   "systems": [
   #     {"id": "PaymentService", "name": "PaymentService"},
   #     ...
   #   ]
   # }
   ```

2. **Trigger build via API**
   ```bash
   curl -X POST http://localhost:8081/api/v1/build \
     -H "Content-Type: application/json" \
     -d '{"format": "html"}'
   
   # Expected: Build initiated
   # Response:
   # {
   #   "status": "building",
   #   "build_id": "abc123"
   # }
   ```

3. **Check build status**
   ```bash
   curl -X GET http://localhost:8081/api/v1/build/abc123
   
   # Expected: Build complete
   # Response:
   # {
   #   "status": "complete",
   #   "duration_ms": 2847,
   #   "output_dir": "dist/"
   # }
   ```

4. **Query validation**
   ```bash
   curl -X GET http://localhost:8081/api/v1/validate
   
   # Expected: Validation results
   # Response:
   # {
   #   "errors": [],
   #   "warnings": [...]
   # }
   ```

5. **Test auth (v0.1.0 foundation only)**
   ```bash
   # Without API key
   curl -X POST http://localhost:8081/api/v1/build
   
   # Expected: If auth enabled, 401 error
   # Otherwise: Works (auth optional in v0.1.0)
   ```

**Expected Result**: ✅ PASS
- Query endpoints return JSON
- Build API triggers rebuild
- Status API shows progress
- Validation API works
- CI/CD integration ready

---

### US-5: Multi-Format Export

**Test**: Export to multiple formats

**Setup**:

```bash
# Build project first
./loko build
```

**Test Steps**:

1. **Export to HTML (default)**
   ```bash
   ./loko build --format html
   
   # Expected: dist/index.html created
   ls -la dist/index.html
   ```

2. **Verify HTML quality**
   ```bash
   # Open in browser
   open dist/index.html
   
   # Expected:
   # - Navigable sidebar
   # - Breadcrumbs working
   # - Diagrams displayed
   # - Search functional
   # - Mobile-friendly responsive design
   ```

3. **Export to Markdown**
   ```bash
   ./loko build --format markdown
   
   # Expected: dist/README.md created
   ls -la dist/README.md
   wc -l dist/README.md  # Should have content
   ```

4. **Verify Markdown content**
   ```bash
   head -100 dist/README.md
   
   # Expected:
   # - All systems documented
   # - All containers listed
   # - D2 diagrams embedded or referenced
   # - Readable structure
   ```

5. **Export to PDF (if veve-cli installed)**
   ```bash
   # Check if veve-cli available
   which veve-cli
   
   # If available:
   ./loko build --format pdf
   
   # Expected: dist/*.pdf files created
   ls dist/*.pdf
   ```

6. **Build all formats at once**
   ```bash
   ./loko build
   
   # Expected: All enabled formats in loko.toml
   ls dist/index.html
   ls dist/README.md
   ```

**Expected Result**: ✅ PASS
- HTML generation works
- Markdown export works
- PDF generation works (if veve-cli available)
- All formats contain same information
- Graceful degradation if PDF unavailable

---

### US-6: Token-Efficient Architecture Queries

**Test**: LLM can query with minimal token overhead

**Setup**:

```bash
# Create 20-system project for token testing
# (Use helper script if available, or repeat system creation 20x)

# Start MCP server
./loko mcp &
```

**Test Steps**:

1. **Query summary level**
   ```json
   // MCP Tool Call
   {
     "method": "query_architecture",
     "params": {
       "detail": "summary"
     }
   }
   
   // Expected response:
   {
     "type": "summary",
     "systems": 20,
     "containers": 80,
     "system_names": [...]
   }
   
   // Token count: ~200 tokens
   ```

2. **Query structure level**
   ```json
   {
     "method": "query_architecture",
     "params": {
       "detail": "structure"
     }
   }
   
   // Token count: ~500 tokens
   ```

3. **Query specific system (full)**
   ```json
   {
     "method": "query_architecture",
     "params": {
       "scope": "system",
       "target": "PaymentService",
       "detail": "full"
     }
   }
   
   // Token count: Variable, focused on one system
   ```

4. **Compare JSON vs TOON (v0.2.0)**
   ```json
   // With format: "json"
   // vs
   // With format: "toon"
   
   // Expected: TOON 30-40% fewer tokens
   ```

5. **Benchmark token efficiency**
   ```bash
   # Create helper script to count tokens
   # Run queries multiple times
   # Average token consumption
   
   # Expected:
   # - summary: <300 tokens
   # - structure: <600 tokens
   # - full (single system): <400 tokens
   ```

**Expected Result**: ✅ PASS
- Progressive loading works
- Token counts within budget
- TOON format reduces tokens by 30%+
- LLM can get context without excessive overhead

---

## Performance Benchmarks

These benchmarks verify NFR requirements:

### NFR-001: Build 100 diagrams in <30 seconds

```bash
# Create large project (100 diagrams)
time ./loko build

# Expected: Total time < 30 seconds
# Measured: real  0m25.123s (example)
```

### NFR-002: Watch mode rebuild <500ms

```bash
# Start watch mode with timing
./loko watch &

# Edit file and measure time from save to refresh
# Measured in browser DevTools or manual stopwatch

# Expected: <500ms
```

### NFR-003: Memory usage <100MB

```bash
# Monitor memory usage
# Top command or Memory/Activity monitor

# Open project and build
# Expected: RSS < 100MB
```

---

## Failure Scenarios (Edge Cases)

### Missing D2 Binary

```bash
# Rename d2 to simulate missing binary
mv /usr/local/bin/d2 /usr/local/bin/d2.bak

# Try to build
./loko build

# Expected: Clear error message
# "D2 binary not found. Install from https://d2lang.com"

# Restore
mv /usr/local/bin/d2.bak /usr/local/bin/d2
```

### Corrupted YAML Frontmatter

```bash
# Edit system.md with invalid YAML
vim src/PaymentService/system.md

# Insert bad YAML:
# ---
# name: "Invalid (unclosed quote
# ---

# Run validate
./loko validate

# Expected: Error with line number
# "Invalid YAML frontmatter in system.md, line 2"
```

### Circular Dependencies

```bash
# Create circular references manually (if possible)
# System A references System B, B references A

# Run validate
./loko validate

# Expected: Detected and reported
# "Circular system dependency detected"
```

### File System Permissions

```bash
# Remove read permission from system.md
chmod 000 src/PaymentService/system.md

# Try to build
./loko build

# Expected: Clear error
# "Permission denied: src/PaymentService/system.md"

# Restore
chmod 644 src/PaymentService/system.md
```

---
