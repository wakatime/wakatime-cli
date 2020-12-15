package heartbeat

import (
	"fmt"
	"strings"
)

// Language represents a programming language.
type Language int

const (
	// LanguageUnknown represents the Unknown programming language.
	LanguageUnknown Language = iota
	// LanguageAda represents the Ada programming language.
	LanguageAda
	// LanguageActionScript represents the ActionScript programming language.
	LanguageActionScript
	// LanguageAgda represents the Agda programming language.
	LanguageAgda
	// LanguageAnsible represents the Ansible programming language.
	LanguageAnsible
	// LanguageAppleScript represents the AppleScript programming language.
	LanguageAppleScript
	// LanguageApacheConf represents the ApacheConf programming language.
	LanguageApacheConf
	// LanguageASP represents the ASP programming language.
	LanguageASP
	// LanguageAssembly represents the Assembly programming language.
	LanguageAssembly
	// LanguageAutoconf represents the Autoconf programming language.
	LanguageAutoconf
	// LanguageAwk represents the Awk programming language.
	LanguageAwk
	// LanguageBash represents the Bash programming language.
	LanguageBash
	// LanguageBasic represents the Basic programming language.
	LanguageBasic
	// LanguageBatchScript represents the BatchScript programming language.
	LanguageBatchScript
	// LanguageBibTeX represents the BibTeX programming language.
	LanguageBibTeX
	// LanguageBrightScript represents the BrightScript programming language.
	LanguageBrightScript
	// LanguageC represents the C programming language.
	LanguageC
	// LanguageClojure represents the Clojure programming language.
	LanguageClojure
	// LanguageCMake represents the CMake programming language.
	LanguageCMake
	// LanguageCocoa represents the Cocoa programming language.
	LanguageCocoa
	// LanguageCoffeeScript represents the CoffeeScript programming language.
	LanguageCoffeeScript
	// LanguageColdfusionHTML represents the ColdfusionHTML programming language.
	LanguageColdfusionHTML
	// LanguageCommonLisp represents the CommonLisp programming language.
	LanguageCommonLisp
	// LanguageCoq represents the Coq programming language.
	LanguageCoq
	// LanguageCPerl represents the CPerl programming language.
	LanguageCPerl
	// LanguageCPP represents the CPP programming language.
	LanguageCPP
	// LanguageCSharp represents the CSharp programming language.
	LanguageCSharp
	// LanguageCSHTML represents the CSHTML programming language.
	LanguageCSHTML
	// LanguageCVS represents the CVS programming language.
	LanguageCVS
	// LanguageCrontab represents the Crontab programming language.
	LanguageCrontab
	// LanguageCrystal represents the Crystal programming language.
	LanguageCrystal
	// LanguageCSS represents the CSS programming language.
	LanguageCSS
	// LanguageDart represents the Dart programming language.
	LanguageDart
	// LanguageDCL represents the DCL programming language.
	LanguageDCL
	// LanguageDelphi represents the Delphi programming language.
	LanguageDelphi
	// LanguageDhall represents the Dhall programming language.
	LanguageDhall
	// LanguageDiff represents the Diff programming language.
	LanguageDiff
	// LanguageDocker represents the Docker programming language.
	LanguageDocker
	// LanguageDocTeX represents the DocTeX programming language.
	LanguageDocTeX
	// LanguageElixir represents the Elixir programming language.
	LanguageElixir
	// LanguageElm represents the Elm programming language.
	LanguageElm
	// LanguageEmacsLisp represents the EmacsLisp programming language.
	LanguageEmacsLisp
	// LanguageErlang represents the Erlang programming language.
	LanguageErlang
	// LanguageEshell represents the Eshell programming language.
	LanguageEshell
	// LanguageFish represents the Fish programming language.
	LanguageFish
	// LanguageFSharp represents the FSharp programming language.
	LanguageFSharp
	// LanguageFortran represents the Fortran programming language.
	LanguageFortran
	// LanguageGo represents the Go programming language.
	LanguageGo
	// LanguageGosu represents the Gosu programming language.
	LanguageGosu
	// LanguageGroovy represents the Groovy programming language.
	LanguageGroovy
	// LanguageHAML represents the HAML programming language.
	LanguageHAML
	// LanguageHaskell represents the Haskell programming language.
	LanguageHaskell
	// LanguageHaxe represents the Haxe programming language.
	LanguageHaxe
	// LanguageHCL represents the HCL programming language.
	LanguageHCL
	// LanguageHTML represents the HTML programming language.
	LanguageHTML
	// LanguageINI represents the INI programming language.
	LanguageINI
	// LanguageJade represents the Jade programming language.
	LanguageJade
	// LanguageJava represents the Java programming language.
	LanguageJava
	// LanguageJavaScript represents the JavaScript programming language.
	LanguageJavaScript
	// LanguageJSON represents the JSON programming language.
	LanguageJSON
	// LanguageJSX represents the JSX programming language.
	LanguageJSX
	// LanguageKotlin represents the Kotlin programming language.
	LanguageKotlin
	// LanguageLasso represents the Lasso programming language.
	LanguageLasso
	// LanguageLaTeX represents the LaTeX programming language.
	LanguageLaTeX
	// LanguageLess represents the Less programming language.
	LanguageLess
	// LanguageLinkerScript represents the LinkerScript programming language.
	LanguageLinkerScript
	// LanguageLiquid represents the Liquid programming language.
	LanguageLiquid
	// LanguageLua represents the Lua programming language.
	LanguageLua
	// LanguageMakefile represents the Makefile programming language.
	LanguageMakefile
	// LanguageMako represents the Mako programming language.
	LanguageMako
	// LanguageMan represents the Man programming language.
	LanguageMan
	// LanguageMarkdown represents the Markdown programming language.
	LanguageMarkdown
	// LanguageMarko represents the Marko programming language.
	LanguageMarko
	// LanguageMatlab represents the Matlab programming language.
	LanguageMatlab
	// LanguageMetafont represents the Metafont programming language.
	LanguageMetafont
	// LanguageMetapost represents the Metapost programming language.
	LanguageMetapost
	// LanguageModelica represents the Modelica programming language.
	LanguageModelica
	// LanguageModula2 represents the Modula2 programming language.
	LanguageModula2
	// LanguageMustache represents the Mustache programming language.
	LanguageMustache
	// LanguageNewLisp represents the NewLisp programming language.
	LanguageNewLisp
	// LanguageNix represents the Nix programming language.
	LanguageNix
	// LanguageObjectiveC represents the ObjectiveC programming language.
	LanguageObjectiveC
	// LanguageObjectiveCPP represents the ObjectiveC++ programming language.
	LanguageObjectiveCPP
	// LanguageObjectiveJ represents the ObjectiveJ programming language.
	LanguageObjectiveJ
	// LanguageOCaml represents the OCaml programming language.
	LanguageOCaml
	// LanguageOrg represents the Org programming language.
	LanguageOrg
	// LanguagePascal represents the Pascal programming language.
	LanguagePascal
	// LanguagePawn represents the Pawn programming language.
	LanguagePawn
	// LanguagePerl represents the Perl programming language.
	LanguagePerl
	// LanguagePHP represents the PHP programming language.
	LanguagePHP
	// LanguagePostScript represents the PostScript programming language.
	LanguagePostScript
	// LanguagePOVRay represents the POVRay programming language.
	LanguagePOVRay
	// LanguagePowerShell represents the PowerShell programming language.
	LanguagePowerShell
	// LanguageProlog represents the Prolog programming language.
	LanguageProlog
	// LanguageProtocolBuffer represents the ProtocolBuffer programming language.
	LanguageProtocolBuffer
	// LanguagePug represents the Pug programming language.
	LanguagePug
	// LanguagePuppet represents the Puppet programming language.
	LanguagePuppet
	// LanguagePureScript represents the PureScript programming language.
	LanguagePureScript
	// LanguagePython represents the Python programming language.
	LanguagePython
	// LanguageQML represents the QML programming language.
	LanguageQML
	// LanguageR represents the R programming language.
	LanguageR
	// LanguageReasonML represents the ReasonML programming language.
	LanguageReasonML
	// LanguageReStructuredText represents the ReStructuredText programming language.
	LanguageReStructuredText
	// LanguageRPMSpec represents the RPMSpec programming language.
	LanguageRPMSpec
	// LanguageRuby represents the Ruby programming language.
	LanguageRuby
	// LanguageRust represents the Rust programming language.
	LanguageRust
	// LanguageSalt represents the Salt programming language.
	LanguageSalt
	// LanguageSass represents the Sass programming language.
	LanguageSass
	// LanguageScala represents the Scala programming language.
	LanguageScala
	// LanguageScheme represents the Scheme programming language.
	LanguageScheme
	// LanguageScribe represents the Scribe programming language.
	LanguageScribe
	// LanguageSCSS represents the SCSS programming language.
	LanguageSCSS
	// LanguageSGML represents the SGML programming language.
	LanguageSGML
	// LanguageShell represents the Shell programming language.
	LanguageShell
	// LanguageSimula represents the Simula programming language.
	LanguageSimula
	// LanguageSingularity represents the Singularity programming language.
	LanguageSingularity
	// LanguageSketchDrawing represents the SketchDrawing programming language.
	LanguageSketchDrawing
	// LanguageSKILL represents the SKILL programming language.
	LanguageSKILL
	// LanguageSlim represents the Slim programming language.
	LanguageSlim
	// LanguageSmali represents the Smali programming language.
	LanguageSmali
	// LanguageSmalltalk represents the Smalltalk programming language.
	LanguageSmalltalk
	// LanguageSMIME represents the SMIME programming language.
	LanguageSMIME
	// LanguageSourcePawn represents the SourcePawn programming language.
	LanguageSourcePawn
	// LanguageSQL represents the SQL programming language.
	LanguageSQL
	// LanguageSublimeTextConfig represents the SublimeTextConfig programming language.
	LanguageSublimeTextConfig
	// LanguageSvelte represents the Svelte programming language.
	LanguageSvelte
	// LanguageSwift represents the Swift programming language.
	LanguageSwift
	// LanguageSWIG represents the SWIG programming language.
	LanguageSWIG
	// LanguageSystemVerilog represents the SystemVerilog programming language.
	LanguageSystemVerilog
	// LanguageTeX represents the TeX programming language.
	LanguageTeX
	// LanguageText represents the Text programming language.
	LanguageText
	// LanguageThrift represents the Thrift programming language.
	LanguageThrift
	// LanguageTOML represents the TOML programming language.
	LanguageTOML
	// LanguageTuring represents the Turing programming language.
	LanguageTuring
	// LanguageTwig represents the Twig programming language.
	LanguageTwig
	// LanguageTypeScript represents the TypeScript programming language.
	LanguageTypeScript
	// LanguageTypoScript represents the TypoScript programming language.
	LanguageTypoScript
	// LanguageVB represents the VB programming language.
	LanguageVB
	// LanguageVBNet represents the VB.net programming language.
	LanguageVBNet
	// LanguageVCL represents the VCL programming language.
	LanguageVCL
	// LanguageVelocity represents the Velocity programming language.
	LanguageVelocity
	// LanguageVerilog represents the Verilog programming language.
	LanguageVerilog
	// LanguageVHDL represents the VHDL programming language.
	LanguageVHDL
	// LanguageVimL represents the VimL programming language.
	LanguageVimL
	// LanguageVueJS represents the VueJS programming language.
	LanguageVueJS
	// LanguageXAML represents the XAML programming language.
	LanguageXAML
	// LanguageXML represents the XML programming language.
	LanguageXML
	// LanguageXSLT represents the XSLT programming language.
	LanguageXSLT
	// LanguageYAML represents the YAML programming language.
	LanguageYAML
	// LanguageZig represents the Zig programming language.
	LanguageZig
)

const (
	languageUnkownStr            = "Unknown"
	languageAdaStr               = "Ada"
	languageActionScriptStr      = "ActionScript"
	languageAgdaStr              = "Agda"
	languageAnsibleStr           = "Ansible"
	languageASPStr               = "ASP"
	languageAppleScriptStr       = "AppleScript"
	languageApacheConfStr        = "ApacheConf"
	languageAssemblyStr          = "Assembly"
	languageAutoconfStr          = "Autoconf"
	languageAwkStr               = "AWK"
	languageBashStr              = "Bash"
	languageBasicStr             = "Basic"
	languageBatchScriptStr       = "Batch Script"
	languageBibTeXStr            = "BibTeX"
	languageBrightScriptStr      = "BrightScript"
	languageCStr                 = "C"
	languageClojureStr           = "Clojure"
	languageCMakeStr             = "CMake"
	languageCocoaStr             = "Cocoa"
	languageCoqStr               = "Coq"
	languageCoffeeScriptStr      = "CoffeeScript"
	languageColdfusionHTMLStr    = "Coldfusion"
	languageCommonLispStr        = "Common Lisp"
	languageCPerlStr             = "cperl"
	languageCPPStr               = "C++"
	languageCrontabStr           = "Crontab"
	languageCrystalStr           = "Crystal"
	languageCSharpStr            = "C#"
	languageCSHTMLStr            = "CSHTML"
	languageCSSStr               = "CSS"
	languageCVSStr               = "CVS"
	languageDartStr              = "Dart"
	languageDCLStr               = "DCL"
	languageDelphiStr            = "Delphi"
	languageDhallStr             = "Dhall"
	languageDiffStr              = "Diff"
	languageDockerStr            = "Docker"
	languageDocTeXStr            = "DocTeX"
	languageElixirStr            = "Elixir"
	languageElmStr               = "Elm"
	languageEmacsLispStr         = "Emacs Lisp"
	languageErlangStr            = "Erlang"
	languageEshellStr            = "Eshell"
	languageFSharpStr            = "F#"
	languageFishStr              = "Fish"
	languageFortranStr           = "Fortran"
	languageGoStr                = "Go"
	languageGosuStr              = "Gosu"
	languageGroovyStr            = "Groovy"
	languageHAMLStr              = "Haml"
	languageHaskellStr           = "Haskell"
	languageHaxeStr              = "Haxe"
	languageHCLStr               = "HCL"
	languageHTMLStr              = "HTML"
	languageINIStr               = "INI"
	languageJadeStr              = "Jade"
	languageJavaStr              = "Java"
	languageJavaScriptStr        = "JavaScript"
	languageJSONStr              = "JSON"
	languageJSXStr               = "JSX"
	languageKotlinStr            = "Kotlin"
	languageLassoStr             = "Lasso"
	languageLaTeXStr             = "LaTeX"
	languageLessStr              = "LESS"
	languageLinkerScriptStr      = "Linker Script"
	languageLiquidStr            = "liquid"
	languageLuaStr               = "Lua"
	languageMakefileStr          = "Makefile"
	languageMakoStr              = "Mako"
	languageManStr               = "Man"
	languageMarkdownStr          = "Markdown"
	languageMarkoStr             = "Marko"
	languageMatlabStr            = "Matlab"
	languageMetafontStr          = "Metafont"
	languageMetapostStr          = "Metapost"
	languageModelicaStr          = "Modelica"
	languageModula2Str           = "Modula-2"
	languageMustacheStr          = "Mustache"
	languageNewLispStr           = "NewLisp"
	languageNixStr               = "Nix"
	languageObjectiveCStr        = "Objective-C"
	languageObjectiveCPPStr      = "Objective-C++"
	languageObjectiveJStr        = "Objective-J"
	languageOCamlStr             = "OCaml"
	languageOrgStr               = "Org"
	languagePascalStr            = "Pascal"
	languagePawnStr              = "Pawn"
	languagePerlStr              = "Perl"
	languagePHPStr               = "PHP"
	languagePostScriptStr        = "PostScript"
	languagePOVRayStr            = "POVRay"
	languagePowerShellStr        = "PowerShell"
	languagePrologStr            = "Prolog"
	languageProtocolBufferStr    = "Protocol Buffer"
	languagePugStr               = "Pug"
	languagePuppetStr            = "Puppet"
	languagePureScriptStr        = "PureScript"
	languagePythonStr            = "Python"
	languageQMLStr               = "QML"
	languageRStr                 = "R"
	languageReasonMLStr          = "ReasonML"
	languageReStructuredTextStr  = "reStructuredText"
	languageRPMSpecStr           = "RPMSpec"
	languageRubyStr              = "Ruby"
	languageRustStr              = "Rust"
	languageSaltStr              = "Salt"
	languageSassStr              = "Sass"
	languageScalaStr             = "Scala"
	languageSchemeStr            = "Scheme"
	languageScribeStr            = "Scribe"
	languageSCSSStr              = "SCSS"
	languageSGMLStr              = "SGML"
	languageShellStr             = "Shell"
	languageSimulaStr            = "Simula"
	languageSingularityStr       = "Singularity"
	languageSketchDrawingStr     = "Sketch Drawing"
	languageSKILLStr             = "SKILL"
	languageSlimStr              = "Slim"
	languageSmaliStr             = "Smali"
	languageSmalltalkStr         = "Smalltalk"
	languageSMIMEStr             = "S/MIME"
	languageSourcePawnStr        = "SourcePawn"
	languageSQLStr               = "SQL"
	languageSublimeTextConfigStr = "Sublime Text Config"
	languageSvelteStr            = "Svelte"
	languageSwiftStr             = "Swift"
	languageSWIGStr              = "SWIG"
	languageSystemVerilogStr     = "systemverilog"
	languageTeXStr               = "TeX"
	languageTextStr              = "Text"
	languageThriftStr            = "Thrift"
	languageTOMLStr              = "TOML"
	languageTuringStr            = "Turing"
	languageTwigStr              = "Twig"
	languageTypeScriptStr        = "TypeScript"
	languageTypoScriptStr        = "TypoScript"
	languageVBStr                = "VB"
	languageVBNetStr             = "VB.net"
	languageVCLStr               = "VCL"
	languageVelocityStr          = "Velocity"
	languageVerilogStr           = "Verilog"
	languageVHDLStr              = "VHDL"
	languageVimLStr              = "VimL"
	languageVueJSStr             = "Vue.js"
	languageXAMLStr              = "XAML"
	languageXMLStr               = "XML"
	languageXSLTStr              = "XSLT"
	languageYAMLStr              = "YAML"
	languageZigStr               = "Zig"
)

const (
	languageAssemblyChromaStr       = "GAS"
	languageColdfusionHTMLChromaStr = "Coldfusion HTML"
	languageFSharpChromaStr         = "FSharp"
	languageGosuChromaStr           = "Gosu Template"
	languageEmacsLispChromaStr      = "EmacsLisp"
	languageJSXChromaStr            = "react"
	languageLessChromaStr           = "LessCss"
	languageMakefileChromaStr       = "Base Makefile"
	languageMarkdownChromaStr       = "markdown"
	languageTextChromaStr           = "plaintext"
	languageVueJSChromaStr          = "vue"
)

// ParseLanguage parses a language from a string. Will return false
// as second parameter, if language could not be parsed.
// nolint:gocyclo
func ParseLanguage(s string) (Language, bool) {
	switch normalizeString(s) {
	case normalizeString(languageAdaStr):
		return LanguageAda, true
	case normalizeString(languageActionScriptStr):
		return LanguageActionScript, true
	case normalizeString(languageAgdaStr):
		return LanguageAgda, true
	case normalizeString(languageAnsibleStr):
		return LanguageAnsible, true
	case normalizeString(languageAppleScriptStr):
		return LanguageAppleScript, true
	case normalizeString(languageApacheConfStr):
		return LanguageApacheConf, true
	case normalizeString(languageASPStr):
		return LanguageASP, true
	case normalizeString(languageAssemblyStr):
		return LanguageAssembly, true
	case normalizeString(languageAutoconfStr):
		return LanguageAutoconf, true
	case normalizeString(languageAwkStr):
		return LanguageAwk, true
	case normalizeString(languageBasicStr):
		return LanguageBasic, true
	case normalizeString(languageBashStr):
		return LanguageBash, true
	case normalizeString(languageBatchScriptStr):
		return LanguageBatchScript, true
	case normalizeString(languageBibTeXStr):
		return LanguageBibTeX, true
	case normalizeString(languageBrightScriptStr):
		return LanguageBrightScript, true
	case normalizeString(languageCStr):
		return LanguageC, true
	case normalizeString(languageClojureStr):
		return LanguageClojure, true
	case normalizeString(languageCMakeStr):
		return LanguageCMake, true
	case normalizeString(languageCocoaStr):
		return LanguageCocoa, true
	case normalizeString(languageCoffeeScriptStr):
		return LanguageCoffeeScript, true
	case normalizeString(languageColdfusionHTMLStr):
		return LanguageColdfusionHTML, true
	case normalizeString(languageCommonLispStr):
		return LanguageCommonLisp, true
	case normalizeString(languageCoqStr):
		return LanguageCoq, true
	case normalizeString(languageCPerlStr):
		return LanguageCPerl, true
	case normalizeString(languageCPPStr):
		return LanguageCPP, true
	case normalizeString(languageCrontabStr):
		return LanguageCrontab, true
	case normalizeString(languageCrystalStr):
		return LanguageCrystal, true
	case normalizeString(languageCSharpStr):
		return LanguageCSharp, true
	case normalizeString(languageCSHTMLStr):
		return LanguageCSHTML, true
	case normalizeString(languageCSSStr):
		return LanguageCSS, true
	case normalizeString(languageCVSStr):
		return LanguageCVS, true
	case normalizeString(languageDartStr):
		return LanguageDart, true
	case normalizeString(languageDCLStr):
		return LanguageDCL, true
	case normalizeString(languageDelphiStr):
		return LanguageDelphi, true
	case normalizeString(languageDhallStr):
		return LanguageDhall, true
	case normalizeString(languageDiffStr):
		return LanguageDiff, true
	case normalizeString(languageDockerStr):
		return LanguageDocker, true
	case normalizeString(languageDocTeXStr):
		return LanguageDocTeX, true
	case normalizeString(languageElixirStr):
		return LanguageElixir, true
	case normalizeString(languageElmStr):
		return LanguageElm, true
	case normalizeString(languageEmacsLispStr):
		return LanguageEmacsLisp, true
	case normalizeString(languageErlangStr):
		return LanguageErlang, true
	case normalizeString(languageEshellStr):
		return LanguageEshell, true
	case normalizeString(languageFishStr):
		return LanguageFish, true
	case normalizeString(languageFSharpStr):
		return LanguageFSharp, true
	case normalizeString(languageFortranStr):
		return LanguageFortran, true
	case normalizeString(languageGoStr):
		return LanguageGo, true
	case normalizeString(languageGosuStr):
		return LanguageGosu, true
	case normalizeString(languageGroovyStr):
		return LanguageGroovy, true
	case normalizeString(languageHAMLStr):
		return LanguageHAML, true
	case normalizeString(languageHaskellStr):
		return LanguageHaskell, true
	case normalizeString(languageHaxeStr):
		return LanguageHaxe, true
	case normalizeString(languageHCLStr):
		return LanguageHCL, true
	case normalizeString(languageHTMLStr):
		return LanguageHTML, true
	case normalizeString(languageINIStr):
		return LanguageINI, true
	case normalizeString(languageJadeStr):
		return LanguageJade, true
	case normalizeString(languageJavaStr):
		return LanguageJava, true
	case normalizeString(languageJavaScriptStr):
		return LanguageJavaScript, true
	case normalizeString(languageJSONStr):
		return LanguageJSON, true
	case normalizeString(languageJSXStr):
		return LanguageJSX, true
	case normalizeString(languageKotlinStr):
		return LanguageKotlin, true
	case normalizeString(languageLassoStr):
		return LanguageLasso, true
	case normalizeString(languageLaTeXStr):
		return LanguageLaTeX, true
	case normalizeString(languageLessStr):
		return LanguageLess, true
	case normalizeString(languageLinkerScriptStr):
		return LanguageLinkerScript, true
	case normalizeString(languageLiquidStr):
		return LanguageLiquid, true
	case normalizeString(languageLuaStr):
		return LanguageLua, true
	case normalizeString(languageMakefileStr):
		return LanguageMakefile, true
	case normalizeString(languageMakoStr):
		return LanguageMako, true
	case normalizeString(languageManStr):
		return LanguageMan, true
	case normalizeString(languageMarkdownStr):
		return LanguageMarkdown, true
	case normalizeString(languageMarkoStr):
		return LanguageMarko, true
	case normalizeString(languageMatlabStr):
		return LanguageMatlab, true
	case normalizeString(languageMetafontStr):
		return LanguageMetafont, true
	case normalizeString(languageMetapostStr):
		return LanguageMetapost, true
	case normalizeString(languageModelicaStr):
		return LanguageModelica, true
	case normalizeString(languageModula2Str):
		return LanguageModula2, true
	case normalizeString(languageMustacheStr):
		return LanguageMustache, true
	case normalizeString(languageNewLispStr):
		return LanguageNewLisp, true
	case normalizeString(languageNixStr):
		return LanguageNix, true
	case normalizeString(languageObjectiveCStr):
		return LanguageObjectiveC, true
	case normalizeString(languageObjectiveCPPStr):
		return LanguageObjectiveCPP, true
	case normalizeString(languageObjectiveJStr):
		return LanguageObjectiveJ, true
	case normalizeString(languageOCamlStr):
		return LanguageOCaml, true
	case normalizeString(languageOrgStr):
		return LanguageOrg, true
	case normalizeString(languagePascalStr):
		return LanguagePascal, true
	case normalizeString(languagePawnStr):
		return LanguagePawn, true
	case normalizeString(languagePerlStr):
		return LanguagePerl, true
	case normalizeString(languagePHPStr):
		return LanguagePHP, true
	case normalizeString(languagePostScriptStr):
		return LanguagePostScript, true
	case normalizeString(languagePOVRayStr):
		return LanguagePOVRay, true
	case normalizeString(languagePowerShellStr):
		return LanguagePowerShell, true
	case normalizeString(languagePrologStr):
		return LanguageProlog, true
	case normalizeString(languageProtocolBufferStr):
		return LanguageProtocolBuffer, true
	case normalizeString(languagePugStr):
		return LanguagePug, true
	case normalizeString(languagePuppetStr):
		return LanguagePuppet, true
	case normalizeString(languagePureScriptStr):
		return LanguagePureScript, true
	case normalizeString(languagePythonStr):
		return LanguagePython, true
	case normalizeString(languageQMLStr):
		return LanguageQML, true
	case normalizeString(languageRStr):
		return LanguageR, true
	case normalizeString(languageReasonMLStr):
		return LanguageReasonML, true
	case normalizeString(languageReStructuredTextStr):
		return LanguageReStructuredText, true
	case normalizeString(languageRPMSpecStr):
		return LanguageRPMSpec, true
	case normalizeString(languageRubyStr):
		return LanguageRuby, true
	case normalizeString(languageRustStr):
		return LanguageRust, true
	case normalizeString(languageSaltStr):
		return LanguageSalt, true
	case normalizeString(languageSassStr):
		return LanguageSass, true
	case normalizeString(languageScalaStr):
		return LanguageScala, true
	case normalizeString(languageSchemeStr):
		return LanguageScheme, true
	case normalizeString(languageScribeStr):
		return LanguageScribe, true
	case normalizeString(languageSCSSStr):
		return LanguageSCSS, true
	case normalizeString(languageSGMLStr):
		return LanguageSGML, true
	case normalizeString(languageShellStr):
		return LanguageShell, true
	case normalizeString(languageSimulaStr):
		return LanguageSimula, true
	case normalizeString(languageSingularityStr):
		return LanguageSingularity, true
	case normalizeString(languageSketchDrawingStr):
		return LanguageSketchDrawing, true
	case normalizeString(languageSKILLStr):
		return LanguageSKILL, true
	case normalizeString(languageSlimStr):
		return LanguageSlim, true
	case normalizeString(languageSmaliStr):
		return LanguageSmali, true
	case normalizeString(languageSmalltalkStr):
		return LanguageSmalltalk, true
	case normalizeString(languageSMIMEStr):
		return LanguageSMIME, true
	case normalizeString(languageSourcePawnStr):
		return LanguageSourcePawn, true
	case normalizeString(languageSQLStr):
		return LanguageSQL, true
	case normalizeString(languageSublimeTextConfigStr):
		return LanguageSublimeTextConfig, true
	case normalizeString(languageSvelteStr):
		return LanguageSvelte, true
	case normalizeString(languageSwiftStr):
		return LanguageSwift, true
	case normalizeString(languageSWIGStr):
		return LanguageSWIG, true
	case normalizeString(languageSystemVerilogStr):
		return LanguageSystemVerilog, true
	case normalizeString(languageTeXStr):
		return LanguageTeX, true
	case normalizeString(languageTextStr):
		return LanguageText, true
	case normalizeString(languageThriftStr):
		return LanguageThrift, true
	case normalizeString(languageTOMLStr):
		return LanguageTOML, true
	case normalizeString(languageTuringStr):
		return LanguageTuring, true
	case normalizeString(languageTwigStr):
		return LanguageTwig, true
	case normalizeString(languageTypeScriptStr):
		return LanguageTypeScript, true
	case normalizeString(languageTypoScriptStr):
		return LanguageTypoScript, true
	case normalizeString(languageVBStr):
		return LanguageVB, true
	case normalizeString(languageVBNetStr):
		return LanguageVBNet, true
	case normalizeString(languageVCLStr):
		return LanguageVCL, true
	case normalizeString(languageVelocityStr):
		return LanguageVelocity, true
	case normalizeString(languageVerilogStr):
		return LanguageVerilog, true
	case normalizeString(languageVHDLStr):
		return LanguageVHDL, true
	case normalizeString(languageVimLStr):
		return LanguageVimL, true
	case normalizeString(languageVueJSStr):
		return LanguageVueJS, true
	case normalizeString(languageXAMLStr):
		return LanguageXAML, true
	case normalizeString(languageXMLStr):
		return LanguageXML, true
	case normalizeString(languageXSLTStr):
		return LanguageXSLT, true
	case normalizeString(languageYAMLStr):
		return LanguageYAML, true
	case normalizeString(languageZigStr):
		return LanguageZig, true
	default:
		return LanguageUnknown, false
	}
}

// ParseLanguageFromChroma parses a language from a chroma lexer name.
// Will return false as second parameter, if language could not be parsed.
// nolint:gocyclo
func ParseLanguageFromChroma(lexerName string) (Language, bool) {
	switch normalizeString(lexerName) {
	case normalizeString(languageAssemblyChromaStr):
		return LanguageAssembly, true
	case normalizeString(languageColdfusionHTMLChromaStr):
		return LanguageColdfusionHTML, true
	case normalizeString(languageEmacsLispChromaStr):
		return LanguageEmacsLisp, true
	case normalizeString(languageFSharpChromaStr):
		return LanguageFSharp, true
	case normalizeString(languageGosuChromaStr):
		return LanguageGosu, true
	case normalizeString(languageJSXChromaStr):
		return LanguageJSX, true
	case normalizeString(languageLessChromaStr):
		return LanguageLess, true
	case normalizeString(languageMakefileChromaStr):
		return LanguageMakefile, true
	case normalizeString(languageMarkdownChromaStr):
		return LanguageMarkdown, true
	case normalizeString(languageTextChromaStr):
		return LanguageText, true
	case normalizeString(languageVueJSChromaStr):
		return LanguageVueJS, true
	default:
		return ParseLanguage(lexerName)
	}
}

// MarshalJSON implements json.Marshaler interface.
func (l Language) MarshalJSON() ([]byte, error) {
	if l == LanguageUnknown {
		return []byte(`null`), nil
	}

	s := l.String()
	if s == "" {
		return nil, fmt.Errorf("invalid language %v", l)
	}

	return []byte(`"` + s + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (l *Language) UnmarshalJSON(v []byte) error {
	trimmed := strings.Trim(string(v), "\"")

	lang, _ := ParseLanguage(trimmed)

	*l = lang

	return nil
}

// String implements fmt.Stringer interface.
// nolint:gocyclo
func (l Language) String() string {
	switch l {
	case LanguageAda:
		return languageAdaStr
	case LanguageActionScript:
		return languageActionScriptStr
	case LanguageAgda:
		return languageAgdaStr
	case LanguageAnsible:
		return languageAnsibleStr
	case LanguageAppleScript:
		return languageAppleScriptStr
	case LanguageApacheConf:
		return languageApacheConfStr
	case LanguageASP:
		return languageASPStr
	case LanguageAssembly:
		return languageAssemblyStr
	case LanguageAutoconf:
		return languageAutoconfStr
	case LanguageAwk:
		return languageAwkStr
	case LanguageBasic:
		return languageBasicStr
	case LanguageBash:
		return languageBashStr
	case LanguageBatchScript:
		return languageBatchScriptStr
	case LanguageBibTeX:
		return languageBibTeXStr
	case LanguageBrightScript:
		return languageBrightScriptStr
	case LanguageC:
		return languageCStr
	case LanguageClojure:
		return languageClojureStr
	case LanguageCMake:
		return languageCMakeStr
	case LanguageCocoa:
		return languageCocoaStr
	case LanguageCoffeeScript:
		return languageCoffeeScriptStr
	case LanguageColdfusionHTML:
		return languageColdfusionHTMLStr
	case LanguageCommonLisp:
		return languageCommonLispStr
	case LanguageCoq:
		return languageCoqStr
	case LanguageCPerl:
		return languageCPerlStr
	case LanguageCPP:
		return languageCPPStr
	case LanguageCrontab:
		return languageCrontabStr
	case LanguageCrystal:
		return languageCrystalStr
	case LanguageCSharp:
		return languageCSharpStr
	case LanguageCSHTML:
		return languageCSHTMLStr
	case LanguageCSS:
		return languageCSSStr
	case LanguageCVS:
		return languageCVSStr
	case LanguageDart:
		return languageDartStr
	case LanguageDCL:
		return languageDCLStr
	case LanguageDelphi:
		return languageDelphiStr
	case LanguageDhall:
		return languageDhallStr
	case LanguageDiff:
		return languageDiffStr
	case LanguageDocker:
		return languageDockerStr
	case LanguageDocTeX:
		return languageDocTeXStr
	case LanguageElixir:
		return languageElixirStr
	case LanguageElm:
		return languageElmStr
	case LanguageEmacsLisp:
		return languageEmacsLispStr
	case LanguageErlang:
		return languageErlangStr
	case LanguageEshell:
		return languageEshellStr
	case LanguageFish:
		return languageFishStr
	case LanguageFortran:
		return languageFortranStr
	case LanguageFSharp:
		return languageFSharpStr
	case LanguageGo:
		return languageGoStr
	case LanguageGosu:
		return languageGosuStr
	case LanguageGroovy:
		return languageGroovyStr
	case LanguageHAML:
		return languageHAMLStr
	case LanguageHaskell:
		return languageHaskellStr
	case LanguageHaxe:
		return languageHaxeStr
	case LanguageHCL:
		return languageHCLStr
	case LanguageHTML:
		return languageHTMLStr
	case LanguageINI:
		return languageINIStr
	case LanguageJade:
		return languageJadeStr
	case LanguageJava:
		return languageJavaStr
	case LanguageJavaScript:
		return languageJavaScriptStr
	case LanguageJSON:
		return languageJSONStr
	case LanguageJSX:
		return languageJSXStr
	case LanguageKotlin:
		return languageKotlinStr
	case LanguageLasso:
		return languageLassoStr
	case LanguageLaTeX:
		return languageLaTeXStr
	case LanguageLess:
		return languageLessStr
	case LanguageLinkerScript:
		return languageLinkerScriptStr
	case LanguageLiquid:
		return languageLiquidStr
	case LanguageLua:
		return languageLuaStr
	case LanguageMakefile:
		return languageMakefileStr
	case LanguageMako:
		return languageMakoStr
	case LanguageMan:
		return languageManStr
	case LanguageMarkdown:
		return languageMarkdownStr
	case LanguageMarko:
		return languageMarkoStr
	case LanguageMatlab:
		return languageMatlabStr
	case LanguageMetafont:
		return languageMetafontStr
	case LanguageMetapost:
		return languageMetapostStr
	case LanguageModelica:
		return languageModelicaStr
	case LanguageModula2:
		return languageModula2Str
	case LanguageMustache:
		return languageMustacheStr
	case LanguageNewLisp:
		return languageNewLispStr
	case LanguageNix:
		return languageNixStr
	case LanguageObjectiveC:
		return languageObjectiveCStr
	case LanguageObjectiveCPP:
		return languageObjectiveCPPStr
	case LanguageObjectiveJ:
		return languageObjectiveJStr
	case LanguageOCaml:
		return languageOCamlStr
	case LanguageOrg:
		return languageOrgStr
	case LanguagePascal:
		return languagePascalStr
	case LanguagePawn:
		return languagePawnStr
	case LanguagePerl:
		return languagePerlStr
	case LanguagePHP:
		return languagePHPStr
	case LanguagePostScript:
		return languagePostScriptStr
	case LanguagePOVRay:
		return languagePOVRayStr
	case LanguagePowerShell:
		return languagePowerShellStr
	case LanguageProlog:
		return languagePrologStr
	case LanguageProtocolBuffer:
		return languageProtocolBufferStr
	case LanguagePug:
		return languagePugStr
	case LanguagePuppet:
		return languagePuppetStr
	case LanguagePureScript:
		return languagePureScriptStr
	case LanguagePython:
		return languagePythonStr
	case LanguageQML:
		return languageQMLStr
	case LanguageR:
		return languageRStr
	case LanguageReasonML:
		return languageReasonMLStr
	case LanguageReStructuredText:
		return languageReStructuredTextStr
	case LanguageRPMSpec:
		return languageRPMSpecStr
	case LanguageRuby:
		return languageRubyStr
	case LanguageRust:
		return languageRustStr
	case LanguageSalt:
		return languageSaltStr
	case LanguageSass:
		return languageSassStr
	case LanguageScala:
		return languageScalaStr
	case LanguageScheme:
		return languageSchemeStr
	case LanguageScribe:
		return languageScribeStr
	case LanguageSCSS:
		return languageSCSSStr
	case LanguageSGML:
		return languageSGMLStr
	case LanguageShell:
		return languageShellStr
	case LanguageSingularity:
		return languageSingularityStr
	case LanguageSimula:
		return languageSimulaStr
	case LanguageSketchDrawing:
		return languageSketchDrawingStr
	case LanguageSKILL:
		return languageSKILLStr
	case LanguageSlim:
		return languageSlimStr
	case LanguageSmali:
		return languageSmaliStr
	case LanguageSmalltalk:
		return languageSmalltalkStr
	case LanguageSMIME:
		return languageSMIMEStr
	case LanguageSourcePawn:
		return languageSourcePawnStr
	case LanguageSQL:
		return languageSQLStr
	case LanguageSublimeTextConfig:
		return languageSublimeTextConfigStr
	case LanguageSvelte:
		return languageSvelteStr
	case LanguageSwift:
		return languageSwiftStr
	case LanguageSWIG:
		return languageSWIGStr
	case LanguageSystemVerilog:
		return languageSystemVerilogStr
	case LanguageTeX:
		return languageTeXStr
	case LanguageText:
		return languageTextStr
	case LanguageThrift:
		return languageThriftStr
	case LanguageTOML:
		return languageTOMLStr
	case LanguageTuring:
		return languageTuringStr
	case LanguageTwig:
		return languageTwigStr
	case LanguageTypeScript:
		return languageTypeScriptStr
	case LanguageTypoScript:
		return languageTypoScriptStr
	case LanguageVB:
		return languageVBStr
	case LanguageVBNet:
		return languageVBNetStr
	case LanguageVCL:
		return languageVCLStr
	case LanguageVelocity:
		return languageVelocityStr
	case LanguageVerilog:
		return languageVerilogStr
	case LanguageVHDL:
		return languageVHDLStr
	case LanguageVimL:
		return languageVimLStr
	case LanguageVueJS:
		return languageVueJSStr
	case LanguageXAML:
		return languageXAMLStr
	case LanguageXML:
		return languageXMLStr
	case LanguageXSLT:
		return languageXSLTStr
	case LanguageYAML:
		return languageYAMLStr
	case LanguageZig:
		return languageZigStr
	default:
		return languageUnkownStr
	}
}

// StringChroma returns the corresponding chroma lexer name.
// nolint:gocyclo
func (l Language) StringChroma() string {
	switch l {
	case LanguageAssembly:
		return languageAssemblyChromaStr
	case LanguageColdfusionHTML:
		return languageColdfusionHTMLChromaStr
	case LanguageEmacsLisp:
		return languageEmacsLispChromaStr
	case LanguageFSharp:
		return languageFSharpChromaStr
	case LanguageGosu:
		return languageGosuChromaStr
	case LanguageJSX:
		return languageJSXChromaStr
	case LanguageLess:
		return languageLessChromaStr
	case LanguageMakefile:
		return languageMakefileChromaStr
	case LanguageMarkdown:
		return languageMarkdownChromaStr
	case LanguageText:
		return languageTextChromaStr
	case LanguageVueJS:
		return languageVueJSChromaStr
	default:
		return l.String()
	}
}

func normalizeString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "#", "sharp")
	s = strings.ReplaceAll(s, "++", "pp")

	return s
}
