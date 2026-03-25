# SPA Wiki Viewer Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a self-contained Go binary that serves a wiki viewer SPA for the eudemics conceptual model.

**Architecture:** Single `main.go` using stdlib `net/http`, goldmark for markdown rendering, fsnotify for hot-reload. SPA is a single `views/index.html` embedded via `go:embed`. Binary can clone the repo if content doesn't exist locally.

**Tech Stack:** Go stdlib, goldmark, yaml.v3, fsnotify. Raw HTML/CSS/JS (inline). Canvas API for graph rendering.

**Fonts:** Copy `terminus.ttf` and `Monocraft.ttf` from `/home/requiem/Projects/WebUI/requiem-new/assets/fonts/` into `views/fonts/` and embed them.

---

### Task 1: Go Module + Flag Parsing

**Files:**
- Create: `main.go`
- Create: `go.mod`

**Step 1: Initialize the Go module**

Run: `cd /home/requiem/Projects/ideas/eudemics && go mod init github.com/yourusername/eudemics`

**Step 2: Write main.go with flag parsing and placeholder server**

```go
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"context"
)

func main() {
	port := flag.Int("port", 8080, "server port")
	contentDir := flag.String("content", ".", "path to content directory")
	repo := flag.String("repo", "", "git repo URL to clone if content dir is empty")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	slog.Info("starting eudemics", "port", *port, "content", *contentDir, "repo", *repo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("eudemics"))
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		slog.Info("listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*1e9)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
```

**Step 3: Verify it builds and runs**

Run: `go build -o eudemics . && ./eudemics --port 9090 &`
Run: `curl -s http://localhost:9090/`
Expected: `eudemics`
Run: `kill %1`

**Step 4: Commit**

```bash
git add main.go go.mod
git commit -m "feat: go module with flag parsing and placeholder server"
```

---

### Task 2: Git Clone Bootstrap

**Files:**
- Modify: `main.go`

**Step 1: Write the cloneRepo function**

Add before `main()`:

```go
func cloneRepo(repo, dest string) error {
	slog.Info("cloning repo", "repo", repo, "dest", dest)
	cmd := exec.Command("git", "clone", repo, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clone repo: %w", err)
	}
	return nil
}
```

Add to `main()` after flag parsing, before mux setup:

```go
if *repo != "" {
	entries, err := os.ReadDir(*contentDir)
	if err != nil || len(entries) == 0 {
		if err := cloneRepo(*repo, *contentDir); err != nil {
			slog.Error("failed to clone", "err", err)
			os.Exit(1)
		}
	}
}
```

Add `"os/exec"` to imports.

**Step 2: Verify it builds**

Run: `go build -o eudemics .`
Expected: clean build, no errors

**Step 3: Commit**

```bash
git add main.go
git commit -m "feat: git clone bootstrap when content dir is empty"
```

---

### Task 3: Frontmatter Parsing + Manifest Builder

**Files:**
- Modify: `main.go`
- Create: `main_test.go`

**Step 1: Add dependencies**

Run: `go get github.com/yuin/goldmark gopkg.in/yaml.v3 github.com/fsnotify/fsnotify`

**Step 2: Write the failing test for frontmatter parsing**

Create `main_test.go`:

```go
package main

import (
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	input := `---
status: draft
tags: [autonomy, rights]
builds_on: [foundations/ethics]
constrains: [justice/enforcement]
related: [law/constitutional]
---

# Liberty

Freedom from coercion, freedom to act.
`
	fm, body, err := parseFrontmatter([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if fm.Status != "draft" {
		t.Errorf("status = %q, want draft", fm.Status)
	}
	if len(fm.Tags) != 2 || fm.Tags[0] != "autonomy" {
		t.Errorf("tags = %v, want [autonomy rights]", fm.Tags)
	}
	if len(fm.BuildsOn) != 1 || fm.BuildsOn[0] != "foundations/ethics" {
		t.Errorf("builds_on = %v", fm.BuildsOn)
	}
	if len(fm.Constrains) != 1 {
		t.Errorf("constrains = %v", fm.Constrains)
	}
	if len(fm.Related) != 1 {
		t.Errorf("related = %v", fm.Related)
	}
	if len(body) == 0 {
		t.Error("body is empty")
	}
}
```

**Step 3: Run test to verify it fails**

Run: `go test -v -run TestParseFrontmatter`
Expected: FAIL - `parseFrontmatter` not defined

**Step 4: Implement frontmatter parsing**

Add to `main.go`:

```go
import (
	"bytes"
	"gopkg.in/yaml.v3"
)

type Frontmatter struct {
	Status     string   `yaml:"status" json:"status"`
	Tags       []string `yaml:"tags" json:"tags"`
	BuildsOn   []string `yaml:"builds_on" json:"builds_on,omitempty"`
	Constrains []string `yaml:"constrains" json:"constrains,omitempty"`
	Implements []string `yaml:"implements" json:"implements,omitempty"`
	Related    []string `yaml:"related" json:"related,omitempty"`
}

func parseFrontmatter(data []byte) (Frontmatter, []byte, error) {
	var fm Frontmatter
	if !bytes.HasPrefix(data, []byte("---\n")) {
		return fm, data, nil
	}
	end := bytes.Index(data[4:], []byte("\n---"))
	if end == -1 {
		return fm, data, nil
	}
	if err := yaml.Unmarshal(data[4:4+end], &fm); err != nil {
		return fm, nil, fmt.Errorf("parse frontmatter: %w", err)
	}
	body := data[4+end+4:]
	return fm, body, nil
}
```

**Step 5: Run test to verify it passes**

Run: `go test -v -run TestParseFrontmatter`
Expected: PASS

**Step 6: Write the failing test for manifest building**

Add to `main_test.go`:

```go
func TestBuildManifest(t *testing.T) {
	manifest, err := buildManifest(".")
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Packages) == 0 {
		t.Error("no packages found")
	}
	found := false
	for _, pkg := range manifest.Packages {
		if pkg.Name == "autonomy" {
			found = true
			if len(pkg.Children) == 0 {
				t.Error("autonomy has no children")
			}
		}
	}
	if !found {
		t.Error("autonomy package not found")
	}
}
```

**Step 7: Run test to verify it fails**

Run: `go test -v -run TestBuildManifest`
Expected: FAIL - `buildManifest` not defined

**Step 8: Implement manifest builder**

Add to `main.go`:

```go
import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type ConceptNode struct {
	Name       string         `json:"name"`
	Path       string         `json:"path"`
	Status     string         `json:"status"`
	Tags       []string       `json:"tags"`
	BuildsOn   []string       `json:"builds_on,omitempty"`
	Constrains []string       `json:"constrains,omitempty"`
	Implements []string       `json:"implements,omitempty"`
	Related    []string       `json:"related,omitempty"`
	Children   []*ConceptNode `json:"children,omitempty"`
}

type Manifest struct {
	Packages []*ConceptNode `json:"packages"`
}

var (
	manifestCache     []byte
	manifestCacheLock sync.RWMutex
)

var skipDirs = map[string]bool{
	".git": true, "docs": true, "views": true, ".github": true, "cmd": true,
}

func buildManifest(root string) (Manifest, error) {
	var manifest Manifest
	packages := map[string]*ConceptNode{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		if rel == "." {
			return nil
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if skipDirs[parts[0]] {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") || d.Name() == "CLAUDE.md" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		fm, _, err := parseFrontmatter(data)
		if err != nil {
			slog.Warn("bad frontmatter", "path", rel, "err", err)
			return nil
		}

		relPath := filepath.ToSlash(rel)
		dirPath := filepath.ToSlash(filepath.Dir(rel))

		node := &ConceptNode{
			Name:       strings.TrimSuffix(d.Name(), ".md"),
			Path:       relPath,
			Status:     fm.Status,
			Tags:       fm.Tags,
			BuildsOn:   fm.BuildsOn,
			Constrains: fm.Constrains,
			Implements: fm.Implements,
			Related:    fm.Related,
		}

		if len(parts) == 2 {
			pkgName := parts[0]
			if _, ok := packages[pkgName]; !ok {
				packages[pkgName] = &ConceptNode{
					Name: pkgName,
					Path: pkgName,
					Tags: fm.Tags,
				}
			}
			pkg := packages[pkgName]
			if d.Name() == "README.md" {
				pkg.Status = fm.Status
				pkg.Tags = fm.Tags
				pkg.BuildsOn = fm.BuildsOn
				pkg.Constrains = fm.Constrains
				pkg.Implements = fm.Implements
				pkg.Related = fm.Related
			}
		} else if len(parts) == 3 {
			pkgName := parts[0]
			if _, ok := packages[pkgName]; !ok {
				packages[pkgName] = &ConceptNode{
					Name: pkgName,
					Path: pkgName,
				}
			}
			subName := parts[1]
			pkg := packages[pkgName]
			if d.Name() == "README.md" {
				child := &ConceptNode{
					Name:       subName,
					Path:       dirPath,
					Status:     fm.Status,
					Tags:       fm.Tags,
					BuildsOn:   fm.BuildsOn,
					Constrains: fm.Constrains,
					Implements: fm.Implements,
					Related:    fm.Related,
				}
				pkg.Children = append(pkg.Children, child)
			} else {
				node.Path = relPath
				pkg.Children = append(pkg.Children, node)
			}
		}
		return nil
	})
	if err != nil {
		return manifest, fmt.Errorf("build manifest: %w", err)
	}

	for _, pkg := range packages {
		manifest.Packages = append(manifest.Packages, pkg)
	}
	sort.Slice(manifest.Packages, func(i, j int) bool {
		return manifest.Packages[i].Name < manifest.Packages[j].Name
	})
	for _, pkg := range manifest.Packages {
		sort.Slice(pkg.Children, func(i, j int) bool {
			return pkg.Children[i].Name < pkg.Children[j].Name
		})
	}
	return manifest, nil
}

func cacheManifest(root string) error {
	m, err := buildManifest(root)
	if err != nil {
		return err
	}
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	manifestCacheLock.Lock()
	manifestCache = data
	manifestCacheLock.Unlock()
	return nil
}
```

**Step 9: Run tests**

Run: `go test -v`
Expected: both tests PASS

**Step 10: Commit**

```bash
git add main.go main_test.go go.mod go.sum
git commit -m "feat: frontmatter parsing and manifest builder"
```

---

### Task 4: Markdown Rendering Endpoint

**Files:**
- Modify: `main.go`
- Modify: `main_test.go`

**Step 1: Write the failing test**

Add to `main_test.go`:

```go
func TestRenderMarkdown(t *testing.T) {
	input := []byte("# Hello\n\nThis is a **test** with a [[wiki link]].\n")
	html, err := renderMarkdown(input)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "<h1>Hello</h1>") {
		t.Errorf("missing h1: %s", html)
	}
	if !strings.Contains(html, "<strong>test</strong>") {
		t.Errorf("missing strong: %s", html)
	}
}
```

Add `"strings"` to test imports.

**Step 2: Run test to verify it fails**

Run: `go test -v -run TestRenderMarkdown`
Expected: FAIL

**Step 3: Implement markdown rendering**

Add to `main.go`:

```go
import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

func renderMarkdown(source []byte) (string, error) {
	var buf bytes.Buffer
	if err := md.Convert(source, &buf); err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}
	return buf.String(), nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v -run TestRenderMarkdown`
Expected: PASS

**Step 5: Commit**

```bash
git add main.go main_test.go
git commit -m "feat: goldmark markdown rendering"
```

---

### Task 5: API Endpoints

**Files:**
- Modify: `main.go`

**Step 1: Wire up the manifest and content endpoints in main()**

Replace the placeholder mux handler in `main()` with:

```go
if err := cacheManifest(*contentDir); err != nil {
	slog.Error("failed to build manifest", "err", err)
	os.Exit(1)
}

mux := http.NewServeMux()

mux.HandleFunc("GET /api/manifest", func(w http.ResponseWriter, r *http.Request) {
	manifestCacheLock.RLock()
	data := manifestCache
	manifestCacheLock.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
})

mux.HandleFunc("GET /api/content/{path...}", func(w http.ResponseWriter, r *http.Request) {
	reqPath := r.PathValue("path")
	clean := filepath.Clean(reqPath)
	if strings.Contains(clean, "..") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	full := filepath.Join(*contentDir, clean)
	info, err := os.Stat(full)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if info.IsDir() {
		full = filepath.Join(full, "README.md")
	}
	if !strings.HasSuffix(full, ".md") {
		full = full + ".md"
	}

	data, err := os.ReadFile(full)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	_, body, err := parseFrontmatter(data)
	if err != nil {
		http.Error(w, "parse error", http.StatusInternalServerError)
		return
	}

	html, err := renderMarkdown(body)
	if err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
})

mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("eudemics"))
})
```

**Step 2: Build and manual test**

Run: `go build -o eudemics . && ./eudemics --port 9090 &`
Run: `curl -s http://localhost:9090/api/manifest | head -c 200`
Expected: JSON starting with `{"packages":[`
Run: `curl -s http://localhost:9090/api/content/autonomy/liberty`
Expected: HTML with `<h1>Liberty</h1>`
Run: `kill %1`

**Step 3: Commit**

```bash
git add main.go
git commit -m "feat: manifest and content API endpoints"
```

---

### Task 6: File Watcher

**Files:**
- Modify: `main.go`

**Step 1: Add fsnotify watcher in main() after cacheManifest**

```go
watcher, err := fsnotify.NewWatcher()
if err != nil {
	slog.Error("failed to create watcher", "err", err)
	os.Exit(1)
}
defer watcher.Close()

go func() {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
				if strings.HasSuffix(event.Name, ".md") {
					slog.Info("content changed, rebuilding manifest", "file", event.Name)
					if err := cacheManifest(*contentDir); err != nil {
						slog.Error("rebuild manifest failed", "err", err)
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			slog.Error("watcher error", "err", err)
		}
	}
}()

filepath.WalkDir(*contentDir, func(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if d.IsDir() {
		rel, _ := filepath.Rel(*contentDir, path)
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) > 0 && skipDirs[parts[0]] {
			return filepath.SkipDir
		}
		watcher.Add(path)
	}
	return nil
})
```

Add `"github.com/fsnotify/fsnotify"` to imports.

**Step 2: Build and verify**

Run: `go build -o eudemics .`
Expected: clean build

**Step 3: Commit**

```bash
git add main.go
git commit -m "feat: fsnotify file watcher for manifest hot-reload"
```

---

### Task 7: Fonts + Embedded SPA Skeleton

**Files:**
- Create: `views/index.html`
- Create: `views/fonts/terminus.ttf` (copy)
- Create: `views/fonts/Monocraft.ttf` (copy)
- Modify: `main.go`

**Step 1: Copy fonts**

Run: `mkdir -p views/fonts`
Run: `cp /home/requiem/Projects/WebUI/requiem-new/assets/fonts/terminus.ttf views/fonts/`
Run: `cp /home/requiem/Projects/WebUI/requiem-new/assets/fonts/Monocraft.ttf views/fonts/`

**Step 2: Add go:embed and serve SPA + fonts in main.go**

Add near top of file:

```go
import "embed"

//go:embed views/index.html
var indexHTML []byte

//go:embed views/fonts
var fontsFS embed.FS
```

Replace the `GET /` handler and add font serving before it:

```go
mux.Handle("GET /fonts/", http.StripPrefix("/fonts/", http.FileServerFS(fontsFS)))

mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
})
```

Note: the font FileServer serves from the `views/fonts` embedded directory. Adjust the StripPrefix and FS subdir:

```go
fontsDir, _ := fs.Sub(fontsFS, "views/fonts")
mux.Handle("GET /fonts/", http.StripPrefix("/fonts/", http.FileServerFS(fontsDir)))
```

Add `"io/fs"` to imports if not present.

**Step 3: Create views/index.html with CSS foundation**

Create `views/index.html` - this is the full SPA skeleton with styles, layout regions, and no JS yet:

```html
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>eudemics</title>
<style>
@font-face {
  font-family: "terminus";
  src: url("/fonts/terminus.ttf");
}
@font-face {
  font-family: "Monocraft";
  src: url("/fonts/Monocraft.ttf");
}
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}
:root {
  --bg: #F5F5F0;
  --primary: #1a6b5a;
  --bright: #00e5c3;
  --dark: #0a2e24;
  --text: #2A2A2A;
  --text-mid: #5A5A5A;
  --card-bg: rgba(255, 255, 255, 0.7);
  --card-border: rgba(26, 107, 90, 0.2);
  --nav-hover: rgba(0, 229, 195, 0.15);
  --grid-color: rgba(26, 107, 90, 0.06);
}
body {
  background: var(--bg);
  color: var(--text);
  font-family: 'terminus', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.8;
  overflow-x: hidden;
}
.grid-bg {
  position: fixed;
  top: 0;
  left: 0;
  width: 200%;
  height: 200%;
  background-image:
    linear-gradient(var(--grid-color) 1px, transparent 1px),
    linear-gradient(90deg, var(--grid-color) 1px, transparent 1px);
  background-size: 40px 40px;
  animation: gridSlide 60s linear infinite;
  pointer-events: none;
  z-index: 0;
}
@keyframes gridSlide {
  0% { transform: translate(0, 0); }
  100% { transform: translate(-40px, -40px); }
}
#app {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: 1fr 240px;
  grid-template-rows: 1fr auto;
  min-height: 100vh;
  gap: 0;
}
#content {
  grid-column: 1;
  grid-row: 1 / -1;
  padding: 40px;
  overflow-y: auto;
}
#content .card {
  backdrop-filter: blur(10px);
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  padding: 32px;
  max-width: 800px;
  opacity: 1;
  transform: translateY(0);
  transition: opacity 0.3s ease, transform 0.3s ease;
}
#content .card.fade-out {
  opacity: 0;
  transform: translateY(8px);
}
#nav {
  grid-column: 2;
  grid-row: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 40px 16px;
  overflow-y: auto;
  max-height: 100vh;
  position: sticky;
  top: 0;
}
.nav-pkg {
  font-family: 'Monocraft', monospace;
  font-size: 12px;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  padding: 10px 14px;
  background: transparent;
  border: 1px solid transparent;
  color: var(--dark);
  cursor: pointer;
  transition: all 0.3s ease;
  text-align: left;
  width: 100%;
}
.nav-pkg:hover {
  background: var(--nav-hover);
  border-color: var(--card-border);
  transform: translateX(-4px);
  box-shadow: 8px 0 20px rgba(0, 229, 195, 0.1);
}
.nav-pkg.active {
  background: var(--nav-hover);
  border-color: var(--primary);
  color: var(--primary);
}
.nav-sub {
  font-family: 'terminus', monospace;
  font-size: 12px;
  padding: 6px 14px 6px 28px;
  background: transparent;
  border: none;
  color: var(--text-mid);
  cursor: pointer;
  transition: all 0.3s ease;
  text-align: left;
  width: 100%;
  display: none;
}
.nav-sub.visible {
  display: block;
}
.nav-sub:hover {
  color: var(--primary);
  transform: translateX(-2px);
}
.nav-sub.active {
  color: var(--primary);
  font-weight: bold;
}
#graph-panel {
  position: fixed;
  bottom: 16px;
  left: 16px;
  width: 280px;
  height: 220px;
  backdrop-filter: blur(10px);
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  z-index: 10;
  overflow: hidden;
}
#graph-panel canvas {
  width: 100%;
  height: 100%;
}
#graph-label {
  position: absolute;
  top: 6px;
  left: 10px;
  font-family: 'Monocraft', monospace;
  font-size: 10px;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--text-mid);
  pointer-events: none;
}
.content-title {
  font-family: 'Monocraft', monospace;
  font-size: 24px;
  letter-spacing: 0.1em;
  color: var(--dark);
  margin-bottom: 8px;
}
.content-meta {
  font-size: 12px;
  color: var(--text-mid);
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--card-border);
}
.content-body h1 { font-family: 'Monocraft', monospace; font-size: 20px; color: var(--dark); margin: 24px 0 12px; }
.content-body h2 { font-family: 'Monocraft', monospace; font-size: 16px; color: var(--primary); margin: 20px 0 10px; }
.content-body h3 { font-family: 'Monocraft', monospace; font-size: 14px; color: var(--text); margin: 16px 0 8px; }
.content-body p { margin: 0 0 12px; }
.content-body a { color: var(--primary); text-decoration: none; border-bottom: 1px solid var(--bright); }
.content-body a:hover { color: var(--bright); }
.content-body code { background: rgba(26, 107, 90, 0.08); padding: 2px 6px; font-size: 13px; }
.content-body ul, .content-body ol { margin: 0 0 12px 24px; }
.homepage-title {
  font-family: 'Monocraft', monospace;
  font-size: 32px;
  letter-spacing: 0.15em;
  color: var(--dark);
  text-transform: lowercase;
  margin-bottom: 8px;
}
.homepage-desc {
  color: var(--text-mid);
  max-width: 600px;
  margin-bottom: 32px;
}
@media (max-width: 1024px) {
  #app { grid-template-columns: 1fr; }
  #nav {
    position: fixed;
    right: 0;
    top: 0;
    width: 240px;
    height: 100vh;
    background: var(--bg);
    border-left: 1px solid var(--card-border);
    transform: translateX(100%);
    transition: transform 0.3s ease;
    z-index: 20;
  }
  #nav.open { transform: translateX(0); }
  #graph-panel { width: 200px; height: 160px; }
}
</style>
</head>
<body>
<div class="grid-bg"></div>
<div id="app">
  <div id="content"></div>
  <nav id="nav"></nav>
</div>
<div id="graph-panel">
  <span id="graph-label">relationships</span>
  <canvas id="graph-canvas"></canvas>
</div>
<script>
</script>
</body>
</html>
```

**Step 4: Build and verify the SPA loads**

Run: `go build -o eudemics . && ./eudemics --port 9090 &`
Run: `curl -s http://localhost:9090/ | head -5`
Expected: `<!DOCTYPE html>`
Run: `curl -s http://localhost:9090/fonts/terminus.ttf | wc -c`
Expected: non-zero byte count
Run: `kill %1`

**Step 5: Commit**

```bash
git add views/ main.go
git commit -m "feat: embedded SPA skeleton with CSS and fonts"
```

---

### Task 8: SPA JavaScript - Router + Nav + Content Loading

**Files:**
- Modify: `views/index.html` (add JS inside the `<script>` tag)

**Step 1: Write the full JS application**

Replace the empty `<script>` tag content in `views/index.html` with:

```javascript
var state = {
  manifest: null,
  currentPath: null,
  expandedPkg: null
};

function navigate(path) {
  window.location.hash = path ? "/" + path : "/";
}

function getHashPath() {
  var h = window.location.hash.slice(1);
  if (h.startsWith("/")) h = h.slice(1);
  return h;
}

async function fetchManifest() {
  var resp = await fetch("/api/manifest");
  state.manifest = await resp.json();
}

async function fetchContent(path) {
  var resp = await fetch("/api/content/" + path);
  if (!resp.ok) return "<p>not found</p>";
  return await resp.text();
}

function renderNav() {
  var nav = document.getElementById("nav");
  nav.innerHTML = "";
  if (!state.manifest) return;

  var homeBtn = document.createElement("button");
  homeBtn.className = "nav-pkg" + (!state.currentPath ? " active" : "");
  homeBtn.textContent = "home";
  homeBtn.onclick = function() { navigate(""); };
  nav.appendChild(homeBtn);

  state.manifest.packages.forEach(function(pkg) {
    var btn = document.createElement("button");
    btn.className = "nav-pkg";
    if (state.currentPath && state.currentPath.startsWith(pkg.path)) btn.className += " active";
    btn.textContent = pkg.name;
    btn.onclick = function() {
      if (state.expandedPkg === pkg.name) {
        navigate(pkg.path);
      } else {
        state.expandedPkg = pkg.name;
        renderNav();
      }
    };
    nav.appendChild(btn);

    if (pkg.children && state.expandedPkg === pkg.name) {
      pkg.children.forEach(function(child) {
        var sub = document.createElement("button");
        sub.className = "nav-sub visible";
        if (state.currentPath === child.path) sub.className += " active";
        sub.textContent = child.name;
        sub.onclick = function() { navigate(child.path); };
        nav.appendChild(sub);
      });
    }
  });
}

function renderHomepage() {
  var content = document.getElementById("content");
  var card = document.createElement("div");
  card.className = "card";
  card.innerHTML =
    '<div class="homepage-title">eudemics</div>' +
    '<div class="homepage-desc">a conceptual governance model optimized toward the human condition. ' +
    'select a package from the nav or click a node in the graph to explore.</div>';
  content.innerHTML = "";
  content.appendChild(card);
}

async function renderDocument(path) {
  var content = document.getElementById("content");
  var existing = content.querySelector(".card");
  if (existing) {
    existing.classList.add("fade-out");
    await new Promise(function(r) { setTimeout(r, 300); });
  }

  var html = await fetchContent(path);
  var pkg = null;
  var concept = null;
  if (state.manifest) {
    state.manifest.packages.forEach(function(p) {
      if (path === p.path || path.startsWith(p.path + "/")) {
        pkg = p;
        if (p.children) {
          p.children.forEach(function(c) {
            if (c.path === path) concept = c;
          });
        }
      }
    });
  }

  var meta = concept || pkg;
  var metaHtml = "";
  if (meta) {
    var parts = [];
    if (meta.status) parts.push("status: " + meta.status);
    if (meta.tags && meta.tags.length) parts.push("tags: " + meta.tags.join(", "));
    if (parts.length) metaHtml = '<div class="content-meta">' + parts.join(" / ") + "</div>";
  }

  var card = document.createElement("div");
  card.className = "card fade-out";
  card.innerHTML = metaHtml + '<div class="content-body">' + html + "</div>";
  content.innerHTML = "";
  content.appendChild(card);
  requestAnimationFrame(function() {
    requestAnimationFrame(function() {
      card.classList.remove("fade-out");
    });
  });
}

async function onRouteChange() {
  var path = getHashPath();
  state.currentPath = path || null;
  if (state.currentPath) {
    var parts = state.currentPath.split("/");
    state.expandedPkg = parts[0];
  }
  renderNav();
  if (!path) {
    renderHomepage();
    renderGraph(null);
  } else {
    await renderDocument(path);
    renderGraph(path);
  }
}

window.addEventListener("hashchange", onRouteChange);

fetchManifest().then(function() {
  onRouteChange();
});
```

**Step 2: Build and verify navigation works**

Run: `go build -o eudemics . && ./eudemics --port 9090 &`
Open browser to `http://localhost:9090/` - verify homepage loads.
Navigate to `http://localhost:9090/#/autonomy/liberty` - verify document loads.
Run: `kill %1`

**Step 3: Commit**

```bash
git add views/index.html
git commit -m "feat: SPA router, nav, and content loading with transitions"
```

---

### Task 9: Graph Renderer - Force-Directed Engine

**Files:**
- Modify: `views/index.html` (add graph code before the route handler)

**Step 1: Add the graph rendering engine inside the `<script>` tag, before the `renderGraph` call**

```javascript
var graphState = {
  nodes: [],
  edges: [],
  hoveredNode: null,
  animFrame: null,
  canvas: null,
  ctx: null,
  width: 0,
  height: 0,
  dpr: 1
};

function initGraphCanvas() {
  graphState.canvas = document.getElementById("graph-canvas");
  graphState.ctx = graphState.canvas.getContext("2d");
  graphState.dpr = window.devicePixelRatio || 1;
  resizeGraph();
}

function resizeGraph() {
  var panel = document.getElementById("graph-panel");
  graphState.width = panel.clientWidth;
  graphState.height = panel.clientHeight;
  graphState.canvas.width = graphState.width * graphState.dpr;
  graphState.canvas.height = graphState.height * graphState.dpr;
  graphState.ctx.scale(graphState.dpr, graphState.dpr);
}

function buildGraphData(focusPath) {
  var nodes = [];
  var edges = [];
  var nodeMap = {};

  if (!state.manifest) return { nodes: nodes, edges: edges };

  if (!focusPath) {
    var count = state.manifest.packages.length;
    state.manifest.packages.forEach(function(pkg, i) {
      var angle = (2 * Math.PI * i) / count - Math.PI / 2;
      var rx = graphState.width * 0.35;
      var ry = graphState.height * 0.35;
      var cx = graphState.width / 2;
      var cy = graphState.height / 2;
      nodes.push({
        name: pkg.name,
        path: pkg.path,
        x: cx + rx * Math.cos(angle),
        y: cy + ry * Math.sin(angle),
        vx: 0, vy: 0,
        radius: 6,
        color: "#1a6b5a",
        fixed: true
      });
      nodeMap[pkg.path] = nodes[nodes.length - 1];
    });
    return { nodes: nodes, edges: edges };
  }

  var focusPkg = null;
  var focusConcept = null;
  state.manifest.packages.forEach(function(pkg) {
    if (focusPath === pkg.path) focusPkg = pkg;
    if (pkg.children) {
      pkg.children.forEach(function(c) {
        if (c.path === focusPath) { focusPkg = pkg; focusConcept = c; }
      });
    }
  });

  var source = focusConcept || focusPkg;
  if (!source) return { nodes: nodes, edges: edges };

  var cx = graphState.width / 2;
  var cy = graphState.height / 2;
  nodes.push({
    name: source.name,
    path: source.path || focusPath,
    x: cx, y: cy,
    vx: 0, vy: 0,
    radius: 8,
    color: "#1a6b5a",
    fixed: true
  });
  nodeMap[focusPath] = nodes[0];

  var rels = [];
  if (source.builds_on) source.builds_on.forEach(function(t) { rels.push({ target: t, type: "builds_on" }); });
  if (source.constrains) source.constrains.forEach(function(t) { rels.push({ target: t, type: "constrains" }); });
  if (source.implements) source.implements.forEach(function(t) { rels.push({ target: t, type: "implements" }); });
  if (source.related) source.related.forEach(function(t) { rels.push({ target: t, type: "related" }); });

  var relCount = rels.length || 1;
  rels.forEach(function(rel, i) {
    var angle = (2 * Math.PI * i) / relCount - Math.PI / 2;
    var dist = Math.min(graphState.width, graphState.height) * 0.3;
    if (!nodeMap[rel.target]) {
      var parts = rel.target.split("/");
      nodes.push({
        name: parts[parts.length - 1],
        path: rel.target,
        x: cx + dist * Math.cos(angle),
        y: cy + dist * Math.sin(angle),
        vx: 0, vy: 0,
        radius: 5,
        color: "#5A5A5A",
        fixed: false
      });
      nodeMap[rel.target] = nodes[nodes.length - 1];
    }
    edges.push({
      source: nodeMap[focusPath],
      target: nodeMap[rel.target],
      type: rel.type
    });
  });

  return { nodes: nodes, edges: edges };
}

function drawGraph() {
  var ctx = graphState.ctx;
  ctx.clearRect(0, 0, graphState.width, graphState.height);

  graphState.edges.forEach(function(edge) {
    ctx.beginPath();
    ctx.moveTo(edge.source.x, edge.source.y);
    ctx.lineTo(edge.target.x, edge.target.y);
    ctx.strokeStyle = "rgba(26, 107, 90, 0.3)";
    ctx.lineWidth = 1;
    if (edge.type === "constrains") ctx.setLineDash([4, 4]);
    else if (edge.type === "related") ctx.setLineDash([2, 2]);
    else if (edge.type === "implements") ctx.setLineDash([]);
    else ctx.setLineDash([]);
    ctx.stroke();
    ctx.setLineDash([]);

    if (edge.type === "implements") {
      var dx = edge.target.x - edge.source.x;
      var dy = edge.target.y - edge.source.y;
      var len = Math.sqrt(dx * dx + dy * dy);
      if (len > 0) {
        var ux = dx / len;
        var uy = dy / len;
        var ax = edge.target.x - ux * (edge.target.radius + 4);
        var ay = edge.target.y - uy * (edge.target.radius + 4);
        ctx.beginPath();
        ctx.moveTo(ax, ay);
        ctx.lineTo(ax - ux * 6 + uy * 3, ay - uy * 6 - ux * 3);
        ctx.lineTo(ax - ux * 6 - uy * 3, ay - uy * 6 + ux * 3);
        ctx.closePath();
        ctx.fillStyle = "rgba(26, 107, 90, 0.3)";
        ctx.fill();
      }
    }
  });

  graphState.nodes.forEach(function(node) {
    var hovered = graphState.hoveredNode === node;
    ctx.beginPath();
    ctx.arc(node.x, node.y, hovered ? node.radius + 2 : node.radius, 0, 2 * Math.PI);
    ctx.fillStyle = hovered ? "#00e5c3" : node.color;
    ctx.fill();
    if (hovered) {
      ctx.strokeStyle = "#00e5c3";
      ctx.lineWidth = 2;
      ctx.stroke();
    }

    ctx.fillStyle = "var(--text)";
    ctx.font = "10px terminus, Courier New, monospace";
    ctx.textAlign = "center";
    ctx.fillStyle = hovered ? "#0a2e24" : "#5A5A5A";
    ctx.fillText(node.name, node.x, node.y + node.radius + 12);
  });
}

function renderGraph(focusPath) {
  if (!graphState.canvas) initGraphCanvas();
  resizeGraph();
  var data = buildGraphData(focusPath);
  graphState.nodes = data.nodes;
  graphState.edges = data.edges;
  drawGraph();
}

function graphHitTest(x, y) {
  for (var i = 0; i < graphState.nodes.length; i++) {
    var n = graphState.nodes[i];
    var dx = x - n.x;
    var dy = y - n.y;
    if (dx * dx + dy * dy <= (n.radius + 4) * (n.radius + 4)) return n;
  }
  return null;
}

document.getElementById("graph-canvas").addEventListener("mousemove", function(e) {
  var rect = graphState.canvas.getBoundingClientRect();
  var x = e.clientX - rect.left;
  var y = e.clientY - rect.top;
  var node = graphHitTest(x, y);
  graphState.hoveredNode = node;
  graphState.canvas.style.cursor = node ? "pointer" : "default";
  drawGraph();
});

document.getElementById("graph-canvas").addEventListener("click", function(e) {
  var rect = graphState.canvas.getBoundingClientRect();
  var x = e.clientX - rect.left;
  var y = e.clientY - rect.top;
  var node = graphHitTest(x, y);
  if (node) navigate(node.path);
});
```

**Step 2: Build and verify the graph renders**

Run: `go build -o eudemics . && ./eudemics --port 9090 &`
Open browser to `http://localhost:9090/` - verify homepage shows radial graph in bottom-left.
Navigate to a concept with relationships - verify neighborhood graph.
Run: `kill %1`

**Step 3: Commit**

```bash
git add views/index.html
git commit -m "feat: canvas graph renderer with radial and neighborhood views"
```

---

### Task 10: GitHub Actions Workflow

**Files:**
- Create: `.github/workflows/release.yml`

**Step 1: Create the workflow**

```yaml
name: release
on:
  push:
    tags:
      - "v*"
permissions:
  contents: write
jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, darwin, windows]
        goarch: [amd64, arm64]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24"
      - name: build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          ext=""
          if [ "$GOOS" = "windows" ]; then ext=".exe"; fi
          go build -o eudemics-${GOOS}-${GOARCH}${ext} .
      - name: upload
        uses: actions/upload-artifact@v4
        with:
          name: eudemics-${{ matrix.goos }}-${{ matrix.goarch }}
          path: eudemics-*
  release:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@v4
        with:
          merge-multiple: true
      - name: create release
        uses: softprops/action-gh-release@v2
        with:
          files: eudemics-*
          generate_release_notes: true
```

**Step 2: Verify YAML syntax**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))" 2>&1`
Expected: no output (valid YAML)

**Step 3: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "ci: cross-platform release workflow"
```

---

### Task 11: Final Integration + Polish

**Files:**
- Modify: `main.go` (ensure all imports clean)
- Modify: `views/index.html` (any CSS tweaks)

**Step 1: Run all tests**

Run: `go test -v ./...`
Expected: all PASS

**Step 2: Run go vet**

Run: `go vet ./...`
Expected: clean

**Step 3: Full manual smoke test**

Run: `go build -o eudemics . && ./eudemics --port 9090`

Verify in browser:
- Homepage loads with title and description
- Bottom-left graph shows all packages in radial layout
- Right nav lists all packages
- Clicking a package expands sub-topics
- Clicking a sub-topic loads document with fade transition
- Graph updates to show neighborhood
- Clicking graph nodes navigates
- Hash URLs update and browser back/forward works
- `curl http://localhost:9090/api/manifest` returns valid JSON
- `curl http://localhost:9090/api/content/autonomy` returns HTML

**Step 4: Commit any polish**

```bash
git add -A
git commit -m "feat: spa wiki viewer complete"
```
