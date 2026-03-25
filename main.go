package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
)

//go:embed views/index.html
var indexHTML []byte

//go:embed views/fonts
var fontsFS embed.FS

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
	rest := data[4:]
	end := bytes.Index(rest, []byte("\n---\n"))
	if end == -1 {
		return fm, data, nil
	}
	yamlBlock := rest[:end]
	body := rest[end+5:]
	if err := yaml.Unmarshal(yamlBlock, &fm); err != nil {
		return fm, data, fmt.Errorf("parse frontmatter: %w", err)
	}
	return fm, body, nil
}

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

var wikiLinkRe = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

func renderMarkdown(source []byte) (string, error) {
	source = wikiLinkRe.ReplaceAll(source, []byte(`[$1](#/$1)`))
	var buf bytes.Buffer
	if err := md.Convert(source, &buf); err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}
	return buf.String(), nil
}

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

var skipDirs = map[string]bool{
	".git":    true,
	"docs":    true,
	"views":   true,
	".github": true,
	"cmd":     true,
}

func buildManifest(root string) (Manifest, error) {
	packages := map[string]*ConceptNode{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "CLAUDE.md" {
			return nil
		}
		parts := strings.Split(rel, string(filepath.Separator))
		// depth 2: <pkg>/README.md
		// depth 3: <pkg>/<child>/README.md
		if len(parts) < 2 {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		fm, _, err := parseFrontmatter(data)
		if err != nil {
			return err
		}

		pkgName := parts[0]
		if _, ok := packages[pkgName]; !ok {
			packages[pkgName] = &ConceptNode{Name: pkgName, Path: pkgName}
		}
		pkg := packages[pkgName]

		if len(parts) == 2 {
			// package-level README
			pkg.Status = fm.Status
			pkg.Tags = fm.Tags
			pkg.BuildsOn = fm.BuildsOn
			pkg.Constrains = fm.Constrains
			pkg.Implements = fm.Implements
			pkg.Related = fm.Related
		} else if len(parts) == 3 {
			// child concept README
			childName := parts[1]
			childPath := pkgName + "/" + childName
			child := &ConceptNode{
				Name:       childName,
				Path:       childPath,
				Status:     fm.Status,
				Tags:       fm.Tags,
				BuildsOn:   fm.BuildsOn,
				Constrains: fm.Constrains,
				Implements: fm.Implements,
				Related:    fm.Related,
			}
			pkg.Children = append(pkg.Children, child)
		}
		return nil
	})
	if err != nil {
		return Manifest{}, fmt.Errorf("walk %s: %w", root, err)
	}

	result := make([]*ConceptNode, 0, len(packages))
	for _, pkg := range packages {
		sort.Slice(pkg.Children, func(i, j int) bool {
			return pkg.Children[i].Name < pkg.Children[j].Name
		})
		result = append(result, pkg)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return Manifest{Packages: result}, nil
}

var (
	manifestCache     []byte
	manifestCacheLock sync.RWMutex
)

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

func main() {
	port := flag.Int("port", 8080, "server port")
	contentDir := flag.String("content", ".", "path to content directory")
	repo := flag.String("repo", "", "git repo URL to clone if content dir is empty")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	if *repo != "" {
		entries, err := os.ReadDir(*contentDir)
		if err != nil || len(entries) == 0 {
			if err := cloneRepo(*repo, *contentDir); err != nil {
				slog.Error("failed to clone", "err", err)
				os.Exit(1)
			}
		}
	}

	if err := cacheManifest(*contentDir); err != nil {
		slog.Error("failed to build manifest", "err", err)
		os.Exit(1)
	}

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

	slog.Info("starting eudemics", "port", *port, "content", *contentDir, "repo", *repo)

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
		if err == nil && info.IsDir() {
			full = filepath.Join(full, "README.md")
		} else if !strings.HasSuffix(full, ".md") {
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

		rendered, err := renderMarkdown(body)
		if err != nil {
			http.Error(w, "render error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(rendered))
	})

	fontsDir, _ := fs.Sub(fontsFS, "views/fonts")
	mux.Handle("GET /fonts/", http.StripPrefix("/fonts/", http.FileServerFS(fontsDir)))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
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
