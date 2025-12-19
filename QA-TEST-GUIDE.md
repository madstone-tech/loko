# QA Test Guide - MRO Scheduler (Phase 3 US-2)

## Project Information

- **Test Project**: MRO Scheduler (Maintenance Resource Optimization)
- **Location**: `/Users/andhi/code/mdstn/loko/test-phase3/`
- **Phase**: 3/8 (File Editing & Build Pipeline)
- **Status**: Production Ready

## Quick Start

```bash
cd /Users/andhi/code/mdstn/loko

# 1. Validate project structure
./loko validate -project test-phase3

# 2. Build documentation
./loko build -project test-phase3 -output test-phase3/dist

# 3. Serve locally
./loko serve -output test-phase3/dist
# Browse: http://localhost:8080

# 4. Watch for changes (in separate terminal)
./loko watch -project test-phase3 -output test-phase3/dist
```

## Test Commands

### 1. Validate Project Structure

```bash
./loko validate -project test-phase3
```

**Expected Output:**
- 3 systems discovered: API System, AWS Infrastructure, UI System
- 7 containers discovered across systems
- 6 warnings about missing components (expected for test project)
- No fatal errors

**Tests:**
- ✅ Project loading from loko.toml
- ✅ System discovery and enumeration
- ✅ Container discovery within systems
- ✅ Warning categorization

### 2. Build Documentation

```bash
./loko build -project test-phase3 -output test-phase3/dist
```

**Expected Output:**
```
ℹ Starting documentation build...
[100%] Generating HTML documentation...
✓ Documentation built successfully in test-phase3/dist
✓ Build completed in 0s
✓ Output: test-phase3/dist
```

**Verify Generated Files:**
```bash
find test-phase3/dist -type f | sort
```

**Expected Files:**
- `dist/index.html` - Project overview
- `dist/systems/api-system.html` - API System page
- `dist/systems/aws-infrastructure.html` - AWS Infrastructure page
- `dist/systems/ui-system.html` - UI System page
- `dist/styles/style.css` - Embedded CSS
- `dist/js/main.js` - JavaScript for search
- `dist/search.json` - Search index

**Tests:**
- ✅ Full build pipeline execution
- ✅ HTML generation
- ✅ System page generation
- ✅ CSS embedding
- ✅ JavaScript asset generation
- ✅ Search index creation
- ✅ Build timing <1 second

### 3. Inspect Generated HTML

```bash
# View index.html
head -50 test-phase3/dist/index.html

# View system page
head -50 test-phase3/dist/systems/api-system.html

# Check CSS
head -30 test-phase3/dist/styles/style.css

# View search index
cat test-phase3/dist/search.json | head -50
```

**Tests:**
- ✅ DOCTYPE declaration
- ✅ Proper HTML5 structure
- ✅ Navigation sidebar
- ✅ System list rendering
- ✅ Container details
- ✅ CSS responsive design
- ✅ Dark mode CSS variables
- ✅ Search functionality

### 4. Serve Documentation Locally

```bash
./loko serve -output test-phase3/dist
```

**Expected Output:**
```
🚀 Server starting on http://localhost:8080
   Serving documentation from: test-phase3/dist
   Press Ctrl+C to stop
```

**Manual Testing in Browser:**

1. **Open Homepage**
   - URL: http://localhost:8080
   - ✅ Should see MRO Scheduler project page
   - ✅ Sidebar with all systems
   - ✅ System list in main area
   - ✅ Search box functional

2. **Test Navigation**
   - Click "API System" → Should show API System page
   - Click "AWS Infrastructure" → Should show AWS Infrastructure page
   - Click project name in sidebar → Should return to index

3. **Test Search**
   - Type "API" in search box → Should filter systems/containers
   - Type "AWS" → Should highlight infrastructure items
   - Clear search → Should show all items

4. **Test Responsive Design**
   - Resize browser window
   - Should adapt to mobile, tablet, desktop sizes
   - Sidebar should collapse on mobile

5. **Test Dark Mode**
   - Open browser developer console
   - Toggle `prefers-color-scheme` in DevTools
   - Should switch between light and dark themes

**Tests:**
- ✅ Server starts on port 8080
- ✅ Static files serve correctly
- ✅ HTML renders in browser
- ✅ Navigation works
- ✅ Search functions
- ✅ CSS loads properly
- ✅ JavaScript works
- ✅ Responsive design
- ✅ Graceful shutdown with Ctrl+C

### 5. Watch Mode (Hot-Reload)

**Terminal 1 - Start Watch:**
```bash
cd /Users/andhi/code/mdstn/loko/test-phase3
../loko watch -project . -output dist -debounce 500
```

**Expected Output:**
```
👁 Watching for changes...
   Project: .
   Output: dist
   Press Ctrl+C to stop

🔨 Initial build...
[100%] Generating HTML documentation...
✓ Documentation built successfully in dist
✓ Initial build complete
```

**Terminal 2 - Make Changes:**
```bash
cd /Users/andhi/code/mdstn/loko/test-phase3
echo "## New Section" >> src/api-system/system.md
```

**Watch Terminal Should Show:**
```
📝 Change detected: src/api-system/system.md
🔨 Rebuilding...
[100%] Generating HTML documentation...
✓ Rebuild complete (0.5s)
```

**Tests:**
- ✅ File changes detected
- ✅ Changes debounced (500ms window)
- ✅ Auto-rebuild triggered
- ✅ Rebuild completes <500ms
- ✅ Multiple edits batched
- ✅ Progress reported

## Project Structure

```
test-phase3/
├── loko.toml                    # Project configuration
└── src/
    ├── api-system/              # API System
    │   ├── system.md
    │   ├── rest-api/
    │   │   └── container.md
    │   ├── optimization/
    │   │   └── container.md
    │   └── database/
    │       └── container.md
    ├── aws-infrastructure/      # AWS Infrastructure
    │   ├── system.md
    │   ├── system.d2            # D2 Diagram
    │   ├── compute/
    │   │   └── container.md
    │   ├── storage/
    │   │   └── container.md
    │   ├── network/
    │   │   └── container.md
    │   └── security/
    │       └── container.md
    └── ui-system/               # UI System
        ├── system.md
        └── system.d2
```

## Test Project Contents

### Systems

1. **AWS Infrastructure** (4 containers)
   - ECS Fargate Compute
   - Data Storage (RDS, S3)
   - Network & Load Balancing
   - Security & Authentication

2. **API System** (3 containers)
   - REST API Service
   - Optimization Engine
   - Data Access Layer

3. **UI System** (0 containers - test data)
   - React web interface

### Metadata

- Systems have YAML frontmatter with name, description, tags
- Containers have technology stacks defined
- D2 diagrams included for some systems

## Performance Benchmarks

Expected performance:
- Build time: <1 second
- Watch debounce: 500ms
- File change detection: <100ms
- HTML generation: <500ms
- Rebuild on watch: <500ms total

## Pass/Fail Criteria

### Validation ✅
- [ ] Validate command succeeds
- [ ] 3 systems detected
- [ ] 7 containers detected
- [ ] Warnings reported correctly

### Build ✅
- [ ] Build command succeeds
- [ ] 7 HTML files generated
- [ ] CSS embedded
- [ ] JavaScript included
- [ ] Search index created

### Server ✅
- [ ] Server starts on :8080
- [ ] Index page loads
- [ ] System pages load
- [ ] Navigation works
- [ ] Search functional
- [ ] CSS loads
- [ ] Graceful shutdown

### Watch ✅
- [ ] Watch starts
- [ ] Initial build succeeds
- [ ] File changes detected
- [ ] Rebuild triggered
- [ ] Rebuild <500ms
- [ ] Progress reported

## Troubleshooting

### Server Port Already in Use
```bash
lsof -i :8080
kill -9 <PID>
./loko serve -output test-phase3/dist
```

### D2 Not Found
D2 diagrams are optional for Phase 3 - if d2 CLI not installed, system works fine without rendering D2 files.

```bash
# Install d2
brew install d2lang/d2/d2
```

### Watch Not Detecting Changes
- Ensure you're in the correct directory
- Check file paths are relative to project root
- Watch debounce window is 500ms (wait 500ms after last change)

## Next Steps

After QA testing passes:
1. Commit changes to git
2. Push to remote
3. Deploy to staging
4. Begin Phase 4 (MCP Server for Claude)

## References

- **Documentation**: PHASE3-TEST-RESULTS.md
- **Source Code**: /Users/andhi/code/mdstn/loko/
- **Test Project**: /Users/andhi/code/mdstn/loko/test-phase3/

---

**Last Updated**: 2025-12-19
**Phase**: 3/8 (File Editing & Build Pipeline)
**Status**: ✅ Production Ready
