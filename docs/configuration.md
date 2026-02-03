# Configuration Reference

loko uses TOML configuration files. Configuration is loaded from two locations:

1. **Global config**: `~/.loko/config.toml` (user defaults)
2. **Project config**: `./loko.toml` (project-specific settings)

Project settings override global settings.

## Full Configuration Example

```toml
# loko.toml - Project configuration

[project]
name = "My Architecture"
description = "Architecture documentation for my system"
version = "1.0.0"

[paths]
source = "./src"        # Source directory for architecture files
output = "./dist"       # Output directory for generated docs

[d2]
theme = "neutral-default"   # D2 theme for diagrams
layout = "elk"              # D2 layout engine
cache = true                # Cache rendered diagrams

[outputs]
html = true             # Generate HTML documentation
markdown = false        # Generate README.md
pdf = false             # Generate PDF (requires veve-cli)

[build]
parallel = true         # Parallel diagram rendering
max_workers = 4         # Maximum parallel workers

[server]
serve_port = 8080       # Preview server port
api_port = 8081         # API server port
hot_reload = true       # Auto-reload on changes
```

## Configuration Sections

### [project]

Project metadata displayed in documentation.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `name` | string | - | Project name |
| `description` | string | - | Project description |
| `version` | string | - | Documentation version |

### [paths]

Directory configuration.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `source` | string | `"./src"` | Source directory for architecture files |
| `output` | string | `"./dist"` | Output directory for generated documentation |

### [d2]

D2 diagram rendering settings.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `theme` | string | `"neutral-default"` | D2 theme name |
| `layout` | string | `"elk"` | Layout engine: `elk`, `dagre`, `tala` |
| `cache` | bool | `true` | Cache rendered diagrams for faster rebuilds |

**Available Themes:**
- `neutral-default` - Clean, professional look
- `neutral-grey` - Subtle grey tones
- `flagship-terrastruct` - Terrastruct brand colors
- `cool-classics` - Blue-focused palette
- `mixed-berry-blue` - Purple and blue tones
- `grape-soda` - Purple palette
- `aubergine` - Dark purple theme
- `colorblind-clear` - Accessibility-focused
- `vanilla-nitro-cola` - High contrast
- `orange-creamsicle` - Warm orange tones
- `shirley-temple` - Pink and coral
- `earth-tones` - Natural browns and greens
- `everglade-green` - Green palette
- `buttered-toast` - Warm yellows
- `dark-mauve` - Dark theme with mauve accents
- `dark-flagship-terrastruct` - Dark Terrastruct theme
- `terminal` - Terminal/console style
- `terminal-grayscale` - Grayscale terminal

**Layout Engines:**
- `elk` - ELK layered layout (default, best for hierarchical diagrams)
- `dagre` - Dagre layout (fast, good for most diagrams)
- `tala` - TALA layout (premium, requires license)

### [outputs]

Output format configuration.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `html` | bool | `true` | Generate HTML documentation site |
| `markdown` | bool | `false` | Generate single README.md file |
| `pdf` | bool | `false` | Generate PDF (requires veve-cli) |

**Note:** PDF generation requires [veve-cli](https://github.com/nicholasgriffintn/veve-cli) to be installed.

### [build]

Build process configuration.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `parallel` | bool | `true` | Enable parallel diagram rendering |
| `max_workers` | int | `4` | Maximum number of parallel workers |

### [server]

Development server configuration.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `serve_port` | int | `8080` | Port for preview server (`loko serve`) |
| `api_port` | int | `8081` | Port for API server (`loko api`) |
| `hot_reload` | bool | `true` | Auto-reload browser on changes |

## Environment Variables

Some settings can be overridden with environment variables:

| Variable | Description |
|----------|-------------|
| `LOKO_API_KEY` | API key for HTTP API authentication |
| `LOKO_PROJECT_ROOT` | Override project root directory |
| `D2_LAYOUT` | Override D2 layout engine |
| `D2_THEME` | Override D2 theme |

## Command-Line Overrides

Most options can be overridden via command-line flags:

```bash
# Override output directory
loko build --output ./public

# Override formats
loko build --format html --format markdown

# Override server port
loko serve --port 3000

# Override API port and key
loko api --port 9000 --api-key "my-secret-key"
```

## Minimal Configuration

For simple projects, you only need:

```toml
[project]
name = "My Project"
```

All other settings use sensible defaults.

## Multi-Format Output

To generate multiple output formats:

```toml
[outputs]
html = true
markdown = true
pdf = true
```

Or via command line:

```bash
loko build --format html --format markdown --format pdf
```

## Custom Templates

loko looks for templates in these locations (in order):

1. `.loko/templates/` (project-local)
2. `~/.loko/templates/` (user templates)
3. Built-in templates

To customize templates, copy the built-in templates to one of these directories and modify them.

## Diagram Caching

When `d2.cache = true`:

- Diagrams are cached in `.loko/cache/`
- Cache is invalidated when source `.d2` files change
- Use `loko build --clean` to force rebuild all diagrams

## API Authentication

For the HTTP API (`loko api`):

```bash
# Start with authentication
loko api --api-key "your-secret-key"

# Make authenticated requests
curl -H "Authorization: Bearer your-secret-key" http://localhost:8081/api/v1/project
```

When no API key is set, authentication is disabled (suitable for local development).
