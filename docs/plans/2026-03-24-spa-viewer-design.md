---
status: approved
date: 2026-03-24
---

# SPA Wiki Viewer Design

## Context

Eudemics needs a browsable viewer for its wiki-style content. The viewer must be a single self-contained Go binary that serves a single-page application, with content stored as markdown files with YAML frontmatter. Git history is the audit trail, so content stays as files in the repo, not in a database.

## Architecture

Single `main.go` binary. No internal packages. Stdlib `net/http` for routing. Templates embedded via `go:embed`.

On startup:
1. Check if content directory exists. If not, `git clone` the configured repo URL.
2. Walk the content tree, parse YAML frontmatter from all `.md` files, build the manifest in memory.
3. Watch for file changes via `fsnotify`, rebuild manifest on change.
4. Serve the SPA and API endpoints.

### Endpoints

| Route | Purpose |
|-------|---------|
| `GET /` | Serves the embedded SPA (`views/index.html`) |
| `GET /api/manifest` | Full tree + relationships + frontmatter for all concepts |
| `GET /api/content/{path...}` | Returns server-rendered HTML for a specific markdown document |

### Flags

| Flag | Default | Purpose |
|------|---------|---------|
| `--port` | 8080 | Server port |
| `--content` | `./` | Path to content directory |
| `--repo` | (none) | Git URL to clone if content dir is empty |

### Dependencies

| Dep | Purpose |
|-----|---------|
| `github.com/yuin/goldmark` | Markdown to HTML rendering |
| `gopkg.in/yaml.v3` | YAML frontmatter parsing |
| `github.com/fsnotify/fsnotify` | File watching for manifest hot-reload |

Everything else is stdlib.

### Embedded Assets

```
views/
  index.html    (single file, all CSS/JS inline)
```

Embedded via `go:embed` so the binary is fully self-contained.

## Visual Style

Inspired by vAnima and Tricell Network references.

- Light base: `#F5F5F0` background with subtle animated sliding grid (low opacity)
- Accent palette: muted teal/blue-green civic feel
  - Primary: `#1a6b5a`
  - Bright accent: `#00e5c3`
  - Dark chrome/text: `#0a2e24`
- Typography: Terminus (body), Monocraft (headings/labels), self-hosted .ttf
- Cards: `backdrop-filter: blur(10px)` + semi-transparent white bg + thin border
- No border-radius (sharp edges)
- Transitions: `0.3s ease` consistent
- 2-space indentation for CSS and JS

## Layout

Three persistent regions:

### Right-side vertical nav (Tricell style)
- Package buttons stacked vertically
- Clicking a package expands sub-topics inline or replaces the list
- Hover: glow + translateX shift
- Highlights current location

### Main content area (center/left)
- Rendered markdown document displayed in a card panel
- Animated transitions between documents (fade/slide)
- Homepage state: title, description, radial package overview

### Bottom-left graph canvas (persistent)
- Always visible, always interactive
- On homepage: shows all main package nodes in a radial/circle layout
- On document: shows local neighborhood graph (current concept + direct relationships)
- Clicking any node navigates to that concept
- Hover highlights connected edges

## Graph Rendering

All canvas contexts (homepage radial, document neighborhood) use the same rendering logic with different data:

- Nodes: circles with labels, colored by parent package
- Edges: styled by relationship type
  - `builds-on`: solid line
  - `constrains`: dashed line
  - `related`: dotted line
  - `implements`: solid with arrow
- Interaction: click to navigate, hover to highlight connections
- Physics: simple force-directed simulation from scratch in JS (no libraries)
- Zoom/pan support

## URL Routing

Hash-based routing (`#/autonomy/liberty`, `#/graph`, `#/`) so browser back/forward works and URLs are shareable.

## GitHub Actions

Cross-platform matrix build on tag push:

| OS | Arch |
|----|------|
| linux | amd64, arm64 |
| darwin | amd64, arm64 |
| windows | amd64, arm64 |

6 binaries attached to GitHub releases.

## Manifest Schema

```json
{
  "packages": [
    {
      "name": "autonomy",
      "path": "autonomy",
      "status": "draft",
      "tags": ["autonomy"],
      "children": [
        {
          "name": "liberty",
          "path": "autonomy/liberty",
          "status": "draft",
          "tags": ["autonomy"],
          "builds_on": ["foundations/ethics"],
          "constrains": ["justice/enforcement"],
          "related": ["law/constitutional"]
        }
      ]
    }
  ]
}
```

## Future Work

- Search functionality (full-text across all concepts)
- Diff viewer showing git history for a specific concept
- Export to static HTML for hosting without the binary
