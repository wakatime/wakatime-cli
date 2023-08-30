package language_test

import (
	"fmt"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/language"
	"github.com/wakatime/wakatime-cli/pkg/lexer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithDetection(t *testing.T) {
	opt := language.WithDetection()

	h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Len(t, hh, 1)
		assert.Equal(t, heartbeat.LanguageGo.String(), *hh[0].Language)
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Entity:     "testdata/codefiles/golang.go",
				EntityType: heartbeat.FileType,
				Language:   hh[0].Language,
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := h([]heartbeat.Heartbeat{
		{
			Entity:     "testdata/codefiles/golang.go",
			EntityType: heartbeat.FileType,
		},
	})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestWithDetection_Override(t *testing.T) {
	opt := language.WithDetection()

	h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Len(t, hh, 1)
		assert.Equal(t, heartbeat.LanguagePython.String(), *hh[0].Language)
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Entity:     "testdata/codefiles/golang.go",
				EntityType: heartbeat.FileType,
				Language:   hh[0].Language,
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := h([]heartbeat.Heartbeat{
		{
			Entity:     "testdata/codefiles/golang.go",
			EntityType: heartbeat.FileType,
			Language:   heartbeat.PointerTo("Python"),
		},
	})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestWithDetection_NonExistingEntity_Override(t *testing.T) {
	opt := language.WithDetection()

	h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Len(t, hh, 1)
		assert.Equal(t, heartbeat.LanguagePython.String(), hh[0].LanguageAlternate)
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Entity:            "nonexisting",
				EntityType:        heartbeat.FileType,
				Language:          heartbeat.PointerTo(hh[0].LanguageAlternate),
				LanguageAlternate: hh[0].LanguageAlternate,
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := h([]heartbeat.Heartbeat{
		{
			Entity:            "nonexisting",
			EntityType:        heartbeat.FileType,
			LanguageAlternate: "Python",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestWithDetection_Alternate(t *testing.T) {
	opt := language.WithDetection()

	h := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Len(t, hh, 1)
		assert.Equal(t, []heartbeat.Heartbeat{
			{
				Entity:            "testdata/codefiles/unknown.xyz",
				EntityType:        heartbeat.FileType,
				Language:          heartbeat.PointerTo("Golang"),
				LanguageAlternate: "Golang",
			},
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
			},
		}, nil
	})

	result, err := h([]heartbeat.Heartbeat{
		{
			Entity:            "testdata/codefiles/unknown.xyz",
			EntityType:        heartbeat.FileType,
			LanguageAlternate: "Golang",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
		},
	}, result)
}

func TestDetect_HeaderFile_Corresponding_C_File(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/h_with_c_file/empty.h")
	require.NoError(t, err)
	assert.Equal(t, heartbeat.LanguageC, lang)
}

func TestDetect_HeaderFile_With_C_Files(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/h_with_any_c_file/empty.h")
	require.NoError(t, err)
	assert.Equal(t, heartbeat.LanguageC, lang)
}

func TestDetect_HeaderFile_With_C_And_CPP_Files(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/h_with_any_c_and_cpp_files/cpp.h")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageCPP, lang)
}

func TestDetect_HeaderFile_With_C_And_CXX_Files(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/h_with_any_c_and_cxx_files/cpp.h")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageCPP, lang)
}

func TestDetect_ObjectiveC_Over_Matlab_MatchingHeader(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/with_mat_file/objective-c.m")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageObjectiveC, lang)
}

func TestDetect_ObjectiveC_M_FileInFolder(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/with_mat_file/objective-c.h")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageObjectiveC, lang)
}

func TestDetect_ObjectiveCPP_MatchingHeader(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/with_mat_file/objective-cpp.mm")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageObjectiveCPP, lang)
}

func TestDetect_ObjectiveCPP_MM_FileInFolder(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/with_mat_file/objective-cpp.h")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageObjectiveCPP, lang)
}

func TestDetect_ObjectiveC(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/objective-c.m")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageObjectiveC, lang)
}

func TestDetect_Matlab_Over_ObjectiveC_Mat_FileInFolder(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/with_mat_file/empty.m")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageMatlab, lang)
}

func TestDetect_ObjectiveC_Over_Matlab_NonMatchingHeader(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/matlab_with_headers/empty.m")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageObjectiveC, lang)
}

func TestDetect_NonHeaderFile_C_FilesInFolder(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/py_with_c_files/see.py")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguagePython, lang)
}

func TestDetect_Perl_Over_Prolog(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/perl.pl")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguagePerl, lang)
}

func TestDetect_FSharp_Over_Forth(t *testing.T) {
	lang, err := language.Detect("testdata/codefiles/fsharp.fs")
	require.NoError(t, err)

	assert.Equal(t, heartbeat.LanguageFSharp, lang)
}

func TestDetect_ChromaTopLanguagesRetrofit(t *testing.T) {
	err := lexer.RegisterAll()
	require.NoError(t, err)

	tests := map[string]struct {
		Filepaths []string
		Expected  heartbeat.Language
	}{
		"apache config": {
			Filepaths: []string{
				"path/to/.htaccess",
				"path/to/apache.conf",
				"path/to/apache2.conf",
			},
			Expected: heartbeat.LanguageApacheConfig,
		},
		"applescript": {
			Filepaths: []string{
				"path/to/file.applescript",
			},
			Expected: heartbeat.LanguageAppleScript,
		},
		"assembly not gas": {
			Filepaths: []string{
				"path/to/file.s",
			},
			Expected: heartbeat.LanguageAssembly,
		},
		"awk": {
			Filepaths: []string{
				"path/to/file.awk",
			},
			Expected: heartbeat.LanguageAwk,
		},
		"bash": {
			Filepaths: []string{
				"path/to/file.sh",
				"path/to/file.ksh",
				"path/to/file.bash",
				"path/to/file.ebuild",
				"path/to/file.eclass",
				"path/to/file.exheres-0",
				"path/to/file.exlib",
				"path/to/file.zsh",
				"path/to/.bashrc",
				"path/to/bashrc",
				"path/to/.bash_history*",
				"path/to/.bash_xxx*",
				"path/to/bash_history",
				"path/to/bash_xxx",
				"path/to/zshrc",
				"path/to/.zshrc",
				"path/to/PKGBUILD",
				"testdata/bash",
			},
			Expected: heartbeat.LanguageBash,
		},
		"c": {
			Filepaths: []string{"path/to/file.c"},
			Expected:  heartbeat.LanguageC,
		},
		"c++": {
			Filepaths: []string{
				"path/to/file.cpp",
				"path/to/file.cxx",
			},
			Expected: heartbeat.LanguageCPP,
		},
		"c#": {
			Filepaths: []string{"path/to/file.cs"},
			Expected:  heartbeat.LanguageCSharp,
		},
		"clojure": {
			Filepaths: []string{"path/to/file.clj"},
			Expected:  heartbeat.LanguageClojure,
		},
		"cmake": {
			Filepaths: []string{
				"path/to/file.cmake",
				"path/to/CMmakeLists.txt",
			},
			Expected: heartbeat.LanguageCMake,
		},
		"coffescript": {
			Filepaths: []string{"path/to/file.coffee"},
			Expected:  heartbeat.LanguageCoffeeScript,
		},
		"coldfusion": {
			Filepaths: []string{
				"path/to/file.cfm",
				"path/to/file.cfml",
			},
			Expected: heartbeat.LanguageColdfusionHTML,
		},
		"common lisp": {
			Filepaths: []string{
				"path/to/file.cl",
				"path/to/file.lisp",
			},
			Expected: heartbeat.LanguageCommonLisp,
		},
		"crontab": {
			Filepaths: []string{"path/to/crontab"},
			Expected:  heartbeat.LanguageCrontab,
		},
		"crystal": {
			Filepaths: []string{"path/to/file.cr"},
			Expected:  heartbeat.LanguageCrystal,
		},
		"css": {
			Filepaths: []string{"path/to/file.css"},
			Expected:  heartbeat.LanguageCSS,
		},
		"dart": {
			Filepaths: []string{"path/to/file.dart"},
			Expected:  heartbeat.LanguageDart,
		},
		"delphi": {
			Filepaths: []string{
				"path/to/file.pas",
				"path/to/file.dpr",
			},
			Expected: heartbeat.LanguageDelphi,
		},
		"docker": {
			Filepaths: []string{
				"path/to/Dockerfile",
				"path/to/file.docker",
			},
			Expected: heartbeat.LanguageDocker,
		},
		"elixir": {
			Filepaths: []string{
				"path/to/file.ex",
				"path/to/file.eex",
				"path/to/file.exs",
			},
			Expected: heartbeat.LanguageElixir,
		},
		"elm": {
			Filepaths: []string{"path/to/file.elm"},
			Expected:  heartbeat.LanguageElm,
		},
		"emacs lisp": {
			Filepaths: []string{"path/to/file.el"},
			Expected:  heartbeat.LanguageEmacsLisp,
		},
		"erlang": {
			Filepaths: []string{
				"path/to/file.erl",
				"path/to/file.hrl",
				"path/to/file.es",
				"path/to/file.escript",
			},
			Expected: heartbeat.LanguageErlang,
		},
		"f sharp": {
			Filepaths: []string{"path/to/fsharp.fsi"},
			Expected:  heartbeat.LanguageFSharp,
		},
		"fortran": {
			Filepaths: []string{
				"path/to/file.f03",
				"path/to/file.f90",
				"path/to/file.F03",
				"path/to/file.F90",
			},
			Expected: heartbeat.LanguageFortran,
		},
		"golang": {
			Filepaths: []string{"path/to/go.go"},
			Expected:  heartbeat.LanguageGo,
		},
		"gosu": {
			Filepaths: []string{
				"path/to/file.gs",
				"path/to/file.gsp",
				"path/to/file.gsx",
				"path/to/file.vark",
			},
			Expected: heartbeat.LanguageGosu,
		},
		"groovy": {
			Filepaths: []string{
				"path/to/file.groovy",
				"path/to/file.gradle",
			},
			Expected: heartbeat.LanguageGroovy,
		},
		"haskell": {
			Filepaths: []string{"path/to/haskell.hs"},
			Expected:  heartbeat.LanguageHaskell,
		},
		"haxe": {
			Filepaths: []string{
				"path/to/haxe.hx",
				"path/to/haxe.hxsl",
			},
			Expected: heartbeat.LanguageHaxe,
		},
		"html": {
			Filepaths: []string{
				"path/to/html.html",
				"path/to/html.htm",
				"path/to/html.xhtml",
			},
			Expected: heartbeat.LanguageHTML,
		},
		"ini": {
			Filepaths: []string{
				"path/to/file.ini",
				"path/to/file.inf",
				"path/to/file.cfg",
			},
			Expected: heartbeat.LanguageINI,
		},
		"java": {
			Filepaths: []string{"path/to/java.java"},
			Expected:  heartbeat.LanguageJava,
		},
		"javascript": {
			Filepaths: []string{
				"path/to/es6.js",
				"path/to/file.jsm",
			},
			Expected: heartbeat.LanguageJavaScript,
		},
		"javascript module extension": {
			Filepaths: []string{"path/to/javascript_module.mjs"},
			Expected:  heartbeat.LanguageJavaScript,
		},
		"json": {
			Filepaths: []string{"path/to/bower.json"},
			Expected:  heartbeat.LanguageJSON,
		},
		"jsx": {
			Filepaths: []string{"path/to/file.jsx"},
			Expected:  heartbeat.LanguageJSX,
		},
		"kotlin": {
			Filepaths: []string{"path/to/kotlin.kt"},
			Expected:  heartbeat.LanguageKotlin,
		},
		"lasso": {
			Filepaths: []string{
				"path/to/file.lasso",
				"path/to/file.lasso9",
				"path/to/file.lasso8",
			},
			Expected: heartbeat.LanguageLasso,
		},
		"latex": {
			Filepaths: []string{
				"path/to/file.tex",
				"path/to/file.aux",
				"path/to/file.toc",
			},
			Expected: heartbeat.LanguageTeX,
		},
		"less": {
			Filepaths: []string{"path/to/file.less"},
			Expected:  heartbeat.LanguageLess,
		},
		"liquid": {
			Filepaths: []string{"path/to/file.liquid"},
			Expected:  heartbeat.LanguageLiquid,
		},
		"lua": {
			Filepaths: []string{
				"path/to/file.lua",
				"path/to/file.wlua",
			},
			Expected: heartbeat.LanguageLua,
		},
		"mako": {
			Filepaths: []string{"path/to/file.mao"},
			Expected:  heartbeat.LanguageMako,
		},
		"markdown": {
			Filepaths: []string{
				"path/to/file.md",
				"path/to/file.markdown",
			},
			Expected: heartbeat.LanguageMarkdown,
		},
		"marko": {
			Filepaths: []string{"path/to/file.marko"},
			Expected:  heartbeat.LanguageMarko,
		},
		"modelica": {
			Filepaths: []string{"path/to/file.mo"},
			Expected:  heartbeat.LanguageModelica,
		},
		"modula 2": {
			Filepaths: []string{
				"testdata/codefiles/chroma_unsupported_top/modula2.mod",
				"testdata/codefiles/chroma_unsupported_top/modula2.def",
			},
			Expected: heartbeat.LanguageModula2,
		},
		"mustache": {
			Filepaths: []string{"path/to/file.mustache"},
			Expected:  heartbeat.LanguageMustache,
		},
		"new lisp": {
			Filepaths: []string{
				"path/to/file.lsp",
				"path/to/file.nl",
				"path/to/file.kif",
			},
			Expected: heartbeat.LanguageNewLisp,
		},
		"nix": {
			Filepaths: []string{"path/to/file.nix"},
			Expected:  heartbeat.LanguageNix,
		},
		"objective j": {
			Filepaths: []string{"path/to/file.j"},
			Expected:  heartbeat.LanguageObjectiveJ,
		},
		"ocaml": {
			Filepaths: []string{
				"path/to/file.ml",
				"path/to/file.mli",
				"path/to/file.mll",
				"path/to/file.mly",
			},
			Expected: heartbeat.LanguageOCaml,
		},
		"pawn": {
			Filepaths: []string{"path/to/file.pwn"},
			Expected:  heartbeat.LanguagePawn,
		},
		"perl not prolog": {
			Filepaths: []string{
				"testdata/codefiles/chroma_unsupported_top/perl.pl",
				"testdata/perl",
			},
			Expected: heartbeat.LanguagePerl,
		},
		"php": {
			Filepaths: []string{
				"path/to/file.php",
				"path/to/file.php3",
				"path/to/file.php4",
				"path/to/file.php5",
			},
			Expected: heartbeat.LanguagePHP,
		},
		"postscript": {
			Filepaths: []string{
				"path/to/file.ps",
				"path/to/file.eps",
			},

			Expected: heartbeat.LanguagePostScript,
		},
		"povray": {
			Filepaths: []string{"path/to/file.pov"},
			Expected:  heartbeat.LanguagePOVRay,
		},
		"powershell": {
			Filepaths: []string{
				"path/to/file.ps1",
				"path/to/file.psm1",
			},

			Expected: heartbeat.LanguagePowerShell,
		},
		"prolog": {
			Filepaths: []string{
				"path/to/file.ecl",
				"path/to/file.prolog",
				"path/to/file.pro",
			},
			Expected: heartbeat.LanguageProlog,
		},
		"protobuf": {
			Filepaths: []string{"path/to/file.proto"},
			Expected:  heartbeat.LanguageProtocolBuffer,
		},
		"pug": {
			Filepaths: []string{
				"path/to/file.pug",
				"path/to/file.jade",
			},
			Expected: heartbeat.LanguagePug,
		},
		"puppet": {
			Filepaths: []string{"path/to/file.pp"},
			Expected:  heartbeat.LanguagePuppet,
		},
		"python": {
			Filepaths: []string{
				"path/to/file.py",
				"path/to/file.pyw",
				"path/to/file.jy",
				"path/to/file.sage",
				"path/to/SConstruct",
				"path/to/SConscript",
				"path/to/file.bzl",
				"path/to/BUCK",
				"path/to/BUILD",
				"path/to/BUILD.bazel",
				"path/to/WORKSPACE",
				"path/to/file.tac",
				"testdata/python3",
			},
			Expected: heartbeat.LanguagePython,
		},
		"qml": {
			Filepaths: []string{
				"path/to/file.qml",
				"path/to/file.qbs",
			},
			Expected: heartbeat.LanguageQML,
		},
		"reason": {
			Filepaths: []string{
				"path/to/file.re",
				"path/to/file.rei",
			},
			Expected: heartbeat.LanguageReasonML,
		},
		"restructuredtext": {
			Filepaths: []string{
				"path/to/file.rst",
				"path/to/file.rest",
			},
			Expected: heartbeat.LanguageReStructuredText,
		},
		"rpm spec": {
			Filepaths: []string{"path/to/file.spec"},
			Expected:  heartbeat.LanguageRPMSpec,
		},
		"ruby": {
			Filepaths: []string{
				"path/to/file.rb",
				"path/to/file.rbw",
				"path/to/Rakefile",
				"path/to/file.rake",
				"path/to/file.gemspec",
				"path/to/file.rbx",
				"path/to/file.duby",
				"path/to/Gemfile",
			},
			Expected: heartbeat.LanguageRuby,
		},
		"rust": {
			Filepaths: []string{
				"path/to/file.rs",
				"path/to/file.rs.in",
			},
			Expected: heartbeat.LanguageRust,
		},
		"s": {
			Filepaths: []string{
				"path/to/file.r",
				"path/to/file.R",
				"path/to/.Rhistory",
				"path/to/.Rprofile",
				"path/to/.Renviron",
			},
			Expected: heartbeat.LanguageS,
		},
		"sass": {
			Filepaths: []string{"path/to/file.sass"},
			Expected:  heartbeat.LanguageSass,
		},
		"scala": {
			Filepaths: []string{"path/to/file.scala"},
			Expected:  heartbeat.LanguageScala,
		},
		"scheme": {
			Filepaths: []string{
				"path/to/file.scm",
				"path/to/file.ss",
			},
			Expected: heartbeat.LanguageScheme,
		},
		"scss": {
			Filepaths: []string{"path/to/file.scss"},
			Expected:  heartbeat.LanguageSCSS,
		},
		"sketch": {
			Filepaths: []string{"path/to/file.sketch"},
			Expected:  heartbeat.LanguageSketchDrawing,
		},
		"slim": {
			Filepaths: []string{"path/to/file.slim"},
			Expected:  heartbeat.LanguageSlim,
		},
		"smali": {
			Filepaths: []string{"path/to/file.smali"},
			Expected:  heartbeat.LanguageSmali,
		},
		"smalltalk": {
			Filepaths: []string{"path/to/file.st"},
			Expected:  heartbeat.LanguageSmalltalk,
		},
		"sourcepawn": {
			Filepaths: []string{"path/to/file.sp"},
			Expected:  heartbeat.LanguageSourcePawn,
		},
		"sql": {
			Filepaths: []string{"path/to/file.sql"},
			Expected:  heartbeat.LanguageSQL,
		},
		"sublime settings": {
			Filepaths: []string{"path/to/file.sublime-settings"},
			Expected:  heartbeat.LanguageSublimeTextConfig,
		},
		"svelte": {
			Filepaths: []string{"path/to/file.svelte"},
			Expected:  heartbeat.LanguageSvelte,
		},
		"swift": {
			Filepaths: []string{"path/to/file.swift"},
			Expected:  heartbeat.LanguageSwift,
		},
		"swig": {
			Filepaths: []string{
				"path/to/file.swg",
				"path/to/file.i",
			},
			Expected: heartbeat.LanguageSWIG,
		},
		"system verilog": {
			Filepaths: []string{
				"path/to/file.sv",
				"path/to/file.svh",
			},
			Expected: heartbeat.LanguageSystemVerilog,
		},
		"textfile": {
			Filepaths: []string{"path/to/file.txt"},
			Expected:  heartbeat.LanguageText,
		},
		"thrift": {
			Filepaths: []string{"path/to/file.thrift"},
			Expected:  heartbeat.LanguageThrift,
		},
		"toml": {
			Filepaths: []string{
				"path/to/file.toml",
				"path/to/Pipfile",
				"path/to/poetry.lock",
			},
			Expected: heartbeat.LanguageTOML,
		},
		"twig": {
			Filepaths: []string{"path/to/file.twig"},
			Expected:  heartbeat.LanguageTwig,
		},
		"typescript": {
			Filepaths: []string{
				"testdata/codefiles/typescript.ts",
				"path/to/file.tsx",
			},
			Expected: heartbeat.LanguageTypeScript,
		},
		"vb.net": {
			Filepaths: []string{"path/to/file.vb"},
			Expected:  heartbeat.LanguageVBNet,
		},
		"vcl": {
			Filepaths: []string{"path/to/file.vcl"},
			Expected:  heartbeat.LanguageVCL,
		},
		"velocity": {
			Filepaths: []string{
				"path/to/file.vm",
				"path/to/file.fhtml",
			},
			Expected: heartbeat.LanguageVelocity,
		},
		"viml": {
			Filepaths: []string{
				"path/to/file.vim",
				"path/to/.vimrc",
				"path/to/.exrc",
				"path/to/.gvimrc",
				"path/to/_vimrc",
				"path/to/_exrc",
				"path/to/_gvimrc",
				"path/to/vimrc",
				"path/to/gvimrc",
			},
			Expected: heartbeat.LanguageVimL,
		},
		"vue.js": {
			Filepaths: []string{"path/to/file.vue"},
			Expected:  heartbeat.LanguageVueJS,
		},
		"xaml": {
			Filepaths: []string{"path/to/file.xaml"},
			Expected:  heartbeat.LanguageXAML,
		},
		"xml": {
			Filepaths: []string{
				"path/to/file.xml",
				"path/to/file.rss",
				"path/to/file.xsd",
				"path/to/file.wsdl",
				"path/to/file.wsf",
			},
			Expected: heartbeat.LanguageXML,
		},
		"xslt": {
			Filepaths: []string{"path/to/file.xpl"},
			Expected:  heartbeat.LanguageXSLT,
		},
		"yaml": {
			Filepaths: []string{
				"path/to/file.yaml",
				"path/to/file.yml",
			},
			Expected: heartbeat.LanguageYAML,
		},
		"zig": {
			Filepaths: []string{"path/to/file.zig"},
			Expected:  heartbeat.LanguageZig,
		},
		"case-insensitive": {
			Filepaths: []string{"path/to/FILE.GO"},
			Expected:  heartbeat.LanguageGo,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			for _, filepath := range test.Filepaths {
				lang, err := language.Detect(filepath)
				require.NoError(t, err)

				assert.Equal(t, test.Expected, lang, fmt.Sprintf("Got: %q, want: %q", lang, test.Expected))
			}
		})
	}
}
