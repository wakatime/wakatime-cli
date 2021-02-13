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
	// LanguageABAP represents the ABAP programming language.
	LanguageABAP
	// LanguageABNF represents the ABNF programming language.
	LanguageABNF
	// LanguageActionScript represents the ActionScript programming language.
	LanguageActionScript
	// LanguageActionScript3 represents the ActionScript3 programming language.
	LanguageActionScript3
	// LanguageAda represents the Ada programming language.
	LanguageAda
	// LanguageADL represents the ADL programming language.
	LanguageADL
	// LanguageAdvPL represents the AdvPL programming language.
	LanguageAdvPL
	// LanguageAgda represents the Agda programming language.
	LanguageAgda
	// LanguageAheui represents the Aheui programming language.
	LanguageAheui
	// LanguageAlloy represents the Alloy programming language.
	LanguageAlloy
	// LanguageAmbientTalk represents the AmbientTalk programming language.
	LanguageAmbientTalk
	// LanguageAmpl represents the Ampl programming language.
	LanguageAmpl
	// LanguageAngular2 represents the Angular2 programming language.
	LanguageAngular2
	// LanguageAnsible represents the Ansible programming language.
	LanguageAnsible
	// LanguageANTLR represents the ANTLR programming language.
	LanguageANTLR
	// LanguageApacheConfig represents the Apache Config programming language.
	LanguageApacheConfig
	// LanguageApex represents the Apex programming language.
	LanguageApex
	// LanguageAPL represents the APL programming language.
	LanguageAPL
	// LanguageAppleScript represents the AppleScript programming language.
	LanguageAppleScript
	// LanguageArc represents the Arc programming language.
	LanguageArc
	// LanguageArduino represents the Arduino programming language.
	LanguageArduino
	// LanguageArrow represents the Arrow programming language.
	LanguageArrow
	// LanguageASPClassic represents the ASP Classic programming language.
	LanguageASPClassic
	// LanguageASPDotNet represents the ASPDotNet programming language.
	LanguageASPDotNet
	// LanguageAspectJ represents the AspectJ programming language.
	LanguageAspectJ
	// LanguageAspxCSharp represents the CSharpAspx programming language.
	LanguageAspxCSharp
	// LanguageAspxVBNet represents the VBNetAspx programming language.
	LanguageAspxVBNet
	// LanguageAssembly represents the Assembly programming language.
	LanguageAssembly
	// LanguageAsymptote represents the Asymptote programming language.
	LanguageAsymptote
	// LanguageAugeas represents the Augeas programming language.
	LanguageAugeas
	// LanguageAutoconf represents the Autoconf programming language.
	LanguageAutoconf
	// LanguageAutoHotkey represents the AutoHotkey programming language.
	LanguageAutoHotkey
	// LanguageAutoIt represents the AutoIt programming language.
	LanguageAutoIt
	// LanguageAwk represents the Awk programming language.
	LanguageAwk
	// LanguageBallerina represents the Ballerina programming language.
	LanguageBallerina
	// LanguageBARE represents the BARE programming language.
	LanguageBARE
	// LanguageBash represents the Bash programming language.
	LanguageBash
	// LanguageBashSession represents the BashSession programming language.
	LanguageBashSession
	// LanguageBasic represents the Basic programming language.
	LanguageBasic
	// LanguageBatchfile represents the Batchfile programming language.
	LanguageBatchfile
	// LanguageBatchScript represents the BatchScript programming language.
	LanguageBatchScript
	// LanguageBBCBasic represents the BBCBasic programming language.
	LanguageBBCBasic
	// LanguageBBCode represents the BBCode programming language.
	LanguageBBCode
	// LanguageBC represents the BC programming language.
	LanguageBC
	// LanguageBefunge represents the Befunge programming language.
	LanguageBefunge
	// LanguageBibTeX represents the BibTeX programming language.
	LanguageBibTeX
	// LanguageBladeTemplate represents the BladeTemplate programming language.
	LanguageBladeTemplate
	// LanguageBlazor represent the Blazor programming language.
	LanguageBlazor
	// LanguageBlitzBasic represents the BlitzBasic programming language.
	LanguageBlitzBasic
	// LanguageBlitzMax represents the BlitzMax programming language.
	LanguageBlitzMax
	// LanguageBNF represents the BNF programming language.
	LanguageBNF
	// LanguageBoa represents the Boa programming language.
	LanguageBoa
	// LanguageBoo represents the Boo programming language.
	LanguageBoo
	// LanguageBoogie represents the Boogie programming language.
	LanguageBoogie
	// LanguageBrainfuck represents the Brainfuck programming language.
	LanguageBrainfuck
	// LanguageBrightScript represents the BrightScript programming language.
	LanguageBrightScript
	// LanguageBro represents the Bro programming language.
	LanguageBro
	// LanguageBST represents the BST programming language.
	LanguageBST
	// LanguageBUGS represents the BUGS programming language.
	LanguageBUGS
	// LanguageC represents the C programming language.
	LanguageC
	// LanguageCa65Assembler represents the ca65 assembler programming language.
	LanguageCa65Assembler
	// LanguageCaddyfileDirectives represents the Caddyfile Directives programming language.
	LanguageCaddyfileDirectives
	// LanguageCaddyfile represents the Caddyfile programming language.
	LanguageCaddyfile
	// LanguageCADL represents the CADL programming language.
	LanguageCADL
	// LanguageCAmkES represents the CAmkES programming language.
	LanguageCAmkES
	// LanguageCapDL represents the CapDL programming language.
	LanguageCapDL
	// LanguageCapNProto represents the CapNProto programming language.
	LanguageCapNProto
	// LanguageCassandraCQL represents the CassandraCQL programming language.
	LanguageCassandraCQL
	// LanguageCBMBasicV2 represents the CBMBasicV2 programming language.
	LanguageCBMBasicV2
	// LanguageCeylon represents the Ceylon programming language.
	LanguageCeylon
	// LanguageCFEngine3 represents the CFEngine3 programming language.
	LanguageCFEngine3
	// LanguageCfstatement represents the Cfstatement programming language.
	LanguageCfstatement
	// LanguageChaiScript represents the ChaiScript programming language.
	LanguageChaiScript
	// LanguageChapel represents the Chapel programming language.
	LanguageChapel
	// LanguageCharmci represents the Charmci programming language.
	LanguageCharmci
	// LanguageCheetah represents the Cheetah programming language.
	LanguageCheetah
	// LanguageCirru represents the Cirru programming language.
	LanguageCirru
	// LanguageClay represents the Clay programming language.
	LanguageClay
	// LanguageClean represents the Clean programming language.
	LanguageClean
	// LanguageClojure represents the Clojure programming language.
	LanguageClojure
	// LanguageClojureScript represents the ClojureScript programming language.
	LanguageClojureScript
	// LanguageCMake represents the CMake programming language.
	LanguageCMake
	// LanguageCObjdump represents the CObjdump programming language.
	LanguageCObjdump
	// LanguageCOBOL represents the COBOL programming language.
	LanguageCOBOL
	// LanguageCOBOLFree represents the COBOLFree programming language.
	LanguageCOBOLFree
	// LanguageCocoa represents the Cocoa programming language.
	LanguageCocoa
	// LanguageCoffeeScript represents the CoffeeScript programming language.
	LanguageCoffeeScript
	// LanguageColdfusionCFC represents the ColdfusionCFC programming language.
	LanguageColdfusionCFC
	// LanguageColdfusionHTML represents the ColdfusionHTML programming language.
	LanguageColdfusionHTML
	// LanguageCommonLisp represents the CommonLisp programming language.
	LanguageCommonLisp
	// LanguageComponentPascal represents the ComponentPascal programming language.
	LanguageComponentPascal
	// LanguageCoq represents the Coq programming language.
	LanguageCoq
	// LanguageCPerl represents the CPerl programming language.
	LanguageCPerl
	// LanguageCPP represents the CPP programming language.
	LanguageCPP
	// LanguageCppObjdump represents the CppObjdump programming language.
	LanguageCppObjdump
	// LanguageCPSA represents the CPSA programming language.
	LanguageCPSA
	// LanguageCrmsh represents the Crmsh programming language.
	LanguageCrmsh
	// LanguageCroc represents the Croc programming language.
	LanguageCroc
	// LanguageCryptol represents the Cryptol programming language.
	LanguageCryptol
	// LanguageCSharp represents the CSharp programming language.
	LanguageCSharp
	// LanguageCSHTML represents the CSHTML programming language.
	LanguageCSHTML
	// LanguageCrontab represents the Crontab programming language.
	LanguageCrontab
	// LanguageCrystal represents the Crystal programming language.
	LanguageCrystal
	// LanguageCSON represents the CSON programming language.
	LanguageCSON
	// LanguageCsoundDocument represents the CsoundDocument programming language.
	LanguageCsoundDocument
	// LanguageCsoundOrchestra represents the CsoundOrchestra programming language.
	LanguageCsoundOrchestra
	// LanguageCsoundScore represents the CsoundScore programming language.
	LanguageCsoundScore
	// LanguageCSS represents the CSS programming language.
	LanguageCSS
	// LanguageCSV represents the CSV programming language.
	LanguageCSV
	// LanguageCUDA represents the CUDA programming language.
	LanguageCUDA
	// LanguageCVS represents the CVS programming language.
	LanguageCVS
	// LanguageCypher represents the Cypher programming language.
	LanguageCypher
	// LanguageCython represents the Cython programming language.
	LanguageCython
	// LanguageD represents the D programming language.
	LanguageD
	// LanguageDarcsPatch represents the DarcsPatch programming language.
	LanguageDarcsPatch
	// LanguageDart represents the Dart programming language.
	LanguageDart
	// LanguageDASM16 represents the DASM16 programming language.
	LanguageDASM16
	// LanguageDCL represents the DCL programming language.
	LanguageDCL
	// LanguageDCPU16Asm represents the DCPU16Asm programming language.
	LanguageDCPU16Asm
	// LanguageDebianControlFile represents the DebianControlFile programming language.
	LanguageDebianControlFile
	// LanguageDelphi represents the Delphi programming language.
	LanguageDelphi
	// LanguageDevicetree represents the Devicetree programming language.
	LanguageDevicetree
	// LanguageDG represents the DG programming language.
	LanguageDG
	// LanguageDhall represents the Dhall programming language.
	LanguageDhall
	// LanguageDiff represents the Diff programming language.
	LanguageDiff
	// LanguageDjangoJinja represents the DjangoJinja programming language.
	LanguageDjangoJinja
	// LanguageDObjdump represents the DObjdump programming language.
	LanguageDObjdump
	// LanguageDocker represents the Docker programming language.
	LanguageDocker
	// LanguageDocTeX represents the DocTeX programming language.
	LanguageDocTeX
	// LanguageDTD represents the DTD programming language.
	LanguageDTD
	// LanguageDuel represents the Duel programming language.
	LanguageDuel
	// LanguageDylan represents the Dylan programming language.
	LanguageDylan
	// LanguageDylanLID represents the DylanLID programming language.
	LanguageDylanLID
	// LanguageDylanSession represents the DylanSession programming language.
	LanguageDylanSession
	// LanguageDynASM represents the DynASM programming language.
	LanguageDynASM
	// LanguageEMail represents the EMail programming language.
	LanguageEMail
	// LanguageEarlGrey represents the EarlGrey programming language.
	LanguageEarlGrey
	// LanguageEasytrieve represents the Easytrieve programming language.
	LanguageEasytrieve
	// LanguageEBNF represents the EBNF programming language.
	LanguageEBNF
	// LanguageEC represents the EC programming language.
	LanguageEC
	// LanguageECL represents the ECL programming language.
	LanguageECL
	// LanguageEiffel represents the Eiffel programming language.
	LanguageEiffel
	// LanguageEJS represents the EJS programming language.
	LanguageEJS
	// LanguageElixir represents the Elixir programming language.
	LanguageElixir
	// LanguageElixirIexSession represents the ElixirIexSession programming language.
	LanguageElixirIexSession
	// LanguageElm represents the Elm programming language.
	LanguageElm
	// LanguageEmacsLisp represents the EmacsLisp programming language.
	LanguageEmacsLisp
	// LanguageERB represents the ERB programming language.
	LanguageERB
	// LanguageErlang represents the Erlang programming language.
	LanguageErlang
	// LanguageErlangErlSession represents the ErlangErlSession programming language.
	LanguageErlangErlSession
	// LanguageEshell represents the Eshell programming language.
	LanguageEshell
	// LanguageEvoque represents the Evoque programming language.
	LanguageEvoque
	// LanguageExecline represents the Execline programming language.
	LanguageExecline
	// LanguageEzhil represents the Ezhil programming language.
	LanguageEzhil
	// LanguageFactor represents the Factor programming language.
	LanguageFactor
	// LanguageFancy represents the Fancy programming language.
	LanguageFancy
	// LanguageFantom represents the Fantom programming language.
	LanguageFantom
	// LanguageFelix represents the Felix programming language.
	LanguageFelix
	// LanguageFennel represents the Fennel programming language.
	LanguageFennel
	// LanguageFlatline represents the Flatline programming language.
	LanguageFlatline
	// LanguageFloScript represents the FloScript programming language.
	LanguageFloScript
	// LanguageForth represents the Forth programming language.
	LanguageForth
	// LanguageFish represents the Fish programming language.
	LanguageFish
	// LanguageFont represents the Font programming language.
	LanguageFont
	// LanguageFortran represents the Fortran programming language.
	LanguageFortran
	// LanguageFortranFixed represents the FortranFixed programming language.
	LanguageFortranFixed
	// LanguageFSharp represents the FSharp programming language.
	LanguageFSharp
	// LanguageFoxPro represents the FoxPro programming language.
	LanguageFoxPro
	// LanguageFreefem represents the Freefem programming language.
	LanguageFreefem
	// LanguageFStar represents the FStar programming language.
	LanguageFStar
	// LanguageGap represents the Gap programming language.
	LanguageGap
	// LanguageGas represents the Gas programming language.
	LanguageGas
	// LanguageGDNative represents the GDNative programming language.
	LanguageGDNative
	// LanguageGDScript represents the GDScript programming language.
	LanguageGDScript
	// LanguageGenshi represents the Genshi programming language.
	LanguageGenshi
	// LanguageGenshiHTML represents the Genshi HTML programming language.
	LanguageGenshiHTML
	// LanguageGenshiText represents the Genshi Text programming language.
	LanguageGenshiText
	// LanguageGettextCatalog represents the Gettext Catalog programming language.
	LanguageGettextCatalog
	// LanguageGherkin represents the Gherkin programming language.
	LanguageGherkin
	// LanguageGit represents the Git programming language.
	LanguageGit
	// LanguageGitConfig represents the Git Config programming language.
	LanguageGitConfig
	// LanguageGLSL represents the GLSL programming language.
	LanguageGLSL
	// LanguageGnuplot represents the Gnuplot programming language.
	LanguageGnuplot
	// LanguageGo represents the Go programming language.
	LanguageGo
	// LanguageGoHTMLTemplate represents the Go HTML Template programming language.
	LanguageGoHTMLTemplate
	// LanguageGoTextTemplate represents the Go Text Template programming language.
	LanguageGoTextTemplate
	// LanguageGolo represents the Golo programming language.
	LanguageGolo
	// LanguageGoodDataCL represents the GoodData-CL programming language.
	LanguageGoodDataCL
	// LanguageGosu represents the Gosu programming language.
	LanguageGosu
	// LanguageGosuTemplate represents the Gosu Template programming language.
	LanguageGosuTemplate
	// LanguageGradleConfig represents the GradleConfig programming language.
	LanguageGradleConfig
	// LanguageGraphQL represents the GraphQL programming language.
	LanguageGraphQL
	// LanguageGroff represents the Groff programming language.
	LanguageGroff
	// LanguageGroovy represents the Groovy programming language.
	LanguageGroovy
	// LanguageHaml represents the Haml programming language.
	LanguageHaml
	// LanguageHandlebars represents the Handlebars programming language.
	LanguageHandlebars
	// LanguageHaskell represents the Haskell programming language.
	LanguageHaskell
	// LanguageHaxe represents the Haxe programming language.
	LanguageHaxe
	// LanguageHCL represents the HCL programming language.
	LanguageHCL
	// LanguageHexdump represents the Hexdump programming language.
	LanguageHexdump
	// LanguageHLB represents the HLB programming language.
	LanguageHLB
	// LanguageHLSL represents the HLSL programming language.
	LanguageHLSL
	// LanguageHSAIL represents the HSAIL programming language.
	LanguageHSAIL
	// LanguageHspec represents the Hspec programming language.
	LanguageHspec
	// LanguageHTML represents the HTML programming language.
	LanguageHTML
	// LanguageHTTP represents the HTTP programming language.
	LanguageHTTP
	// LanguageHxml represents the Hxml programming language.
	LanguageHxml
	// LanguageHy represents the Hy programming language.
	LanguageHy
	// LanguageHybris represents the Hybris programming language.
	LanguageHybris
	// LanguageIcon represents the Icon programming language.
	LanguageIcon
	// LanguageIDL represents the IDL programming language.
	LanguageIDL
	// LanguageIdris represents the Idris programming language.
	LanguageIdris
	// LanguageIgor represents the Igor programming language.
	LanguageIgor
	// LanguageImageJPEG represents the Image (jpeg) programming language.
	LanguageImageJPEG
	// LanguageImagePNG represents the Image (png) programming language.
	LanguageImagePNG
	// LanguageInform6 represents the Inform 6 programming language.
	LanguageInform6
	// LanguageInform6Template represents the Inform 6 template programming language.
	LanguageInform6Template
	// LanguageInform7 represents the Inform 7 programming language.
	LanguageInform7
	// LanguageINI represents the INI programming language.
	LanguageINI
	// LanguageIo represents the Io programming language.
	LanguageIo
	// LanguageIoke represents the Ioke programming language.
	LanguageIoke
	// LanguageIRCLogs represents the IRC Logs programming language.
	LanguageIRCLogs
	// LanguageIsabelle represents the Isabelle programming language.
	LanguageIsabelle
	// LanguageJ represents the J programming language.
	LanguageJ
	// LanguageJAGS represents the JAGS programming language.
	LanguageJAGS
	// LanguageJade represents the Jade programming language.
	LanguageJade
	// LanguageJasmin represents the Jasmin programming language.
	LanguageJasmin
	// LanguageJava represents the Java programming language.
	LanguageJava
	// LanguageJavaScript represents the JavaScript programming language.
	LanguageJavaScript
	// LanguageJCL represents the JCL programming language.
	LanguageJCL
	// LanguageJSGF represents the JSGF programming language.
	LanguageJSGF
	// LanguageJSON represents the JSON programming language.
	LanguageJSON
	// LanguageJSONLD represents the JSON-LD programming language.
	LanguageJSONLD
	// LanguageJSP represents the Java Server Page programming language.
	LanguageJSP
	// LanguageJSX represents the JSX programming language.
	LanguageJSX
	// LanguageJulia represents the Julia programming language.
	LanguageJulia
	// LanguageJuliaConsole represents the Julia console programming language.
	LanguageJuliaConsole
	// LanguageJungle represents the Jungle programming language.
	LanguageJungle
	// LanguageJuttle represents the Juttle console programming language.
	LanguageJuttle
	// LanguageKal represents the Kal console programming language.
	LanguageKal
	// LanguageKconfig represents the Kconfig console programming language.
	LanguageKconfig
	// LanguageKernelLog represents the Kernel log console programming language.
	LanguageKernelLog
	// LanguageKoka represents the Koka console programming language.
	LanguageKoka
	// LanguageKotlin represents the Kotlin programming language.
	LanguageKotlin
	// LanguageLaravelTemplate represents the Laravel Template programming language.
	LanguageLaravelTemplate
	// LanguageLasso represents the Lasso programming language.
	LanguageLasso
	// LanguageLaTeX represents the LaTeX programming language.
	LanguageLaTeX
	// LanguageLatte represents the Latte programming language.
	LanguageLatte
	// LanguageLean represents the Lean programming language.
	LanguageLean
	// LanguageLess represents the Less programming language.
	LanguageLess
	// LanguageLighttpd represents the Lighttpd configuration file programming language.
	LanguageLighttpd
	// LanguageLimbo represents the Limbo configuration file programming language.
	LanguageLimbo
	// LanguageLinkerScript represents the LinkerScript programming language.
	LanguageLinkerScript
	// LanguageLiquid represents the Liquid programming language.
	LanguageLiquid
	// LanguageLiterateAgda represents the Literate Agda programming language.
	LanguageLiterateAgda
	// LanguageLiterateCryptol represents the Literate Cryptol programming language.
	LanguageLiterateCryptol
	// LanguageLiterateHaskell represents the Literate Haskell programming language.
	LanguageLiterateHaskell
	// LanguageLiterateIdris represents the Literate Idris programming language.
	LanguageLiterateIdris
	// LanguageLiveScript represents the LiveScript programming language.
	LanguageLiveScript
	// LanguageLLVM represents the LLVM programming language.
	LanguageLLVM
	// LanguageLLVMMIR represents the LLVM-MIR programming language.
	LanguageLLVMMIR
	// LanguageLLVMMIRBody represents the LLVM-MIR Body programming language.
	LanguageLLVMMIRBody
	// LanguageLogos represents the Logos programming language.
	LanguageLogos
	// LanguageLogFile represents the LogFile programming language.
	LanguageLogFile
	// LanguageLogtalk represents the Logtalk programming language.
	LanguageLogtalk
	// LanguageLSL represents the LSL programming language.
	LanguageLSL
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
	languageUnkownStr              = "Unknown"
	languageABAPStr                = "ABAP"
	languageABNFStr                = "ABNF"
	languageActionScriptStr        = "ActionScript"
	languageActionScript3Str       = "ActionScript 3"
	languageAdaStr                 = "Ada"
	languageADLStr                 = "ADL"
	languageAdvPLStr               = "AdvPL"
	languageAgdaStr                = "Agda"
	languageAheuiStr               = "Aheui"
	languageAlloyStr               = "Alloy"
	languageAmbientTalkStr         = "AmbientTalk"
	languageAmplStr                = "Ampl"
	languageAngular2Str            = "Angular2"
	languageAnsibleStr             = "Ansible"
	languageANTLRStr               = "ANTLR"
	languageApacheConfigStr        = "Apache Config"
	languageApexStr                = "Apex"
	languageAPLStr                 = "APL"
	languageAppleScriptStr         = "AppleScript"
	languageArcStr                 = "Arc"
	languageArduinoStr             = "Arduino"
	languageArrowStr               = "Arrow"
	languageASPClassicStr          = "ASP Classic"
	languageASPDotNetStr           = "ASP.NET"
	languageAspectJStr             = "AspectJ"
	languageAspxCSharpStr          = "aspx-cs"
	languageAspxVBNetStr           = "aspx-vb"
	languageAssemblyStr            = "Assembly"
	languageAsymptoteStr           = "Asymptote"
	languageAugeasStr              = "Augeas"
	languageAutoconfStr            = "Autoconf"
	languageAutoHotkeyStr          = "AutoHotkey"
	languageAutoItStr              = "AutoIt"
	languageAwkStr                 = "AWK"
	languageBallerinaStr           = "Ballerina"
	languageBAREStr                = "BARE"
	languageBashStr                = "Bash"
	languageBashSessionStr         = "Bash Session"
	languageBasicStr               = "Basic"
	languageBatchfileStr           = "Batchfile"
	languageBatchScriptStr         = "Batch Script"
	languageBBCBasicStr            = "BBC Basic"
	languageBBCodeStr              = "BBCode"
	languageBCStr                  = "BC"
	languageBefungeStr             = "Befunge"
	languageBibTeXStr              = "BibTeX"
	languageBladeTemplateStr       = "Blade Template"
	languageBlazorStr              = "Blazor"
	languageBlitzBasicStr          = "BlitzBasic"
	languageBlitzMaxStr            = "BlitzMax"
	languageBNFStr                 = "BNF"
	languageBoaStr                 = "Boa"
	languageBooStr                 = "Boo"
	languageBoogieStr              = "Boogie"
	languageBrainfuckStr           = "Brainfuck"
	languageBrightScriptStr        = "BrightScript"
	languageBroStr                 = "Bro"
	languageBSTStr                 = "BST"
	languageBUGSStr                = "BUGS"
	languageCStr                   = "C"
	languageCa65AssemblerStr       = "ca65 assembler"
	languageCaddyfileStr           = "Caddyfile"
	languageCaddyfileDirectivesStr = "Caddyfile Directives"
	languageCADLStr                = "cADL"
	languageCAmkESStr              = "CAmkES"
	languageCapDLStr               = "CapDL"
	languageCapNProtoStr           = "Cap'n Proto"
	languageCassandraCQLStr        = "Cassandra CQL"
	languageCBMBasicV2Str          = "CBM BASIC V2"
	languageCeylonStr              = "Ceylon"
	languageCFEngine3Str           = "CFEngine3"
	languageCfstatementStr         = "cfstatement"
	languageChaiScriptStr          = "ChaiScript"
	languageChapelStr              = "Chapel"
	languageCharmciStr             = "Charmci"
	languageCheetahStr             = "Cheetah"
	languageCirruStr               = "Cirru"
	languageClayStr                = "Clay"
	languageCleanStr               = "Clean"
	languageClojureStr             = "Clojure"
	languageClojureScriptStr       = "ClojureScript"
	languageCMakeStr               = "CMake"
	languageCObjdumpStr            = "c-objdump"
	languageCOBOLStr               = "COBOL"
	languageCOBOLFreeStr           = "COBOLFree"
	languageCocoaStr               = "Cocoa"
	languageCoqStr                 = "Coq"
	languageCoffeeScriptStr        = "CoffeeScript"
	languageColdfusionHTMLStr      = "Coldfusion"
	languageColdfusionCFCStr       = "Coldfusion CFC"
	languageCommonLispStr          = "Common Lisp"
	languageComponentPascalStr     = "Component Pascal"
	languageCPerlStr               = "cperl"
	languageCPPStr                 = "C++"
	languageCppObjdumpStr          = "cpp-objdump"
	languageCPSAStr                = "CPSA"
	languageCrmshStr               = "Crmsh"
	languageCrocStr                = "Croc"
	languageCrontabStr             = "Crontab"
	languageCryptolStr             = "Cryptol"
	languageCrystalStr             = "Crystal"
	languageCSharpStr              = "C#"
	languageCSHTMLStr              = "CSHTML"
	languageCSONStr                = "CSON"
	languageCsoundDocumentStr      = "Csound Document"
	languageCsoundOrchestraStr     = "Csound Orchestra"
	languageCsoundScoreStr         = "Csound Score"
	languageCSSStr                 = "CSS"
	languageCSVStr                 = "CSV"
	languageCUDAStr                = "CUDA"
	languageCVSStr                 = "CVS"
	languageCypherStr              = "Cypher"
	languageCythonStr              = "Cython"
	languageDStr                   = "D"
	languageDarcsPatchStr          = "Darcs Patch"
	languageDartStr                = "Dart"
	languageDASM16Str              = "DASM16"
	languageDCLStr                 = "DCL"
	languageDCPU16AsmStr           = "DCPU-16 ASM"
	languageDebianControlFileStr   = "Debian Control file"
	languageDelphiStr              = "Delphi"
	languageDevicetreeStr          = "Devicetree"
	languageDGStr                  = "dg"
	languageDhallStr               = "Dhall"
	languageDiffStr                = "Diff"
	languageDjangoJinjaStr         = "Django/Jinja"
	languageDObjdumpStr            = "d-objdump"
	languageDockerStr              = "Docker"
	languageDocTeXStr              = "DocTeX"
	languageDTDStr                 = "DTD"
	languageDuelStr                = "Duel"
	languageDylanStr               = "Dylan"
	languageDylanLIDStr            = "DylanLID"
	languageDylanSessionStr        = "Dylan session"
	languageDynASMStr              = "DynASM"
	languageEarlGreyStr            = "Earl Grey"
	languageEasytrieveStr          = "Easytrieve"
	languageEBNFStr                = "EBNF"
	languageECStr                  = "eC"
	languageECLStr                 = "ECL"
	languageEiffelStr              = "Eiffel"
	languageEJSStr                 = "EJS"
	languageElixirIexSessionStr    = "Elixir iex session"
	languageElixirStr              = "Elixir"
	languageElmStr                 = "Elm"
	languageEmacsLispStr           = "Emacs Lisp"
	languageEMailStr               = "E-mail"
	languageERBStr                 = "ERB"
	languageErlangStr              = "Erlang"
	languageErlangErlSessionStr    = "Erlang erl session"
	languageEshellStr              = "Eshell"
	languageEvoqueStr              = "Evoque"
	languageExeclineStr            = "execline"
	languageEzhilStr               = "Ezhil"
	languageFactorStr              = "Factor"
	languageFancyStr               = "Fancy"
	languageFantomStr              = "Fantom"
	languageFelixStr               = "Felix"
	languageFennelStr              = "Fennel"
	languageFlatlineStr            = "Flatline"
	languageFloScriptStr           = "FloScript"
	languageForthStr               = "Forth"
	languageFishStr                = "Fish"
	languageFontStr                = "Font"
	languageFortranStr             = "Fortran"
	languageFortranFixedStr        = "FortranFixed"
	languageFSharpStr              = "F#"
	languageFoxProStr              = "FoxPro"
	languageFreefemStr             = "Freefem"
	languageFStarStr               = "FStar"
	languageGapStr                 = "GAP"
	languageGasStr                 = "GAS"
	languageGDNativeStr            = "GDNative"
	languageGDScriptStr            = "GDScript"
	languageGenshiStr              = "Genshi"
	languageGenshiHTMLStr          = "Genshi HTML"
	languageGenshiTextStr          = "Genshi Text"
	languageGettextCatalogStr      = "Gettext Catalog"
	languageGherkinStr             = "Gherkin"
	languageGitStr                 = "Git"
	languageGitConfigStr           = "Git Config"
	languageGLSLStr                = "GLSL"
	languageGnuplotStr             = "Gnuplot"
	languageGoStr                  = "Go"
	languageGoloStr                = "Golo"
	languageGoodDataCLStr          = "GoodData-CL"
	languageGosuStr                = "Gosu"
	languageGosuTemplateStr        = "Gosu Template"
	languageGradleConfigStr        = "Gradle Config"
	languageGraphQLStr             = "GraphQL"
	languageGroffStr               = "Groff"
	languageGroovyStr              = "Groovy"
	languageHamlStr                = "Haml"
	languageHandlebarsStr          = "Handlebars"
	languageHaskellStr             = "Haskell"
	languageHaxeStr                = "Haxe"
	languageHCLStr                 = "HCL"
	languageHexdumpStr             = "Hexdump"
	languageHLBStr                 = "HLB"
	languageHLSLStr                = "HLSL"
	languageHSAILStr               = "HSAIL"
	languageHspecStr               = "Hspec"
	languageHTMLStr                = "HTML"
	languageHTTPStr                = "HTTP"
	languageHxmlStr                = "Hxml"
	languageHyStr                  = "Hy"
	languageHybrisStr              = "Hybris"
	languageIconStr                = "Icon"
	languageIDLStr                 = "IDL"
	languageIdrisStr               = "Idris"
	languageIgorStr                = "Igor"
	languageImageJPEGStr           = "Image (jpeg)"
	languageImagePNGStr            = "Image (png)"
	languageInform6Str             = "Inform 6"
	languageInform6TemplateStr     = "Inform 6 template"
	languageInform7Str             = "Inform 7"
	languageINIStr                 = "INI"
	languageIoStr                  = "Io"
	languageIokeStr                = "Ioke"
	languageIRCLogsStr             = "IRC Logs"
	languageIsabelleStr            = "Isabelle"
	languageJStr                   = "J"
	languageJAGSStr                = "JAGS"
	languageJadeStr                = "Jade"
	languageJasminStr              = "Jasmin"
	languageJavaStr                = "Java"
	languageJavaScriptStr          = "JavaScript"
	languageJCLStr                 = "JCL"
	languageJSGFStr                = "JSGF"
	languageJSONStr                = "JSON"
	languageJSONLDStr              = "JSON-LD"
	languageJSPStr                 = "Java Server Page"
	languageJSXStr                 = "JSX"
	languageJuliaStr               = "Julia"
	languageJuliaConsoleStr        = "Julia console"
	languageJungleStr              = "Jungle"
	languageJuttleStr              = "Juttle"
	languageKalStr                 = "Kal"
	languageKconfigStr             = "Kconfig"
	languageKernelLogStr           = "Kernel log"
	languageKokaStr                = "Koka"
	languageKotlinStr              = "Kotlin"
	languageLaravelTemplateStr     = "Laravel Template"
	languageLassoStr               = "Lasso"
	languageLaTeXStr               = "LaTeX"
	languageLatteStr               = "Latte"
	languageLeanStr                = "Lean"
	languageLessStr                = "LESS"
	languageLighttpdStr            = "Lighttpd configuration file"
	languageLimboStr               = "Limbo"
	languageLinkerScriptStr        = "Linker Script"
	languageLiquidStr              = "Liquid"
	languageLiterateAgdaStr        = "Literate Agda"
	languageLiterateCryptolStr     = "Literate Cryptol"
	languageLiterateHaskellStr     = "Literate Haskell"
	languageLiterateIdrisStr       = "Literate Idris"
	languageLiveScriptStr          = "LiveScript"
	languageLLVMStr                = "LLVM"
	languageLLVMMIRStr             = "LLVM-MIR"
	languageLLVMMIRBodyStr         = "LLVM-MIR Body"
	languageLogosStr               = "Logos"
	languageLogFileStr             = "Log File"
	languageLogtalkStr             = "Logtalk"
	languageLSLStr                 = "LSL"
	languageLuaStr                 = "Lua"
	languageMakefileStr            = "Makefile"
	languageMakoStr                = "Mako"
	languageManStr                 = "Man"
	languageMarkdownStr            = "Markdown"
	languageMarkoStr               = "Marko"
	languageMatlabStr              = "Matlab"
	languageMetafontStr            = "Metafont"
	languageMetapostStr            = "Metapost"
	languageModelicaStr            = "Modelica"
	languageModula2Str             = "Modula-2"
	languageMustacheStr            = "Mustache"
	languageNewLispStr             = "NewLisp"
	languageNixStr                 = "Nix"
	languageObjectiveCStr          = "Objective-C"
	languageObjectiveCPPStr        = "Objective-C++"
	languageObjectiveJStr          = "Objective-J"
	languageOCamlStr               = "OCaml"
	languageOrgStr                 = "Org"
	languagePascalStr              = "Pascal"
	languagePawnStr                = "Pawn"
	languagePerlStr                = "Perl"
	languagePHPStr                 = "PHP"
	languagePOVRayStr              = "POVRay"
	languagePostScriptStr          = "PostScript"
	languagePowerShellStr          = "PowerShell"
	languagePrologStr              = "Prolog"
	languageProtocolBufferStr      = "Protocol Buffer"
	languagePugStr                 = "Pug"
	languagePuppetStr              = "Puppet"
	languagePureScriptStr          = "PureScript"
	languagePythonStr              = "Python"
	languageQMLStr                 = "QML"
	languageRStr                   = "R"
	languageReasonMLStr            = "ReasonML"
	languageReStructuredTextStr    = "reStructuredText"
	languageRPMSpecStr             = "RPMSpec"
	languageRubyStr                = "Ruby"
	languageRustStr                = "Rust"
	languageSaltStr                = "Salt"
	languageSassStr                = "Sass"
	languageScalaStr               = "Scala"
	languageSchemeStr              = "Scheme"
	languageScribeStr              = "Scribe"
	languageSCSSStr                = "SCSS"
	languageSGMLStr                = "SGML"
	languageShellStr               = "Shell"
	languageSimulaStr              = "Simula"
	languageSingularityStr         = "Singularity"
	languageSketchDrawingStr       = "Sketch Drawing"
	languageSKILLStr               = "SKILL"
	languageSlimStr                = "Slim"
	languageSmaliStr               = "Smali"
	languageSmalltalkStr           = "Smalltalk"
	languageSMIMEStr               = "S/MIME"
	languageSourcePawnStr          = "SourcePawn"
	languageSQLStr                 = "SQL"
	languageSublimeTextConfigStr   = "Sublime Text Config"
	languageSvelteStr              = "Svelte"
	languageSwiftStr               = "Swift"
	languageSWIGStr                = "SWIG"
	languageSystemVerilogStr       = "systemverilog"
	languageTeXStr                 = "TeX"
	languageTextStr                = "Text"
	languageThriftStr              = "Thrift"
	languageTOMLStr                = "TOML"
	languageTuringStr              = "Turing"
	languageTwigStr                = "Twig"
	languageTypeScriptStr          = "TypeScript"
	languageTypoScriptStr          = "TypoScript"
	languageVBStr                  = "VB"
	languageVBNetStr               = "VB.net"
	languageVCLStr                 = "VCL"
	languageVelocityStr            = "Velocity"
	languageVerilogStr             = "Verilog"
	languageVHDLStr                = "vhdl"
	languageVimLStr                = "VimL"
	languageVueJSStr               = "Vue.js"
	languageXAMLStr                = "XAML"
	languageXMLStr                 = "XML"
	languageXSLTStr                = "XSLT"
	languageYAMLStr                = "YAML"
	languageZigStr                 = "Zig"
)

const (
	languageApacheConfigChromaStr   = "ApacheConf"
	languageAssemblyChromaStr       = "GAS"
	languageColdfusionHTMLChromaStr = "Coldfusion HTML"
	languageFSharpChromaStr         = "FSharp"
	languageEmacsLispChromaStr      = "EmacsLisp"
	languageGoHTMLTemplateChromaStr = "Go HTML Template"
	languageGoTextTemplateChromaStr = "Go Text Template"
	languageGosuTemplateChromaStr   = "Gosu Template"
	languageJSXChromaStr            = "react"
	languageLessChromaStr           = "LessCss"
	languageLiquidChromaStr         = "liquid"
	languageMakefileChromaStr       = "Base Makefile"
	languageMarkdownChromaStr       = "markdown"
	languageTextChromaStr           = "plaintext"
	languageVHDLChromaStr           = "VHDL"
	languageVueJSChromaStr          = "vue"
)

// ParseLanguage parses a language from a string. Will return false
// as second parameter, if language could not be parsed.
// nolint:gocyclo
func ParseLanguage(s string) (Language, bool) {
	switch normalizeString(s) {
	case normalizeString(languageABNFStr):
		return LanguageABNF, true
	case normalizeString(languageABAPStr):
		return LanguageABAP, true
	case normalizeString(languageAdaStr):
		return LanguageAda, true
	case normalizeString(languageADLStr):
		return LanguageADL, true
	case normalizeString(languageAdvPLStr):
		return LanguageAdvPL, true
	case normalizeString(languageActionScriptStr):
		return LanguageActionScript, true
	case normalizeString(languageActionScript3Str):
		return LanguageActionScript3, true
	case normalizeString(languageAgdaStr):
		return LanguageAgda, true
	case normalizeString(languageAheuiStr):
		return LanguageAheui, true
	case normalizeString(languageAlloyStr):
		return LanguageAlloy, true
	case normalizeString(languageAmbientTalkStr):
		return LanguageAmbientTalk, true
	case normalizeString(languageAmplStr):
		return LanguageAmpl, true
	case normalizeString(languageAngular2Str):
		return LanguageAngular2, true
	case normalizeString(languageAnsibleStr):
		return LanguageAnsible, true
	case normalizeString(languageANTLRStr):
		return LanguageANTLR, true
	case normalizeString(languageApacheConfigStr):
		return LanguageApacheConfig, true
	case normalizeString(languageApexStr):
		return LanguageApex, true
	case normalizeString(languageAPLStr):
		return LanguageAPL, true
	case normalizeString(languageAppleScriptStr):
		return LanguageAppleScript, true
	case normalizeString(languageArcStr):
		return LanguageArc, true
	case normalizeString(languageArduinoStr):
		return LanguageArduino, true
	case normalizeString(languageArrowStr):
		return LanguageArrow, true
	case normalizeString(languageASPClassicStr):
		return LanguageASPClassic, true
	case normalizeString(languageASPDotNetStr):
		return LanguageASPDotNet, true
	case normalizeString(languageAspectJStr):
		return LanguageAspectJ, true
	case normalizeString(languageAspxCSharpStr):
		return LanguageAspxCSharp, true
	case normalizeString(languageAspxVBNetStr):
		return LanguageAspxVBNet, true
	case normalizeString(languageAssemblyStr):
		return LanguageAssembly, true
	case normalizeString(languageAsymptoteStr):
		return LanguageAsymptote, true
	case normalizeString(languageAugeasStr):
		return LanguageAugeas, true
	case normalizeString(languageAutoconfStr):
		return LanguageAutoconf, true
	case normalizeString(languageAutoHotkeyStr):
		return LanguageAutoHotkey, true
	case normalizeString(languageAutoItStr):
		return LanguageAutoIt, true
	case normalizeString(languageAwkStr):
		return LanguageAwk, true
	case normalizeString(languageBallerinaStr):
		return LanguageBallerina, true
	case normalizeString(languageBAREStr):
		return LanguageBARE, true
	case normalizeString(languageBashStr):
		return LanguageBash, true
	case normalizeString(languageBashSessionStr):
		return LanguageBashSession, true
	case normalizeString(languageBasicStr):
		return LanguageBasic, true
	case normalizeString(languageBatchfileStr):
		return LanguageBatchfile, true
	case normalizeString(languageBatchScriptStr):
		return LanguageBatchScript, true
	case normalizeString(languageBBCBasicStr):
		return LanguageBBCBasic, true
	case normalizeString(languageBBCodeStr):
		return LanguageBBCode, true
	case normalizeString(languageBCStr):
		return LanguageBC, true
	case normalizeString(languageBefungeStr):
		return LanguageBefunge, true
	case normalizeString(languageBibTeXStr):
		return LanguageBibTeX, true
	case normalizeString(languageBladeTemplateStr):
		return LanguageBladeTemplate, true
	case normalizeString(languageBlazorStr):
		return LanguageBlazor, true
	case normalizeString(languageBlitzBasicStr):
		return LanguageBlitzBasic, true
	case normalizeString(languageBlitzMaxStr):
		return LanguageBlitzMax, true
	case normalizeString(languageBNFStr):
		return LanguageBNF, true
	case normalizeString(languageBoaStr):
		return LanguageBoa, true
	case normalizeString(languageBooStr):
		return LanguageBoo, true
	case normalizeString(languageBoogieStr):
		return LanguageBoogie, true
	case normalizeString(languageBrainfuckStr):
		return LanguageBrainfuck, true
	case normalizeString(languageBrightScriptStr):
		return LanguageBrightScript, true
	case normalizeString(languageBroStr):
		return LanguageBro, true
	case normalizeString(languageBSTStr):
		return LanguageBST, true
	case normalizeString(languageBUGSStr):
		return LanguageBUGS, true
	case normalizeString(languageCStr):
		return LanguageC, true
	case normalizeString(languageCa65AssemblerStr):
		return LanguageCa65Assembler, true
	case normalizeString(languageCaddyfileStr):
		return LanguageCaddyfile, true
	case normalizeString(languageCaddyfileDirectivesStr):
		return LanguageCaddyfileDirectives, true
	case normalizeString(languageCADLStr):
		return LanguageCADL, true
	case normalizeString(languageCAmkESStr):
		return LanguageCAmkES, true
	case normalizeString(languageCapDLStr):
		return LanguageCapDL, true
	case normalizeString(languageCapNProtoStr):
		return LanguageCapNProto, true
	case normalizeString(languageCassandraCQLStr):
		return LanguageCassandraCQL, true
	case normalizeString(languageCBMBasicV2Str):
		return LanguageCBMBasicV2, true
	case normalizeString(languageCeylonStr):
		return LanguageCeylon, true
	case normalizeString(languageCFEngine3Str):
		return LanguageCFEngine3, true
	case normalizeString(languageCfstatementStr):
		return LanguageCfstatement, true
	case normalizeString(languageChaiScriptStr):
		return LanguageChaiScript, true
	case normalizeString(languageChapelStr):
		return LanguageChapel, true
	case normalizeString(languageCharmciStr):
		return LanguageCharmci, true
	case normalizeString(languageCheetahStr):
		return LanguageCheetah, true
	case normalizeString(languageCirruStr):
		return LanguageCirru, true
	case normalizeString(languageClayStr):
		return LanguageClay, true
	case normalizeString(languageCleanStr):
		return LanguageClean, true
	case normalizeString(languageClojureStr):
		return LanguageClojure, true
	case normalizeString(languageClojureScriptStr):
		return LanguageClojureScript, true
	case normalizeString(languageCMakeStr):
		return LanguageCMake, true
	case normalizeString(languageCObjdumpStr):
		return LanguageCObjdump, true
	case normalizeString(languageCOBOLStr):
		return LanguageCOBOL, true
	case normalizeString(languageCOBOLFreeStr):
		return LanguageCOBOLFree, true
	case normalizeString(languageCocoaStr):
		return LanguageCocoa, true
	case normalizeString(languageCoffeeScriptStr):
		return LanguageCoffeeScript, true
	case normalizeString(languageColdfusionCFCStr):
		return LanguageColdfusionCFC, true
	case normalizeString(languageColdfusionHTMLStr):
		return LanguageColdfusionHTML, true
	case normalizeString(languageCommonLispStr):
		return LanguageCommonLisp, true
	case normalizeString(languageComponentPascalStr):
		return LanguageComponentPascal, true
	case normalizeString(languageCoqStr):
		return LanguageCoq, true
	case normalizeString(languageCPerlStr):
		return LanguageCPerl, true
	case normalizeString(languageCPPStr):
		return LanguageCPP, true
	case normalizeString(languageCppObjdumpStr):
		return LanguageCppObjdump, true
	case normalizeString(languageCPSAStr):
		return LanguageCPSA, true
	case normalizeString(languageCrmshStr):
		return LanguageCrmsh, true
	case normalizeString(languageCrocStr):
		return LanguageCroc, true
	case normalizeString(languageCrontabStr):
		return LanguageCrontab, true
	case normalizeString(languageCryptolStr):
		return LanguageCryptol, true
	case normalizeString(languageCrystalStr):
		return LanguageCrystal, true
	case normalizeString(languageCSharpStr):
		return LanguageCSharp, true
	case normalizeString(languageCSHTMLStr):
		return LanguageCSHTML, true
	case normalizeString(languageCSONStr):
		return LanguageCSON, true
	case normalizeString(languageCsoundDocumentStr):
		return LanguageCsoundDocument, true
	case normalizeString(languageCsoundOrchestraStr):
		return LanguageCsoundOrchestra, true
	case normalizeString(languageCsoundScoreStr):
		return LanguageCsoundScore, true
	case normalizeString(languageCSSStr):
		return LanguageCSS, true
	case normalizeString(languageCSVStr):
		return LanguageCSV, true
	case normalizeString(languageCUDAStr):
		return LanguageCUDA, true
	case normalizeString(languageCVSStr):
		return LanguageCVS, true
	case normalizeString(languageCypherStr):
		return LanguageCypher, true
	case normalizeString(languageCythonStr):
		return LanguageCython, true
	case normalizeString(languageDStr):
		return LanguageD, true
	case normalizeString(languageDarcsPatchStr):
		return LanguageDarcsPatch, true
	case normalizeString(languageDartStr):
		return LanguageDart, true
	case normalizeString(languageDASM16Str):
		return LanguageDASM16, true
	case normalizeString(languageDCLStr):
		return LanguageDCL, true
	case normalizeString(languageDCPU16AsmStr):
		return LanguageDCPU16Asm, true
	case normalizeString(languageDebianControlFileStr):
		return LanguageDebianControlFile, true
	case normalizeString(languageDelphiStr):
		return LanguageDelphi, true
	case normalizeString(languageDevicetreeStr):
		return LanguageDevicetree, true
	case normalizeString(languageDGStr):
		return LanguageDG, true
	case normalizeString(languageDhallStr):
		return LanguageDhall, true
	case normalizeString(languageDiffStr):
		return LanguageDiff, true
	case normalizeString(languageDjangoJinjaStr):
		return LanguageDjangoJinja, true
	case normalizeString(languageDObjdumpStr):
		return LanguageDObjdump, true
	case normalizeString(languageDockerStr):
		return LanguageDocker, true
	case normalizeString(languageDocTeXStr):
		return LanguageDocTeX, true
	case normalizeString(languageDTDStr):
		return LanguageDTD, true
	case normalizeString(languageDuelStr):
		return LanguageDuel, true
	case normalizeString(languageDylanStr):
		return LanguageDylan, true
	case normalizeString(languageDylanLIDStr):
		return LanguageDylanLID, true
	case normalizeString(languageDylanSessionStr):
		return LanguageDylanSession, true
	case normalizeString(languageDynASMStr):
		return LanguageDynASM, true
	case normalizeString(languageEMailStr):
		return LanguageEMail, true
	case normalizeString(languageEarlGreyStr):
		return LanguageEarlGrey, true
	case normalizeString(languageEasytrieveStr):
		return LanguageEasytrieve, true
	case normalizeString(languageEBNFStr):
		return LanguageEBNF, true
	case normalizeString(languageECStr):
		return LanguageEC, true
	case normalizeString(languageECLStr):
		return LanguageECL, true
	case normalizeString(languageEiffelStr):
		return LanguageEiffel, true
	case normalizeString(languageEJSStr):
		return LanguageEJS, true
	case normalizeString(languageElixirStr):
		return LanguageElixir, true
	case normalizeString(languageElixirIexSessionStr):
		return LanguageElixirIexSession, true
	case normalizeString(languageElmStr):
		return LanguageElm, true
	case normalizeString(languageEmacsLispStr):
		return LanguageEmacsLisp, true
	case normalizeString(languageERBStr):
		return LanguageERB, true
	case normalizeString(languageErlangStr):
		return LanguageErlang, true
	case normalizeString(languageErlangErlSessionStr):
		return LanguageErlangErlSession, true
	case normalizeString(languageEshellStr):
		return LanguageEshell, true
	case normalizeString(languageEvoqueStr):
		return LanguageEvoque, true
	case normalizeString(languageExeclineStr):
		return LanguageExecline, true
	case normalizeString(languageEzhilStr):
		return LanguageEzhil, true
	case normalizeString(languageFactorStr):
		return LanguageFactor, true
	case normalizeString(languageFancyStr):
		return LanguageFancy, true
	case normalizeString(languageFantomStr):
		return LanguageFantom, true
	case normalizeString(languageFelixStr):
		return LanguageFelix, true
	case normalizeString(languageFennelStr):
		return LanguageFennel, true
	case normalizeString(languageFlatlineStr):
		return LanguageFlatline, true
	case normalizeString(languageFloScriptStr):
		return LanguageFloScript, true
	case normalizeString(languageForthStr):
		return LanguageForth, true
	case normalizeString(languageFishStr):
		return LanguageFish, true
	case normalizeString(languageFontStr):
		return LanguageFont, true
	case normalizeString(languageFortranStr):
		return LanguageFortran, true
	case normalizeString(languageFortranFixedStr):
		return LanguageFortranFixed, true
	case normalizeString(languageFSharpStr):
		return LanguageFSharp, true
	case normalizeString(languageFoxProStr):
		return LanguageFoxPro, true
	case normalizeString(languageFreefemStr):
		return LanguageFreefem, true
	case normalizeString(languageFStarStr):
		return LanguageFStar, true
	case normalizeString(languageGapStr):
		return LanguageGap, true
	case normalizeString(languageGasStr):
		return LanguageGas, true
	case normalizeString(languageGDNativeStr):
		return LanguageGDNative, true
	case normalizeString(languageGDScriptStr):
		return LanguageGDScript, true
	case normalizeString(languageGenshiStr):
		return LanguageGenshi, true
	case normalizeString(languageGenshiHTMLStr):
		return LanguageGenshiHTML, true
	case normalizeString(languageGenshiTextStr):
		return LanguageGenshiText, true
	case normalizeString(languageGettextCatalogStr):
		return LanguageGettextCatalog, true
	case normalizeString(languageGherkinStr):
		return LanguageGherkin, true
	case normalizeString(languageGitStr):
		return LanguageGit, true
	case normalizeString(languageGitConfigStr):
		return LanguageGitConfig, true
	case normalizeString(languageGLSLStr):
		return LanguageGLSL, true
	case normalizeString(languageGnuplotStr):
		return LanguageGnuplot, true
	case normalizeString(languageGoStr):
		return LanguageGo, true
	case normalizeString(languageGoloStr):
		return LanguageGolo, true
	case normalizeString(languageGoodDataCLStr):
		return LanguageGoodDataCL, true
	case normalizeString(languageGosuStr):
		return LanguageGosu, true
	case normalizeString(languageGosuTemplateStr):
		return LanguageGosuTemplate, true
	case normalizeString(languageGradleConfigStr):
		return LanguageGradleConfig, true
	case normalizeString(languageGraphQLStr):
		return LanguageGraphQL, true
	case normalizeString(languageGroffStr):
		return LanguageGroff, true
	case normalizeString(languageGroovyStr):
		return LanguageGroovy, true
	case normalizeString(languageHamlStr):
		return LanguageHaml, true
	case normalizeString(languageHandlebarsStr):
		return LanguageHandlebars, true
	case normalizeString(languageHaskellStr):
		return LanguageHaskell, true
	case normalizeString(languageHaxeStr):
		return LanguageHaxe, true
	case normalizeString(languageHCLStr):
		return LanguageHCL, true
	case normalizeString(languageHexdumpStr):
		return LanguageHexdump, true
	case normalizeString(languageHLBStr):
		return LanguageHLB, true
	case normalizeString(languageHLSLStr):
		return LanguageHLSL, true
	case normalizeString(languageHSAILStr):
		return LanguageHSAIL, true
	case normalizeString(languageHspecStr):
		return LanguageHspec, true
	case normalizeString(languageHTMLStr):
		return LanguageHTML, true
	case normalizeString(languageHTTPStr):
		return LanguageHTTP, true
	case normalizeString(languageHxmlStr):
		return LanguageHxml, true
	case normalizeString(languageHyStr):
		return LanguageHy, true
	case normalizeString(languageHybrisStr):
		return LanguageHybris, true
	case normalizeString(languageIconStr):
		return LanguageIcon, true
	case normalizeString(languageIDLStr):
		return LanguageIDL, true
	case normalizeString(languageIdrisStr):
		return LanguageIdris, true
	case normalizeString(languageIgorStr):
		return LanguageIgor, true
	case normalizeString(languageImageJPEGStr):
		return LanguageImageJPEG, true
	case normalizeString(languageImagePNGStr):
		return LanguageImagePNG, true
	case normalizeString(languageInform6Str):
		return LanguageInform6, true
	case normalizeString(languageInform6TemplateStr):
		return LanguageInform6Template, true
	case normalizeString(languageInform7Str):
		return LanguageInform7, true
	case normalizeString(languageINIStr):
		return LanguageINI, true
	case normalizeString(languageIoStr):
		return LanguageIo, true
	case normalizeString(languageIokeStr):
		return LanguageIoke, true
	case normalizeString(languageIRCLogsStr):
		return LanguageIRCLogs, true
	case normalizeString(languageIsabelleStr):
		return LanguageIsabelle, true
	case normalizeString(languageJStr):
		return LanguageJ, true
	case normalizeString(languageJAGSStr):
		return LanguageJAGS, true
	case normalizeString(languageJadeStr):
		return LanguageJade, true
	case normalizeString(languageJasminStr):
		return LanguageJasmin, true
	case normalizeString(languageJavaStr):
		return LanguageJava, true
	case normalizeString(languageJavaScriptStr):
		return LanguageJavaScript, true
	case normalizeString(languageJCLStr):
		return LanguageJCL, true
	case normalizeString(languageJSGFStr):
		return LanguageJSGF, true
	case normalizeString(languageJSONStr):
		return LanguageJSON, true
	case normalizeString(languageJSONLDStr):
		return LanguageJSONLD, true
	case normalizeString(languageJSPStr):
		return LanguageJSP, true
	case normalizeString(languageJSXStr):
		return LanguageJSX, true
	case normalizeString(languageJuliaStr):
		return LanguageJulia, true
	case normalizeString(languageJuliaConsoleStr):
		return LanguageJuliaConsole, true
	case normalizeString(languageJungleStr):
		return LanguageJungle, true
	case normalizeString(languageJuttleStr):
		return LanguageJuttle, true
	case normalizeString(languageKalStr):
		return LanguageKal, true
	case normalizeString(languageKconfigStr):
		return LanguageKconfig, true
	case normalizeString(languageKernelLogStr):
		return LanguageKernelLog, true
	case normalizeString(languageKokaStr):
		return LanguageKoka, true
	case normalizeString(languageKotlinStr):
		return LanguageKotlin, true
	case normalizeString(languageLaravelTemplateStr):
		return LanguageLaravelTemplate, true
	case normalizeString(languageLassoStr):
		return LanguageLasso, true
	case normalizeString(languageLaTeXStr):
		return LanguageLaTeX, true
	case normalizeString(languageLatteStr):
		return LanguageLatte, true
	case normalizeString(languageLeanStr):
		return LanguageLean, true
	case normalizeString(languageLessStr):
		return LanguageLess, true
	case normalizeString(languageLighttpdStr):
		return LanguageLighttpd, true
	case normalizeString(languageLimboStr):
		return LanguageLimbo, true
	case normalizeString(languageLinkerScriptStr):
		return LanguageLinkerScript, true
	case normalizeString(languageLiquidStr):
		return LanguageLiquid, true
	case normalizeString(languageLiterateAgdaStr):
		return LanguageLiterateAgda, true
	case normalizeString(languageLiterateCryptolStr):
		return LanguageLiterateCryptol, true
	case normalizeString(languageLiterateHaskellStr):
		return LanguageLiterateHaskell, true
	case normalizeString(languageLiterateIdrisStr):
		return LanguageLiterateIdris, true
	case normalizeString(languageLiveScriptStr):
		return LanguageLiveScript, true
	case normalizeString(languageLLVMStr):
		return LanguageLLVM, true
	case normalizeString(languageLLVMMIRStr):
		return LanguageLLVMMIR, true
	case normalizeString(languageLLVMMIRBodyStr):
		return LanguageLLVMMIRBody, true
	case normalizeString(languageLogosStr):
		return LanguageLogos, true
	case normalizeString(languageLogFileStr):
		return LanguageLogFile, true
	case normalizeString(languageLogtalkStr):
		return LanguageLogtalk, true
	case normalizeString(languageLSLStr):
		return LanguageLSL, true
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
	case normalizeString(languageApacheConfigChromaStr):
		return LanguageApacheConfig, true
	case normalizeString(languageAssemblyChromaStr):
		return LanguageAssembly, true
	case normalizeString(languageColdfusionHTMLChromaStr):
		return LanguageColdfusionHTML, true
	case normalizeString(languageEmacsLispChromaStr):
		return LanguageEmacsLisp, true
	case normalizeString(languageGoHTMLTemplateChromaStr):
		return LanguageGo, true
	case normalizeString(languageGoTextTemplateChromaStr):
		return LanguageGo, true
	case normalizeString(languageGosuTemplateChromaStr):
		return LanguageGosu, true
	case normalizeString(languageFSharpChromaStr):
		return LanguageFSharp, true
	case normalizeString(languageJSXChromaStr):
		return LanguageJSX, true
	case normalizeString(languageLessChromaStr):
		return LanguageLess, true
	case normalizeString(languageLiquidChromaStr):
		return LanguageLiquid, true
	case normalizeString(languageMakefileChromaStr):
		return LanguageMakefile, true
	case normalizeString(languageMarkdownChromaStr):
		return LanguageMarkdown, true
	case normalizeString(languageTextChromaStr):
		return LanguageText, true
	case normalizeString(languageVHDLChromaStr):
		return LanguageVHDL, true
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
	case LanguageABAP:
		return languageABAPStr
	case LanguageABNF:
		return languageABNFStr
	case LanguageAda:
		return languageAdaStr
	case LanguageADL:
		return languageADLStr
	case LanguageAdvPL:
		return languageAdvPLStr
	case LanguageActionScript:
		return languageActionScriptStr
	case LanguageActionScript3:
		return languageActionScript3Str
	case LanguageAgda:
		return languageAgdaStr
	case LanguageAheui:
		return languageAheuiStr
	case LanguageAlloy:
		return languageAlloyStr
	case LanguageAmbientTalk:
		return languageAmbientTalkStr
	case LanguageAmpl:
		return languageAmplStr
	case LanguageAngular2:
		return languageAngular2Str
	case LanguageAnsible:
		return languageAnsibleStr
	case LanguageANTLR:
		return languageANTLRStr
	case LanguageApacheConfig:
		return languageApacheConfigStr
	case LanguageApex:
		return languageApexStr
	case LanguageAPL:
		return languageAPLStr
	case LanguageAppleScript:
		return languageAppleScriptStr
	case LanguageArc:
		return languageArcStr
	case LanguageArduino:
		return languageArduinoStr
	case LanguageArrow:
		return languageArrowStr
	case LanguageASPClassic:
		return languageASPClassicStr
	case LanguageASPDotNet:
		return languageASPDotNetStr
	case LanguageAspectJ:
		return languageAspectJStr
	case LanguageAspxCSharp:
		return languageAspxCSharpStr
	case LanguageAspxVBNet:
		return languageAspxVBNetStr
	case LanguageAssembly:
		return languageAssemblyStr
	case LanguageAsymptote:
		return languageAsymptoteStr
	case LanguageAugeas:
		return languageAugeasStr
	case LanguageAutoconf:
		return languageAutoconfStr
	case LanguageAutoHotkey:
		return languageAutoHotkeyStr
	case LanguageAutoIt:
		return languageAutoItStr
	case LanguageAwk:
		return languageAwkStr
	case LanguageBallerina:
		return languageBallerinaStr
	case LanguageBARE:
		return languageBAREStr
	case LanguageBash:
		return languageBashStr
	case LanguageBashSession:
		return languageBashSessionStr
	case LanguageBasic:
		return languageBasicStr
	case LanguageBatchfile:
		return languageBatchfileStr
	case LanguageBatchScript:
		return languageBatchScriptStr
	case LanguageBBCBasic:
		return languageBBCBasicStr
	case LanguageBBCode:
		return languageBBCodeStr
	case LanguageBC:
		return languageBCStr
	case LanguageBefunge:
		return languageBefungeStr
	case LanguageBibTeX:
		return languageBibTeXStr
	case LanguageBladeTemplate:
		return languageBladeTemplateStr
	case LanguageBlazor:
		return languageBlazorStr
	case LanguageBlitzBasic:
		return languageBlitzBasicStr
	case LanguageBlitzMax:
		return languageBlitzMaxStr
	case LanguageBNF:
		return languageBNFStr
	case LanguageBoa:
		return languageBoaStr
	case LanguageBoo:
		return languageBooStr
	case LanguageBoogie:
		return languageBoogieStr
	case LanguageBrainfuck:
		return languageBrainfuckStr
	case LanguageBrightScript:
		return languageBrightScriptStr
	case LanguageBro:
		return languageBroStr
	case LanguageBST:
		return languageBSTStr
	case LanguageBUGS:
		return languageBUGSStr
	case LanguageC:
		return languageCStr
	case LanguageCa65Assembler:
		return languageCa65AssemblerStr
	case LanguageCaddyfile:
		return languageCaddyfileStr
	case LanguageCaddyfileDirectives:
		return languageCaddyfileDirectivesStr
	case LanguageCADL:
		return languageCADLStr
	case LanguageCAmkES:
		return languageCAmkESStr
	case LanguageCapDL:
		return languageCapDLStr
	case LanguageCapNProto:
		return languageCapNProtoStr
	case LanguageCassandraCQL:
		return languageCassandraCQLStr
	case LanguageCBMBasicV2:
		return languageCBMBasicV2Str
	case LanguageCeylon:
		return languageCeylonStr
	case LanguageCFEngine3:
		return languageCFEngine3Str
	case LanguageCfstatement:
		return languageCfstatementStr
	case LanguageChaiScript:
		return languageChaiScriptStr
	case LanguageChapel:
		return languageChapelStr
	case LanguageCharmci:
		return languageCharmciStr
	case LanguageCheetah:
		return languageCheetahStr
	case LanguageCirru:
		return languageCirruStr
	case LanguageClay:
		return languageClayStr
	case LanguageClean:
		return languageCleanStr
	case LanguageClojure:
		return languageClojureStr
	case LanguageClojureScript:
		return languageClojureScriptStr
	case LanguageCMake:
		return languageCMakeStr
	case LanguageCObjdump:
		return languageCObjdumpStr
	case LanguageCOBOL:
		return languageCOBOLStr
	case LanguageCOBOLFree:
		return languageCOBOLFreeStr
	case LanguageCocoa:
		return languageCocoaStr
	case LanguageCoffeeScript:
		return languageCoffeeScriptStr
	case LanguageColdfusionCFC:
		return languageColdfusionCFCStr
	case LanguageColdfusionHTML:
		return languageColdfusionHTMLStr
	case LanguageCommonLisp:
		return languageCommonLispStr
	case LanguageComponentPascal:
		return languageComponentPascalStr
	case LanguageCoq:
		return languageCoqStr
	case LanguageCPerl:
		return languageCPerlStr
	case LanguageCPP:
		return languageCPPStr
	case LanguageCppObjdump:
		return languageCppObjdumpStr
	case LanguageCPSA:
		return languageCPSAStr
	case LanguageCrmsh:
		return languageCrmshStr
	case LanguageCroc:
		return languageCrocStr
	case LanguageCrontab:
		return languageCrontabStr
	case LanguageCryptol:
		return languageCryptolStr
	case LanguageCrystal:
		return languageCrystalStr
	case LanguageCSharp:
		return languageCSharpStr
	case LanguageCSHTML:
		return languageCSHTMLStr
	case LanguageCSON:
		return languageCSONStr
	case LanguageCsoundDocument:
		return languageCsoundDocumentStr
	case LanguageCsoundOrchestra:
		return languageCsoundOrchestraStr
	case LanguageCsoundScore:
		return languageCsoundScoreStr
	case LanguageCSS:
		return languageCSSStr
	case LanguageCSV:
		return languageCSVStr
	case LanguageCUDA:
		return languageCUDAStr
	case LanguageCVS:
		return languageCVSStr
	case LanguageCypher:
		return languageCypherStr
	case LanguageCython:
		return languageCythonStr
	case LanguageD:
		return languageDStr
	case LanguageDarcsPatch:
		return languageDarcsPatchStr
	case LanguageDart:
		return languageDartStr
	case LanguageDASM16:
		return languageDASM16Str
	case LanguageDCL:
		return languageDCLStr
	case LanguageDCPU16Asm:
		return languageDCPU16AsmStr
	case LanguageDebianControlFile:
		return languageDebianControlFileStr
	case LanguageDelphi:
		return languageDelphiStr
	case LanguageDevicetree:
		return languageDevicetreeStr
	case LanguageDG:
		return languageDGStr
	case LanguageDhall:
		return languageDhallStr
	case LanguageDiff:
		return languageDiffStr
	case LanguageDjangoJinja:
		return languageDjangoJinjaStr
	case LanguageDObjdump:
		return languageDObjdumpStr
	case LanguageDocker:
		return languageDockerStr
	case LanguageDocTeX:
		return languageDocTeXStr
	case LanguageDTD:
		return languageDTDStr
	case LanguageDuel:
		return languageDuelStr
	case LanguageDylan:
		return languageDylanStr
	case LanguageDylanLID:
		return languageDylanLIDStr
	case LanguageDylanSession:
		return languageDylanSessionStr
	case LanguageDynASM:
		return languageDynASMStr
	case LanguageEarlGrey:
		return languageEarlGreyStr
	case LanguageEasytrieve:
		return languageEasytrieveStr
	case LanguageEBNF:
		return languageEBNFStr
	case LanguageEC:
		return languageECStr
	case LanguageECL:
		return languageECLStr
	case LanguageEiffel:
		return languageEiffelStr
	case LanguageEJS:
		return languageEJSStr
	case LanguageElixir:
		return languageElixirStr
	case LanguageElixirIexSession:
		return languageElixirIexSessionStr
	case LanguageElm:
		return languageElmStr
	case LanguageEmacsLisp:
		return languageEmacsLispStr
	case LanguageEMail:
		return languageEMailStr
	case LanguageERB:
		return languageERBStr
	case LanguageErlang:
		return languageErlangStr
	case LanguageErlangErlSession:
		return languageErlangErlSessionStr
	case LanguageEshell:
		return languageEshellStr
	case LanguageEvoque:
		return languageEvoqueStr
	case LanguageExecline:
		return languageExeclineStr
	case LanguageEzhil:
		return languageEzhilStr
	case LanguageFactor:
		return languageFactorStr
	case LanguageFancy:
		return languageFancyStr
	case LanguageFantom:
		return languageFantomStr
	case LanguageFelix:
		return languageFelixStr
	case LanguageFennel:
		return languageFennelStr
	case LanguageFlatline:
		return languageFlatlineStr
	case LanguageFloScript:
		return languageFloScriptStr
	case LanguageForth:
		return languageForthStr
	case LanguageFish:
		return languageFishStr
	case LanguageFont:
		return languageFontStr
	case LanguageFortran:
		return languageFortranStr
	case LanguageFortranFixed:
		return languageFortranFixedStr
	case LanguageFSharp:
		return languageFSharpStr
	case LanguageFoxPro:
		return languageFoxProStr
	case LanguageFreefem:
		return languageFreefemStr
	case LanguageFStar:
		return languageFStarStr
	case LanguageGap:
		return languageGapStr
	case LanguageGas:
		return languageGasStr
	case LanguageGDScript:
		return languageGDScriptStr
	case LanguageGenshi:
		return languageGenshiStr
	case LanguageGenshiHTML:
		return languageGenshiHTMLStr
	case LanguageGenshiText:
		return languageGenshiTextStr
	case LanguageGettextCatalog:
		return languageGettextCatalogStr
	case LanguageGherkin:
		return languageGherkinStr
	case LanguageGDNative:
		return languageGDNativeStr
	case LanguageGit:
		return languageGitStr
	case LanguageGitConfig:
		return languageGitConfigStr
	case LanguageGLSL:
		return languageGLSLStr
	case LanguageGnuplot:
		return languageGnuplotStr
	case LanguageGo, LanguageGoHTMLTemplate, LanguageGoTextTemplate:
		return languageGoStr
	case LanguageGolo:
		return languageGoloStr
	case LanguageGoodDataCL:
		return languageGoodDataCLStr
	case LanguageGosu:
		return languageGosuStr
	case LanguageGosuTemplate:
		return languageGosuTemplateStr
	case LanguageGradleConfig:
		return languageGradleConfigStr
	case LanguageGraphQL:
		return languageGraphQLStr
	case LanguageGroff:
		return languageGroffStr
	case LanguageGroovy:
		return languageGroovyStr
	case LanguageHaml:
		return languageHamlStr
	case LanguageHandlebars:
		return languageHandlebarsStr
	case LanguageHaskell:
		return languageHaskellStr
	case LanguageHaxe:
		return languageHaxeStr
	case LanguageHCL:
		return languageHCLStr
	case LanguageHexdump:
		return languageHexdumpStr
	case LanguageHLB:
		return languageHLBStr
	case LanguageHLSL:
		return languageHLSLStr
	case LanguageHSAIL:
		return languageHSAILStr
	case LanguageHspec:
		return languageHspecStr
	case LanguageHTML:
		return languageHTMLStr
	case LanguageHTTP:
		return languageHTTPStr
	case LanguageHxml:
		return languageHxmlStr
	case LanguageHy:
		return languageHyStr
	case LanguageHybris:
		return languageHybrisStr
	case LanguageIcon:
		return languageIconStr
	case LanguageIDL:
		return languageIDLStr
	case LanguageIdris:
		return languageIdrisStr
	case LanguageIgor:
		return languageIgorStr
	case LanguageImageJPEG:
		return languageImageJPEGStr
	case LanguageImagePNG:
		return languageImagePNGStr
	case LanguageInform6:
		return languageInform6Str
	case LanguageInform6Template:
		return languageInform6TemplateStr
	case LanguageInform7:
		return languageInform7Str
	case LanguageINI:
		return languageINIStr
	case LanguageIo:
		return languageIoStr
	case LanguageIoke:
		return languageIokeStr
	case LanguageIRCLogs:
		return languageIRCLogsStr
	case LanguageIsabelle:
		return languageIsabelleStr
	case LanguageJ:
		return languageJStr
	case LanguageJAGS:
		return languageJAGSStr
	case LanguageJade:
		return languageJadeStr
	case LanguageJasmin:
		return languageJasminStr
	case LanguageJava:
		return languageJavaStr
	case LanguageJavaScript:
		return languageJavaScriptStr
	case LanguageJCL:
		return languageJCLStr
	case LanguageJSGF:
		return languageJSGFStr
	case LanguageJSON:
		return languageJSONStr
	case LanguageJSONLD:
		return languageJSONLDStr
	case LanguageJSP:
		return languageJSPStr
	case LanguageJSX:
		return languageJSXStr
	case LanguageJulia:
		return languageJuliaStr
	case LanguageJuliaConsole:
		return languageJuliaConsoleStr
	case LanguageJungle:
		return languageJungleStr
	case LanguageJuttle:
		return languageJuttleStr
	case LanguageKal:
		return languageKalStr
	case LanguageKconfig:
		return languageKconfigStr
	case LanguageKernelLog:
		return languageKernelLogStr
	case LanguageKoka:
		return languageKokaStr
	case LanguageKotlin:
		return languageKotlinStr
	case LanguageLaravelTemplate:
		return languageLaravelTemplateStr
	case LanguageLasso:
		return languageLassoStr
	case LanguageLaTeX:
		return languageLaTeXStr
	case LanguageLatte:
		return languageLatteStr
	case LanguageLean:
		return languageLeanStr
	case LanguageLess:
		return languageLessStr
	case LanguageLighttpd:
		return languageLighttpdStr
	case LanguageLimbo:
		return languageLimboStr
	case LanguageLinkerScript:
		return languageLinkerScriptStr
	case LanguageLiquid:
		return languageLiquidStr
	case LanguageLiterateAgda:
		return languageLiterateAgdaStr
	case LanguageLiterateCryptol:
		return languageLiterateCryptolStr
	case LanguageLiterateHaskell:
		return languageLiterateHaskellStr
	case LanguageLiterateIdris:
		return languageLiterateIdrisStr
	case LanguageLiveScript:
		return languageLiveScriptStr
	case LanguageLLVM:
		return languageLLVMStr
	case LanguageLLVMMIR:
		return languageLLVMMIRStr
	case LanguageLLVMMIRBody:
		return languageLLVMMIRBodyStr
	case LanguageLogos:
		return languageLogosStr
	case LanguageLogFile:
		return languageLogFileStr
	case LanguageLogtalk:
		return languageLogtalkStr
	case LanguageLSL:
		return languageLSLStr
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
	case LanguageVHDL:
		return languageVHDLChromaStr
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
