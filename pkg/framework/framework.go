// Package framework detects the JavaScript/TypeScript framework used by a project, by
// inspecting the "dependencies" and "devDependencies" of the nearest package.json file.
package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/wakatime/wakatime-cli/pkg/file"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// packageJSONFilename is the name of the npm package manifest file used to detect frameworks.
const packageJSONFilename = "package.json"

// Framework represents a JavaScript/TypeScript framework that can be detected from a
// package.json dependency.
type Framework struct {
	// Name is the human readable framework name reported on the heartbeat.
	Name string
	// Dependency is the package name looked up in "dependencies" and "devDependencies".
	Dependency string
}

// frameworks is the registry of detectable frameworks, ordered from most to least specific.
// Meta-frameworks (Next.js, Nuxt, SvelteKit, Remix, Qwik, Astro) are listed before the base
// libraries they are commonly built on top of (React, Vue, Svelte), so a project depending on
// both is attributed to the more specific framework. Adding support for another framework only
// requires appending an entry here.
//
// nolint:gochecknoglobals
var frameworks = []Framework{
	{Name: "Next.js", Dependency: "next"},
	{Name: "Nuxt", Dependency: "nuxt"},
	{Name: "SvelteKit", Dependency: "@sveltejs/kit"},
	{Name: "Remix", Dependency: "@remix-run/react"},
	{Name: "Qwik", Dependency: "@builder.io/qwik"},
	{Name: "Astro", Dependency: "astro"},
	{Name: "NestJS", Dependency: "@nestjs/core"},
	{Name: "Angular", Dependency: "@angular/core"},
	{Name: "SolidJS", Dependency: "solid-js"},
	{Name: "Preact", Dependency: "preact"},
	{Name: "React", Dependency: "react"},
	{Name: "Vue", Dependency: "vue"},
	{Name: "Svelte", Dependency: "svelte"},
}

// supportedLanguages restricts framework detection to languages where a JS/TS framework is
// meaningful. Files detected as any other language are left untouched, even if a package.json
// happens to be nearby.
//
// nolint:gochecknoglobals
var supportedLanguages = map[string]struct{}{
	heartbeat.LanguageJavaScript.String(): {},
	heartbeat.LanguageTypeScript.String(): {},
	heartbeat.LanguageJSX.String():        {},
	heartbeat.LanguageTSX.String():        {},
	heartbeat.LanguageVueJS.String():      {},
	heartbeat.LanguageSvelte.String():     {},
}

// packageJSON is the subset of a package.json file relevant for framework detection.
type packageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// cacheEntry stores the outcome of parsing a single package.json file, including a failed
// attempt, so it's never parsed more than once per run.
type cacheEntry struct {
	pkg *packageJSON
	err error
}

// cache memoizes parsed package.json files by absolute path, so heartbeats for files within
// the same project only pay the cost of reading and parsing package.json once. It is safe for
// concurrent use.
type cache struct {
	mu    sync.Mutex
	items map[string]cacheEntry
}

func newCache() *cache {
	return &cache{items: make(map[string]cacheEntry)}
}

// WithDetection initializes and returns a heartbeat handle option, which can be used in a
// heartbeat processing pipeline to detect and add the JavaScript/TypeScript framework used by
// a project to heartbeats of entity type 'file'.
//
// It relies on the heartbeat's already-detected language, so it must run after
// language.WithDetection in the pipeline. If no framework is detected, the heartbeat is left
// unchanged.
func WithDetection() heartbeat.HandleOption {
	c := newCache()

	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(ctx context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			logger := log.Extract(ctx)

			for n, h := range hh {
				name, ok := detectFramework(ctx, c, h)
				if !ok {
					continue
				}

				logger.Debugf("detected framework %q for file %q", name, h.Entity)

				hh[n].Framework = heartbeat.PointerTo(name)
			}

			return next(ctx, hh)
		}
	}
}

// detectFramework returns the framework name for a single heartbeat, or false if none applies
// or none was detected. It never returns an error, so a missing, malformed, or unreadable
// package.json simply results in no framework being set, leaving existing behavior unchanged.
func detectFramework(ctx context.Context, c *cache, h heartbeat.Heartbeat) (string, bool) {
	if h.Framework != nil || h.Language == nil {
		return "", false
	}

	if _, ok := supportedLanguages[*h.Language]; !ok {
		return "", false
	}

	fp := h.Entity
	if h.LocalFile != "" {
		fp = h.LocalFile
	}

	return detect(ctx, c, fp)
}

// detect searches upward from fp for the nearest package.json and returns the name of the
// first known framework found among its dependencies and devDependencies.
func detect(ctx context.Context, c *cache, fp string) (string, bool) {
	pkgPath, found := file.Find(ctx, filepath.Dir(fp), packageJSONFilename)
	if !found {
		return "", false
	}

	pkg, err := loadPackageJSON(c, pkgPath)
	if err != nil {
		log.Extract(ctx).Debugf("failed to parse %q: %s", pkgPath, err)
		return "", false
	}

	for _, fw := range frameworks {
		if _, ok := pkg.Dependencies[fw.Dependency]; ok {
			return fw.Name, true
		}

		if _, ok := pkg.DevDependencies[fw.Dependency]; ok {
			return fw.Name, true
		}
	}

	return "", false
}

// loadPackageJSON reads and parses the package.json file at path, using c to make sure it's
// never read or parsed more than once per run, regardless of how many heartbeats share it.
func loadPackageJSON(c *cache, path string) (*packageJSON, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.items[path]; ok {
		return entry.pkg, entry.err
	}

	pkg, err := parsePackageJSON(path)

	c.items[path] = cacheEntry{pkg: pkg, err: err}

	return pkg, err
}

// parsePackageJSON reads and unmarshals a package.json file from disk.
func parsePackageJSON(path string) (*packageJSON, error) {
	data, err := os.ReadFile(path) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return &pkg, nil
}
