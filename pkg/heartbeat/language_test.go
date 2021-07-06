package heartbeat_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/lexers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func languageTests() map[string]heartbeat.Language {
	return map[string]heartbeat.Language{
		"1C Enterprise":                    heartbeat.Language1CEnterprise,
		"4D":                               heartbeat.Language4D,
		"ABAP":                             heartbeat.LanguageABAP,
		"ABNF":                             heartbeat.LanguageABNF,
		"ActionScript":                     heartbeat.LanguageActionScript,
		"ActionScript 3":                   heartbeat.LanguageActionScript3,
		"Ada":                              heartbeat.LanguageAda,
		"ADL":                              heartbeat.LanguageADL,
		"Adobe Font Metrics":               heartbeat.LanguageAdobeFontMetrics,
		"AdvPL":                            heartbeat.LanguageAdvPL,
		"Agda":                             heartbeat.LanguageAgda,
		"AGS Script":                       heartbeat.LanguageAGSScript,
		"Aheui":                            heartbeat.LanguageAheui,
		"AL":                               heartbeat.LanguageAL,
		"Alloy":                            heartbeat.LanguageAlloy,
		"Alpine Abuild":                    heartbeat.LanguageAlpineAbuild,
		"Altium Designer":                  heartbeat.LanguageAltiumDesigner,
		"AmbientTalk":                      heartbeat.LanguageAmbientTalk,
		"AMPL":                             heartbeat.LanguageAMPL,
		"AngelScript":                      heartbeat.LanguageAngelScript,
		"Angular2":                         heartbeat.LanguageAngular2,
		"Ansible":                          heartbeat.LanguageAnsible,
		"Ant Build System":                 heartbeat.LanguageAntBuildSystem,
		"ANTLR":                            heartbeat.LanguageANTLR,
		"APL":                              heartbeat.LanguageAPL,
		"AppleScript":                      heartbeat.LanguageAppleScript,
		"Apache Config":                    heartbeat.LanguageApacheConfig,
		"Apex":                             heartbeat.LanguageApex,
		"API Blueprint":                    heartbeat.LanguageAPIBlueprint,
		"Apollo Guidance Computer":         heartbeat.LanguageApolloGuidanceComputer,
		"Arc":                              heartbeat.LanguageArc,
		"Arduino":                          heartbeat.LanguageArduino,
		"Arrow":                            heartbeat.LanguageArrow,
		"AsciiDoc":                         heartbeat.LanguageASCIIDoc,
		"ASL":                              heartbeat.LanguageASL,
		"ASN.1":                            heartbeat.LanguageASN1,
		"ASP.NET":                          heartbeat.LanguageASPDotNet,
		"AspectJ":                          heartbeat.LanguageAspectJ,
		"aspx-cs":                          heartbeat.LanguageAspxCSharp,
		"aspx-vb":                          heartbeat.LanguageAspxVBNet,
		"Assembly":                         heartbeat.LanguageAssembly,
		"Asymptote":                        heartbeat.LanguageAsymptote,
		"ATS":                              heartbeat.LanguageATS,
		"Augeas":                           heartbeat.LanguageAugeas,
		"Autoconf":                         heartbeat.LanguageAutoconf,
		"AutoHotkey":                       heartbeat.LanguageAutoHotkey,
		"AutoIt":                           heartbeat.LanguageAutoIt,
		"Avro IDL":                         heartbeat.LanguageAvroIDL,
		"Awk":                              heartbeat.LanguageAwk,
		"Ballerina":                        heartbeat.LanguageBallerina,
		"BARE":                             heartbeat.LanguageBARE,
		"Bash":                             heartbeat.LanguageBash,
		"Bash Session":                     heartbeat.LanguageBashSession,
		"Batchfile":                        heartbeat.LanguageBatchfile,
		"Basic":                            heartbeat.LanguageBasic,
		"BBC Basic":                        heartbeat.LanguageBBCBasic,
		"BBCode":                           heartbeat.LanguageBBCode,
		"BC":                               heartbeat.LanguageBC,
		"Beef":                             heartbeat.LanguageBeef,
		"Befunge":                          heartbeat.LanguageBefunge,
		"BibTeX":                           heartbeat.LanguageBibTeX,
		"Bison":                            heartbeat.LanguageBison,
		"BitBake":                          heartbeat.LanguageBitBake,
		"Blade":                            heartbeat.LanguageBlade,
		"Blade Template":                   heartbeat.LanguageBladeTemplate,
		"Blazor":                           heartbeat.LanguageBlazor,
		"BlitzBasic":                       heartbeat.LanguageBlitzBasic,
		"BlitzMax":                         heartbeat.LanguageBlitzMax,
		"Bluespec":                         heartbeat.LanguageBluespec,
		"BNF":                              heartbeat.LanguageBNF,
		"Boa":                              heartbeat.LanguageBoa,
		"Boo":                              heartbeat.LanguageBoo,
		"Boogie":                           heartbeat.LanguageBoogie,
		"Brainfuck":                        heartbeat.LanguageBrainfuck,
		"BrightScript":                     heartbeat.LanguageBrightScript,
		"Browserslist":                     heartbeat.LanguageBrowserslist,
		"Bro":                              heartbeat.LanguageBro,
		"BST":                              heartbeat.LanguageBST,
		"BUGS":                             heartbeat.LanguageBUGS,
		"C":                                heartbeat.LanguageC,
		"C++":                              heartbeat.LanguageCPP,
		"C#":                               heartbeat.LanguageCSharp,
		"C2hs Haskell":                     heartbeat.LanguageC2hsHaskell,
		"ca65 assembler":                   heartbeat.LanguageCa65Assembler,
		"Cabal Config":                     heartbeat.LanguageCabalConfig,
		"Caddyfile":                        heartbeat.LanguageCaddyfile,
		"Caddyfile Directives":             heartbeat.LanguageCaddyfileDirectives,
		"cADL":                             heartbeat.LanguageCADL,
		"CAmkES":                           heartbeat.LanguageCAmkES,
		"CapDL":                            heartbeat.LanguageCapDL,
		"Cap'n Proto":                      heartbeat.LanguageCapNProto,
		"CartoCSS":                         heartbeat.LanguageCartoCSS,
		"Cassandra CQL":                    heartbeat.LanguageCassandraCQL,
		"CBM BASIC V2":                     heartbeat.LanguageCBMBasicV2,
		"Ceylon":                           heartbeat.LanguageCeylon,
		"CFEngine3":                        heartbeat.LanguageCFEngine3,
		"cfstatement":                      heartbeat.LanguageCfstatement,
		"ChaiScript":                       heartbeat.LanguageChaiScript,
		"Chapel":                           heartbeat.LanguageChapel,
		"Charity":                          heartbeat.LanguageCharity,
		"Charmci":                          heartbeat.LanguageCharmci,
		"Cheetah":                          heartbeat.LanguageCheetah,
		"ChucK":                            heartbeat.LanguageChucK,
		"Cirru":                            heartbeat.LanguageCirru,
		"Clarion":                          heartbeat.LanguageClarion,
		"Classic ASP":                      heartbeat.LanguageClassicASP,
		"Clay":                             heartbeat.LanguageClay,
		"Clean":                            heartbeat.LanguageClean,
		"Click":                            heartbeat.LanguageClick,
		"CLIPS":                            heartbeat.LanguageCLIPS,
		"Clojure":                          heartbeat.LanguageClojure,
		"ClojureScript":                    heartbeat.LanguageClojureScript,
		"Closure Templates":                heartbeat.LanguageClosureTemplates,
		"Cloud Firestore Security Rules":   heartbeat.LanguageCloudFirestoreSecurityRules,
		"C-ObjDump":                        heartbeat.LanguageCObjdump,
		"CMake":                            heartbeat.LanguageCMake,
		"COBOL":                            heartbeat.LanguageCOBOL,
		"COBOLFree":                        heartbeat.LanguageCOBOLFree,
		"Cocoa":                            heartbeat.LanguageCocoa,
		"CodeQL":                           heartbeat.LanguageCodeQL,
		"CoffeeScript":                     heartbeat.LanguageCoffeeScript,
		"ColdFusion":                       heartbeat.LanguageColdfusionHTML,
		"ColdFusion CFC":                   heartbeat.LanguageColdfusionCFC,
		"COLLADA":                          heartbeat.LanguageCOLLADA,
		"Common Lisp":                      heartbeat.LanguageCommonLisp,
		"Common Workflow Language":         heartbeat.LanguageCommonWorkflowLanguage,
		"Component Pascal":                 heartbeat.LanguageComponentPascal,
		"Config":                           heartbeat.LanguageConfig,
		"CoNLL-U":                          heartbeat.LanguageCoNLLU,
		"Cool":                             heartbeat.LanguageCool,
		"Coq":                              heartbeat.LanguageCoq,
		"cperl":                            heartbeat.LanguageCPerl,
		"Cpp-ObjDump":                      heartbeat.LanguageCppObjdump,
		"CPSA":                             heartbeat.LanguageCPSA,
		"Creole":                           heartbeat.LanguageCreole,
		"Crmsh":                            heartbeat.LanguageCrmsh,
		"Croc":                             heartbeat.LanguageCroc,
		"Crontab":                          heartbeat.LanguageCrontab,
		"Cryptol":                          heartbeat.LanguageCryptol,
		"Crystal":                          heartbeat.LanguageCrystal,
		"CSHTML":                           heartbeat.LanguageCSHTML,
		"CSON":                             heartbeat.LanguageCSON,
		"Csound":                           heartbeat.LanguageCsound,
		"Csound Document":                  heartbeat.LanguageCsoundDocument,
		"Csound Orchestra":                 heartbeat.LanguageCsoundOrchestra,
		"Csound Score":                     heartbeat.LanguageCsoundScore,
		"CSS":                              heartbeat.LanguageCSS,
		"CSV":                              heartbeat.LanguageCSV,
		"Cuda":                             heartbeat.LanguageCUDA,
		"cURL Config":                      heartbeat.LanguagecURLConfig,
		"CVS":                              heartbeat.LanguageCVS,
		"CWeb":                             heartbeat.LanguageCWeb,
		"Cycript":                          heartbeat.LanguageCycript,
		"Cypher":                           heartbeat.LanguageCypher,
		"Cython":                           heartbeat.LanguageCython,
		"D":                                heartbeat.LanguageD,
		"Dafny":                            heartbeat.LanguageDafny,
		"d-objdump":                        heartbeat.LanguageDObjdump,
		"Darcs Patch":                      heartbeat.LanguageDarcsPatch,
		"Dart":                             heartbeat.LanguageDart,
		"DataWeave":                        heartbeat.LanguageDataWeave,
		"DASM16":                           heartbeat.LanguageDASM16,
		"DCL":                              heartbeat.LanguageDCL,
		"DCPU-16 ASM":                      heartbeat.LanguageDCPU16Asm,
		"Debian Control file":              heartbeat.LanguageDebianControlFile,
		"Debian Sourcelist":                heartbeat.LanguageSourcesList,
		"Delphi":                           heartbeat.LanguageDelphi,
		"desktop":                          heartbeat.LanguageDesktop,
		"Devicetree":                       heartbeat.LanguageDevicetree,
		"dg":                               heartbeat.LanguageDG,
		"Dhall":                            heartbeat.LanguageDhall,
		"Diff":                             heartbeat.LanguageDiff,
		"DIGITAL Command Language":         heartbeat.LanguageDigitalCommand,
		"dircolors":                        heartbeat.LanguageDircolors,
		"DirectX 3D File":                  heartbeat.LanguageDirectX3DFile,
		"Django/Jinja":                     heartbeat.LanguageDjangoJinja,
		"DM":                               heartbeat.LanguageDM,
		"DNS Zone":                         heartbeat.LanguageDNSZone,
		"Docker":                           heartbeat.LanguageDocker,
		"Dockerfile":                       heartbeat.LanguageDockerfile,
		"DocTeX":                           heartbeat.LanguageDocTeX,
		"Documentation":                    heartbeat.LanguageDocumentation,
		"Dogescript":                       heartbeat.LanguageDogescript,
		"DTD":                              heartbeat.LanguageDTD,
		"DTrace":                           heartbeat.LanguageDTrace,
		"Duel":                             heartbeat.LanguageDuel,
		"Dylan":                            heartbeat.LanguageDylan,
		"DylanLID":                         heartbeat.LanguageDylanLID,
		"Dylan session":                    heartbeat.LanguageDylanSession,
		"DynASM":                           heartbeat.LanguageDynASM,
		"E":                                heartbeat.LanguageE,
		"Eagle":                            heartbeat.LanguageEagle,
		"Earl Grey":                        heartbeat.LanguageEarlGrey,
		"Easybuild":                        heartbeat.LanguageEasybuild,
		"Easytrieve":                       heartbeat.LanguageEasytrieve,
		"EBNF":                             heartbeat.LanguageEBNF,
		"eC":                               heartbeat.LanguageEC,
		"Ecere Projects":                   heartbeat.LanguageEcereProjects,
		"ECL":                              heartbeat.LanguageECL,
		"ECLiPSe":                          heartbeat.LanguageEclipse,
		"EditorConfig":                     heartbeat.LanguageEditorConfig,
		"Edje Data Collection":             heartbeat.LanguageEdjeDataCollection,
		"edn":                              heartbeat.LanguageEdn,
		"Eiffel":                           heartbeat.LanguageEiffel,
		"EJS":                              heartbeat.LanguageEJS,
		"Elixir":                           heartbeat.LanguageElixir,
		"Elixir iex session":               heartbeat.LanguageElixirIexSession,
		"Elm":                              heartbeat.LanguageElm,
		"Emacs Lisp":                       heartbeat.LanguageEmacsLisp,
		"E-mail":                           heartbeat.LanguageEMail,
		"EmberScript":                      heartbeat.LanguageEmberScript,
		"EML":                              heartbeat.LanguageEML,
		"EQ":                               heartbeat.LanguageEQ,
		"ERB":                              heartbeat.LanguageERB,
		"Erlang":                           heartbeat.LanguageErlang,
		"Erlang erl session":               heartbeat.LanguageErlangErlSession,
		"Eshell":                           heartbeat.LanguageEshell,
		"Evoque":                           heartbeat.LanguageEvoque,
		"execline":                         heartbeat.LanguageExecline,
		"Ezhil":                            heartbeat.LanguageEzhil,
		"F*":                               heartbeat.LanguageFStar,
		"F#":                               heartbeat.LanguageFSharp,
		"Factor":                           heartbeat.LanguageFactor,
		"Fancy":                            heartbeat.LanguageFancy,
		"Fantom":                           heartbeat.LanguageFantom,
		"Faust":                            heartbeat.LanguageFaust,
		"Felix":                            heartbeat.LanguageFelix,
		"Fennel":                           heartbeat.LanguageFennel,
		"FIGlet Font":                      heartbeat.LanguageFIGletFont,
		"Filebench WML":                    heartbeat.LanguageFilebenchWML,
		"Filterscript":                     heartbeat.LanguageFilterscript,
		"Fish":                             heartbeat.LanguageFish,
		"FLUX":                             heartbeat.LanguageFLUX,
		"Flatline":                         heartbeat.LanguageFlatline,
		"FloScript":                        heartbeat.LanguageFloScript,
		"Font":                             heartbeat.LanguageFont,
		"Formatted":                        heartbeat.LanguageFormatted,
		"Forth":                            heartbeat.LanguageForth,
		"Fortran":                          heartbeat.LanguageFortran,
		"FortranFixed":                     heartbeat.LanguageFortranFixed,
		"Fortran Free Form":                heartbeat.LanguageFortranFreeForm,
		"FoxPro":                           heartbeat.LanguageFoxPro,
		"Freefem":                          heartbeat.LanguageFreefem,
		"FreeMarker":                       heartbeat.LangauageFreeMarker,
		"Frege":                            heartbeat.LangauageFrege,
		"Futhark":                          heartbeat.LanguageFuthark,
		"Game Maker Language":              heartbeat.LanguageGameMakerLanguage,
		"GAML":                             heartbeat.LanguageGAML,
		"GAMS":                             heartbeat.LanguageGAMS,
		"GAP":                              heartbeat.LanguageGap,
		"GAS":                              heartbeat.LanguageGas,
		"GCC Machine Description":          heartbeat.LanguageGCCMachineDescription,
		"G-code":                           heartbeat.LanguageGCode,
		"GDB":                              heartbeat.LanguageGDB,
		"GDNative":                         heartbeat.LanguageGDNative,
		"GDScript":                         heartbeat.LanguageGDScript,
		"GEDCOM":                           heartbeat.LanguageGEDCOM,
		"Genie":                            heartbeat.LanguageGenie,
		"Genshi":                           heartbeat.LanguageGenshi,
		"Genshi HTML":                      heartbeat.LanguageGenshiHTML,
		"Genshi Text":                      heartbeat.LanguageGenshiText,
		"Gentoo Ebuild":                    heartbeat.LanguageGentooEbuild,
		"Gentoo Eclass":                    heartbeat.LanguageGentooEclass,
		"Gerber Image":                     heartbeat.LanguageGerberImage,
		"Gettext Catalog":                  heartbeat.LanguageGettextCatalog,
		"Gherkin":                          heartbeat.LanguageGherkin,
		"Git":                              heartbeat.LanguageGit,
		"Git Attributes":                   heartbeat.LanguageGitAttributes,
		"Git Config":                       heartbeat.LanguageGitConfig,
		"GLSL":                             heartbeat.LanguageGLSL,
		"Glyph":                            heartbeat.LanguageGlyph,
		"Glyph Bitmap Distribution Format": heartbeat.LanguageGlyphBitmap,
		"GN":                               heartbeat.LanguageGN,
		"Gnuplot":                          heartbeat.LanguageGnuplot,
		"Go":                               heartbeat.LanguageGo,
		"Golo":                             heartbeat.LanguageGolo,
		"GoodData-CL":                      heartbeat.LanguageGoodDataCL,
		"Gosu":                             heartbeat.LanguageGosu,
		"Gosu Template":                    heartbeat.LanguageGosuTemplate,
		"Grace":                            heartbeat.LanguageGrace,
		"Gradle":                           heartbeat.LanguageGradle,
		"Gradle Config":                    heartbeat.LanguageGradleConfig,
		"Grammatical Framework":            heartbeat.LanguageGrammaticalFramework,
		"Graph Modeling Language":          heartbeat.LanguageGraphModelingLanguage,
		"GraphQL":                          heartbeat.LanguageGraphQL,
		"Graphviz (DOT)":                   heartbeat.LanguageGraphvizDOT,
		"Groff":                            heartbeat.LanguageGroff,
		"Groovy":                           heartbeat.LanguageGroovy,
		"Groovy Server Pages":              heartbeat.LanguageGroovyServerPages,
		"Hack":                             heartbeat.LanguageHack,
		"Haml":                             heartbeat.LanguageHaml,
		"Handlebars":                       heartbeat.LanguageHandlebars,
		"HAProxy":                          heartbeat.LanguageHAProxy,
		"Harbour":                          heartbeat.LanguageHarbour,
		"Haskell":                          heartbeat.LanguageHaskell,
		"Haxe":                             heartbeat.LanguageHaxe,
		"HCL":                              heartbeat.LanguageHCL,
		"Hexdump":                          heartbeat.LanguageHexdump,
		"HiveQL":                           heartbeat.LanguageHiveQL,
		"HLB":                              heartbeat.LanguageHLB,
		"HLSL":                             heartbeat.LanguageHLSL,
		"HolyC":                            heartbeat.LanguageHolyC,
		"HSAIL":                            heartbeat.LanguageHSAIL,
		"Hspec":                            heartbeat.LanguageHspec,
		"HTML":                             heartbeat.LanguageHTML,
		"HTML+Django":                      heartbeat.LanguageHTMLDjango,
		"HTML+ECR":                         heartbeat.LanguageHTMLECR,
		"HTML+EEX":                         heartbeat.LanguageHTMLEEX,
		"HTML+ERB":                         heartbeat.LanguageHTMLERB,
		"HTML+PHP":                         heartbeat.LanguageHTMLPHP,
		"HTML+Razor":                       heartbeat.LanguageHTMLRazor,
		"HTTP":                             heartbeat.LanguageHTTP,
		"HXML":                             heartbeat.LanguageHxml,
		"Hy":                               heartbeat.LanguageHy,
		"Hybris":                           heartbeat.LanguageHybris,
		"HyPhy":                            heartbeat.LanguageHyPhy,
		"Icon":                             heartbeat.LanguageIcon,
		"IDL":                              heartbeat.LanguageIDL,
		"Idris":                            heartbeat.LanguageIdris,
		"Ignore List":                      heartbeat.LanguageIgnoreList,
		"Igor":                             heartbeat.LanguageIgor,
		"IGOR Pro":                         heartbeat.LanguageIGORPro,
		"Image (jpeg)":                     heartbeat.LanguageImageJPEG,
		"Image (png)":                      heartbeat.LanguageImagePNG,
		"ImageJ Macro":                     heartbeat.LanguageImageJMacro,
		"Inform 6":                         heartbeat.LanguageInform6,
		"Inform 6 template":                heartbeat.LanguageInform6Template,
		"Inform 7":                         heartbeat.LanguageInform7,
		"INI":                              heartbeat.LanguageINI,
		"Inno Setup":                       heartbeat.LanguageInnoSetup,
		"Io":                               heartbeat.LanguageIo,
		"Ioke":                             heartbeat.LanguageIoke,
		"IRC Logs":                         heartbeat.LanguageIRCLogs,
		"Isabelle":                         heartbeat.LanguageIsabelle,
		"Isabelle ROOT":                    heartbeat.LanguageIsabelleRoot,
		"J":                                heartbeat.LanguageJ,
		"Jade":                             heartbeat.LanguageJade,
		"JAGS":                             heartbeat.LanguageJAGS,
		"Jasmin":                           heartbeat.LanguageJasmin,
		"Java":                             heartbeat.LanguageJava,
		"Java Properties":                  heartbeat.LanguageJavaProperties,
		"Java Server Page":                 heartbeat.LanguageJSP,
		"JavaScript":                       heartbeat.LanguageJavaScript,
		"JavaScript+ERB":                   heartbeat.LanguageJavaScriptERB,
		"JCL":                              heartbeat.LanguageJCL,
		"JFlex":                            heartbeat.LanguageJFlex,
		"Jison":                            heartbeat.LanguageJison,
		"Jison Lex":                        heartbeat.LanguageJisonLex,
		"Jolie":                            heartbeat.LanguageJolie,
		"JSGF":                             heartbeat.LanguageJSGF,
		"JSON":                             heartbeat.LanguageJSON,
		"JSON5":                            heartbeat.LanguageJSON5,
		"JSONiq":                           heartbeat.LanguageJSONiq,
		"JSONLD":                           heartbeat.LanguageJSONLD,
		"Jsonnet":                          heartbeat.LanguageJsonnet,
		"JSON with Comments":               heartbeat.LanguageJSONWithComments,
		"JSX":                              heartbeat.LanguageJSX,
		"Julia":                            heartbeat.LanguageJulia,
		"Julia console":                    heartbeat.LanguageJuliaConsole,
		"Jungle":                           heartbeat.LanguageJungle,
		"Jupyter Notebook":                 heartbeat.LanguageJupyterNotebook,
		"Juttle":                           heartbeat.LanguageJuttle,
		"Kaitai Struct":                    heartbeat.LanguageKaitai,
		"Kal":                              heartbeat.LanguageKal,
		"Kconfig":                          heartbeat.LanguageKconfig,
		"Kernel log":                       heartbeat.LanguageKernelLog,
		"KiCad Layout":                     heartbeat.LanguageKiCadLayout,
		"KiCad Legacy Layout":              heartbeat.LanguageKiCadLegacyLayout,
		"KiCad Schematic":                  heartbeat.LanguageKiCadSchematic,
		"Kit":                              heartbeat.LanguageKit,
		"Koka":                             heartbeat.LanguageKoka,
		"Kotlin":                           heartbeat.LanguageKotlin,
		"KRL":                              heartbeat.LanguageKRL,
		"LabVIEW":                          heartbeat.LanguageLabVIEW,
		"Laravel Template":                 heartbeat.LanguageLaravelTemplate,
		"Lark":                             heartbeat.LanguageLark,
		"Lasso":                            heartbeat.LanguageLasso,
		"LaTeX":                            heartbeat.LanguageLaTeX,
		"Latte":                            heartbeat.LanguageLatte,
		"Lean":                             heartbeat.LanguageLean,
		"LESS":                             heartbeat.LanguageLess,
		"Lex":                              heartbeat.LanguageLex,
		"LFE":                              heartbeat.LanguageLFE,
		"Lighttpd configuration file":      heartbeat.LanguageLighttpd,
		"LilyPond":                         heartbeat.LanguageLilyPond,
		"Limbo":                            heartbeat.LanguageLimbo,
		"Linker Script":                    heartbeat.LanguageLinkerScript,
		"Linux Kernel Module":              heartbeat.LanguageLinuxKernelModule,
		"Liquid":                           heartbeat.LanguageLiquid,
		"Literate Agda":                    heartbeat.LanguageLiterateAgda,
		"Literate CoffeeScript":            heartbeat.LanguageLiterateCoffeeScript,
		"Literate Cryptol":                 heartbeat.LanguageLiterateCryptol,
		"Literate Haskell":                 heartbeat.LanguageLiterateHaskell,
		"Literate Idris":                   heartbeat.LanguageLiterateIdris,
		"LiveScript":                       heartbeat.LanguageLiveScript,
		"LLVM":                             heartbeat.LanguageLLVM,
		"LLVM-MIR":                         heartbeat.LanguageLLVMMIR,
		"LLVM-MIR Body":                    heartbeat.LanguageLLVMMIRBody,
		"Log File":                         heartbeat.LanguageLogFile,
		"Logos":                            heartbeat.LanguageLogos,
		"Logtalk":                          heartbeat.LanguageLogtalk,
		"LOLCODE":                          heartbeat.LanguageLOLCODE,
		"LookML":                           heartbeat.LanguageLookML,
		"LoomScript":                       heartbeat.LanguageLoomScript,
		"LSL":                              heartbeat.LanguageLSL,
		"LTspice Symbol":                   heartbeat.LanguageLTspiceSymbol,
		"Lua":                              heartbeat.LanguageLua,
		"Makefile":                         heartbeat.LanguageMakefile,
		"Mako":                             heartbeat.LanguageMako,
		"Man":                              heartbeat.LanguageMan,
		"MAQL":                             heartbeat.LanguageMAQL,
		"Markdown":                         heartbeat.LanguageMarkdown,
		"Marko":                            heartbeat.LanguageMarko,
		"Mask":                             heartbeat.LanguageMask,
		"Mason":                            heartbeat.LanguageMason,
		"Mathematica":                      heartbeat.LanguageMathematica,
		"Matlab":                           heartbeat.LanguageMatlab,
		"Matlab session":                   heartbeat.LanguageMatlabSession,
		"Max":                              heartbeat.LanguageMax,
		"Max/MSP":                          heartbeat.LanguageMaxMSP,
		"Meson":                            heartbeat.LanguageMeson,
		"Metafont":                         heartbeat.LanguageMetafont,
		"Metapost":                         heartbeat.LanguageMetapost,
		"MIME":                             heartbeat.LanguageMIME,
		"MiniD":                            heartbeat.LanguageMiniD,
		"MiniScript":                       heartbeat.LanguageMiniScript,
		"MiniZinc":                         heartbeat.LanguageMiniZinc,
		"Mirah":                            heartbeat.LanguageMirah,
		"MLIR":                             heartbeat.LanguageMLIR,
		"Modelica":                         heartbeat.LanguageModelica,
		"Modula-2":                         heartbeat.LanguageModula2,
		"MoinMoin/Trac Wiki markup":        heartbeat.LanguageMoinWiki,
		"Monkey":                           heartbeat.LanguageMonkey,
		"MonkeyC":                          heartbeat.LanguageMonkeyC,
		"Monte":                            heartbeat.LanguageMonte,
		"MOOCode":                          heartbeat.LanguageMOOCode,
		"MoonScript":                       heartbeat.LanguageMoonScript,
		"MorrowindScript":                  heartbeat.LanguageMorrowindScript,
		"Mosel":                            heartbeat.LanguageMosel,
		"mozhashpreproc":                   heartbeat.LanguageMozPreprocHash,
		"mozpercentpreproc":                heartbeat.LanguageMozPreprocPercent,
		"MQL":                              heartbeat.LanguageMQL,
		"Mscgen":                           heartbeat.LanguageMscgen,
		"MSDOS Session":                    heartbeat.LanguageMSDOSSession,
		"MuPAD":                            heartbeat.LanguageMuPAD,
		"Mustache":                         heartbeat.LanguageMustache,
		"MXML":                             heartbeat.LanguageMXML,
		"Myghty":                           heartbeat.LanguageMyghty,
		"MySQL":                            heartbeat.LanguageMySQL,
		"NASM":                             heartbeat.LanguageNASM,
		"NCL":                              heartbeat.LanguageNCL,
		"Nemerle":                          heartbeat.LanguageNemerle,
		"Neon":                             heartbeat.LanguageNeon,
		"nesC":                             heartbeat.LanguageNesC,
		"newLisp":                          heartbeat.LanguageNewLisp,
		"Newspeak":                         heartbeat.LanguageNewspeak,
		"Nginx":                            heartbeat.LanguageNginx,
		"Nginx configuration file":         heartbeat.LanguageNginxConfig,
		"Nimrod":                           heartbeat.LanguageNimrod,
		"Nit":                              heartbeat.LanguageNit,
		"Nix":                              heartbeat.LanguageNix,
		"Notmuch":                          heartbeat.LanguageNotmuch,
		"Nu":                               heartbeat.LanguageNu,
		"NSIS":                             heartbeat.LanguageNSIS,
		"NumPy":                            heartbeat.LanguageNumPy,
		"NuSMV":                            heartbeat.LanguageNuSMV,
		"objdump":                          heartbeat.LanguageObjdump,
		"objdump-nasm":                     heartbeat.LanguageNASMObjdump,
		"Objective-C":                      heartbeat.LanguageObjectiveC,
		"Objective-C++":                    heartbeat.LanguageObjectiveCPP,
		"Objective-J":                      heartbeat.LanguageObjectiveJ,
		"OCaml":                            heartbeat.LanguageOCaml,
		"Octave":                           heartbeat.LanguageOctave,
		"ODIN":                             heartbeat.LanguageODIN,
		"ooc":                              heartbeat.LanguageOoc,
		"Opa":                              heartbeat.LanguageOpa,
		"OpenEdge ABL":                     heartbeat.LanguageOpenEdgeABL,
		"OpenSCAD":                         heartbeat.LanguageOpenSCAD,
		"Org":                              heartbeat.LanguageOrg,
		"PacmanConf":                       heartbeat.LanguagePacmanConf,
		"Pan":                              heartbeat.LanguagePan,
		"ParaSail":                         heartbeat.LanguageParaSail,
		"Parrot":                           heartbeat.LanguageParrot,
		"Pascal":                           heartbeat.LanguagePascal,
		"Pawn":                             heartbeat.LanguagePawn,
		"PEG":                              heartbeat.LanguagePEG,
		"Perl":                             heartbeat.LanguagePerl,
		"Perl6":                            heartbeat.LanguagePerl6,
		"PHP":                              heartbeat.LanguagePHP,
		"PHTML":                            heartbeat.LanguagePHTML,
		"Pig":                              heartbeat.LanguagePig,
		"Pike":                             heartbeat.LanguagePike,
		"PkgConfig":                        heartbeat.LanguagePkgConfig,
		"PL/pgSQL":                         heartbeat.LanguagePLpgSQL,
		"Pointless":                        heartbeat.LanguagePointless,
		"Pony":                             heartbeat.LanguagePony,
		"PostgreSQL console (psql)":        heartbeat.LanguagePostgresConsole,
		"PostgreSQL SQL dialect":           heartbeat.LanguagePostgres,
		"PostScript":                       heartbeat.LanguagePostScript,
		"POVRay":                           heartbeat.LanguagePOVRay,
		"PowerShell":                       heartbeat.LanguagePowerShell,
		"PowerShell Session":               heartbeat.LanguagePowerShellSession,
		"Praat":                            heartbeat.LanguagePraat,
		"Prolog":                           heartbeat.LanguageProlog,
		"PromQL":                           heartbeat.LanguagePromQL,
		"Properties":                       heartbeat.LanguagePropertiesJava,
		"Protocol Buffer":                  heartbeat.LanguageProtocolBuffer,
		"PsySH console session for PHP":    heartbeat.LanguagePsyShPHP,
		"Pug":                              heartbeat.LanguagePug,
		"Puppet":                           heartbeat.LanguagePuppet,
		"Pure Data":                        heartbeat.LanguagePureData,
		"PureScript":                       heartbeat.LanguagePureScript,
		"PyPy Log":                         heartbeat.LanguagePyPyLog,
		"Python":                           heartbeat.LanguagePython,
		"Python 2.x":                       heartbeat.LanguagePython2,
		"Python 2.x Traceback":             heartbeat.LanguagePython2Traceback,
		"Python Traceback":                 heartbeat.LanguagePythonTraceback,
		"Python console session":           heartbeat.LanguagePythonConsole,
		"QBasic":                           heartbeat.LanguageQBasic,
		"QML":                              heartbeat.LanguageQML,
		"QVTO":                             heartbeat.LanguageQVTO,
		"R":                                heartbeat.LanguageR,
		"Racket":                           heartbeat.LanguageRacket,
		"Ragel":                            heartbeat.LanguageRagel,
		"Embedded Ragel":                   heartbeat.LanguageRagelEmbedded,
		"Raku":                             heartbeat.LanguageRaku,
		"RAML":                             heartbeat.LanguageRAML,
		"Rascal":                           heartbeat.LanguageRascal,
		"Raw token data":                   heartbeat.LanguageRawToken,
		"RConsole":                         heartbeat.LanguageRConsole,
		"Rd":                               heartbeat.LanguageRd,
		"RDoc":                             heartbeat.LanguageRDoc,
		"Readline Config":                  heartbeat.LanguageReadlineConfig,
		"REALbasic":                        heartbeat.LanguageREALbasic,
		"Reason":                           heartbeat.LanguageReasonML,
		"Rebol":                            heartbeat.LanguageREBOL,
		"Record Jar":                       heartbeat.LanguageRecordJar,
		"Red":                              heartbeat.LanguageRed,
		"Redcode":                          heartbeat.LanguageRedcode,
		"reg":                              heartbeat.LanguageRegistry,
		"Regular Expression":               heartbeat.LanguageRegularExpression,
		"RenderScript":                     heartbeat.LanguageRenderScript,
		"Ren'Py":                           heartbeat.LanguageRenPy,
		"ReScript":                         heartbeat.LanguageReScript,
		"ResourceBundle":                   heartbeat.LanguageResourceBundle,
		"reStructuredText":                 heartbeat.LanguageReStructuredText,
		"REXX":                             heartbeat.LanguageRexx,
		"RHTML":                            heartbeat.LanguageRHTML,
		"Rich Text Format":                 heartbeat.LanguageRichTextFormat,
		"Ride":                             heartbeat.LanguageRide,
		"Ring":                             heartbeat.LanguageRing,
		"Riot":                             heartbeat.LanguageRiot,
		"RMarkdown":                        heartbeat.LanguageRMarkdown,
		"Relax-NG Compact":                 heartbeat.LanguageRNGCompact,
		"Roboconf Graph":                   heartbeat.LanguageRoboconfGraph,
		"Roboconf Instances":               heartbeat.LanguageRoboconfInstances,
		"RobotFramework":                   heartbeat.LanguageRobotFramework,
		"Roff":                             heartbeat.LanguageRoff,
		"Roff Manpage":                     heartbeat.LanguageRoffManpage,
		"Rouge":                            heartbeat.LanguageRouge,
		"RPC":                              heartbeat.LanguageRPC,
		"RPMSpec":                          heartbeat.LanguageRPMSpec,
		"RQL":                              heartbeat.LanguageRQL,
		"RSL":                              heartbeat.LanguageRSL,
		"Ruby":                             heartbeat.LanguageRuby,
		"Ruby irb session":                 heartbeat.LanguageRubyIRBSession,
		"RUNOFF":                           heartbeat.LanguageRUNOFF,
		"Rust":                             heartbeat.LanguageRust,
		"S":                                heartbeat.LanguageS,
		"Sage":                             heartbeat.LanguageSage,
		"Salt":                             heartbeat.LanguageSalt,
		"SARL":                             heartbeat.LanguageSARL,
		"SAS":                              heartbeat.LanguageSAS,
		"Sass":                             heartbeat.LanguageSass,
		"Scala":                            heartbeat.LanguageScala,
		"Scalate Server Page":              heartbeat.LanguageSSP,
		"Scaml":                            heartbeat.LanguageScaml,
		"scdoc":                            heartbeat.LanguageScdoc,
		"Scheme":                           heartbeat.LanguageScheme,
		"Scilab":                           heartbeat.LanguageScilab,
		"Scribe":                           heartbeat.LanguageScribe,
		"SCSS":                             heartbeat.LanguageSCSS,
		"sed":                              heartbeat.LanguageSed,
		"Self":                             heartbeat.LanguageSelf,
		"ShaderLab":                        heartbeat.LanguageShaderLab,
		"Shell":                            heartbeat.LanguageShell,
		"ShellSession":                     heartbeat.LanguageShellSession,
		"Shen":                             heartbeat.LanguageShen,
		"ShExC":                            heartbeat.LanguageShExC,
		"Sieve":                            heartbeat.LanguageSieve,
		"Silver":                           heartbeat.LanguageSilver,
		"Singularity":                      heartbeat.LanguageSingularity,
		"Sketch Drawing":                   heartbeat.LanguageSketchDrawing,
		"Slash":                            heartbeat.LanguageSlash,
		"Slice":                            heartbeat.LanguageSlice,
		"Slim":                             heartbeat.LanguageSlim,
		"Slurm":                            heartbeat.LanguageSlurm,
		"Smali":                            heartbeat.LanguageSmali,
		"Smalltalk":                        heartbeat.LanguageSmalltalk,
		"SmartGameFormat":                  heartbeat.LanguageSmartGameFormat,
		"Smarty":                           heartbeat.LanguageSmarty,
		"S/MIME":                           heartbeat.LanguageSMIME,
		"SmPL":                             heartbeat.LanguageSmPL,
		"SMT":                              heartbeat.LanguageSMT,
		"Snobol":                           heartbeat.LanguageSnobol,
		"Snowball":                         heartbeat.LanguageSnowball,
		"Solidity":                         heartbeat.LanguageSolidity,
		"SourcePawn":                       heartbeat.LanguageSourcePawn,
		"SPARQL":                           heartbeat.LanguageSPARQL,
		"Spline Font Database":             heartbeat.LanguageSplineFontDatabase,
		"SQF":                              heartbeat.LanguageSQF,
		"SQL":                              heartbeat.LanguageSQL,
		"SQLPL":                            heartbeat.LanguageSQLPL,
		"sqlite3con":                       heartbeat.LanguageSqlite3con,
		"SquidConf":                        heartbeat.LanguageSquidConf,
		"Squirrel":                         heartbeat.LanguageSquirrel,
		"SRecode Template":                 heartbeat.LanguageSRecodeTemplate,
		"SSH Config":                       heartbeat.LanguageSSHConfig,
		"Stan":                             heartbeat.LanguageStan,
		"Starlark":                         heartbeat.LanguageStarlark,
		"Stata":                            heartbeat.LanguageStata,
		"Standard ML":                      heartbeat.LanguageSML,
		"STON":                             heartbeat.LanguageSTON,
		"Stylus":                           heartbeat.LanguageStylus,
		"Sublime Text Config":              heartbeat.LanguageSublimeTextConfig,
		"SubRip Text":                      heartbeat.LanguageSubRipText,
		"SugarSS":                          heartbeat.LanguageSugarSS,
		"SuperCollider":                    heartbeat.LanguageSuperCollider,
		"Svelte":                           heartbeat.LanguageSvelte,
		"SVG":                              heartbeat.LanguageSVG,
		"Swift":                            heartbeat.LanguageSwift,
		"SWIG":                             heartbeat.LanguageSWIG,
		"systemd":                          heartbeat.LanguageSYSTEMD,
		"SystemVerilog":                    heartbeat.LanguageSystemVerilog,
		"TableGen":                         heartbeat.LanguageTableGen,
		"TADS 3":                           heartbeat.LanguageTADS3,
		"TAP":                              heartbeat.LanguageTAP,
		"TASM":                             heartbeat.LanguageTASM,
		"Tcl":                              heartbeat.LanguageTcl,
		"Tcsh":                             heartbeat.LanguageTcsh,
		"Tcsh Session":                     heartbeat.LanguageTcshSession,
		"Tea":                              heartbeat.LanguageTea,
		"Tera Term macro":                  heartbeat.LanguageTeraTerm,
		"Termcap":                          heartbeat.LanguageTermcap,
		"Terminfo":                         heartbeat.LanguageTerminfo,
		"Terra":                            heartbeat.LanguageTerra,
		"Terraform":                        heartbeat.LanguageTerraform,
		"TeX":                              heartbeat.LanguageTeX,
		"Texinfo":                          heartbeat.LanguageTexinfo,
		"Text":                             heartbeat.LanguageText,
		"Textile":                          heartbeat.LanguageTextile,
		"Thrift":                           heartbeat.LanguageThrift,
		"tiddler":                          heartbeat.LanguageTiddler,
		"TI Program":                       heartbeat.LanguageTIProgram,
		"TLA":                              heartbeat.LanguageTLA,
		"Todotxt":                          heartbeat.LanguageTodotxt,
		"TOML":                             heartbeat.LanguageTOML,
		"TradingView":                      heartbeat.LanguageTradingView,
		"TrafficScript":                    heartbeat.LanguageTrafficScript,
		"TSQL":                             heartbeat.LanguageTransactSQL,
		"Treetop":                          heartbeat.LanguageTreetop,
		"TSV":                              heartbeat.LanguageTSV,
		"TSX":                              heartbeat.LanguageTSX,
		"Turing":                           heartbeat.LanguageTuring,
		"Turtle":                           heartbeat.LanguageTurtle,
		"Twig":                             heartbeat.LanguageTwig,
		"TXL":                              heartbeat.LanguageTXL,
		"Type Language":                    heartbeat.LanguageTypeLanguage,
		"Typographic Number Theory":        heartbeat.LanguageTNT,
		"TypeScript":                       heartbeat.LanguageTypeScript,
		"TypoScript":                       heartbeat.LanguageTypoScript,
		"ucode":                            heartbeat.LanguageUcode,
		"Unicon":                           heartbeat.LanguageUnicon,
		"Unified Parallel C":               heartbeat.LanguageUnifiedParallelC,
		"Unity3D Asset":                    heartbeat.LanguageUnity3DAsset,
		"Unix Assembly":                    heartbeat.LanguageUnixAssembly,
		"Uno":                              heartbeat.LanguageUno,
		"UnrealScript":                     heartbeat.LanguageUnrealScript,
		"UrbiScript":                       heartbeat.LanguageUrbiScript,
		"UrWeb":                            heartbeat.LanguageUrWeb,
		"USD":                              heartbeat.LanguageUSD,
		"V":                                heartbeat.LanguageV,
		"Vala":                             heartbeat.LanguageVala,
		"VB":                               heartbeat.LanguageVB,
		"VBA":                              heartbeat.LanguageVBA,
		"VB.NET":                           heartbeat.LanguageVBNet,
		"VBScript":                         heartbeat.LanguageVBScript,
		"VCL":                              heartbeat.LanguageVCL,
		"VCLSnippets":                      heartbeat.LanguageVCLSnippets,
		"VCTreeStatus":                     heartbeat.LanguageVCTreeStatus,
		"Velocity":                         heartbeat.LanguageVelocity,
		"Verilog":                          heartbeat.LanguageVerilog,
		"VGL":                              heartbeat.LanguageVGL,
		"VHDL":                             heartbeat.LanguageVHDL,
		"Vim Help File":                    heartbeat.LanguageVimHelpFile,
		"VimL":                             heartbeat.LanguageVimL,
		"Vim script":                       heartbeat.LanguageVimScript,
		"Vim Snippet":                      heartbeat.LanguageVimSnippet,
		"Volt":                             heartbeat.LanguageVolt,
		"Vue.js":                           heartbeat.LanguageVueJS,
		"Wavefront Material":               heartbeat.LanguageWavefrontMaterial,
		"Wavefront Object":                 heartbeat.LanguageWavefrontObject,
		"wdl":                              heartbeat.LanguageWdl,
		"WDTE":                             heartbeat.LanguageWDTE,
		"WDiff":                            heartbeat.LanguageWDiff,
		"WebAssembly":                      heartbeat.LanguageWebAssembly,
		"WebIDL":                           heartbeat.LanguageWebIDL,
		"Web Ontology Language":            heartbeat.LanguageWebOntologyLanguage,
		"WebVTT":                           heartbeat.LanguageWebVTT,
		"Wget Config":                      heartbeat.LanguageWgetConfig,
		"Whiley":                           heartbeat.LanguageWhiley,
		"Windows Registry Entries":         heartbeat.LanguageWindowsRegistryEntries,
		"wisp":                             heartbeat.LanguageWisp,
		"Wollok":                           heartbeat.LanguageWollok,
		"World of Warcraft Addon Data":     heartbeat.LanguageWowAddonData,
		"X10":                              heartbeat.LanguageX10,
		"XAML":                             heartbeat.LanguageXAML,
		"xBase":                            heartbeat.LanguageXBase,
		"X BitMap":                         heartbeat.LanguageXBitMap,
		"XC":                               heartbeat.LanguageXC,
		"XCompose":                         heartbeat.LanguageXCompose,
		"X Font Directory Index":           heartbeat.LanguageXFontDirectoryIndex,
		"XML":                              heartbeat.LanguageXML,
		"XML Property List":                heartbeat.LanguageXMLPropertyList,
		"Xojo":                             heartbeat.LanguageXojo,
		"Xorg":                             heartbeat.LanguageXorg,
		"XPages":                           heartbeat.LanguageXPages,
		"X PixMap":                         heartbeat.LanguageXPixMap,
		"XProc":                            heartbeat.LanguageXProc,
		"XQuery":                           heartbeat.LanguageXQuery,
		"XS":                               heartbeat.LanguageXS,
		"XSLT":                             heartbeat.LanguageXSLT,
		"Xtend":                            heartbeat.LanguageXtend,
		"xtlang":                           heartbeat.LanguageXtlang,
		"Yacc":                             heartbeat.LanguageYacc,
		"YAML":                             heartbeat.LanguageYAML,
		"YANG":                             heartbeat.LanguageYANG,
		"YARA":                             heartbeat.LanguageYARA,
		"YASnippet":                        heartbeat.LanguageYASnippet,
		"ZAP":                              heartbeat.LanguageZAP,
		"Zeek":                             heartbeat.LanguageZeek,
		"ZenScript":                        heartbeat.LanguageZenScript,
		"Zephir":                           heartbeat.LanguageZephir,
		"Zig":                              heartbeat.LanguageZig,
		"ZIL":                              heartbeat.LanguageZIL,
		"Zimpl":                            heartbeat.LanguageZimpl,
	}
}

func languageTestsAliases() map[string]heartbeat.Language {
	return map[string]heartbeat.Language{
		"ApacheConf":        heartbeat.LanguageApacheConfig,
		"ASP Classic":       heartbeat.LanguageClassicASP,
		"Batch Script":      heartbeat.LanguageBatchfile,
		"FStar":             heartbeat.LanguageFStar,
		"Golang":            heartbeat.LanguageGo,
		"JSON-LD":           heartbeat.LanguageJSONLD,
		"SaltStack":         heartbeat.LanguageSalt,
		"Transact-SQL":      heartbeat.LanguageTransactSQL,
		"Visual Basic .NET": heartbeat.LanguageVBNet,
	}
}

func TestParseLanguage(t *testing.T) {
	// standard language names
	for value, language := range languageTests() {
		t.Run(value, func(t *testing.T) {
			parsed, ok := heartbeat.ParseLanguage(value)
			assert.True(t, ok)

			assert.Equal(t, language, parsed, fmt.Sprintf("got: %q, want: %q", parsed, language))
		})
	}

	// alias language names
	for value, language := range languageTestsAliases() {
		t.Run(value, func(t *testing.T) {
			parsed, ok := heartbeat.ParseLanguage(value)
			assert.True(t, ok)

			assert.Equal(t, language, parsed, fmt.Sprintf("got: %q, want: %q", parsed, language))
		})
	}

	t.Run("lower case", func(t *testing.T) {
		parsed, ok := heartbeat.ParseLanguage("go")
		assert.True(t, ok)

		assert.Equal(t, heartbeat.LanguageGo, parsed)
	})

	t.Run("hash", func(t *testing.T) {
		parsed, ok := heartbeat.ParseLanguage("CSharp")
		assert.True(t, ok)

		assert.Equal(t, heartbeat.LanguageCSharp, parsed)
	})

	t.Run("plus sign", func(t *testing.T) {
		parsed, ok := heartbeat.ParseLanguage("CPP")
		assert.True(t, ok)

		assert.Equal(t, heartbeat.LanguageCPP, parsed)
	})

	t.Run("leading space", func(t *testing.T) {
		parsed, ok := heartbeat.ParseLanguage(" Go")
		assert.True(t, ok)

		assert.Equal(t, heartbeat.LanguageGo, parsed)
	})

	t.Run("trailing space", func(t *testing.T) {
		parsed, ok := heartbeat.ParseLanguage("Go ")
		assert.True(t, ok)

		assert.Equal(t, heartbeat.LanguageGo, parsed)
	})

	t.Run("missing hyphen", func(t *testing.T) {
		parsed, ok := heartbeat.ParseLanguage("ObjectiveC")
		assert.True(t, ok)

		assert.Equal(t, heartbeat.LanguageObjectiveC, parsed)
	})

	t.Run("missing space", func(t *testing.T) {
		parsed, ok := heartbeat.ParseLanguage("Sublime Text Config")
		assert.True(t, ok)

		assert.Equal(t, heartbeat.LanguageSublimeTextConfig, parsed)
	})
}

func TestParseLanguage_Unknown(t *testing.T) {
	parsed, ok := heartbeat.ParseLanguage("invalid")

	assert.False(t, ok)
	assert.Equal(t, heartbeat.LanguageUnknown, parsed)
}

func TestParseLanguageFromChroma(t *testing.T) {
	tests := map[string]heartbeat.Language{
		"Ampl":               heartbeat.LanguageAMPL,
		"autohotkey":         heartbeat.LanguageAutoHotkey,
		"Base Makefile":      heartbeat.LanguageMakefile,
		"c-objdump":          heartbeat.LanguageCObjdump,
		"Coldfusion CFC":     heartbeat.LanguageColdfusionCFC,
		"Coldfusion HTML":    heartbeat.LanguageColdfusionHTML,
		"cpp-objdump":        heartbeat.LanguageCppObjdump,
		"CUDA":               heartbeat.LanguageCUDA,
		"EmacsLisp":          heartbeat.LanguageEmacsLisp,
		"Go HTML Template":   heartbeat.LanguageGo,
		"Go Text Template":   heartbeat.LanguageGo,
		"FSharp":             heartbeat.LanguageFSharp,
		"GAS":                heartbeat.LanguageAssembly,
		"Hxml":               heartbeat.LanguageHxml,
		"JSON-LD":            heartbeat.LanguageJSONLD,
		"LessCss":            heartbeat.LanguageLess,
		"liquid":             heartbeat.LanguageLiquid,
		"markdown":           heartbeat.LanguageMarkdown,
		"NewLisp":            heartbeat.LanguageNewLisp,
		"Nim":                heartbeat.LanguageNimrod,
		"Ooc":                heartbeat.LanguageOoc,
		"Org Mode":           heartbeat.LanguageOrg,
		"plaintext":          heartbeat.LanguageText,
		"Python 3":           heartbeat.LanguagePython,
		"R":                  heartbeat.LanguageS,
		"react":              heartbeat.LanguageJSX,
		"ReasonML":           heartbeat.LanguageReasonML,
		"REBOL":              heartbeat.LanguageREBOL,
		"Rexx":               heartbeat.LanguageRexx,
		"SWIG":               heartbeat.LanguageSWIG,
		"systemverilog":      heartbeat.LanguageSystemVerilog,
		"Transact-SQL":       heartbeat.LanguageTransactSQL,
		"TypoScriptCssData":  heartbeat.LanguageTypoScript,
		"TypoScriptHtmlData": heartbeat.LanguageTypoScript,
		"VB.net":             heartbeat.LanguageVBNet,
		"verilog":            heartbeat.LanguageVerilog,
		"vue":                heartbeat.LanguageVueJS,
		"Web IDL":            heartbeat.LanguageWebIDL,
		// lowercase
		"zig": heartbeat.LanguageZig,
		// missing blank space
		"ProtocolBuffer": heartbeat.LanguageProtocolBuffer,
		// missing hyphen
		"ObjectiveC": heartbeat.LanguageObjectiveC,
		// plus sign
		"CPP": heartbeat.LanguageCPP,
		// hash
		"CSharp": heartbeat.LanguageCSharp,
	}

	for lexerName, language := range tests {
		t.Run(lexerName, func(t *testing.T) {
			parsed, ok := heartbeat.ParseLanguageFromChroma(lexerName)

			assert.True(t, ok)
			assert.Equal(t, language, parsed, fmt.Sprintf("got: %q, want: %q", parsed, language))
		})
	}
}

func TestParseLanguageFromChroma_Unknown(t *testing.T) {
	parsed, ok := heartbeat.ParseLanguageFromChroma("invalid")

	assert.False(t, ok)
	assert.Equal(t, heartbeat.LanguageUnknown, parsed)
}

func TestParseLanguageFromChroma_AllLexersSupported(t *testing.T) {
	for _, lexer := range lexers.Registry.Lexers {
		config := lexer.Config()

		parsed, ok := heartbeat.ParseLanguageFromChroma(config.Name)

		assert.True(t, ok, fmt.Sprintf("failed parsing language from lexer %q", config.Name))
		assert.NotEqual(t, heartbeat.LanguageUnknown, parsed, fmt.Sprintf(
			"parsed language.Unknown. Failed parsing language from lexer %q",
			config.Name,
		))
	}
}

func TestLanguage_MarshalJSON(t *testing.T) {
	for value, language := range languageTests() {
		t.Run(value, func(t *testing.T) {
			data, err := json.Marshal(language)
			require.NoError(t, err)

			assert.JSONEq(t, `"`+value+`"`, string(data))
		})
	}
}

func TestLanguage_MarshalJSON_UnknownLanguage(t *testing.T) {
	data, err := json.Marshal(heartbeat.LanguageUnknown)
	require.NoError(t, err)

	assert.JSONEq(t, `null`, string(data))
}

func TestLanguage_UnmarshalJSON(t *testing.T) {
	for value, language := range languageTests() {
		t.Run(value, func(t *testing.T) {
			var l heartbeat.Language
			require.NoError(t, json.Unmarshal([]byte(`"`+value+`"`), &l))

			assert.Equal(t, language, l)
		})
	}
}

func TestLanguage_String(t *testing.T) {
	for value, language := range languageTests() {
		t.Run(value, func(t *testing.T) {
			assert.Equal(t, value, language.String())
		})
	}
}

func TestLanguage_String_UnknownLanguage(t *testing.T) {
	assert.Equal(t, "Unknown", heartbeat.LanguageUnknown.String())
}

func TestLanguage_StringChroma(t *testing.T) {
	tests := map[string]heartbeat.Language{
		"ApacheConf":      heartbeat.LanguageApacheConfig,
		"Ampl":            heartbeat.LanguageAMPL,
		"autohotkey":      heartbeat.LanguageAutoHotkey,
		"Base Makefile":   heartbeat.LanguageMakefile,
		"c-objdump":       heartbeat.LanguageCObjdump,
		"Coldfusion CFC":  heartbeat.LanguageColdfusionCFC,
		"Coldfusion HTML": heartbeat.LanguageColdfusionHTML,
		"cpp-objdump":     heartbeat.LanguageCppObjdump,
		"CUDA":            heartbeat.LanguageCUDA,
		"EmacsLisp":       heartbeat.LanguageEmacsLisp,
		"GAS":             heartbeat.LanguageAssembly,
		"FSharp":          heartbeat.LanguageFSharp,
		"FStar":           heartbeat.LanguageFStar,
		"Go":              heartbeat.LanguageGo,
		"Hxml":            heartbeat.LanguageHxml,
		"JSON-LD":         heartbeat.LanguageJSONLD,
		"LessCss":         heartbeat.LanguageLess,
		"liquid":          heartbeat.LanguageLiquid,
		"markdown":        heartbeat.LanguageMarkdown,
		"Nim":             heartbeat.LanguageNimrod,
		"Ooc":             heartbeat.LanguageOoc,
		"Org Mode":        heartbeat.LanguageOrg,
		"plaintext":       heartbeat.LanguageText,
		"R":               heartbeat.LanguageS,
		"react":           heartbeat.LanguageJSX,
		"ReasonML":        heartbeat.LanguageReasonML,
		"REBOL":           heartbeat.LanguageREBOL,
		"Rexx":            heartbeat.LanguageRexx,
		"SWIG":            heartbeat.LanguageSWIG,
		"systemverilog":   heartbeat.LanguageSystemVerilog,
		"VB.net":          heartbeat.LanguageVBNet,
		"verilog":         heartbeat.LanguageVerilog,
		"vue":             heartbeat.LanguageVueJS,
		"Web IDL":         heartbeat.LanguageWebIDL,
	}

	for lexerName, language := range tests {
		t.Run(lexerName, func(t *testing.T) {
			assert.Equal(t, lexerName, language.StringChroma())
		})
	}
}

func TestLanguage_StringChroma_AllLexersSupported(t *testing.T) {
	for _, lexer := range lexers.Registry.Lexers {
		config := lexer.Config()

		// Aliases, which match in addition to standard spelling of languages are ignored here.
		switch config.Name {
		case "Go HTML Template", "Go Text Template":
			continue
		case "Python 3":
			continue
		case "TypoScriptCssData", "TypoScriptHtmlData":
			continue
		}

		parsed, ok := heartbeat.ParseLanguageFromChroma(config.Name)
		require.True(t, ok, fmt.Sprintf("Failed parsing language from lexer %q", config.Name))
		require.NotEqual(t, heartbeat.LanguageUnknown, parsed, fmt.Sprintf(
			"Parsed language.Unknown. Failed parsing language from lexer %q",
			config.Name,
		))

		assert.Equal(t, config.Name, parsed.StringChroma())
	}
}
