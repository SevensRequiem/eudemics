package main

import (
	"strings"
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

func TestParseFrontmatterNoFrontmatter(t *testing.T) {
	input := []byte("# Just a heading\n\nNo frontmatter here.\n")
	fm, body, err := parseFrontmatter(input)
	if err != nil {
		t.Fatal(err)
	}
	if fm.Status != "" {
		t.Errorf("expected empty status, got %q", fm.Status)
	}
	if string(body) != string(input) {
		t.Error("body should equal input when no frontmatter")
	}
}

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
