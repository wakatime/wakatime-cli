// language.go contains programming language definitions and mappings.
// To override the priority of languages when multiple languages map to the same
// string, use the corresponding priority map defined in language_gen.go.
//
// priority:
// 1. LanguageTransactSQL -> languageTSQLStr
// 2. LanguageGo -> languageGoStr
// 3. LanguageSYSTEMD -> languageSYSTEMDStr
// 4. LanguageApacheConfig -> languageApacheConfigStr
// 5. LanguageBatchfile -> languageBatchfileStr
// 6. LanguageBatchfile -> languageBatchScriptStr
// 7. LanguageClassicASP -> languageClassicASPStr
// 8. LanguageClassicASP -> languageASPClassicStr
// 9. LanguageDesktopFile -> languageDesktopFileStr
// 10. LanguageFStar -> languageFStarStr
// 11. LanguageSalt -> languageSaltStr
// 12. LanguageSalt -> languageSaltStackStr
// 13. LanguageMaterializeSQLDialect -> languageMaterializeSQLDialectStr
// 14. LanguageMaterializeSQLDialect -> languageMzqlStr
// 15. LanguageVBNet -> languageVBNetStr
// 16. LanguageVBNet -> languageVisualBasicNetStr
//
// chromaPriority: (chroma lexer name -> Language constant)
// These override the default Chroma lexer to Language mapping.
// 1. Ampl -> LanguageAMPL
// 2. ApacheConf -> LanguageApacheConfig
// 3. ArangoDB AQL -> LanguageArangoDBQueryLanguage
// 4. c-objdump -> LanguageCObjdump
// 5. Coldfusion CFC -> LanguageColdfusionCFC
// 6. Coldfusion HTML -> LanguageColdfusionHTML
// 7. cpp-objdump -> LanguageCppObjdump
// 8. CUDA -> LanguageCUDA
// 9. dns -> LanguageDNSZone
// 10. EmacsLisp -> LanguageEmacsLisp
// 11. FSharp -> LanguageFSharp
// 12. GAS -> LanguageAssembly
// 13. Go HTML Template -> LanguageGo
// 14. Go Text Template -> LanguageGo
// 15. Hxml -> LanguageHxml
// 16. JSON-LD -> LanguageJSONLD
// 17. ISCdhcpd -> LanguageISCdhcpd
// 18. LessCss -> LanguageLess
// 19. liquid -> LanguageLiquid
// 20. markdown -> LanguageMarkdown
// 21. NewLisp -> LanguageNewLisp
// 22. Nim -> LanguageNimrod
// 23. Ooc -> LanguageOoc
// 24. Org Mode -> LanguageOrg
// 25. plaintext -> LanguageText
// 26. properties -> LanguageJavaProperties
// 27. PSL -> LanguagePSL
// 28. R -> LanguageS
// 29. react -> LanguageJSX
// 30. ReasonML -> LanguageReasonML
// 31. REBOL -> LanguageREBOL
// 32. Rexx -> LanguageRexx
// 33. Sed -> LanguageSed
// 34. stas -> LanguageStas
// 35. SYSTEMD -> LanguageSYSTEMD
// 36. systemverilog -> LanguageSystemVerilog
// 37. Tal -> LanguageUxntal
// 38. Transact-SQL -> LanguageTransactSQL
// 39. TypoScriptCssData -> LanguageTypoScript
// 40. TypoScriptHtmlData -> LanguageTypoScript
// 41. VB.net -> LanguageVBNet
// 42. verilog -> LanguageVerilog
// 43. vue -> LanguageVueJS
// 44. Web IDL -> LanguageWebIDL
// 45. FStar -> LanguageFStar
// 46. Protocol Buffer Text Format -> LanguageProtocolBuffer
// 47. WebAssembly Text Format -> LanguageWebAssembly
// 48. MDX -> LanguageMDX
// 49. Arturo -> LanguageArturo
// 50. Gemfile.lock -> LanguageGemfileLock
// 51. Gettext -> LanguageGettext
// 52. KDL -> LanguageKDL
// 53. Lateralus -> LanguageLateralus
// 54. Luau -> LanguageLuau
// 55. Markless -> LanguageMarkless
// 56. microcad -> LanguageMicrocad
// 57. MoonBit -> LanguageMoonBit
// 58. Spade -> LanguageSpade
// 59. Templ -> LanguageTempl

package heartbeat

//go:generate go run language_gen.go

import (
	"strings"
)

// Language represents a programming language.
type Language int

const (
	// LanguageUnknown represents the Unknown programming language.
	LanguageUnknown Language = iota
	// Language1CEnterprise represents the 1CEnterprise programming language.
	Language1CEnterprise
	// Language4D represents the 4D programming language.
	Language4D
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
	// LanguageAdobeFontMetrics represents the AdobeFontMetrics programming language.
	LanguageAdobeFontMetrics
	// LanguageAdvPL represents the AdvPL programming language.
	LanguageAdvPL
	// LanguageAgda represents the Agda programming language.
	LanguageAgda
	// LanguageAGSScript represents the AGSScript programming language.
	LanguageAGSScript
	// LanguageAheui represents the Aheui programming language.
	LanguageAheui
	// LanguageAL represents the AL programming language.
	LanguageAL
	// LanguageAlloy represents the Alloy programming language.
	LanguageAlloy
	// LanguageAlpineAbuild represents the AlpineAbuild programming language.
	LanguageAlpineAbuild
	// LanguageAltiumDesigner represents the AltiumDesigner programming language.
	LanguageAltiumDesigner
	// LanguageAmbientTalk represents the AmbientTalk programming language.
	LanguageAmbientTalk
	// LanguageAMPL represents the AMPL programming language.
	LanguageAMPL
	// LanguageAngelScript represents the AngelScript programming language.
	LanguageAngelScript
	// LanguageAngular2 represents the Angular2 programming language.
	LanguageAngular2
	// LanguageAnsible represents the Ansible programming language.
	LanguageAnsible
	// LanguageAntBuildSystem represents the AntBuildSystem programming language.
	LanguageAntBuildSystem
	// LanguageANTLR represents the ANTLR programming language.
	LanguageANTLR
	// LanguageApacheConfig represents the Apache Config programming language.
	LanguageApacheConfig
	// LanguageApex represents the Apex programming language.
	LanguageApex
	// LanguageAPIBlueprint represents the APIBlueprint programming language.
	LanguageAPIBlueprint
	// LanguageAPL represents the APL programming language.
	LanguageAPL
	// LanguageApolloGuidanceComputer represents the ApolloGuidanceComputer programming language.
	LanguageApolloGuidanceComputer
	// LanguageAppleScript represents the AppleScript programming language.
	LanguageAppleScript
	// LanguageArangoDBQueryLanguage represents the ArangoDB Query Language programming language.
	LanguageArangoDBQueryLanguage
	// LanguageArc represents the Arc programming language.
	LanguageArc
	// LanguageArduino represents the Arduino programming language.
	LanguageArduino
	// LanguageArmAsm represents the ArmAsm programming language.
	LanguageArmAsm
	// LanguageArrow represents the Arrow programming language.
	LanguageArrow
	// LanguageArturo represents the Arturo programming language.
	LanguageArturo
	// LanguageASCIIDoc represents the ASCIIDoc programming language.
	LanguageASCIIDoc
	// LanguageASL represents the ASL programming language.
	LanguageASL
	// LanguageASN1 represents the ASN1 programming language.
	LanguageASN1
	// LanguageASPDotNet represents the ASPDotNet programming language.
	LanguageASPDotNet
	// LanguageAspectJ represents the AspectJ programming language.
	LanguageAspectJ
	// LanguageAspxCSharp represents the Aspx C# programming language.
	LanguageAspxCSharp
	// LanguageAspxVBNet represents the Aspx VB.Net programming language.
	LanguageAspxVBNet
	// LanguageAssembly represents the Assembly programming language.
	LanguageAssembly
	// LanguageAstro represents the Astro programming language.
	LanguageAstro
	// LanguageAsymptote represents the Asymptote programming language.
	LanguageAsymptote
	// LanguageATL represents the ATL programming language.
	LanguageATL
	// LanguageATS represents the ATS programming language.
	LanguageATS
	// LanguageAugeas represents the Augeas programming language.
	LanguageAugeas
	// LanguageAutoconf represents the Autoconf programming language.
	LanguageAutoconf
	// LanguageAutoHotkey represents the AutoHotkey programming language.
	LanguageAutoHotkey
	// LanguageAutoIt represents the AutoIt programming language.
	LanguageAutoIt
	// LanguageAvroIDL represents the AvroIDL programming language.
	LanguageAvroIDL
	// LanguageAwk represents the Awk programming language.
	LanguageAwk
	// LanguageBallerina represents the Ballerina programming language.
	LanguageBallerina
	// LanguageBARE represents the BARE programming language.
	LanguageBARE
	// LanguageBash represents the Bash programming language.
	LanguageBash
	// LanguageBashSession represents the Bash Session programming language.
	LanguageBashSession
	// LanguageBasic represents the Basic programming language.
	LanguageBasic
	// LanguageBatchfile represents the Batchfile programming language.
	LanguageBatchfile
	// LanguageBBCBasic represents the BBCBasic programming language.
	LanguageBBCBasic
	// LanguageBBCode represents the BBCode programming language.
	LanguageBBCode
	// LanguageBC represents the BC programming language.
	LanguageBC
	// LanguageBeef represents the Beef programming language.
	LanguageBeef
	// LanguageBefunge represents the Befunge programming language.
	LanguageBefunge
	// LanguageBibTeX represents the BibTeX programming language.
	LanguageBibTeX
	// LanguageBicep represents the Bicep programming language.
	LanguageBicep
	// LanguageBison represents the Bison programming language.
	LanguageBison
	// LanguageBitBake represents the BitBake programming language.
	LanguageBitBake
	// LanguageBlade represents the Blade programming language.
	LanguageBlade
	// LanguageBladeTemplate represents the BladeTemplate programming language.
	LanguageBladeTemplate
	// LanguageBlazor represent the Blazor programming language.
	LanguageBlazor
	// LanguageBlitzBasic represents the BlitzBasic programming language.
	LanguageBlitzBasic
	// LanguageBlitzMax represents the BlitzMax programming language.
	LanguageBlitzMax
	// LanguageBluespec represents the Bluespec programming language.
	LanguageBluespec
	// LanguageBNF represents the BNF programming language.
	LanguageBNF
	// LanguageBoa represents the Boa programming language.
	LanguageBoa
	// LanguageBoo represents the Boo programming language.
	LanguageBoo
	// LanguageBoogie represents the Boogie programming language.
	LanguageBoogie
	// LanguageBQN represents the BQN programming language.
	LanguageBQN
	// LanguageBrainfuck represents the Brainfuck programming language.
	LanguageBrainfuck
	// LanguageBrightScript represents the BrightScript programming language.
	LanguageBrightScript
	// LanguageBro represents the Bro programming language.
	LanguageBro
	// LanguageBrowserslist represents the Browserslist programming language.
	LanguageBrowserslist
	// LanguageBST represents the BST programming language.
	LanguageBST
	// LanguageBUGS represents the BUGS programming language.
	LanguageBUGS
	// LanguageC represents the C programming language.
	LanguageC
	// LanguageC2hsHaskell represents the C2hsHaskell programming language.
	LanguageC2hsHaskell
	// LanguageC3 represents the C3 programming language.
	LanguageC3
	// LanguageCa65Assembler represents the ca65 assembler programming language.
	LanguageCa65Assembler
	// LanguageCabalConfig represents the CabalConfig programming language.
	LanguageCabalConfig
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
	// LanguageCartoCSS represents the CartoCSS programming language.
	LanguageCartoCSS
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
	// LanguageCharity represents the Charity programming language.
	LanguageCharity
	// LanguageCharmci represents the Charmci programming language.
	LanguageCharmci
	// LanguageCheetah represents the Cheetah programming language.
	LanguageCheetah
	// LanguageChucK represents the ChucK programming language.
	LanguageChucK
	// LanguageCirru represents the Cirru programming language.
	LanguageCirru
	// LanguageClarion represents the Clarion programming language.
	LanguageClarion
	// LanguageClassicASP represents the ClassicASP programming language.
	LanguageClassicASP
	// LanguageClay represents the Clay programming language.
	LanguageClay
	// LanguageClean represents the Clean programming language.
	LanguageClean
	// LanguageClick represents the Click programming language.
	LanguageClick
	// LanguageCLIPS represents the CLIPS programming language.
	LanguageCLIPS
	// LanguageClojure represents the Clojure programming language.
	LanguageClojure
	// LanguageClojureScript represents the ClojureScript programming language.
	LanguageClojureScript
	// LanguageClosureTemplates represents the ClosureTemplates programming language.
	LanguageClosureTemplates
	// LanguageCloudFirestoreSecurityRules represents the CloudFirestoreSecurityRules programming language.
	LanguageCloudFirestoreSecurityRules
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
	// LanguageCodeQL represents the CodeQL programming language.
	LanguageCodeQL
	// LanguageCoffeeScript represents the CoffeeScript programming language.
	LanguageCoffeeScript
	// LanguageColdfusionCFC represents the ColdfusionCFC programming language.
	LanguageColdfusionCFC
	// LanguageColdfusionHTML represents the ColdfusionHTML programming language.
	LanguageColdfusionHTML
	// LanguageCOLLADA represents the COLLADA programming language.
	LanguageCOLLADA
	// LanguageCommonLisp represents the CommonLisp programming language.
	LanguageCommonLisp
	// LanguageCommonWorkflowLanguage represents the CommonWorkflowLanguage programming language.
	LanguageCommonWorkflowLanguage
	// LanguageComponentPascal represents the ComponentPascal programming language.
	LanguageComponentPascal
	// LanguageConfig represents the Config programming language.
	LanguageConfig
	// LanguageCoNLLU represents the CoNLLU programming language.
	LanguageCoNLLU
	// LanguageCool represents the Cool programming language.
	LanguageCool
	// LanguageCoq represents the Coq programming language.
	LanguageCoq
	// LanguageCore represents the Core programming language.
	LanguageCore
	// LanguageCPerl represents the CPerl programming language.
	LanguageCPerl
	// LanguageCPP represents the CPP programming language.
	LanguageCPP
	// LanguageCppObjdump represents the CppObjdump programming language.
	LanguageCppObjdump
	// LanguageCPSA represents the CPSA programming language.
	LanguageCPSA
	// LanguageCreole represents the Creole programming language.
	LanguageCreole
	// LanguageCrmsh represents the Crmsh programming language.
	LanguageCrmsh
	// LanguageCroc represents the Croc programming language.
	LanguageCroc
	// LanguageCrontab represents the Crontab programming language.
	LanguageCrontab
	// LanguageCryptol represents the Cryptol programming language.
	LanguageCryptol
	// LanguageCSharp represents the CSharp programming language.
	LanguageCSharp
	// LanguageCSHTML represents the CSHTML programming language.
	LanguageCSHTML
	// LanguageCrystal represents the Crystal programming language.
	LanguageCrystal
	// LanguageCSON represents the CSON programming language.
	LanguageCSON
	// LanguageCsound represents the Csound programming language.
	LanguageCsound
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
	// LanguageCUE represents the CUE programming language.
	LanguageCUE
	// LanguagecURLConfig represents the cURLConfig programming language.
	LanguagecURLConfig
	// LanguageCVS represents the CVS programming language.
	LanguageCVS
	// LanguageCWeb represents the CWeb programming language.
	LanguageCWeb
	// LanguageCycript represents the Cycript programming language.
	LanguageCycript
	// LanguageCypher represents the Cypher programming language.
	LanguageCypher
	// LanguageCython represents the Cython programming language.
	LanguageCython
	// LanguageD represents the D programming language.
	LanguageD
	// LanguageDafny represents the Dafny programming language.
	LanguageDafny
	// LanguageDarcsPatch represents the DarcsPatch programming language.
	LanguageDarcsPatch
	// LanguageDart represents the Dart programming language.
	LanguageDart
	// LanguageDataWeave represents the DataWeave programming language.
	LanguageDataWeave
	// LanguageDASM16 represents the DASM16 programming language.
	LanguageDASM16
	// LanguageDax represents the Dax programming language.
	LanguageDax
	// LanguageDCL represents the DCL programming language.
	LanguageDCL
	// LanguageDCPU16Asm represents the DCPU16Asm programming language.
	LanguageDCPU16Asm
	// LanguageDebianControlFile represents the DebianControlFile programming language.
	LanguageDebianControlFile
	// LanguageDelphi represents the Delphi programming language.
	LanguageDelphi
	// LanguageDesktopFile represents the Desktop file programming language.
	LanguageDesktopFile
	// LanguageDevicetree represents the Devicetree programming language.
	LanguageDevicetree
	// LanguageDG represents the DG programming language.
	LanguageDG
	// LanguageDhall represents the Dhall programming language.
	LanguageDhall
	// LanguageDiff represents the Diff programming language.
	LanguageDiff
	// LanguageDigitalCommand represents the DIGITAL Command Language programming language.
	LanguageDigitalCommand
	// LanguageDircolors represents the dircolors programming language.
	LanguageDircolors
	// LanguageDirectX3DFile represents the DirectX 3D File programming language.
	LanguageDirectX3DFile
	// LanguageDjangoJinja represents the DjangoJinja programming language.
	LanguageDjangoJinja
	// LanguageDM represents the DM programming language.
	LanguageDM
	// LanguageDNSZone represents the DNS Zone programming language.
	LanguageDNSZone
	// LanguageDObjdump represents the DObjdump programming language.
	LanguageDObjdump
	// LanguageDocker represents the Docker programming language.
	LanguageDocker
	// LanguageDockerfile represents the Dockerfile programming language.
	LanguageDockerfile
	// LanguageDocTeX represents the DocTeX programming language.
	LanguageDocTeX
	// LanguageDocumentation represents the Documentation programming language.
	LanguageDocumentation
	// LanguageDogescript represents the Dogescript programming language.
	LanguageDogescript
	// LanguageDTD represents the DTD programming language.
	LanguageDTD
	// LanguageDTrace represents the DTrace programming language.
	LanguageDTrace
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
	// LanguageE represents the E programming language.
	LanguageE
	// LanguageEagle represents the Eagle programming language.
	LanguageEagle
	// LanguageEMail represents the EMail programming language.
	LanguageEMail
	// LanguageEarlGrey represents the EarlGrey programming language.
	LanguageEarlGrey
	// LanguageEasybuild represents the Easybuild programming language.
	LanguageEasybuild
	// LanguageEasytrieve represents the Easytrieve programming language.
	LanguageEasytrieve
	// LanguageEBNF represents the EBNF programming language.
	LanguageEBNF
	// LanguageEC represents the EC programming language.
	LanguageEC
	// LanguageEcereProjects represents the Ecere Projects programming language.
	LanguageEcereProjects
	// LanguageECL represents the ECL programming language.
	LanguageECL
	// LanguageEclipse represents the ECLiPSe programming language.
	LanguageEclipse
	// LanguageEditorConfig represents the EditorConfig programming language.
	LanguageEditorConfig
	// LanguageEdjeDataCollection represents the Edje Data Collection programming language.
	LanguageEdjeDataCollection
	// LanguageEdn represents the edn programming language.
	LanguageEdn
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
	// LanguageEmberScript represents the EmberScript programming language.
	LanguageEmberScript
	// LanguageEML represents the EML programming language.
	LanguageEML
	// LanguageEQ represents the EQ programming language.
	LanguageEQ
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
	// LanguageFaust represents the Faust programming language.
	LanguageFaust
	// LanguageFelix represents the Felix programming language.
	LanguageFelix
	// LanguageFennel represents the Fennel programming language.
	LanguageFennel
	// LanguageFIGletFont represents the FIGlet Font programming language.
	LanguageFIGletFont
	// LanguageFilebenchWML represents the Filebench WML programming language.
	LanguageFilebenchWML
	// LanguageFilterscript represents the Filterscript programming language.
	LanguageFilterscript
	// LanguageFlatline represents the Flatline programming language.
	LanguageFlatline
	// LanguageFloScript represents the FloScript programming language.
	LanguageFloScript
	// LanguageFish represents the Fish programming language.
	LanguageFish
	// LanguageFLUX represents the FLUX programming language.
	LanguageFLUX
	// LanguageFont represents the Font programming language.
	LanguageFont
	// LanguageFormatted represents the Formatted programming language.
	LanguageFormatted
	// LanguageForth represents the Forth programming language.
	LanguageForth
	// LanguageFortran represents the Fortran programming language.
	LanguageFortran
	// LanguageFortranFixed represents the FortranFixed programming language.
	LanguageFortranFixed
	// LanguageFortranFreeForm represents the Fortran Free Form programming language.
	LanguageFortranFreeForm
	// LanguageFSharp represents the FSharp programming language.
	LanguageFSharp
	// LanguageFoxPro represents the FoxPro programming language.
	LanguageFoxPro
	// LanguageFreefem represents the Freefem programming language.
	LanguageFreefem
	// LanguageFreeMarker represents the FreeMarker programming language.
	LanguageFreeMarker
	// LanguageFrege represents the Frege programming language.
	LanguageFrege
	// LanguageFStar represents the F* programming language.
	LanguageFStar
	// LanguageFuthark represents the Futhark programming language.
	LanguageFuthark
	// LanguageGameMakerLanguage represents the GameMakerLanguage programming language.
	LanguageGameMakerLanguage
	// LanguageGAML represents the GAML programming language.
	LanguageGAML
	// LanguageGAMS represents the GAMS programming language.
	LanguageGAMS
	// LanguageGap represents the Gap programming language.
	LanguageGap
	// LanguageGas represents the Gas programming language.
	LanguageGas
	// LanguageGCCMachineDescription represents the GCCMachineDescription programming language.
	LanguageGCCMachineDescription
	// LanguageGCode represents the GCode programming language.
	LanguageGCode
	// LanguageGDB represents the GDB programming language.
	LanguageGDB
	// LanguageGDNative represents the GDNative programming language.
	LanguageGDNative
	// LanguageGDScript represents the GDScript programming language.
	LanguageGDScript
	// LanguageGDScript3 represents the GDScript3 programming language.
	LanguageGDScript3
	// LanguageGEDCOM represents the GEDCOM programming language.
	LanguageGEDCOM
	// LanguageGemfileLock represents the Gemfile.lock programming language.
	LanguageGemfileLock
	// LanguageGemtext represents the Gemtext programming language.
	LanguageGemtext
	// LanguageGenie represents the Genie programming language.
	LanguageGenie
	// LanguageGenshi represents the Genshi programming language.
	LanguageGenshi
	// LanguageGenshiHTML represents the Genshi HTML programming language.
	LanguageGenshiHTML
	// LanguageGenshiText represents the Genshi Text programming language.
	LanguageGenshiText
	// LanguageGentooEbuild represents the GentooEbuild programming language.
	LanguageGentooEbuild
	// LanguageGentooEclass represents the GentooEclass programming language.
	LanguageGentooEclass
	// LanguageGerberImage represents the GerberImage programming language.
	LanguageGerberImage
	// LanguageGettext represents the Gettext programming language.
	LanguageGettext
	// LanguageGettextCatalog represents the Gettext Catalog programming language.
	LanguageGettextCatalog
	// LanguageGherkin represents the Gherkin programming language.
	LanguageGherkin
	// LanguageGit represents the Git programming language.
	LanguageGit
	// LanguageGitAttributes represents the GitAttributes programming language.
	LanguageGitAttributes
	// LanguageGitConfig represents the Git Config programming language.
	LanguageGitConfig
	// LanguageGleam represents the Gleam programming language.
	LanguageGleam
	// LanguageGLSL represents the GLSL programming language.
	LanguageGLSL
	// LanguageGlyph represents the Glyph programming language.
	LanguageGlyph
	// LanguageGlyphBitmap represents the GlyphBitmap programming language.
	LanguageGlyphBitmap
	// LanguageGN represents the GN programming language.
	LanguageGN
	// LanguageGnuplot represents the Gnuplot programming language.
	LanguageGnuplot
	// LanguageGo represents the Go programming language.
	LanguageGo
	// LanguageGoHTMLTemplate represents the Go HTML Template programming language.
	LanguageGoHTMLTemplate
	// LanguageGoTextTemplate represents the Go Text Template programming language.
	LanguageGoTextTemplate
	// LanguageGoTemplate represents the Go Template programming language.
	LanguageGoTemplate
	// LanguageGolo represents the Golo programming language.
	LanguageGolo
	// LanguageGoodDataCL represents the GoodData-CL programming language.
	LanguageGoodDataCL
	// LanguageGosu represents the Gosu programming language.
	LanguageGosu
	// LanguageGosuTemplate represents the Gosu Template programming language.
	LanguageGosuTemplate
	// LanguageGrace represents the Grace programming language.
	LanguageGrace
	// LanguageGradle represents the Gradle programming language.
	LanguageGradle
	// LanguageGradleConfig represents the GradleConfig programming language.
	LanguageGradleConfig
	// LanguageGrammaticalFramework represents the GrammaticalFramework programming language.
	LanguageGrammaticalFramework
	// LanguageGraphModelingLanguage represents the GraphModelingLanguage programming language.
	LanguageGraphModelingLanguage
	// LanguageGraphQL represents the GraphQL programming language.
	LanguageGraphQL
	// LanguageGraphvizDOT represents the GraphvizDOT programming language.
	LanguageGraphvizDOT
	// LanguageGroff represents the Groff programming language.
	LanguageGroff
	// LanguageGroovy represents the Groovy programming language.
	LanguageGroovy
	// LanguageGroovyServerPages represents the GroovyServerPages programming language.
	LanguageGroovyServerPages
	// LanguageHack represents the Hack programming language.
	LanguageHack
	// LanguageHaml represents the Haml programming language.
	LanguageHaml
	// LanguageHandlebars represents the Handlebars programming language.
	LanguageHandlebars
	// LanguageHAProxy represents the HAProxy programming language.
	LanguageHAProxy
	// LanguageHarbour represents the Harbour programming language.
	LanguageHarbour
	// LanguageHare represents the Hare programming language.
	LanguageHare
	// LanguageHaskell represents the Haskell programming language.
	LanguageHaskell
	// LanguageHaxe represents the Haxe programming language.
	LanguageHaxe
	// LanguageHCL represents the HCL programming language.
	LanguageHCL
	// LanguageHexdump represents the Hexdump programming language.
	LanguageHexdump
	// LanguageHiveQL represents the HiveQL programming language.
	LanguageHiveQL
	// LanguageHLB represents the HLB programming language.
	LanguageHLB
	// LanguageHLSL represents the HLSL programming language.
	LanguageHLSL
	// LanguageHolyC represents the HolyC programming language.
	LanguageHolyC
	// LanguageHSAIL represents the HSAIL programming language.
	LanguageHSAIL
	// LanguageHspec represents the Hspec programming language.
	LanguageHspec
	// LanguageHTML represents the HTML programming language.
	LanguageHTML
	// LanguageHTMLDjango represents the HTMLDjango programming language.
	LanguageHTMLDjango
	// LanguageHTMLECR represents the HTMLECR programming language.
	LanguageHTMLECR
	// LanguageHTMLEEX represents the HTMLEEX programming language.
	LanguageHTMLEEX
	// LanguageHTMLERB represents the HTMLERB programming language.
	LanguageHTMLERB
	// LanguageHTMLPHP represents the HTMLPHP programming language.
	LanguageHTMLPHP
	// LanguageHTMLRazor represents the HTMLRazor programming language.
	LanguageHTMLRazor
	// LanguageHTTP represents the HTTP programming language.
	LanguageHTTP
	// LanguageHxml represents the Hxml programming language.
	LanguageHxml
	// LanguageHy represents the Hy programming language.
	LanguageHy
	// LanguageHybris represents the Hybris programming language.
	LanguageHybris
	// LanguageHyPhy represents the HyPhy programming language.
	LanguageHyPhy
	// LanguageIcon represents the Icon programming language.
	LanguageIcon
	// LanguageIDA represents the IDA programming language.
	LanguageIDA
	// LanguageIDL represents the IDL programming language.
	LanguageIDL
	// LanguageIdris represents the Idris programming language.
	LanguageIdris
	// LanguageIgnoreList represents the IgnoreList programming language.
	LanguageIgnoreList
	// LanguageIgor represents the Igor programming language.
	LanguageIgor
	// LanguageIGORPro represents the IGORPro programming language.
	LanguageIGORPro
	// LanguageImageJMacro represents the ImageJMacro programming language.
	LanguageImageJMacro
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
	// LanguageInnoSetup represents the InnoSetup programming language.
	LanguageInnoSetup
	// LanguageIo represents the Io programming language.
	LanguageIo
	// LanguageIoke represents the Ioke programming language.
	LanguageIoke
	// LanguageIRCLogs represents the IRC Logs programming language.
	LanguageIRCLogs
	// LanguageIsabelle represents the Isabelle programming language.
	LanguageIsabelle
	// LanguageIsabelleRoot represents the IsabelleRoot programming language.
	LanguageIsabelleRoot
	// LanguageISCdhcpd represents the ISC dhcpd programming language.
	LanguageISCdhcpd
	// LanguageJ represents the J programming language.
	LanguageJ
	// LanguageJade represents the Jade programming language.
	LanguageJade
	// LanguageJAGS represents the JAGS programming language.
	LanguageJAGS
	// LanguageJanet represents the Janet programming language.
	LanguageJanet
	// LanguageJasmin represents the Jasmin programming language.
	LanguageJasmin
	// LanguageJava represents the Java programming language.
	LanguageJava
	// LanguageJavaProperties represents the Java Properties programming language.
	LanguageJavaProperties
	// LanguageJavaScript represents the JavaScript programming language.
	LanguageJavaScript
	// LanguageJavaScriptERB represents the JavaScriptERB programming language.
	LanguageJavaScriptERB
	// LanguageJCL represents the JCL programming language.
	LanguageJCL
	// LanguageJFlex represents the JFlex programming language.
	LanguageJFlex
	// LanguageJison represents the Jison programming language.
	LanguageJison
	// LanguageJisonLex represents the JisonLex programming language.
	LanguageJisonLex
	// LanguageJolie represents the Jolie programming language.
	LanguageJolie
	// LanguageJSGF represents the JSGF programming language.
	LanguageJSGF
	// LanguageJSON represents the JSON programming language.
	LanguageJSON
	// LanguageJSON5 represents the JSON5 programming language.
	LanguageJSON5
	// LanguageJSONata represents the JSONata programming language.
	LanguageJSONata
	// LanguageJSONiq represents the JSONiq programming language.
	LanguageJSONiq
	// LanguageJSONLD represents the JSON-LD programming language.
	LanguageJSONLD
	// LanguageJsonnet represents the Jsonnet programming language.
	LanguageJsonnet
	// LanguageJSONWithComments represents the JSONWithComments programming language.
	LanguageJSONWithComments
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
	// LanguageJupyterNotebook represents the JupyterNotebook programming language.
	LanguageJupyterNotebook
	// LanguageJuttle represents the Juttle programming language.
	LanguageJuttle
	// LanguageKaitai represent the Kaitai programming language.
	LanguageKaitai
	// LanguageKakoune represents the Kakoune programming language.
	LanguageKakoune
	// LanguageKal represents the Kal programming language.
	LanguageKal
	// LanguageKconfig represents the Kconfig programming language.
	LanguageKconfig
	// LanguageKDL represents the KDL programming language.
	LanguageKDL
	// LanguageKernelLog represents the Kernel Log programming language.
	LanguageKernelLog
	// LanguageKiCadLayout represent the KiCadLayout programming language.
	LanguageKiCadLayout
	// LanguageKiCadLegacyLayout represent the KiCadLegacyLayout programming language.
	LanguageKiCadLegacyLayout
	// LanguageKiCadSchematic represent the KiCadSchematic programming language.
	LanguageKiCadSchematic
	// LanguageKit represent the Kit programming language.
	LanguageKit
	// LanguageKoka represents the Koka programming language.
	LanguageKoka
	// LanguageKotlin represents the Kotlin programming language.
	LanguageKotlin
	// LanguageKRL represent the KRL programming language.
	LanguageKRL
	// LanguageLabVIEW represents the LabVIEW programming language.
	LanguageLabVIEW
	// LanguageLaravelTemplate represents the Laravel Template programming language.
	LanguageLaravelTemplate
	// LanguageLark represents the Lark programming language.
	LanguageLark
	// LanguageLasso represents the Lasso programming language.
	LanguageLasso
	// LanguageLateralus represents the Lateralus programming language.
	LanguageLateralus
	// LanguageLaTeX represents the LaTeX programming language.
	LanguageLaTeX
	// LanguageLatte represents the Latte programming language.
	LanguageLatte
	// LanguageLean represents the Lean programming language.
	LanguageLean
	// LanguageLean4 represents the Lean4 programming language.
	LanguageLean4
	// LanguageLess represents the Less programming language.
	LanguageLess
	// LanguageLex represents the Lex programming language.
	LanguageLex
	// LanguageLFE represents the LFE programming language.
	LanguageLFE
	// LanguageLighttpd represents the Lighttpd programming language.
	LanguageLighttpd
	// LanguageLilyPond represents the LilyPond programming language.
	LanguageLilyPond
	// LanguageLimbo represents the Limbo programming language.
	LanguageLimbo
	// LanguageLinkerScript represents the LinkerScript programming language.
	LanguageLinkerScript
	// LanguageLinuxKernelModule represents the LinuxKernelModule programming language.
	LanguageLinuxKernelModule
	// LanguageLiquid represents the Liquid programming language.
	LanguageLiquid
	// LanguageLiterateAgda represents the Literate Agda programming language.
	LanguageLiterateAgda
	// LanguageLiterateCoffeeScript represents the LiterateCoffeeScript programming language.
	LanguageLiterateCoffeeScript
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
	// LanguageLOLCODE represents the LOLCODE programming language.
	LanguageLOLCODE
	// LanguageLookML represents the LookML programming language.
	LanguageLookML
	// LanguageLoomScript represents the LoomScript programming language.
	LanguageLoomScript
	// LanguageLox represents the lox programming language.
	LanguageLox
	// LanguageLSL represents the LSL programming language.
	LanguageLSL
	// LanguageLTspiceSymbol represents the LTspiceSymbol programming language.
	LanguageLTspiceSymbol
	// LanguageLua represents the Lua programming language.
	LanguageLua
	// LanguageLuau represents the Luau programming language.
	LanguageLuau
	// LanguageMakefile represents the Makefile programming language.
	LanguageMakefile
	// LanguageMako represents the Mako programming language.
	LanguageMako
	// LanguageMan represents the Man programming language.
	LanguageMan
	// LanguageMAQL represents the MAQL programming language.
	LanguageMAQL
	// LanguageMarkdown represents the Markdown programming language.
	LanguageMarkdown
	// LanguageMarkless represents the Markless programming language.
	LanguageMarkless
	// LanguageMarko represents the Marko programming language.
	LanguageMarko
	// LanguageMask represents the Mask programming language.
	LanguageMask
	// LanguageMason represents the Mason programming language.
	LanguageMason
	// LanguageMaterializeSQLDialect represents the Materialize SQL dialect programming language.
	LanguageMaterializeSQLDialect
	// LanguageMathematica represents the Mathematica programming language.
	LanguageMathematica
	// LanguageMatlab represents the Matlab programming language.
	LanguageMatlab
	// LanguageMatlabSession represents the MatlabSession programming language.
	LanguageMatlabSession
	// LanguageMax represents the Max programming language.
	LanguageMax
	// LanguageMaxMSP represents the MaxMSP programming language.
	LanguageMaxMSP
	// LanguageMCFunction represents the MCFunction programming language.
	LanguageMCFunction
	// LanguageMDX represents the MDX programming language.
	LanguageMDX
	// LanguageMeson represents the Meson programming language.
	LanguageMeson
	// LanguageMetafont represents the Metafont programming language.
	LanguageMetafont
	// LanguageMetal represents the Metal programming language.
	LanguageMetal
	// LanguageMetapost represents the Metapost programming language.
	LanguageMetapost
	// LanguageMicrocad represents the microcad programming language.
	LanguageMicrocad
	// LanguageMIME represents the MIME programming language.
	LanguageMIME
	// LanguageMiniD represents the MiniD programming language.
	LanguageMiniD
	// LanguageMiniScript represents the MiniScript programming language.
	LanguageMiniScript
	// LanguageMiniZinc represents the MiniZinc programming language.
	LanguageMiniZinc
	// LanguageMirah represents the Mirah programming language.
	LanguageMirah
	// LanguageMLIR represents the MLIR programming language.
	LanguageMLIR
	// LanguageModelica represents the Modelica programming language.
	LanguageModelica
	// LanguageModula2 represents the Modula2 programming language.
	LanguageModula2
	// LanguageMoinWiki represents the MoinWiki programming language.
	LanguageMoinWiki
	// LanguageMojo represents the Mojo programming language.
	LanguageMojo
	// LanguageMonkey represents the Monkey programming language.
	LanguageMonkey
	// LanguageMonkeyC represents the MonkeyC programming language.
	LanguageMonkeyC
	// LanguageMonte represents the Monte programming language.
	LanguageMonte
	// LanguageMOOCode represents the MOOCode programming language.
	LanguageMOOCode
	// LanguageMoonBit represents the MoonBit programming language.
	LanguageMoonBit
	// LanguageMoonScript represents the MoonScript programming language.
	LanguageMoonScript
	// LanguageMorrowindScript represents the MorrowindScript programming language.
	LanguageMorrowindScript
	// LanguageMosel represents the Mosel programming language.
	LanguageMosel
	// LanguageMozPreprocHash represents the MozPreprocHash programming language.
	LanguageMozPreprocHash
	// LanguageMozPreprocPercent represents the MozPreprocPercent programming language.
	LanguageMozPreprocPercent
	// LanguageMQL represents the MQL programming language.
	LanguageMQL
	// LanguageMscgen represents the Mscgen programming language.
	LanguageMscgen
	// LanguageMSDOSSession represents the MSDOSSession programming language.
	LanguageMSDOSSession
	// LanguageMuPAD represents the MuPAD programming language.
	LanguageMuPAD
	// LanguageMXML represents the MXML programming language.
	LanguageMXML
	// LanguageMyghty represents the Myghty programming language.
	LanguageMyghty
	// LanguageMySQL represents the MySQL programming language.
	LanguageMySQL
	// LanguageMustache represents the Mustache programming language.
	LanguageMustache
	// LanguageNASM represents the NASM programming language.
	LanguageNASM
	// LanguageNASMObjdump represents the NASMObjdump programming language.
	LanguageNASMObjdump
	// LanguageNatural represents the Natural programming language.
	LanguageNatural
	// LanguageNCL represents the NCL programming language.
	LanguageNCL
	// LanguageNDISASM represents the NDISASM programming language.
	LanguageNDISASM
	// LanguageNemerle represents the Nemerle programming language.
	LanguageNemerle
	// LanguageNeon represents the Neon programming language.
	LanguageNeon
	// LanguageNesC represents the NesC programming language.
	LanguageNesC
	// LanguageNewLisp represents the NewLisp programming language.
	LanguageNewLisp
	// LanguageNewspeak represents the Newspeak programming language.
	LanguageNewspeak
	// LanguageNginx represents the Nginx programming language.
	LanguageNginx
	// LanguageNginxConfig represents the NginxConfig programming language.
	LanguageNginxConfig
	// LanguageNimrod represents the Nimrod programming language.
	LanguageNimrod
	// LanguageNit represents the Nit programming language.
	LanguageNit
	// LanguageNix represents the Nix programming language.
	LanguageNix
	// LanguageNotmuch represents the Notmuch programming language.
	LanguageNotmuch
	// LanguageNSIS represents the NSIS programming language.
	LanguageNSIS
	// LanguageNu represents the Nu programming language.
	LanguageNu
	// LanguageNumPy represents the NumPy programming language.
	LanguageNumPy
	// LanguageNushell represents the Nushell programming language.
	LanguageNushell
	// LanguageNuSMV represents the NuSMV programming language.
	LanguageNuSMV
	// LanguageObjdump represents the Objdump programming language.
	LanguageObjdump
	// LanguageObjectPascal represents the ObjectPascal programming language.
	LanguageObjectPascal
	// LanguageObjectiveC represents the ObjectiveC programming language.
	LanguageObjectiveC
	// LanguageObjectiveCPP represents the ObjectiveC++ programming language.
	LanguageObjectiveCPP
	// LanguageObjectiveJ represents the ObjectiveJ programming language.
	LanguageObjectiveJ
	// LanguageOCaml represents the OCaml programming language.
	LanguageOCaml
	// LanguageOctave represents the Octave programming language.
	LanguageOctave
	// LanguageODIN represents the ODIN programming language.
	LanguageODIN
	// LanguageOnesEnterprise represents the OnesEnterprise programming language.
	LanguageOnesEnterprise
	// LanguageOoc represents the Ooc programming language.
	LanguageOoc
	// LanguageOpa represents the Opa programming language.
	LanguageOpa
	// LanguageOpenEdgeABL represents the OpenEdgeABL programming language.
	LanguageOpenEdgeABL
	// LanguageOpenSCAD represents the OpenSCAD programming language.
	LanguageOpenSCAD
	// LanguageOrg represents the Org programming language.
	LanguageOrg
	// LanguagePacmanConf represents the PacmanConf programming language.
	LanguagePacmanConf
	// LanguagePan represents the Pan programming language.
	LanguagePan
	// LanguageParaSail represents the ParaSail programming language.
	LanguageParaSail
	// LanguageParrot represents the Parrot programming language.
	LanguageParrot
	// LanguagePascal represents the Pascal programming language.
	LanguagePascal
	// LanguagePawn represents the Pawn programming language.
	LanguagePawn
	// LanguagePEG represents the PEG programming language.
	LanguagePEG
	// LanguagePerl represents the Perl programming language.
	LanguagePerl
	// LanguagePerl6 represents the Perl6 programming language.
	LanguagePerl6
	// LanguagePHP represents the PHP programming language.
	LanguagePHP
	// LanguagePHTML represents the PHTML programming language.
	LanguagePHTML
	// LanguagePig represents the Pig programming language.
	LanguagePig
	// LanguagePike represents the Pike programming language.
	LanguagePike
	// LanguagePkgConfig represents the PkgConfig programming language.
	LanguagePkgConfig
	// LanguagePLpgSQL represents the PLpgSQL programming language.
	LanguagePLpgSQL
	// LanguagePlutusCore represents the Plutus Core programming language.
	LanguagePlutusCore
	// LanguagePointless represents the Pointless programming language.
	LanguagePointless
	// LanguagePony represents the Pony programming language.
	LanguagePony
	// LanguagePostgres represents the Postgres programming language.
	LanguagePostgres
	// LanguagePostgresConsole represents the PostgresConsole programming language.
	LanguagePostgresConsole
	// LanguagePostScript represents the PostScript programming language.
	LanguagePostScript
	// LanguagePOVRay represents the POVRay programming language.
	LanguagePOVRay
	// LanguagePowerQuery represents the PowerQuery programming language.
	LanguagePowerQuery
	// LanguagePowerShell represents the PowerShell programming language.
	LanguagePowerShell
	// LanguagePowerShellSession represents the PowerShellSession programming language.
	LanguagePowerShellSession
	// LanguagePraat represents the Praat programming language.
	LanguagePraat
	// LanguageProcessing represents the Processing programming language.
	LanguageProcessing
	// LanguageProlog represents the Prolog programming language.
	LanguageProlog
	// LanguagePromela represents the Promela programming language.
	LanguagePromela
	// LanguagePromQL represents the PromQL programming language.
	LanguagePromQL
	// LanguageProtocolBuffer represents the ProtocolBuffer programming language.
	LanguageProtocolBuffer
	// LanguagePRQL represents the PRQL programming language.
	LanguagePRQL
	// LanguagePSL represents the Property Specification Language programming language.
	LanguagePSL
	// LanguagePsyShPHP represents the PsySH PHP programming language.
	LanguagePsyShPHP
	// LanguagePug represents the Pug programming language.
	LanguagePug
	// LanguagePuppet represents the Puppet programming language.
	LanguagePuppet
	// LanguagePureData represents the PureData programming language.
	LanguagePureData
	// LanguagePureScript represents the PureScript programming language.
	LanguagePureScript
	// LanguagePyPyLog represents the PyPyLog programming language.
	LanguagePyPyLog
	// LanguagePython represents the Python programming language.
	LanguagePython
	// LanguagePython2 represents the Python2 programming language.
	LanguagePython2
	// LanguagePython2Traceback represents the Python2Traceback programming language.
	LanguagePython2Traceback
	// LanguagePythonConsole represents the PythonConsole programming language.
	LanguagePythonConsole
	// LanguagePythonTraceback represents the PythonTraceback programming language.
	LanguagePythonTraceback
	// LanguageQBasic represents the QBasic programming language.
	LanguageQBasic
	// LanguageQML represents the QML programming language.
	LanguageQML
	// LanguageQVTO represents the QVTO programming language.
	LanguageQVTO
	// LanguageR represents the R programming language.
	LanguageR
	// LanguageRacket represents the Racket programming language.
	LanguageRacket
	// LanguageRagel represents the Ragel programming language.
	LanguageRagel
	// LanguageRagelEmbedded represents the RagelEmbedded programming language.
	LanguageRagelEmbedded
	// LanguageRaku represents the Raku programming language.
	LanguageRaku
	// LanguageRAML represents the RAML programming language.
	LanguageRAML
	// LanguageRascal represents the Rascal programming language.
	LanguageRascal
	// LanguageRawToken represents the RawToken programming language.
	LanguageRawToken
	// LanguageRazor represents the Razor programming language.
	LanguageRazor
	// LanguageRConsole represents the RConsole programming language.
	LanguageRConsole
	// LanguageRd represents the Rd programming language.
	LanguageRd
	// LanguageRDoc represents the RDoc programming language.
	LanguageRDoc
	// LanguageReadlineConfig represents the ReadlineConfig programming language.
	LanguageReadlineConfig
	// LanguageREALbasic represents the REALbasic programming language.
	LanguageREALbasic
	// LanguageReasonML represents the ReasonML programming language.
	LanguageReasonML
	// LanguageREBOL represents the REBOL programming language.
	LanguageREBOL
	// LanguageRecordJar represents the RecordJar programming language.
	LanguageRecordJar
	// LanguageRed represents the Red programming language.
	LanguageRed
	// LanguageRedcode represents the Redcode programming language.
	LanguageRedcode
	// LanguageRegistry represents the Registry programming language.
	LanguageRegistry
	// LanguageRego represents the Rego programming language.
	LanguageRego
	// LanguageRegularExpression represents the RegularExpression programming language.
	LanguageRegularExpression
	// LanguageRGBDSAssembly represents the RGBDS Assembly programming language.
	LanguageRGBDSAssembly
	// LanguageRenderScript represents the RenderScript programming language.
	LanguageRenderScript
	// LanguageRenPy represents the RenPy programming language.
	LanguageRenPy
	// LanguageReScript represents the ReScript programming language.
	LanguageReScript
	// LanguageResourceBundle represents the ResourceBundle programming language.
	LanguageResourceBundle
	// LanguageReStructuredText represents the ReStructuredText programming language.
	LanguageReStructuredText
	// LanguageRexx represents the Rexx programming language.
	LanguageRexx
	// LanguageRHTML represents the RHTML programming language.
	LanguageRHTML
	// LanguageRichTextFormat represents the RichTextFormat programming language.
	LanguageRichTextFormat
	// LanguageRide represents the Ride programming language.
	LanguageRide
	// LanguageRing represents the Ring programming language.
	LanguageRing
	// LanguageRiot represents the Riot programming language.
	LanguageRiot
	// LanguageRMarkdown represents the RMarkdown programming language.
	LanguageRMarkdown
	// LanguageRNGCompact represents the RNGCompact programming language.
	LanguageRNGCompact
	// LanguageRoboconfGraph represents the RoboconfGraph programming language.
	LanguageRoboconfGraph
	// LanguageRoboconfInstances represents the RoboconfInstances programming language.
	LanguageRoboconfInstances
	// LanguageRobotFramework represents the RobotFramework programming language.
	LanguageRobotFramework
	// LanguageRoff represents the Roff programming language.
	LanguageRoff
	// LanguageRoffManpage represents the RoffManpage programming language.
	LanguageRoffManpage
	// LanguageRouge represents the Rouge programming language.
	LanguageRouge
	// LanguageRPC represents the RPC programming language.
	LanguageRPC
	// LanguageRPGLE represents the RPGLE programming language.
	LanguageRPGLE
	// LanguageRPMSpec represents the RPMSpec programming language.
	LanguageRPMSpec
	// LanguageRQL represents the RQL programming language.
	LanguageRQL
	// LanguageRSL represents the RSL programming language.
	LanguageRSL
	// LanguageRuby represents the Ruby programming language.
	LanguageRuby
	// LanguageRubyIRBSession represents the RubyIRBSession programming language.
	LanguageRubyIRBSession
	// LanguageRUNOFF represents the RUNOFF programming language.
	LanguageRUNOFF
	// LanguageRust represents the Rust programming language.
	LanguageRust
	// LanguageS represents the S programming language.
	LanguageS
	// LanguageSage represents the Sage programming language.
	LanguageSage
	// LanguageSalt represents the Salt programming language.
	LanguageSalt
	// LanguageSARL represents the SARL programming language.
	LanguageSARL
	// LanguageSAS represents the SAS programming language.
	LanguageSAS
	// LanguageSass represents the Sass programming language.
	LanguageSass
	// LanguageScala represents the Scala programming language.
	LanguageScala
	// LanguageScaml represents the Scaml programming language.
	LanguageScaml
	// LanguageScdoc represents the Scdoc programming language.
	LanguageScdoc
	// LanguageScheme represents the Scheme programming language.
	LanguageScheme
	// LanguageScilab represents the Scilab programming language.
	LanguageScilab
	// LanguageScribe represents the Scribe programming language.
	LanguageScribe
	// LanguageSCSS represents the SCSS programming language.
	LanguageSCSS
	// LanguageSed represents the Sed programming language.
	LanguageSed
	// LanguageSelf represents the Self programming language.
	LanguageSelf
	// LanguageSGML represents the SGML programming language.
	LanguageSGML
	// LanguageShaderLab represents the ShaderLab programming language.
	LanguageShaderLab
	// LanguageShell represents the Shell programming language.
	LanguageShell
	// LanguageShellSession represents the Shell Session programming language.
	LanguageShellSession
	// LanguageShen represents the Shen programming language.
	LanguageShen
	// LanguageShExC represents the ShExC programming language.
	LanguageShExC
	// LanguageSieve represents the Sieve programming language.
	LanguageSieve
	// LanguageSilver represents the Silver programming language.
	LanguageSilver
	// LanguageSimula represents the Simula programming language.
	LanguageSimula
	// LanguageSingularity represents the Singularity programming language.
	LanguageSingularity
	// LanguageSketchDrawing represents the SketchDrawing programming language.
	LanguageSketchDrawing
	// LanguageSKILL represents the SKILL programming language.
	LanguageSKILL
	// LanguageSlash represents the Slash programming language.
	LanguageSlash
	// LanguageSlice represents the Slice programming language.
	LanguageSlice
	// LanguageSlim represents the Slim programming language.
	LanguageSlim
	// LanguageSlint represents the Slint programming language.
	LanguageSlint
	// LanguageSlurm represents the Slurm programming language.
	LanguageSlurm
	// LanguageSmali represents the Smali programming language.
	LanguageSmali
	// LanguageSmalltalk represents the Smalltalk programming language.
	LanguageSmalltalk
	// LanguageSmartGameFormat represents the SmartGameFormat programming language.
	LanguageSmartGameFormat
	// LanguageSmarty represents the Smarty programming language.
	LanguageSmarty
	// LanguageSMIME represents the SMIME programming language.
	LanguageSMIME
	// LanguageSML represents the Standard ML programming language.
	LanguageSML
	// LanguageSmPL represents the SmPL programming language.
	LanguageSmPL
	// LanguageSMT represents the SMT programming language.
	LanguageSMT
	// LanguageSNBT represents the SNBT programming language.
	LanguageSNBT
	// LanguageSnobol represents the Snobol programming language.
	LanguageSnobol
	// LanguageSnowball represents the Snowball programming language.
	LanguageSnowball
	// LanguageSolidity represents the Solidity programming language.
	LanguageSolidity
	// LanguageSourcePawn represents the SourcePawn programming language.
	LanguageSourcePawn
	// LanguageSpade represents the Spade programming language.
	LanguageSpade
	// LanguageSPARQL represents the SPARQL programming language.
	LanguageSPARQL
	// LanguageSplineFontDatabase represents the Spline Font Database programming language.
	LanguageSplineFontDatabase
	// LanguageSourcesList represents the Debian Sourcelist programming language.
	LanguageSourcesList
	// LanguageSQF represents the SQF programming language.
	LanguageSQF
	// LanguageSQL represents the SQL programming language.
	LanguageSQL
	// LanguageSQLPL represents the SQLPL programming language.
	LanguageSQLPL
	// LanguageSqlite3con represents the sqlite3con programming language.
	LanguageSqlite3con
	// LanguageSquidConf represents the SquidConf programming language.
	LanguageSquidConf
	// LanguageSquirrel represents the Squirrel programming language.
	LanguageSquirrel
	// LanguageSRecodeTemplate represents the SRecode Template programming language.
	LanguageSRecodeTemplate
	// LanguageSSHConfig represents the SSH Config programming language.
	LanguageSSHConfig
	// LanguageSSP represents the Scalate Server Page programming language.
	LanguageSSP
	// LanguageStan represents the Stan programming language.
	LanguageStan
	// LanguageStarlark represents the Starlark programming language.
	LanguageStarlark
	// LanguageStas represents the st(ack) as(sembler) programming language.
	LanguageStas
	// LanguageStata represents the Stata programming language.
	LanguageStata
	// LanguageSTON represents the STON programming language.
	LanguageSTON
	// LanguageSVG represents the SVG programming language.
	LanguageSVG
	// LanguageStylus represents the Stylus programming language.
	LanguageStylus
	// LanguageSublimeTextConfig represents the SublimeTextConfig programming language.
	LanguageSublimeTextConfig
	// LanguageSubRipText represents the SubRip Text programming language.
	LanguageSubRipText
	// LanguageSugarSS represents the SugarSS programming language.
	LanguageSugarSS
	// LanguageSuperCollider represents the SuperCollider programming language.
	LanguageSuperCollider
	// LanguageSvelte represents the Svelte programming language.
	LanguageSvelte
	// LanguageSwift represents the Swift programming language.
	LanguageSwift
	// LanguageSWIG represents the SWIG programming language.
	LanguageSWIG
	// LanguageSYSTEMD represents the SYSTEMD programming language.
	LanguageSYSTEMD
	// LanguageSystemVerilog represents the SystemVerilog programming language.
	LanguageSystemVerilog
	// LanguageTableGen represents the TableGen programming language.
	LanguageTableGen
	// LanguageTADS3 represents the TADS3 programming language.
	LanguageTADS3
	// LanguageTAP represents the TAP programming language.
	LanguageTAP
	// LanguageTASM represents the TASM programming language.
	LanguageTASM
	// LanguageTcl represents the Tcl programming language.
	LanguageTcl
	// LanguageTcsh represents the Tcsh programming language.
	LanguageTcsh
	// LanguageTcshSession represents the TcshSession programming language.
	LanguageTcshSession
	// LanguageTea represents the Tea programming language.
	LanguageTea
	// LanguageTempl represents the Templ programming language.
	LanguageTempl
	// LanguageTeraTerm represents the TeraTerm programming language.
	LanguageTeraTerm
	// LanguageTermcap represents the Termcap programming language.
	LanguageTermcap
	// LanguageTerminfo represents the Terminfo programming language.
	LanguageTerminfo
	// LanguageTerra represent the Terra programming language.
	LanguageTerra
	// LanguageTerraform represents the Terraform programming language.
	LanguageTerraform
	// LanguageTeX represents the TeX programming language.
	LanguageTeX
	// LanguageTexinfo represent the Texinfo programming language.
	LanguageTexinfo
	// LanguageText represents the Text programming language.
	LanguageText
	// LanguageTextile represent the Textile programming language.
	LanguageTextile
	// LanguageThrift represents the Thrift programming language.
	LanguageThrift
	// LanguageTiddler represents the Tiddler programming languages.
	LanguageTiddler
	// LanguageTIProgram represent the TIProgram programming language.
	LanguageTIProgram
	// LanguageTLA represent the TLA programming language.
	LanguageTLA
	// LanguageTNT represents the TNT programming languages.
	LanguageTNT
	// LanguageTodotxt represents the Todotxt programming languages.
	LanguageTodotxt
	// LanguageTOML represents the TOML programming language.
	LanguageTOML
	// LanguageTradingView represents the TradingView programming languages.
	LanguageTradingView
	// LanguageTrafficScript represents the TrafficScript programming languages.
	LanguageTrafficScript
	// LanguageTransactSQL represents the TransactSQL programming languages.
	LanguageTransactSQL
	// LanguageTreetop represents the Treetop programming languages.
	LanguageTreetop
	// LanguageTSV represent the TSV programming language.
	LanguageTSV
	// LanguageTSX represent the TSX programming language.
	LanguageTSX
	// LanguageTuring represents the Turing programming language.
	LanguageTuring
	// LanguageTurtle represents the Turtle programming languages.
	LanguageTurtle
	// LanguageTwig represents the Twig programming language.
	LanguageTwig
	// LanguageTXL represent the TXL programming language.
	LanguageTXL
	// LanguageTypst represents the Typst programming language.
	LanguageTypst
	// LanguageTypeLanguage represent the TypeLanguage programming language.
	LanguageTypeLanguage
	// LanguageTypeScript represents the TypeScript programming language.
	LanguageTypeScript
	// LanguageTypoScript represents the TypoScript programming language.
	LanguageTypoScript
	// LanguageUcode represents the Ucode programming language.
	LanguageUcode
	// LanguageUnicon represents the Unicon programming language.
	LanguageUnicon
	// LanguageUnifiedParallelC represents the UnifiedParallelC programming language.
	LanguageUnifiedParallelC
	// LanguageUnity3DAsset represents the Unity3DAsset programming language.
	LanguageUnity3DAsset
	// LanguageUnixAssembly represents the UnixAssembly programming language.
	LanguageUnixAssembly
	// LanguageUno represents the Uno programming language.
	LanguageUno
	// LanguageUnrealScript represents the UnrealScript programming language.
	LanguageUnrealScript
	// LanguageUrbiScript represents the UrbiScript programming language.
	LanguageUrbiScript
	// LanguageUrWeb represents the UrWeb programming language.
	LanguageUrWeb
	// LanguageUSD represents the USD programming language.
	LanguageUSD
	// LanguageUxntal represents the Uxntal programming language.
	LanguageUxntal
	// LanguageV represents the V programming language.
	LanguageV
	// LanguageVala represents the Vala programming language.
	LanguageVala
	// LanguageVB represents the VB programming language.
	LanguageVB
	// LanguageVBA represents the VBA programming language.
	LanguageVBA
	// LanguageVBNet represents the VB.net programming language.
	LanguageVBNet
	// LanguageVBScript represents the VBScript programming language.
	LanguageVBScript
	// LanguageVCL represents the VCL programming language.
	LanguageVCL
	// LanguageVCLSnippets represents the VCLSnippets programming language.
	LanguageVCLSnippets
	// LanguageVCTreeStatus represents the VCTreeStatus programming language.
	LanguageVCTreeStatus
	// LanguageVelocity represents the Velocity programming language.
	LanguageVelocity
	// LanguageVerilog represents the Verilog programming language.
	LanguageVerilog
	// LanguageVGL represents the VGL programming language.
	LanguageVGL
	// LanguageVHDL represents the VHDL programming language.
	LanguageVHDL
	// LanguageVHS represents the VHS programming language.
	LanguageVHS
	// LanguageVimHelpFile represents the VimHelpFile programming language.
	LanguageVimHelpFile
	// LanguageVimL represents the VimL programming language.
	LanguageVimL
	// LanguageVimScript represents the VimScript programming language.
	LanguageVimScript
	// LanguageVimSnippet represents the VimSnippet programming language.
	LanguageVimSnippet
	// LanguageVolt represents the Volt programming language.
	LanguageVolt
	// LanguageVShell represents the V shell programming language.
	LanguageVShell
	// LanguageVueJS represents the VueJS programming language.
	LanguageVueJS
	// LanguageWavefrontMaterial represents the WavefrontMaterial programming language.
	LanguageWavefrontMaterial
	// LanguageWavefrontObject represents the WavefrontObject programming language.
	LanguageWavefrontObject
	// LanguageWdl represents the Wdl programming language.
	LanguageWdl
	// LanguageWDTE represents the WDTE programming language.
	LanguageWDTE
	// LanguageWDiff represents the WDiff programming language.
	LanguageWDiff
	// LanguageWebAssembly represents the WebAssembly programming language.
	LanguageWebAssembly
	// LanguageWebGPUShadingLanguage represents the WebGPU Shading Language programming language.
	LanguageWebGPUShadingLanguage
	// LanguageWebIDL represents the WebIDL programming language.
	LanguageWebIDL
	// LanguageWebOntologyLanguage represents the WebOntologyLanguage programming language.
	LanguageWebOntologyLanguage
	// LanguageWebVTT represents the WebVTT programming language.
	LanguageWebVTT
	// LanguageWgetConfig represents the WgetConfig programming language.
	LanguageWgetConfig
	// LanguageWhiley represents the Whiley programming language.
	LanguageWhiley
	// LanguageWindowsRegistryEntries represents the WindowsRegistryEntries programming language.
	LanguageWindowsRegistryEntries
	// LanguageWisp represents the wisp programming language.
	LanguageWisp
	// LanguageWollok represents the Wollok programming language.
	LanguageWollok
	// LanguageWowAddonData represents the WowAddonData programming language.
	LanguageWowAddonData
	// LanguageX10 represents the X10 programming language.
	LanguageX10
	// LanguageXAML represents the XAML programming language.
	LanguageXAML
	// LanguageXBase represents the XBase programming language.
	LanguageXBase
	// LanguageXBitMap represents the XBitMap programming language.
	LanguageXBitMap
	// LanguageXC represents the XC programming language.
	LanguageXC
	// LanguageXCompose represents the XCompose programming language.
	LanguageXCompose
	// LanguageXFontDirectoryIndex represents the XFontDirectoryIndex programming language.
	LanguageXFontDirectoryIndex
	// LanguageXML represents the XML programming language.
	LanguageXML
	// LanguageXMLPropertyList represents the XMLPropertyList programming language.
	LanguageXMLPropertyList
	// LanguageXojo represents the Xojo programming language.
	LanguageXojo
	// LanguageXorg represents the Xorg programming language.
	LanguageXorg
	// LanguageXPages represents the XPages programming language.
	LanguageXPages
	// LanguageXPixMap represents the XPixMap programming language.
	LanguageXPixMap
	// LanguageXProc represents the XProc programming language.
	LanguageXProc
	// LanguageXQuery represents the XQuery programming language.
	LanguageXQuery
	// LanguageXS represents the XS programming language.
	LanguageXS
	// LanguageXSLT represents the XSLT programming language.
	LanguageXSLT
	// LanguageXtend represents the Xtend programming language.
	LanguageXtend
	// LanguageXtlang represents the Xtlang programming language.
	LanguageXtlang
	// LanguageYacc represents the Yacc programming language.
	LanguageYacc
	// LanguageYAML represents the YAML programming language.
	LanguageYAML
	// LanguageYANG represents the YANG programming language.
	LanguageYANG
	// LanguageYARA represents the YARA programming language.
	LanguageYARA
	// LanguageYASnippet represents the YASnippet programming language.
	LanguageYASnippet
	// LanguageZ80Assembly represents the Z80 Assembly programming language.
	LanguageZ80Assembly
	// LanguageZAP represents the ZAP programming language.
	LanguageZAP
	// LanguageZed represents the Zed programming language.
	LanguageZed
	// LanguageZeek represents the Zeek programming language.
	LanguageZeek
	// LanguageZenScript represents the ZenScript programming language.
	LanguageZenScript
	// LanguageZephir represents the Zephir programming language.
	LanguageZephir
	// LanguageZig represents the Zig programming language.
	LanguageZig
	// LanguageZIL represents the ZIL programming language.
	LanguageZIL
	// LanguageZimpl represents the Zimpl programming language.
	LanguageZimpl
)

const (
	language1CEnterpriseStr                = "1C Enterprise"
	language4DStr                          = "4D"
	languageABAPStr                        = "ABAP"
	languageABNFStr                        = "ABNF"
	languageActionScript3Str               = "ActionScript 3"
	languageActionScriptStr                = "ActionScript"
	languageAdaStr                         = "Ada"
	languageADLStr                         = "ADL"
	languageAdobeFontMetricsStr            = "Adobe Font Metrics"
	languageAdvPLStr                       = "AdvPL"
	languageAgdaStr                        = "Agda"
	languageAGSScriptStr                   = "AGS Script"
	languageAheuiStr                       = "Aheui"
	languageAlloyStr                       = "Alloy"
	languageAlpineAbuildStr                = "Alpine Abuild"
	languageALStr                          = "AL"
	languageAltiumDesignerStr              = "Altium Designer"
	languageAmbientTalkStr                 = "AmbientTalk"
	languageAMPLStr                        = "AMPL"
	languageAngelScriptStr                 = "AngelScript"
	languageAngular2Str                    = "Angular2"
	languageAnsibleStr                     = "Ansible"
	languageAntBuildSystemStr              = "Ant Build System"
	languageANTLRStr                       = "ANTLR"
	languageApacheConfigStr                = "Apache Config"
	languageApexStr                        = "Apex"
	languageAPIBlueprintStr                = "API Blueprint"
	languageAPLStr                         = "APL"
	languageApolloGuidanceComputerStr      = "Apollo Guidance Computer"
	languageAppleScriptStr                 = "AppleScript"
	languageArangoDBQueryLanguageStr       = "ArangoDB Query Language"
	languageArcStr                         = "Arc"
	languageArduinoStr                     = "Arduino"
	languageArmAsmStr                      = "ArmAsm"
	languageArrowStr                       = "Arrow"
	languageArturoStr                      = "Arturo"
	languageASCIIDocStr                    = "AsciiDoc"
	languageASLStr                         = "ASL"
	languageASN1Str                        = "ASN.1"
	languageASPClassicStr                  = "ASP Classic"
	languageASPDotNetStr                   = "ASP.NET"
	languageAspectJStr                     = "AspectJ"
	languageAspxCSharpStr                  = "aspx-cs"
	languageAspxVBNetStr                   = "aspx-vb"
	languageAssemblyStr                    = "Assembly"
	languageAstroStr                       = "Astro"
	languageAsymptoteStr                   = "Asymptote"
	languageATLStr                         = "ATL"
	languageATSStr                         = "ATS"
	languageAugeasStr                      = "Augeas"
	languageAutoconfStr                    = "Autoconf"
	languageAutoHotkeyStr                  = "AutoHotkey"
	languageAutoItStr                      = "AutoIt"
	languageAvroIDLStr                     = "Avro IDL"
	languageAwkStr                         = "Awk"
	languageBallerinaStr                   = "Ballerina"
	languageBAREStr                        = "BARE"
	languageBashSessionStr                 = "Bash Session"
	languageBashStr                        = "Bash"
	languageBasicStr                       = "Basic"
	languageBatchfileStr                   = "Batchfile"
	languageBatchScriptStr                 = "Batch Script"
	languageBBCBasicStr                    = "BBC Basic"
	languageBBCodeStr                      = "BBCode"
	languageBCStr                          = "BC"
	languageBeefStr                        = "Beef"
	languageBefungeStr                     = "Befunge"
	languageBibTeXStr                      = "BibTeX"
	languageBicepStr                       = "Bicep"
	languageBisonStr                       = "Bison"
	languageBitBakeStr                     = "BitBake"
	languageBladeStr                       = "Blade"
	languageBladeTemplateStr               = "Blade Template"
	languageBlazorStr                      = "Blazor"
	languageBlitzBasicStr                  = "BlitzBasic"
	languageBlitzMaxStr                    = "BlitzMax"
	languageBluespecStr                    = "Bluespec"
	languageBNFStr                         = "BNF"
	languageBoaStr                         = "Boa"
	languageBoogieStr                      = "Boogie"
	languageBooStr                         = "Boo"
	languageBQNStr                         = "BQN"
	languageBrainfuckStr                   = "Brainfuck"
	languageBrightScriptStr                = "BrightScript"
	languageBroStr                         = "Bro"
	languageBrowserslistStr                = "Browserslist"
	languageBSTStr                         = "BST"
	languageBUGSStr                        = "BUGS"
	languageC2hsHaskellStr                 = "C2hs Haskell"
	languageC3Str                          = "C3"
	languageCa65AssemblerStr               = "ca65 assembler"
	languageCabalConfigStr                 = "Cabal Config"
	languageCaddyfileDirectivesStr         = "Caddyfile Directives"
	languageCaddyfileStr                   = "Caddyfile"
	languageCADLStr                        = "cADL"
	languageCAmkESStr                      = "CAmkES"
	languageCapDLStr                       = "CapDL"
	languageCapNProtoStr                   = "Cap'n Proto"
	languageCartoCSSStr                    = "CartoCSS"
	languageCassandraCQLStr                = "Cassandra CQL"
	languageCBMBasicV2Str                  = "CBM BASIC V2"
	languageCeylonStr                      = "Ceylon"
	languageCFEngine3Str                   = "CFEngine3"
	languageCfstatementStr                 = "cfstatement"
	languageChaiScriptStr                  = "ChaiScript"
	languageChapelStr                      = "Chapel"
	languageCharityStr                     = "Charity"
	languageCharmciStr                     = "Charmci"
	languageCheetahStr                     = "Cheetah"
	languageChucKStr                       = "ChucK"
	languageCirruStr                       = "Cirru"
	languageClarionStr                     = "Clarion"
	languageClassicASPStr                  = "Classic ASP"
	languageClayStr                        = "Clay"
	languageCleanStr                       = "Clean"
	languageClickStr                       = "Click"
	languageCLIPSStr                       = "CLIPS"
	languageClojureScriptStr               = "ClojureScript"
	languageClojureStr                     = "Clojure"
	languageClosureTemplatesStr            = "Closure Templates"
	languageCloudFirestoreSecurityRulesStr = "Cloud Firestore Security Rules"
	languageCMakeStr                       = "CMake"
	languageCObjdumpStr                    = "C-ObjDump"
	languageCOBOLFreeStr                   = "COBOLFree"
	languageCOBOLStr                       = "COBOL"
	languageCocoaStr                       = "Cocoa"
	languageCodeQLStr                      = "CodeQL"
	languageCoffeeScriptStr                = "CoffeeScript"
	languageColdfusionCFCStr               = "ColdFusion CFC"
	languageColdfusionHTMLStr              = "ColdFusion"
	languageCOLLADAStr                     = "COLLADA"
	languageCommonLispStr                  = "Common Lisp"
	languageCommonWorkflowLanguageStr      = "Common Workflow Language"
	languageComponentPascalStr             = "Component Pascal"
	languageConfigStr                      = "Config"
	languageCoNLLUStr                      = "CoNLL-U"
	languageCoolStr                        = "Cool"
	languageCoqStr                         = "Coq"
	languageCoreStr                        = "Core"
	languageCPerlStr                       = "cperl"
	languageCppObjdumpStr                  = "Cpp-ObjDump"
	languageCPPStr                         = "C++"
	languageCPSAStr                        = "CPSA"
	languageCreoleStr                      = "Creole"
	languageCrmshStr                       = "Crmsh"
	languageCrocStr                        = "Croc"
	languageCrontabStr                     = "Crontab"
	languageCryptolStr                     = "Cryptol"
	languageCrystalStr                     = "Crystal"
	languageCSharpStr                      = "C#"
	languageCSHTMLStr                      = "CSHTML"
	languageCSONStr                        = "CSON"
	languageCsoundDocumentStr              = "Csound Document"
	languageCsoundOrchestraStr             = "Csound Orchestra"
	languageCsoundScoreStr                 = "Csound Score"
	languageCsoundStr                      = "Csound"
	languageCSSStr                         = "CSS"
	languageCStr                           = "C"
	languageCSVStr                         = "CSV"
	languageCUDAStr                        = "Cuda"
	languageCUEStr                         = "CUE"
	languagecURLConfigStr                  = "cURL Config"
	languageCVSStr                         = "CVS"
	languageCWebStr                        = "CWeb"
	languageCycriptStr                     = "Cycript"
	languageCypherStr                      = "Cypher"
	languageCythonStr                      = "Cython"
	languageDafnyStr                       = "Dafny"
	languageDarcsPatchStr                  = "Darcs Patch"
	languageDartStr                        = "Dart"
	languageDASM16Str                      = "DASM16"
	languageDataWeaveStr                   = "DataWeave"
	languageDaxStr                         = "Dax"
	languageDCLStr                         = "DCL"
	languageDCPU16AsmStr                   = "DCPU-16 ASM"
	languageDebianControlFileStr           = "Debian Control file"
	languageDelphiStr                      = "Delphi"
	languageDesktopFileStr                 = "Desktop file"
	languageDevicetreeStr                  = "Devicetree"
	languageDGStr                          = "dg"
	languageDhallStr                       = "Dhall"
	languageDiffStr                        = "Diff"
	languageDigitalCommandStr              = "DIGITAL Command Language"
	languageDircolorsStr                   = "dircolors"
	languageDirectX3DFileStr               = "DirectX 3D File"
	languageDjangoJinjaStr                 = "Django/Jinja"
	languageDMStr                          = "DM"
	languageDNSZoneStr                     = "DNS Zone"
	languageDObjdumpStr                    = "d-objdump"
	languageDockerfileStr                  = "Dockerfile"
	languageDockerStr                      = "Docker"
	languageDocTeXStr                      = "DocTeX"
	languageDocumentationStr               = "Documentation"
	languageDogescriptStr                  = "Dogescript"
	languageDStr                           = "D"
	languageDTDStr                         = "DTD"
	languageDTraceStr                      = "DTrace"
	languageDuelStr                        = "Duel"
	languageDylanLIDStr                    = "DylanLID"
	languageDylanSessionStr                = "Dylan session"
	languageDylanStr                       = "Dylan"
	languageDynASMStr                      = "DynASM"
	languageEagleStr                       = "Eagle"
	languageEarlGreyStr                    = "Earl Grey"
	languageEasybuildStr                   = "Easybuild"
	languageEasytrieveStr                  = "Easytrieve"
	languageEBNFStr                        = "EBNF"
	languageEcereProjectsStr               = "Ecere Projects"
	languageEclipseStr                     = "ECLiPSe"
	languageECLStr                         = "ECL"
	languageECStr                          = "eC"
	languageEditorConfigStr                = "EditorConfig"
	languageEdjeDataCollectionStr          = "Edje Data Collection"
	languageEdnStr                         = "edn"
	languageEiffelStr                      = "Eiffel"
	languageEJSStr                         = "EJS"
	languageElixirIexSessionStr            = "Elixir iex session"
	languageElixirStr                      = "Elixir"
	languageElmStr                         = "Elm"
	languageEmacsLispStr                   = "Emacs Lisp"
	languageEMailStr                       = "E-mail"
	languageEmberScriptStr                 = "EmberScript"
	languageEMLStr                         = "EML"
	languageEQStr                          = "EQ"
	languageERBStr                         = "ERB"
	languageErlangErlSessionStr            = "Erlang erl session"
	languageErlangStr                      = "Erlang"
	languageEshellStr                      = "Eshell"
	languageEStr                           = "E"
	languageEvoqueStr                      = "Evoque"
	languageExeclineStr                    = "execline"
	languageEzhilStr                       = "Ezhil"
	languageFactorStr                      = "Factor"
	languageFancyStr                       = "Fancy"
	languageFantomStr                      = "Fantom"
	languageFaustStr                       = "Faust"
	languageFelixStr                       = "Felix"
	languageFennelStr                      = "Fennel"
	languageFIGletFontStr                  = "FIGlet Font"
	languageFilebenchWMLStr                = "Filebench WML"
	languageFilterscriptStr                = "Filterscript"
	languageFishStr                        = "Fish"
	languageFlatlineStr                    = "Flatline"
	languageFloScriptStr                   = "FloScript"
	languageFLUXStr                        = "FLUX"
	languageFontStr                        = "Font"
	languageFormattedStr                   = "Formatted"
	languageForthStr                       = "Forth"
	languageFortranFixedStr                = "FortranFixed"
	languageFortranFreeFormStr             = "Fortran Free Form"
	languageFortranStr                     = "Fortran"
	languageFoxProStr                      = "FoxPro"
	languageFreefemStr                     = "Freefem"
	languageFreeMarkerStr                  = "FreeMarker"
	languageFregeStr                       = "Frege"
	languageFSharpStr                      = "F#"
	languageFStarStr                       = "F*"
	languageFutharkStr                     = "Futhark"
	languageGameMakerLanguageStr           = "Game Maker Language"
	languageGAMLStr                        = "GAML"
	languageGAMSStr                        = "GAMS"
	languageGapStr                         = "GAP"
	languageGasStr                         = "GAS"
	languageGCCMachineDescriptionStr       = "GCC Machine Description"
	languageGCodeStr                       = "G-code"
	languageGDBStr                         = "GDB"
	languageGDNativeStr                    = "GDNative"
	languageGDScript3Str                   = "GDScript3"
	languageGDScriptStr                    = "GDScript"
	languageGEDCOMStr                      = "GEDCOM"
	languageGemfileLockStr                 = "Gemfile.lock"
	languageGemtextStr                     = "Gemtext"
	languageGenieStr                       = "Genie"
	languageGenshiHTMLStr                  = "Genshi HTML"
	languageGenshiStr                      = "Genshi"
	languageGenshiTextStr                  = "Genshi Text"
	languageGentooEbuildStr                = "Gentoo Ebuild"
	languageGentooEclassStr                = "Gentoo Eclass"
	languageGerberImageStr                 = "Gerber Image"
	languageGettextCatalogStr              = "Gettext Catalog"
	languageGettextStr                     = "Gettext"
	languageGherkinStr                     = "Gherkin"
	languageGitAttributesStr               = "Git Attributes"
	languageGitConfigStr                   = "Git Config"
	languageGitStr                         = "Git"
	languageGleamStr                       = "Gleam"
	languageGLSLStr                        = "GLSL"
	languageGlyphBitmapStr                 = "Glyph Bitmap Distribution Format"
	languageGlyphStr                       = "Glyph"
	languageGNStr                          = "GN"
	languageGnuplotStr                     = "Gnuplot"
	languageGoloStr                        = "Golo"
	languageGoodDataCLStr                  = "GoodData-CL"
	languageGoStr                          = "Go"
	languageGosuStr                        = "Gosu"
	languageGosuTemplateStr                = "Gosu Template"
	languageGoHTMLTemplateStr              = "Go HTML Template"
	languageGoTemplateStr                  = "Go Template"
	languageGoTextTemplateStr              = "Go Text Template"
	languageGraceStr                       = "Grace"
	languageGradleConfigStr                = "Gradle Config"
	languageGradleStr                      = "Gradle"
	languageGrammaticalFrameworkStr        = "Grammatical Framework"
	languageGraphModelingLanguageStr       = "Graph Modeling Language"
	languageGraphQLStr                     = "GraphQL"
	languageGraphvizDOTStr                 = "Graphviz (DOT)"
	languageGroffStr                       = "Groff"
	languageGroovyServerPagesStr           = "Groovy Server Pages"
	languageGroovyStr                      = "Groovy"
	languageHackStr                        = "Hack"
	languageHamlStr                        = "Haml"
	languageHandlebarsStr                  = "Handlebars"
	languageHAProxyStr                     = "HAProxy"
	languageHarbourStr                     = "Harbour"
	languageHareStr                        = "Hare"
	languageHaskellStr                     = "Haskell"
	languageHaxeStr                        = "Haxe"
	languageHCLStr                         = "HCL"
	languageHexdumpStr                     = "Hexdump"
	languageHiveQLStr                      = "HiveQL"
	languageHLBStr                         = "HLB"
	languageHLSLStr                        = "HLSL"
	languageHolyCStr                       = "HolyC"
	languageHSAILStr                       = "HSAIL"
	languageHspecStr                       = "Hspec"
	languageHTMLDjangoStr                  = "HTML+Django"
	languageHTMLECRStr                     = "HTML+ECR"
	languageHTMLEEXStr                     = "HTML+EEX"
	languageHTMLERBStr                     = "HTML+ERB"
	languageHTMLPHPStr                     = "HTML+PHP"
	languageHTMLRazorStr                   = "HTML+Razor"
	languageHTMLStr                        = "HTML"
	languageHTTPStr                        = "HTTP"
	languageHxmlStr                        = "HXML"
	languageHybrisStr                      = "Hybris"
	languageHyPhyStr                       = "HyPhy"
	languageHyStr                          = "Hy"
	languageIconStr                        = "Icon"
	languageIDAStr                         = "IDA"
	languageIDLStr                         = "IDL"
	languageIdrisStr                       = "Idris"
	languageIgnoreListStr                  = "Ignore List"
	languageIGORProStr                     = "IGOR Pro"
	languageIgorStr                        = "Igor"
	languageImageJMacroStr                 = "ImageJ Macro"
	languageImageJPEGStr                   = "Image (jpeg)"
	languageImagePNGStr                    = "Image (png)"
	languageInform6Str                     = "Inform 6"
	languageInform6TemplateStr             = "Inform 6 template"
	languageInform7Str                     = "Inform 7"
	languageINIStr                         = "INI"
	languageInnoSetupStr                   = "Inno Setup"
	languageIokeStr                        = "Ioke"
	languageIoStr                          = "Io"
	languageIRCLogsStr                     = "IRC Logs"
	languageIsabelleRootStr                = "Isabelle ROOT"
	languageIsabelleStr                    = "Isabelle"
	languageISCdhcpdStr                    = "ISC dhcpd"
	languageJadeStr                        = "Jade"
	languageJAGSStr                        = "JAGS"
	languageJanetStr                       = "Janet"
	languageJasminStr                      = "Jasmin"
	languageJavaPropertiesStr              = "Java Properties"
	languageJavaScriptERBStr               = "JavaScript+ERB"
	languageJavaScriptStr                  = "JavaScript"
	languageJavaStr                        = "Java"
	languageJCLStr                         = "JCL"
	languageJFlexStr                       = "JFlex"
	languageJisonLexStr                    = "Jison Lex"
	languageJisonStr                       = "Jison"
	languageJolieStr                       = "Jolie"
	languageJSGFStr                        = "JSGF"
	languageJSON5Str                       = "JSON5"
	languageJSONataStr                     = "JSONata"
	languageJSONiqStr                      = "JSONiq"
	languageJSONLDStr                      = "JSONLD"
	languageJsonnetStr                     = "Jsonnet"
	languageJSONStr                        = "JSON"
	languageJSONWithCommentsStr            = "JSON with Comments"
	languageJSPStr                         = "Java Server Page"
	languageJStr                           = "J"
	languageJSXStr                         = "JSX"
	languageJuliaConsoleStr                = "Julia console"
	languageJuliaStr                       = "Julia"
	languageJungleStr                      = "Jungle"
	languageJupyterNotebookStr             = "Jupyter Notebook"
	languageJuttleStr                      = "Juttle"
	languageKaitaiStr                      = "Kaitai Struct"
	languageKakouneStr                     = "Kakoune"
	languageKalStr                         = "Kal"
	languageKconfigStr                     = "Kconfig"
	languageKDLStr                         = "KDL"
	languageKernelLogStr                   = "Kernel log"
	languageKiCadLayoutStr                 = "KiCad Layout"
	languageKiCadLegacyLayoutStr           = "KiCad Legacy Layout"
	languageKiCadSchematicStr              = "KiCad Schematic"
	languageKitStr                         = "Kit"
	languageKokaStr                        = "Koka"
	languageKotlinStr                      = "Kotlin"
	languageKRLStr                         = "KRL"
	languageLabVIEWStr                     = "LabVIEW"
	languageLaravelTemplateStr             = "Laravel Template"
	languageLarkStr                        = "Lark"
	languageLassoStr                       = "Lasso"
	languageLateralusStr                   = "Lateralus"
	languageLaTeXStr                       = "LaTeX"
	languageLatteStr                       = "Latte"
	languageLeanStr                        = "Lean"
	languageLean4Str                       = "Lean4"
	languageLessStr                        = "LESS"
	languageLexStr                         = "Lex"
	languageLFEStr                         = "LFE"
	languageLighttpdStr                    = "Lighttpd configuration file"
	languageLilyPondStr                    = "LilyPond"
	languageLimboStr                       = "Limbo"
	languageLinkerScriptStr                = "Linker Script"
	languageLinuxKernelModuleStr           = "Linux Kernel Module"
	languageLiquidStr                      = "Liquid"
	languageLiterateAgdaStr                = "Literate Agda"
	languageLiterateCoffeeScriptStr        = "Literate CoffeeScript"
	languageLiterateCryptolStr             = "Literate Cryptol"
	languageLiterateHaskellStr             = "Literate Haskell"
	languageLiterateIdrisStr               = "Literate Idris"
	languageLiveScriptStr                  = "LiveScript"
	languageLLVMMIRBodyStr                 = "LLVM-MIR Body"
	languageLLVMMIRStr                     = "LLVM-MIR"
	languageLLVMStr                        = "LLVM"
	languageLogFileStr                     = "Log File"
	languageLogosStr                       = "Logos"
	languageLogtalkStr                     = "Logtalk"
	languageLOLCODEStr                     = "LOLCODE"
	languageLookMLStr                      = "LookML"
	languageLoomScriptStr                  = "LoomScript"
	languageLoxStr                         = "lox"
	languageLSLStr                         = "LSL"
	languageLTspiceSymbolStr               = "LTspice Symbol"
	languageLuaStr                         = "Lua"
	languageLuauStr                        = "Luau"
	languageMakefileStr                    = "Makefile"
	languageMakoStr                        = "Mako"
	languageManStr                         = "Man"
	languageMAQLStr                        = "MAQL"
	languageMarkdownStr                    = "Markdown"
	languageMarklessStr                    = "Markless"
	languageMarkoStr                       = "Marko"
	languageMaskStr                        = "Mask"
	languageMasonStr                       = "Mason"
	languageMaterializeSQLDialectStr       = "Materialize SQL dialect"
	languageMathematicaStr                 = "Mathematica"
	languageMatlabSessionStr               = "Matlab session"
	languageMatlabStr                      = "Matlab"
	languageMaxMSPStr                      = "Max/MSP"
	languageMaxStr                         = "Max"
	languageMCFunctionStr                  = "MCFunction"
	languageMDXStr                         = "MDX"
	languageMesonStr                       = "Meson"
	languageMetafontStr                    = "Metafont"
	languageMetalStr                       = "Metal"
	languageMetapostStr                    = "Metapost"
	languageMicrocadStr                    = "microcad"
	languageMIMEStr                        = "MIME"
	languageMiniDStr                       = "MiniD"
	languageMiniScriptStr                  = "MiniScript"
	languageMiniZincStr                    = "MiniZinc"
	languageMirahStr                       = "Mirah"
	languageMLIRStr                        = "MLIR"
	languageModelicaStr                    = "Modelica"
	languageModula2Str                     = "Modula-2"
	languageMoinWikiStr                    = "MoinMoin/Trac Wiki markup"
	languageMojoStr                        = "Mojo"
	languageMonkeyCStr                     = "MonkeyC"
	languageMonkeyStr                      = "Monkey"
	languageMonteStr                       = "Monte"
	languageMOOCodeStr                     = "MOOCode"
	languageMoonBitStr                     = "MoonBit"
	languageMoonScriptStr                  = "MoonScript"
	languageMorrowindScriptStr             = "MorrowindScript"
	languageMoselStr                       = "Mosel"
	languageMozPreprocHashStr              = "mozhashpreproc"
	languageMozPreprocPercentStr           = "mozpercentpreproc"
	languageMQLStr                         = "MQL"
	languageMscgenStr                      = "Mscgen"
	languageMSDOSSessionStr                = "MSDOS Session"
	languageMuPADStr                       = "MuPAD"
	languageMustacheStr                    = "Mustache"
	languageMXMLStr                        = "MXML"
	languageMyghtyStr                      = "Myghty"
	languageMySQLStr                       = "MySQL"
	languageMzqlStr                        = "mzql"
	languageNASMObjdumpStr                 = "objdump-nasm"
	languageNASMStr                        = "NASM"
	languageNaturalStr                     = "Natural"
	languageNCLStr                         = "NCL"
	languageNDISASMStr                     = "NDISASM"
	languageNemerleStr                     = "Nemerle"
	languageNeonStr                        = "Neon"
	languageNesCStr                        = "nesC"
	languageNewLispStr                     = "newLisp"
	languageNewspeakStr                    = "Newspeak"
	languageNginxConfigStr                 = "Nginx configuration file"
	languageNginxStr                       = "Nginx"
	languageNimrodStr                      = "Nimrod"
	languageNitStr                         = "Nit"
	languageNixStr                         = "Nix"
	languageNotmuchStr                     = "Notmuch"
	languageNSISStr                        = "NSIS"
	languageNumPyStr                       = "NumPy"
	languageNushellStr                     = "Nushell"
	languageNuSMVStr                       = "NuSMV"
	languageNuStr                          = "Nu"
	languageObjdumpStr                     = "objdump"
	languageObjectiveCPPStr                = "Objective-C++"
	languageObjectiveCStr                  = "Objective-C"
	languageObjectiveJStr                  = "Objective-J"
	languageObjectPascalStr                = "ObjectPascal"
	languageOCamlStr                       = "OCaml"
	languageOctaveStr                      = "Octave"
	languageODINStr                        = "Odin"
	languageOnesEnterpriseStr              = "OnesEnterprise"
	languageOocStr                         = "ooc"
	languageOpaStr                         = "Opa"
	languageOpenEdgeABLStr                 = "OpenEdge ABL"
	languageOpenSCADStr                    = "OpenSCAD"
	languageOrgStr                         = "Org"
	languagePacmanConfStr                  = "PacmanConf"
	languagePanStr                         = "Pan"
	languageParaSailStr                    = "ParaSail"
	languageParrotStr                      = "Parrot"
	languagePascalStr                      = "Pascal"
	languagePawnStr                        = "Pawn"
	languagePEGStr                         = "PEG"
	languagePerl6Str                       = "Perl6"
	languagePerlStr                        = "Perl"
	languagePHPStr                         = "PHP"
	languagePHTMLStr                       = "PHTML"
	languagePigStr                         = "Pig"
	languagePikeStr                        = "Pike"
	languagePkgConfigStr                   = "PkgConfig"
	languagePLpgSQLStr                     = "PL/pgSQL"
	languagePlutusCoreStr                  = "Plutus Core"
	languagePointlessStr                   = "Pointless"
	languagePonyStr                        = "Pony"
	languagePostgresConsoleStr             = "PostgreSQL console (psql)"
	languagePostgresStr                    = "PostgreSQL SQL dialect"
	languagePostScriptStr                  = "PostScript"
	languagePOVRayStr                      = "POVRay"
	languagePowerQueryStr                  = "PowerQuery"
	languagePowerShellSessionStr           = "PowerShell Session"
	languagePowerShellStr                  = "PowerShell"
	languagePraatStr                       = "Praat"
	languageProcessingStr                  = "Processing"
	languagePrologStr                      = "Prolog"
	languagePromelaStr                     = "Promela"
	languagePromQLStr                      = "PromQL"
	languageProtocolBufferStr              = "Protocol Buffer"
	languagePRQLStr                        = "PRQL"
	languagePSLStr                         = "Property Specification Language"
	languagePsyShPHPStr                    = "PsySH console session for PHP"
	languagePugStr                         = "Pug"
	languagePuppetStr                      = "Puppet"
	languagePureDataStr                    = "Pure Data"
	languagePureScriptStr                  = "PureScript"
	languagePyPyLogStr                     = "PyPy Log"
	languagePython2Str                     = "Python 2"
	languagePython2TracebackStr            = "Python 2.x Traceback"
	languagePythonConsoleStr               = "Python console session"
	languagePythonStr                      = "Python"
	languagePythonTracebackStr             = "Python Traceback"
	languageQBasicStr                      = "QBasic"
	languageQMLStr                         = "QML"
	languageQVTOStr                        = "QVTO"
	languageRacketStr                      = "Racket"
	languageRagelEmbeddedStr               = "Embedded Ragel"
	languageRagelStr                       = "Ragel"
	languageRakuStr                        = "Raku"
	languageRAMLStr                        = "RAML"
	languageRascalStr                      = "Rascal"
	languageRawTokenStr                    = "Raw token data" // nolint:gosec
	languageRazorStr                       = "Razor"
	languageRConsoleStr                    = "RConsole"
	languageRDocStr                        = "RDoc"
	languageRdStr                          = "Rd"
	languageReadlineConfigStr              = "Readline Config"
	languageREALbasicStr                   = "REALbasic"
	languageReasonMLStr                    = "Reason"
	languageREBOLStr                       = "Rebol"
	languageRecordJarStr                   = "Record Jar"
	languageRedcodeStr                     = "Redcode"
	languageRedStr                         = "Red"
	languageRegistryStr                    = "reg"
	languageRegoStr                        = "Rego"
	languageRegularExpressionStr           = "Regular Expression"
	languageRGBDSAssemblyStr               = "RGBDS Assembly"
	languageRenderScriptStr                = "RenderScript"
	languageRenPyStr                       = "Ren'Py"
	languageReScriptStr                    = "ReScript"
	languageResourceBundleStr              = "ResourceBundle"
	languageReStructuredTextStr            = "reStructuredText"
	languageRexxStr                        = "REXX"
	languageRHTMLStr                       = "RHTML"
	languageRichTextFormatStr              = "Rich Text Format"
	languageRideStr                        = "Ride"
	languageRingStr                        = "Ring"
	languageRiotStr                        = "Riot"
	languageRMarkdownStr                   = "RMarkdown"
	languageRNGCompactStr                  = "Relax-NG Compact"
	languageRoboconfGraphStr               = "Roboconf Graph"
	languageRoboconfInstancesStr           = "Roboconf Instances"
	languageRobotFrameworkStr              = "RobotFramework"
	languageRoffManpageStr                 = "Roff Manpage"
	languageRoffStr                        = "Roff"
	languageRougeStr                       = "Rouge"
	languageRPCStr                         = "RPC"
	languageRPGLEStr                       = "RPGLE"
	languageRPMSpecStr                     = "RPMSpec"
	languageRQLStr                         = "RQL"
	languageRSLStr                         = "RSL"
	languageRStr                           = "R"
	languageRubyIRBSessionStr              = "Ruby irb session"
	languageRubyStr                        = "Ruby"
	languageRUNOFFStr                      = "RUNOFF"
	languageRustStr                        = "Rust"
	languageSageStr                        = "Sage"
	languageSaltStackStr                   = "SaltStack"
	languageSaltStr                        = "Salt"
	languageSARLStr                        = "SARL"
	languageSassStr                        = "Sass"
	languageSASStr                         = "SAS"
	languageScalaStr                       = "Scala"
	languageScamlStr                       = "Scaml"
	languageScdocStr                       = "scdoc"
	languageSchemeStr                      = "Scheme"
	languageScilabStr                      = "Scilab"
	languageScribeStr                      = "Scribe"
	languageSCSSStr                        = "SCSS"
	languageSedStr                         = "sed"
	languageSelfStr                        = "Self"
	languageSGMLStr                        = "SGML"
	languageShaderLabStr                   = "ShaderLab"
	languageShellSessionStr                = "ShellSession"
	languageShellStr                       = "Shell"
	languageShenStr                        = "Shen"
	languageShExCStr                       = "ShExC"
	languageSieveStr                       = "Sieve"
	languageSilverStr                      = "Silver"
	languageSimulaStr                      = "Simula"
	languageSingularityStr                 = "Singularity"
	languageSketchDrawingStr               = "Sketch Drawing"
	languageSKILLStr                       = "SKILL"
	languageSlashStr                       = "Slash"
	languageSliceStr                       = "Slice"
	languageSlimStr                        = "Slim"
	languageSlintStr                       = "Slint"
	languageSlurmStr                       = "Slurm"
	languageSmaliStr                       = "Smali"
	languageSmalltalkStr                   = "Smalltalk"
	languageSmartGameFormatStr             = "SmartGameFormat"
	languageSmartyStr                      = "Smarty"
	languageSMIMEStr                       = "S/MIME"
	languageSMLStr                         = "Standard ML"
	languageSmPLStr                        = "SmPL"
	languageSMTStr                         = "SMT"
	languageSNBTStr                        = "SNBT"
	languageSnobolStr                      = "Snobol"
	languageSnowballStr                    = "Snowball"
	languageSolidityStr                    = "Solidity"
	languageSourcePawnStr                  = "SourcePawn"
	languageSourcesListStr                 = "Debian Sourcelist"
	languageSpadeStr                       = "Spade"
	languageSPARQLStr                      = "SPARQL"
	languageSplineFontDatabaseStr          = "Spline Font Database"
	languageSQFStr                         = "SQF"
	languageSqlite3conStr                  = "sqlite3con"
	languageSQLPLStr                       = "SQLPL"
	languageSQLStr                         = "SQL"
	languageSquidConfStr                   = "SquidConf"
	languageSquirrelStr                    = "Squirrel"
	languageSRecodeTemplateStr             = "SRecode Template"
	languageSSHConfigStr                   = "SSH Config"
	languageSSPStr                         = "Scalate Server Page"
	languageSStr                           = "S"
	languageStanStr                        = "Stan"
	languageStarlarkStr                    = "Starlark"
	languageStasStr                        = "st(ack) as(sembler)"
	languageStataStr                       = "Stata"
	languageSTONStr                        = "STON"
	languageStylusStr                      = "Stylus"
	languageSublimeTextConfigStr           = "Sublime Text Config"
	languageSubRipTextStr                  = "SubRip Text"
	languageSugarSSStr                     = "SugarSS"
	languageSuperColliderStr               = "SuperCollider"
	languageSvelteStr                      = "Svelte"
	languageSVGStr                         = "SVG"
	languageSwiftStr                       = "Swift"
	languageSWIGStr                        = "SWIG"
	languageSYSTEMDStr                     = "systemd"
	languageSystemVerilogStr               = "SystemVerilog"
	languageTableGenStr                    = "TableGen"
	languageTADS3Str                       = "TADS 3"
	languageTAPStr                         = "TAP"
	languageTASMStr                        = "TASM"
	languageTclStr                         = "Tcl"
	languageTcshSessionStr                 = "Tcsh Session"
	languageTcshStr                        = "Tcsh"
	languageTeaStr                         = "Tea"
	languageTemplStr                       = "Templ"
	languageTeraTermStr                    = "Tera Term macro"
	languageTermcapStr                     = "Termcap"
	languageTerminfoStr                    = "Terminfo"
	languageTerraformStr                   = "Terraform"
	languageTerraStr                       = "Terra"
	languageTexinfoStr                     = "Texinfo"
	languageTeXStr                         = "TeX"
	languageTextileStr                     = "Textile"
	languageTextStr                        = "Text"
	languageThriftStr                      = "Thrift"
	languageTiddlerStr                     = "tiddler"
	languageTIProgramStr                   = "TI Program"
	languageTLAStr                         = "TLA"
	languageTNTStr                         = "Typographic Number Theory"
	languageTodotxtStr                     = "Todotxt"
	languageTOMLStr                        = "TOML"
	languageTradingViewStr                 = "TradingView"
	languageTrafficScriptStr               = "TrafficScript"
	languageTreetopStr                     = "Treetop"
	languageTSQLStr                        = "TSQL"
	languageTSVStr                         = "TSV"
	languageTSXStr                         = "TSX"
	languageTuringStr                      = "Turing"
	languageTurtleStr                      = "Turtle"
	languageTwigStr                        = "Twig"
	languageTXLStr                         = "TXL"
	languageTypeLanguageStr                = "Type Language"
	languageTypeScriptStr                  = "TypeScript"
	languageTypoScriptStr                  = "TypoScript"
	languageTypstStr                       = "Typst"
	languageUcodeStr                       = "ucode"
	languageUniconStr                      = "Unicon"
	languageUnifiedParallelCStr            = "Unified Parallel C"
	languageUnity3DAssetStr                = "Unity3D Asset"
	languageUnixAssemblyStr                = "Unix Assembly"
	languageUnknownStr                     = "Unknown"
	languageUnoStr                         = "Uno"
	languageUnrealScriptStr                = "UnrealScript"
	languageUrbiScriptStr                  = "UrbiScript"
	languageUrWebStr                       = "UrWeb"
	languageUSDStr                         = "USD"
	languageUxntalStr                      = "Uxntal"
	languageValaStr                        = "Vala"
	languageVBAStr                         = "VBA"
	languageVBNetStr                       = "VB.NET"
	languageVBScriptStr                    = "VBScript"
	languageVBStr                          = "VB"
	languageVCLSnippetsStr                 = "VCLSnippets"
	languageVCLStr                         = "VCL"
	languageVCTreeStatusStr                = "VCTreeStatus"
	languageVelocityStr                    = "Velocity"
	languageVerilogStr                     = "Verilog"
	languageVGLStr                         = "VGL"
	languageVHDLStr                        = "VHDL"
	languageVHSStr                         = "VHS"
	languageVimHelpFileStr                 = "Vim Help File"
	languageVimLStr                        = "VimL"
	languageVimScriptStr                   = "Vim Script"
	languageVimSnippetStr                  = "Vim Snippet"
	languageVisualBasicNetStr              = "Visual Basic .NET"
	languageVoltStr                        = "Volt"
	languageVShellStr                      = "V shell"
	languageVStr                           = "V"
	languageVueJSStr                       = "Vue.js"
	languageWavefrontMaterialStr           = "Wavefront Material"
	languageWavefrontObjectStr             = "Wavefront Object"
	languageWDiffStr                       = "WDiff"
	languageWdlStr                         = "wdl"
	languageWDTEStr                        = "WDTE"
	languageWebAssemblyStr                 = "WebAssembly"
	languageWebGPUShadingLanguageStr       = "WebGPU Shading Language"
	languageWebIDLStr                      = "WebIDL"
	languageWebOntologyLanguageStr         = "Web Ontology Language"
	languageWebVTTStr                      = "WebVTT"
	languageWgetConfigStr                  = "Wget Config"
	languageWhileyStr                      = "Whiley"
	languageWindowsRegistryEntriesStr      = "Windows Registry Entries"
	languageWispStr                        = "wisp"
	languageWollokStr                      = "Wollok"
	languageWowAddonDataStr                = "World of Warcraft Addon Data"
	languageX10Str                         = "X10"
	languageXAMLStr                        = "XAML"
	languageXBaseStr                       = "xBase"
	languageXBitMapStr                     = "X BitMap"
	languageXComposeStr                    = "XCompose"
	languageXCStr                          = "XC"
	languageXFontDirectoryIndexStr         = "X Font Directory Index"
	languageXMLPropertyListStr             = "XML Property List"
	languageXMLStr                         = "XML"
	languageXojoStr                        = "Xojo"
	languageXorgStr                        = "Xorg"
	languageXPagesStr                      = "XPages"
	languageXPixMapStr                     = "X PixMap"
	languageXProcStr                       = "XProc"
	languageXQueryStr                      = "XQuery"
	languageXSLTStr                        = "XSLT"
	languageXSStr                          = "XS"
	languageXtendStr                       = "Xtend"
	languageXtlangStr                      = "xtlang"
	languageYaccStr                        = "Yacc"
	languageYAMLStr                        = "YAML"
	languageYANGStr                        = "YANG"
	languageYARAStr                        = "YARA"
	languageYASnippetStr                   = "YASnippet"
	languageZ80AssemblyStr                 = "Z80 Assembly"
	languageZAPStr                         = "ZAP"
	languageZedStr                         = "Zed"
	languageZeekStr                        = "Zeek"
	languageZenScriptStr                   = "ZenScript"
	languageZephirStr                      = "Zephir"
	languageZigStr                         = "Zig"
	languageZILStr                         = "ZIL"
	languageZimplStr                       = "Zimpl"
)

// ParseLanguage parses a language from a string. Will return false
// as second parameter, if language could not be parsed.
func ParseLanguage(s string) (Language, bool) {
	lang, ok := stringToLanguage[normalizeString(s)]
	return lang, ok
}

// ParseLanguageFromChroma parses a language from a chroma lexer name.
// Will return false as second parameter, if language could not be parsed.
func ParseLanguageFromChroma(lexerName string) (Language, bool) {
	normalized := normalizeString(lexerName)

	if lang, ok := chromaToLanguage[normalized]; ok {
		return lang, true
	}

	return ParseLanguage(lexerName)
}

// String implements fmt.Stringer interface.
func (l Language) String() string {
	if s, ok := languageToString[l]; ok {
		return s
	}

	return languageUnknownStr
}

// StringChroma returns the corresponding chroma lexer name.
func (l Language) StringChroma() string {
	if s, ok := languageToChroma[l]; ok {
		return s
	}

	return l.String()
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
