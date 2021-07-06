package deps_test

import (
	"regexp"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/deps"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/regex"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithDetection(t *testing.T) {
	opt := deps.WithDetection(deps.Config{})

	h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Dependencies: []string{
					"os",
					"github.com/wakatime/wakatime-cli/pkg/heartbeat",
				},
				Entity:     "testdata/golang_minimal.go",
				EntityType: heartbeat.FileType,
				Language:   heartbeat.String("Go"),
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := h([]heartbeat.Heartbeat{{
		Entity:     "testdata/golang_minimal.go",
		EntityType: heartbeat.FileType,
		Language:   heartbeat.String("Go"),
	}})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestWithDetection_SkipSanitized(t *testing.T) {
	opt := deps.WithDetection(deps.Config{
		FilePatterns: []regex.Regex{regexp.MustCompile(".*")},
	})

	h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Len(t, hh[0].Dependencies, 0)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := h([]heartbeat.Heartbeat{{
		Entity:     "testdata/golang.go",
		EntityType: heartbeat.FileType,
		Language:   heartbeat.String("Go"),
	}})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestWithDetection_LocalFile(t *testing.T) {
	opt := deps.WithDetection(deps.Config{})

	h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Dependencies: []string{
					"os",
					"github.com/wakatime/wakatime-cli/pkg/heartbeat",
				},
				Entity:     "testdata/golang.go",
				EntityType: heartbeat.FileType,
				Language:   heartbeat.String("Go"),
				LocalFile:  "testdata/golang_minimal.go",
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := h([]heartbeat.Heartbeat{{
		Entity:     "testdata/golang.go",
		EntityType: heartbeat.FileType,
		Language:   heartbeat.String("Go"),
		LocalFile:  "testdata/golang_minimal.go",
	}})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestWithDetection_NonFileType(t *testing.T) {
	opt := deps.WithDetection(deps.Config{})

	h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Entity:     "testdata/codefiles/golang.go",
				EntityType: heartbeat.AppType,
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := h([]heartbeat.Heartbeat{{
		Entity:     "testdata/codefiles/golang.go",
		EntityType: heartbeat.AppType,
	}})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestDetect(t *testing.T) {
	tests := map[string]struct {
		Filepath     string
		Language     heartbeat.Language
		Dependencies []string
	}{
		"c": {
			Filepath:     "testdata/c_minimal.c",
			Language:     heartbeat.LanguageC,
			Dependencies: []string{"openssl"},
		},
		"cpp": {
			Filepath:     "testdata/cpp_minimal.cpp",
			Language:     heartbeat.LanguageCPP,
			Dependencies: []string{"iostream"},
		},
		"csharp": {
			Filepath:     "testdata/csharp_minimal.cs",
			Language:     heartbeat.LanguageCSharp,
			Dependencies: []string{"WakaTime"},
		},
		"elm": {
			Filepath:     "testdata/elm_minimal.elm",
			Language:     heartbeat.LanguageElm,
			Dependencies: []string{"Html"},
		},
		"golang": {
			Filepath: "testdata/golang_minimal.go",
			Language: heartbeat.LanguageGo,
			Dependencies: []string{
				"os",
				"github.com/wakatime/wakatime-cli/pkg/heartbeat",
			},
		},
		"haskell": {
			Filepath:     "testdata/haskell_minimal.hs",
			Language:     heartbeat.LanguageHaskell,
			Dependencies: []string{"Control"},
		},
		"haxe": {
			Filepath:     "testdata/haxe_minimal.hx",
			Language:     heartbeat.LanguageHaxe,
			Dependencies: []string{"alpha"},
		},
		"html": {
			Filepath:     "testdata/html_minimal.html",
			Language:     heartbeat.LanguageHTML,
			Dependencies: []string{`"https://cdn.wakatime.com/app.min.js"`},
		},
		"java": {
			Filepath:     "testdata/java_minimal.java",
			Language:     heartbeat.LanguageJava,
			Dependencies: []string{"foobar"},
		},
		"javascript": {
			Filepath:     "testdata/es6_minimal.js",
			Language:     heartbeat.LanguageJavaScript,
			Dependencies: []string{"bravo"},
		},
		"json": {
			Filepath:     "testdata/bower_minimal.json",
			Language:     heartbeat.LanguageJSON,
			Dependencies: []string{"bootstrap"},
		},
		"kotlin": {
			Filepath:     "testdata/kotlin_minimal.kt",
			Language:     heartbeat.LanguageKotlin,
			Dependencies: []string{"alpha.time"},
		},
		"objective-c": {
			Filepath:     "testdata/objective_c_minimal.m",
			Language:     heartbeat.LanguageObjectiveC,
			Dependencies: []string{"Foundation"},
		},
		"php": {
			Filepath:     "testdata/php_minimal.php",
			Language:     heartbeat.LanguagePHP,
			Dependencies: []string{"Interop", "FooBarOne"},
		},
		"python": {
			Filepath:     "testdata/python_minimal.py",
			Language:     heartbeat.LanguagePython,
			Dependencies: []string{"flask", "simplejson"},
		},
		"rust": {
			Filepath:     "testdata/rust_minimal.rs",
			Language:     heartbeat.LanguageRust,
			Dependencies: []string{"syn"},
		},
		"scala": {
			Filepath:     "testdata/scala_minimal.scala",
			Language:     heartbeat.LanguageScala,
			Dependencies: []string{"com.alpha.SomeClass"},
		},
		"swift": {
			Filepath:     "testdata/swift_minimal.swift",
			Language:     heartbeat.LanguageSwift,
			Dependencies: []string{"Swift"},
		},
		"typescript": {
			Filepath:     "testdata/typescript_minimal.ts",
			Language:     heartbeat.LanguageTypeScript,
			Dependencies: []string{"bravo"},
		},
		"unknown": {
			Filepath:     "testdata/Gruntfile",
			Language:     heartbeat.LanguageUnknown,
			Dependencies: []string{"grunt"},
		},
		"vb.net": {
			Filepath:     "testdata/vbnet_minimal.vb",
			Language:     heartbeat.LanguageVBNet,
			Dependencies: []string{"WakaTime"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			deps, err := deps.Detect(test.Filepath, test.Language)
			require.NoError(t, err)

			assert.Equal(t, test.Dependencies, deps)
		})
	}
}

func TestDetect_DuplicatesRemoved(t *testing.T) {
	deps, err := deps.Detect(
		"testdata/golang_duplicate.go",
		heartbeat.LanguageGo,
	)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"os",
	}, deps)
}

func TestDetect_LongDependenciesRemoved(t *testing.T) {
	deps, err := deps.Detect(
		"testdata/python_with_long_import.py",
		heartbeat.LanguagePython,
	)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"django",
		"flask",
		// nolint:lll
		"notlongenoughnotlongenoughnotlongenoughnotlongenoughnotlongenoughnotlongenoughnotlongenoughnotlongenoughnotlongenoughnotlongenoughnotlongenoughnotlongenoughnotlongenoughnotlongenoughnotlongenoughnotlo",
	}, deps)
}

func TestDetect_MaxDependenciesCountReached(t *testing.T) {
	deps, err := deps.Detect(
		"testdata/python_with_many_imports.py",
		heartbeat.LanguagePython,
	)
	require.NoError(t, err)

	assert.Len(t, deps, 1000)
}

func TestDetect_EmptyDependenciesRemoved(t *testing.T) {
	deps, err := deps.Detect(
		"testdata/bower_empty_dependency.json",
		heartbeat.LanguageJSON,
	)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"bootstrap",
	}, deps)
}
