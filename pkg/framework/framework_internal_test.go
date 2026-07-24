package framework

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writePackageJSON writes content to <dir>/package.json, creating dir if needed.
func writePackageJSON(t *testing.T, dir string, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(dir, 0750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(content), 0600))
}

// packageJSONWith builds a minimal package.json body from dependency and devDependency names.
func packageJSONWith(deps []string, devDeps []string) string {
	depsObj := "{"
	for i, d := range deps {
		if i > 0 {
			depsObj += ","
		}
		depsObj += `"` + d + `": "^1.0.0"`
	}
	depsObj += "}"

	devDepsObj := "{"
	for i, d := range devDeps {
		if i > 0 {
			devDepsObj += ","
		}
		devDepsObj += `"` + d + `": "^1.0.0"`
	}
	devDepsObj += "}"

	return `{"name": "fixture", "dependencies": ` + depsObj + `, "devDependencies": ` + devDepsObj + `}`
}

func TestDetect_Frameworks(t *testing.T) {
	tests := map[string]struct {
		Deps     []string
		Expected string
	}{
		"React":     {Deps: []string{"react"}, Expected: "React"},
		"Next.js":   {Deps: []string{"react", "next"}, Expected: "Next.js"},
		"Vue":       {Deps: []string{"vue"}, Expected: "Vue"},
		"Nuxt":      {Deps: []string{"vue", "nuxt"}, Expected: "Nuxt"},
		"Angular":   {Deps: []string{"@angular/core"}, Expected: "Angular"},
		"NestJS":    {Deps: []string{"@nestjs/core"}, Expected: "NestJS"},
		"Svelte":    {Deps: []string{"svelte"}, Expected: "Svelte"},
		"SvelteKit": {Deps: []string{"svelte", "@sveltejs/kit"}, Expected: "SvelteKit"},
		"SolidJS":   {Deps: []string{"solid-js"}, Expected: "SolidJS"},
		"Preact":    {Deps: []string{"preact"}, Expected: "Preact"},
		"Qwik":      {Deps: []string{"@builder.io/qwik"}, Expected: "Qwik"},
		"Remix":     {Deps: []string{"react", "@remix-run/react"}, Expected: "Remix"},
		"Astro":     {Deps: []string{"astro"}, Expected: "Astro"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tmpDir := t.TempDir()
			writePackageJSON(t, tmpDir, packageJSONWith(test.Deps, nil))

			entity := filepath.Join(tmpDir, "src", "index.js")
			require.NoError(t, os.MkdirAll(filepath.Dir(entity), 0750))

			got, ok := detect(t.Context(), newCache(), entity)
			require.True(t, ok)
			assert.Equal(t, test.Expected, got)
		})
	}
}

func TestDetect_PlainJavaScriptNoFramework(t *testing.T) {
	tmpDir := t.TempDir()
	writePackageJSON(t, tmpDir, packageJSONWith([]string{"lodash", "axios"}, nil))

	entity := filepath.Join(tmpDir, "index.js")

	name, ok := detect(t.Context(), newCache(), entity)
	assert.False(t, ok)
	assert.Empty(t, name)
}

func TestDetect_PlainJSXNoDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	writePackageJSON(t, tmpDir, packageJSONWith(nil, nil))

	entity := filepath.Join(tmpDir, "index.jsx")

	name, ok := detect(t.Context(), newCache(), entity)
	assert.False(t, ok)
	assert.Empty(t, name)
}

func TestDetect_NoPackageJSON(t *testing.T) {
	tmpDir := t.TempDir()
	entity := filepath.Join(tmpDir, "index.js")

	name, ok := detect(t.Context(), newCache(), entity)
	assert.False(t, ok)
	assert.Empty(t, name)
}

func TestDetect_InvalidPackageJSON(t *testing.T) {
	tmpDir := t.TempDir()
	writePackageJSON(t, tmpDir, `{not valid json`)

	entity := filepath.Join(tmpDir, "index.js")

	name, ok := detect(t.Context(), newCache(), entity)
	assert.False(t, ok)
	assert.Empty(t, name)
}

func TestDetect_NestedPackageJSONLookup(t *testing.T) {
	tmpDir := t.TempDir()
	writePackageJSON(t, tmpDir, packageJSONWith([]string{"react"}, nil))

	// entity lives several directories below the package.json.
	entity := filepath.Join(tmpDir, "src", "components", "deeply", "nested", "Button.jsx")
	require.NoError(t, os.MkdirAll(filepath.Dir(entity), 0750))

	name, ok := detect(t.Context(), newCache(), entity)
	require.True(t, ok)
	assert.Equal(t, "React", name)
}

func TestDetect_DevDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	writePackageJSON(t, tmpDir, packageJSONWith(nil, []string{"svelte"}))

	entity := filepath.Join(tmpDir, "App.svelte")

	name, ok := detect(t.Context(), newCache(), entity)
	require.True(t, ok)
	assert.Equal(t, "Svelte", name)
}

func TestDetect_UsesNearestPackageJSONNotFarthest(t *testing.T) {
	root := t.TempDir()
	writePackageJSON(t, root, packageJSONWith([]string{"vue"}, nil))

	sub := filepath.Join(root, "packages", "app")
	writePackageJSON(t, sub, packageJSONWith([]string{"react"}, nil))

	entity := filepath.Join(sub, "src", "index.jsx")
	require.NoError(t, os.MkdirAll(filepath.Dir(entity), 0750))

	name, ok := detect(t.Context(), newCache(), entity)
	require.True(t, ok)
	assert.Equal(t, "React", name)
}

func TestLoadPackageJSON_CachesResult(t *testing.T) {
	tmpDir := t.TempDir()
	pkgPath := filepath.Join(tmpDir, "package.json")
	require.NoError(t, os.WriteFile(pkgPath, []byte(packageJSONWith([]string{"react"}, nil)), 0600))

	c := newCache()

	pkg1, err := loadPackageJSON(c, pkgPath)
	require.NoError(t, err)
	require.NotNil(t, pkg1)

	// Remove the file from disk. If loadPackageJSON reads from disk again instead of using
	// the cache, this second call would fail.
	require.NoError(t, os.Remove(pkgPath))

	pkg2, err := loadPackageJSON(c, pkgPath)
	require.NoError(t, err)
	assert.Same(t, pkg1, pkg2)
}

func TestLoadPackageJSON_CachesFailure(t *testing.T) {
	tmpDir := t.TempDir()
	pkgPath := filepath.Join(tmpDir, "package.json")
	require.NoError(t, os.WriteFile(pkgPath, []byte(`{not valid`), 0600))

	c := newCache()

	_, err1 := loadPackageJSON(c, pkgPath)
	require.Error(t, err1)

	// Fix the file on disk. A cache hit should still return the original (cached) error
	// rather than re-reading, proving the failed parse was memoized too.
	require.NoError(t, os.WriteFile(pkgPath, []byte(packageJSONWith([]string{"react"}, nil)), 0600))

	_, err2 := loadPackageJSON(c, pkgPath)
	require.Error(t, err2)
	assert.Equal(t, err1, err2)
}

func TestDetectFramework_SkipsUnsupportedLanguage(t *testing.T) {
	tmpDir := t.TempDir()
	writePackageJSON(t, tmpDir, packageJSONWith([]string{"react"}, nil))

	entity := filepath.Join(tmpDir, "main.py")

	h := heartbeat.Heartbeat{
		Entity:   entity,
		Language: heartbeat.PointerTo(heartbeat.LanguagePython.String()),
	}

	name, ok := detectFramework(t.Context(), newCache(), h)
	assert.False(t, ok)
	assert.Empty(t, name)
}

func TestDetectFramework_SkipsNilLanguage(t *testing.T) {
	tmpDir := t.TempDir()
	writePackageJSON(t, tmpDir, packageJSONWith([]string{"react"}, nil))

	entity := filepath.Join(tmpDir, "index.jsx")

	h := heartbeat.Heartbeat{Entity: entity}

	name, ok := detectFramework(t.Context(), newCache(), h)
	assert.False(t, ok)
	assert.Empty(t, name)
}

func TestDetectFramework_SkipsAlreadySetFramework(t *testing.T) {
	tmpDir := t.TempDir()
	writePackageJSON(t, tmpDir, packageJSONWith([]string{"vue"}, nil))

	entity := filepath.Join(tmpDir, "index.jsx")

	h := heartbeat.Heartbeat{
		Entity:    entity,
		Language:  heartbeat.PointerTo(heartbeat.LanguageJSX.String()),
		Framework: heartbeat.PointerTo("React"),
	}

	name, ok := detectFramework(t.Context(), newCache(), h)
	assert.False(t, ok)
	assert.Empty(t, name)
}

func TestDetectFramework_PrefersLocalFileOverEntity(t *testing.T) {
	tmpDir := t.TempDir()
	writePackageJSON(t, tmpDir, packageJSONWith([]string{"react"}, nil))

	entity := filepath.Join(t.TempDir(), "index.jsx") // unrelated tree, no package.json
	localFile := filepath.Join(tmpDir, "index.jsx")

	h := heartbeat.Heartbeat{
		Entity:    entity,
		LocalFile: localFile,
		Language:  heartbeat.PointerTo(heartbeat.LanguageJSX.String()),
	}

	name, ok := detectFramework(t.Context(), newCache(), h)
	require.True(t, ok)
	assert.Equal(t, "React", name)
}

func TestWithDetection_SetsFrameworkOnHeartbeats(t *testing.T) {
	tmpDir := t.TempDir()
	writePackageJSON(t, tmpDir, packageJSONWith([]string{"next", "react"}, nil))

	entity := filepath.Join(tmpDir, "pages", "index.tsx")
	require.NoError(t, os.MkdirAll(filepath.Dir(entity), 0750))

	var received []heartbeat.Heartbeat

	next := func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		received = hh
		return []heartbeat.Result{}, nil
	}

	handle := WithDetection()(next)

	hh := []heartbeat.Heartbeat{
		{
			Entity:   entity,
			Language: heartbeat.PointerTo(heartbeat.LanguageTSX.String()),
		},
	}

	_, err := handle(t.Context(), hh)
	require.NoError(t, err)

	require.Len(t, received, 1)
	require.NotNil(t, received[0].Framework)
	assert.Equal(t, "Next.js", *received[0].Framework)
}

func TestWithDetection_LeavesHeartbeatUnchangedWhenNoFrameworkDetected(t *testing.T) {
	tmpDir := t.TempDir()

	entity := filepath.Join(tmpDir, "main.py")

	var received []heartbeat.Heartbeat

	next := func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		received = hh
		return []heartbeat.Result{}, nil
	}

	handle := WithDetection()(next)

	hh := []heartbeat.Heartbeat{
		{
			Entity:   entity,
			Language: heartbeat.PointerTo(heartbeat.LanguagePython.String()),
		},
	}

	_, err := handle(t.Context(), hh)
	require.NoError(t, err)

	require.Len(t, received, 1)
	assert.Nil(t, received[0].Framework)
}
