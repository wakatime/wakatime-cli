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
	// LanguageArc represents the Arc programming language.
	LanguageArc
	// LanguageArduino represents the Arduino programming language.
	LanguageArduino
	// LanguageArrow represents the Arrow programming language.
	LanguageArrow
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
	// LanguageAspxCSharp represents the CSharpAspx programming language.
	LanguageAspxCSharp
	// LanguageAspxVBNet represents the VBNetAspx programming language.
	LanguageAspxVBNet
	// LanguageAssembly represents the Assembly programming language.
	LanguageAssembly
	// LanguageAsymptote represents the Asymptote programming language.
	LanguageAsymptote
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
	// LanguageBashSession represents the BashSession programming language.
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
	// LanguageDCL represents the DCL programming language.
	LanguageDCL
	// LanguageDCPU16Asm represents the DCPU16Asm programming language.
	LanguageDCPU16Asm
	// LanguageDebianControlFile represents the DebianControlFile programming language.
	LanguageDebianControlFile
	// LanguageDelphi represents the Delphi programming language.
	LanguageDelphi
	// LanguageDesktop represents the desktop programming language.
	LanguageDesktop
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
	// LangauageFreeMarker represents the FreeMarker programming language.
	LangauageFreeMarker
	// LangauageFrege represents the Frege programming language.
	LangauageFrege
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
	// LanguageGEDCOM represents the GEDCOM programming language.
	LanguageGEDCOM
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
	// LanguageJavaProperties represents the JavaProperties programming language.
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
	// LanguageJuttle represents the Juttle console programming language.
	LanguageJuttle
	// LanguageKaitai represent the Kaitai programming language.
	LanguageKaitai
	// LanguageKal represents the Kal console programming language.
	LanguageKal
	// LanguageKconfig represents the Kconfig console programming language.
	LanguageKconfig
	// LanguageKernelLog represents the Kernel log console programming language.
	LanguageKernelLog
	// LanguageKiCadLayout represent the KiCadLayout programming language.
	LanguageKiCadLayout
	// LanguageKiCadLegacyLayout represent the KiCadLegacyLayout programming language.
	LanguageKiCadLegacyLayout
	// LanguageKiCadSchematic represent the KiCadSchematic programming language.
	LanguageKiCadSchematic
	// LanguageKit represent the Kit programming language.
	LanguageKit
	// LanguageKoka represents the Koka console programming language.
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
	// LanguageLaTeX represents the LaTeX programming language.
	LanguageLaTeX
	// LanguageLatte represents the Latte programming language.
	LanguageLatte
	// LanguageLean represents the Lean programming language.
	LanguageLean
	// LanguageLess represents the Less programming language.
	LanguageLess
	// LanguageLex represents the Lex programming language.
	LanguageLex
	// LanguageLFE represents the LFE programming language.
	LanguageLFE
	// LanguageLighttpd represents the Lighttpd configuration file programming language.
	LanguageLighttpd
	// LanguageLilyPond represents the LilyPond programming language.
	LanguageLilyPond
	// LanguageLimbo represents the Limbo configuration file programming language.
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
	// LanguageLSL represents the LSL programming language.
	LanguageLSL
	// LanguageLTspiceSymbol represents the LTspiceSymbol programming language.
	LanguageLTspiceSymbol
	// LanguageLua represents the Lua programming language.
	LanguageLua
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
	// LanguageMarko represents the Marko programming language.
	LanguageMarko
	// LanguageMask represents the Mask programming language.
	LanguageMask
	// LanguageMason represents the Mason programming language.
	LanguageMason
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
	// LanguageMeson represents the Meson programming language.
	LanguageMeson
	// LanguageMetafont represents the Metafont programming language.
	LanguageMetafont
	// LanguageMetapost represents the Metapost programming language.
	LanguageMetapost
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
	// LanguageMonkey represents the Monkey programming language.
	LanguageMonkey
	// LanguageMonkeyC represents the MonkeyC programming language.
	LanguageMonkeyC
	// LanguageMonte represents the Monte programming language.
	LanguageMonte
	// LanguageMOOCode represents the MOOCode programming language.
	LanguageMOOCode
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
	// LanguageNCL represents the NCL programming language.
	LanguageNCL
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
	// LanguageNuSMV represents the NuSMV programming language.
	LanguageNuSMV
	// LanguageObjdump represents the Objdump programming language.
	LanguageObjdump
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
	// LanguagePowerShell represents the PowerShell programming language.
	LanguagePowerShell
	// LanguagePowerShellSession represents the PowerShellSession programming language.
	LanguagePowerShellSession
	// LanguagePraat represents the Praat programming language.
	LanguagePraat
	// LanguageProlog represents the Prolog programming language.
	LanguageProlog
	// LanguagePromQL represents the PromQL programming language.
	LanguagePromQL
	// LanguagePropertiesJava represents the JavaProperties programming language.
	LanguagePropertiesJava
	// LanguageProtocolBuffer represents the ProtocolBuffer programming language.
	LanguageProtocolBuffer
	// LanguagePsyShPHP represents the PHPPsySH programming language.
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
	// LanguageRegularExpression represents the RegularExpression programming language.
	LanguageRegularExpression
	// LanguageRenderScript represents the RenderScript programming language.
	LanguageRenderScript
	// LanguageRenPy represents the RenPy programming language.
	LanguageRenPy
	// LanguageRegistry represents the Registry programming language.
	LanguageRegistry
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
	// LanguageSage represents the LanguageSage programming language.
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
	// LanguageSed represents the LanguageSed programming language.
	LanguageSed
	// LanguageSelf represents the Self programming language.
	LanguageSelf
	// LanguageSGML represents the SGML programming language.
	LanguageSGML
	// LanguageShaderLab represents the LanguageShaderLab programming language.
	LanguageShaderLab
	// LanguageShell represents the Shell programming language.
	LanguageShell
	// LanguageShellSession represents the LanguageShellSession programming language.
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
	// LanguageSlice represents the LanguageSlice programming language.
	LanguageSlice
	// LanguageSlim represents the Slim programming language.
	LanguageSlim
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
	// LanguageSmPL represents the LanguageSmPL programming language.
	LanguageSmPL
	// LanguageSMT represents the LanguageSMT programming language.
	LanguageSMT
	// LanguageSnobol represents the Snobol programming language.
	LanguageSnobol
	// LanguageSnowball represents the Snowball programming language.
	LanguageSnowball
	// LanguageSolidity represents the Solidity programming language.
	LanguageSolidity
	// LanguageSourcePawn represents the SourcePawn programming language.
	LanguageSourcePawn
	// LanguageSPARQL represents the SPARQL programming language.
	LanguageSPARQL
	// LanguageSplineFontDatabase represents the LanguageSplineFontDatabase programming language.
	LanguageSplineFontDatabase
	// LanguageSourcesList represents the Debian Sourcelist programming language.
	LanguageSourcesList
	// LanguageSQF represents the LanguageSQF programming language.
	LanguageSQF
	// LanguageSQL represents the SQL programming language.
	LanguageSQL
	// LanguageSQLPL represents the LanguageSQLPL programming language.
	LanguageSQLPL
	// LanguageSqlite3con represents the sqlite3con programming language.
	LanguageSqlite3con
	// LanguageSquidConf represents the SquidConf programming language.
	LanguageSquidConf
	// LanguageSquirrel represents the LanguageSquirrel programming language.
	LanguageSquirrel
	// LanguageSRecodeTemplate represents the LanguageSRecodeTemplate programming language.
	LanguageSRecodeTemplate
	// LanguageSSHConfig represents the LanguageSSHConfig programming language.
	LanguageSSHConfig
	// LanguageSSP represents the Scalate Server Page programming language.
	LanguageSSP
	// LanguageStan represents the Stan programming language.
	LanguageStan
	// LanguageStarlark represents the LanguageStarlark programming language.
	LanguageStarlark
	// LanguageStata represents the Stata programming language.
	LanguageStata
	// LanguageSTON represents the LanguageSTON programming language.
	LanguageSTON
	// LanguageSVG represents the LanguageSVG programming language.
	LanguageSVG
	// LanguageStylus represents the Stylus programming language.
	LanguageStylus
	// LanguageSublimeTextConfig represents the SublimeTextConfig programming language.
	LanguageSublimeTextConfig
	// LanguageSubRipText represents the LanguageSubRipText programming language.
	LanguageSubRipText
	// LanguageSugarSS represents the LanguageSugarSS programming language.
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
	// LanguageZAP represents the ZAP programming language.
	LanguageZAP
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
	languageUnkownStr                      = "Unknown"
	language1CEnterpriseStr                = "1C Enterprise"
	language4DStr                          = "4D"
	languageABAPStr                        = "ABAP"
	languageABNFStr                        = "ABNF"
	languageActionScriptStr                = "ActionScript"
	languageActionScript3Str               = "ActionScript 3"
	languageAdaStr                         = "Ada"
	languageADLStr                         = "ADL"
	languageAdobeFontMetricsStr            = "Adobe Font Metrics"
	languageAdvPLStr                       = "AdvPL"
	languageAgdaStr                        = "Agda"
	languageAGSScriptStr                   = "AGS Script"
	languageAheuiStr                       = "Aheui"
	languageALStr                          = "AL"
	languageAlloyStr                       = "Alloy"
	languageAlpineAbuildStr                = "Alpine Abuild"
	languageAltiumDesignerStr              = "Altium Designer"
	languageAmbientTalkStr                 = "AmbientTalk"
	languageAMPLStr                        = "AMPL"
	languageAngelScriptStr                 = "AngelScript"
	languageAngular2Str                    = "Angular2"
	languageAnsibleStr                     = "Ansible"
	languageAntBuildSystemStr              = "Ant Build System"
	languageANTLRStr                       = "ANTLR"
	languageApacheConfStr                  = "ApacheConf"
	languageApacheConfigStr                = "Apache Config"
	languageApexStr                        = "Apex"
	languageAPIBlueprintStr                = "API Blueprint"
	languageAPLStr                         = "APL"
	languageApolloGuidanceComputerStr      = "Apollo Guidance Computer"
	languageAppleScriptStr                 = "AppleScript"
	languageArcStr                         = "Arc"
	languageArduinoStr                     = "Arduino"
	languageArrowStr                       = "Arrow"
	languageASCIIDocStr                    = "AsciiDoc"
	languageASLStr                         = "ASL"
	languageASN1Str                        = "ASN.1"
	languageASPClassicStr                  = "ASP Classic"
	languageASPDotNetStr                   = "ASP.NET"
	languageAspectJStr                     = "AspectJ"
	languageAspxCSharpStr                  = "aspx-cs"
	languageAspxVBNetStr                   = "aspx-vb"
	languageAssemblyStr                    = "Assembly"
	languageAsymptoteStr                   = "Asymptote"
	languageATSStr                         = "ATS"
	languageAugeasStr                      = "Augeas"
	languageAutoconfStr                    = "Autoconf"
	languageAutoHotkeyStr                  = "AutoHotkey"
	languageAutoItStr                      = "AutoIt"
	languageAvroIDLStr                     = "Avro IDL"
	languageAwkStr                         = "Awk"
	languageBallerinaStr                   = "Ballerina"
	languageBAREStr                        = "BARE"
	languageBashStr                        = "Bash"
	languageBashSessionStr                 = "Bash Session"
	languageBasicStr                       = "Basic"
	languageBatchfileStr                   = "Batchfile"
	languageBatchScriptStr                 = "Batch Script"
	languageBBCBasicStr                    = "BBC Basic"
	languageBBCodeStr                      = "BBCode"
	languageBCStr                          = "BC"
	languageBeefStr                        = "Beef"
	languageBefungeStr                     = "Befunge"
	languageBibTeXStr                      = "BibTeX"
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
	languageBooStr                         = "Boo"
	languageBoogieStr                      = "Boogie"
	languageBrainfuckStr                   = "Brainfuck"
	languageBrightScriptStr                = "BrightScript"
	languageBroStr                         = "Bro"
	languageBrowserslistStr                = "Browserslist"
	languageBSTStr                         = "BST"
	languageBUGSStr                        = "BUGS"
	languageCStr                           = "C"
	languageC2hsHaskellStr                 = "C2hs Haskell"
	languageCa65AssemblerStr               = "ca65 assembler"
	languageCabalConfigStr                 = "Cabal Config"
	languageCaddyfileStr                   = "Caddyfile"
	languageCaddyfileDirectivesStr         = "Caddyfile Directives"
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
	languageClojureStr                     = "Clojure"
	languageClojureScriptStr               = "ClojureScript"
	languageClosureTemplatesStr            = "Closure Templates"
	languageCloudFirestoreSecurityRulesStr = "Cloud Firestore Security Rules"
	languageCMakeStr                       = "CMake"
	languageCObjdumpStr                    = "C-ObjDump"
	languageCOBOLStr                       = "COBOL"
	languageCOBOLFreeStr                   = "COBOLFree"
	languageCocoaStr                       = "Cocoa"
	languageCodeQLStr                      = "CodeQL"
	languageCoqStr                         = "Coq"
	languageCoffeeScriptStr                = "CoffeeScript"
	languageColdfusionHTMLStr              = "ColdFusion"
	languageColdfusionCFCStr               = "ColdFusion CFC"
	languageCOLLADAStr                     = "COLLADA"
	languageCommonLispStr                  = "Common Lisp"
	languageCommonWorkflowLanguageStr      = "Common Workflow Language"
	languageComponentPascalStr             = "Component Pascal"
	languageConfigStr                      = "Config"
	languageCoNLLUStr                      = "CoNLL-U"
	languageCoolStr                        = "Cool"
	languageCPerlStr                       = "cperl"
	languageCPPStr                         = "C++"
	languageCppObjdumpStr                  = "Cpp-ObjDump"
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
	languageCsoundStr                      = "Csound"
	languageCsoundDocumentStr              = "Csound Document"
	languageCsoundOrchestraStr             = "Csound Orchestra"
	languageCsoundScoreStr                 = "Csound Score"
	languageCSSStr                         = "CSS"
	languageCSVStr                         = "CSV"
	languageCUDAStr                        = "Cuda"
	languagecURLConfigStr                  = "cURL Config"
	languageCVSStr                         = "CVS"
	languageCWebStr                        = "CWeb"
	languageCycriptStr                     = "Cycript"
	languageCypherStr                      = "Cypher"
	languageCythonStr                      = "Cython"
	languageDStr                           = "D"
	languageDafnyStr                       = "Dafny"
	languageDarcsPatchStr                  = "Darcs Patch"
	languageDartStr                        = "Dart"
	languageDataWeaveStr                   = "DataWeave"
	languageDASM16Str                      = "DASM16"
	languageDCLStr                         = "DCL"
	languageDCPU16AsmStr                   = "DCPU-16 ASM"
	languageDebianControlFileStr           = "Debian Control file"
	languageDelphiStr                      = "Delphi"
	languageDesktopStr                     = "desktop"
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
	languageDockerStr                      = "Docker"
	languageDockerfileStr                  = "Dockerfile"
	languageDocTeXStr                      = "DocTeX"
	languageDocumentationStr               = "Documentation"
	languageDogescriptStr                  = "Dogescript"
	languageDTDStr                         = "DTD"
	languageDTraceStr                      = "DTrace"
	languageDuelStr                        = "Duel"
	languageDylanStr                       = "Dylan"
	languageDylanLIDStr                    = "DylanLID"
	languageDylanSessionStr                = "Dylan session"
	languageDynASMStr                      = "DynASM"
	languageEStr                           = "E"
	languageEagleStr                       = "Eagle"
	languageEarlGreyStr                    = "Earl Grey"
	languageEasybuildStr                   = "Easybuild"
	languageEasytrieveStr                  = "Easytrieve"
	languageEBNFStr                        = "EBNF"
	languageECStr                          = "eC"
	languageEcereProjectsStr               = "Ecere Projects"
	languageECLStr                         = "ECL"
	languageEclipseStr                     = "ECLiPSe"
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
	languageErlangStr                      = "Erlang"
	languageErlangErlSessionStr            = "Erlang erl session"
	languageEshellStr                      = "Eshell"
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
	languageFlatlineStr                    = "Flatline"
	languageFloScriptStr                   = "FloScript"
	languageFishStr                        = "Fish"
	languageFLUXStr                        = "FLUX"
	languageFontStr                        = "Font"
	languageFormattedStr                   = "Formatted"
	languageForthStr                       = "Forth"
	languageFortranStr                     = "Fortran"
	languageFortranFixedStr                = "FortranFixed"
	languageFortranFreeFormStr             = "Fortran Free Form"
	languageFSharpStr                      = "F#"
	languageFoxProStr                      = "FoxPro"
	languageFreefemStr                     = "Freefem"
	langauageFreeMarkerStr                 = "FreeMarker"
	langauageFregeStr                      = "Frege"
	languageFutharkStr                     = "Futhark"
	languageFStarStr                       = "F*"
	languageFStarLiteralStr                = "FStar"
	languageGameMakerLanguageStr           = "Game Maker Language"
	languageGAMLStr                        = "GAML"
	languageGAMSStr                        = "GAMS"
	languageGapStr                         = "GAP"
	languageGasStr                         = "GAS"
	languageGCCMachineDescriptionStr       = "GCC Machine Description"
	languageGCodeStr                       = "G-code"
	languageGDBStr                         = "GDB"
	languageGDNativeStr                    = "GDNative"
	languageGDScriptStr                    = "GDScript"
	languageGEDCOMStr                      = "GEDCOM"
	languageGenieStr                       = "Genie"
	languageGenshiStr                      = "Genshi"
	languageGenshiHTMLStr                  = "Genshi HTML"
	languageGenshiTextStr                  = "Genshi Text"
	languageGentooEbuildStr                = "Gentoo Ebuild"
	languageGentooEclassStr                = "Gentoo Eclass"
	languageGerberImageStr                 = "Gerber Image"
	languageGettextCatalogStr              = "Gettext Catalog"
	languageGherkinStr                     = "Gherkin"
	languageGitStr                         = "Git"
	languageGitAttributesStr               = "Git Attributes"
	languageGitConfigStr                   = "Git Config"
	languageGLSLStr                        = "GLSL"
	languageGlyphStr                       = "Glyph"
	languageGlyphBitmapStr                 = "Glyph Bitmap Distribution Format"
	languageGNStr                          = "GN"
	languageGnuplotStr                     = "Gnuplot"
	languageGoStr                          = "Go"
	languageGolangStr                      = "Golang"
	languageGoloStr                        = "Golo"
	languageGoodDataCLStr                  = "GoodData-CL"
	languageGosuStr                        = "Gosu"
	languageGosuTemplateStr                = "Gosu Template"
	languageGraceStr                       = "Grace"
	languageGradleStr                      = "Gradle"
	languageGradleConfigStr                = "Gradle Config"
	languageGrammaticalFrameworkStr        = "Grammatical Framework"
	languageGraphModelingLanguageStr       = "Graph Modeling Language"
	languageGraphQLStr                     = "GraphQL"
	languageGraphvizDOTStr                 = "Graphviz (DOT)"
	languageGroffStr                       = "Groff"
	languageGroovyStr                      = "Groovy"
	languageGroovyServerPagesStr           = "Groovy Server Pages"
	languageHackStr                        = "Hack"
	languageHamlStr                        = "Haml"
	languageHandlebarsStr                  = "Handlebars"
	languageHAProxyStr                     = "HAProxy"
	languageHarbourStr                     = "Harbour"
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
	languageHTMLStr                        = "HTML"
	languageHTMLDjangoStr                  = "HTML+Django"
	languageHTMLECRStr                     = "HTML+ECR"
	languageHTMLEEXStr                     = "HTML+EEX"
	languageHTMLERBStr                     = "HTML+ERB"
	languageHTMLPHPStr                     = "HTML+PHP"
	languageHTMLRazorStr                   = "HTML+Razor"
	languageHTTPStr                        = "HTTP"
	languageHxmlStr                        = "HXML"
	languageHyStr                          = "Hy"
	languageHybrisStr                      = "Hybris"
	languageHyPhyStr                       = "HyPhy"
	languageIconStr                        = "Icon"
	languageIDLStr                         = "IDL"
	languageIdrisStr                       = "Idris"
	languageIgnoreListStr                  = "Ignore List"
	languageIgorStr                        = "Igor"
	languageIGORProStr                     = "IGOR Pro"
	languageImageJMacroStr                 = "ImageJ Macro"
	languageImageJPEGStr                   = "Image (jpeg)"
	languageImagePNGStr                    = "Image (png)"
	languageInform6Str                     = "Inform 6"
	languageInform6TemplateStr             = "Inform 6 template"
	languageInform7Str                     = "Inform 7"
	languageINIStr                         = "INI"
	languageInnoSetupStr                   = "Inno Setup"
	languageIoStr                          = "Io"
	languageIokeStr                        = "Ioke"
	languageIRCLogsStr                     = "IRC Logs"
	languageIsabelleStr                    = "Isabelle"
	languageIsabelleRootStr                = "Isabelle ROOT"
	languageJStr                           = "J"
	languageJAGSStr                        = "JAGS"
	languageJadeStr                        = "Jade"
	languageJasminStr                      = "Jasmin"
	languageJavaStr                        = "Java"
	languageJavaPropertiesStr              = "Java Properties"
	languageJavaScriptStr                  = "JavaScript"
	languageJavaScriptERBStr               = "JavaScript+ERB"
	languageJCLStr                         = "JCL"
	languageJFlexStr                       = "JFlex"
	languageJisonStr                       = "Jison"
	languageJisonLexStr                    = "Jison Lex"
	languageJolieStr                       = "Jolie"
	languageJSGFStr                        = "JSGF"
	languageJSONStr                        = "JSON"
	languageJSON5Str                       = "JSON5"
	languageJSONiqStr                      = "JSONiq"
	languageJSONLDStr                      = "JSONLD"
	languageJsonnetStr                     = "Jsonnet"
	languageJSONWithCommentsStr            = "JSON with Comments"
	languageJSPStr                         = "Java Server Page"
	languageJSXStr                         = "JSX"
	languageJuliaStr                       = "Julia"
	languageJuliaConsoleStr                = "Julia console"
	languageJungleStr                      = "Jungle"
	languageJupyterNotebookStr             = "Jupyter Notebook"
	languageJuttleStr                      = "Juttle"
	languageKaitaiStr                      = "Kaitai Struct"
	languageKalStr                         = "Kal"
	languageKconfigStr                     = "Kconfig"
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
	languageLaTeXStr                       = "LaTeX"
	languageLatteStr                       = "Latte"
	languageLeanStr                        = "Lean"
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
	languageLLVMStr                        = "LLVM"
	languageLLVMMIRStr                     = "LLVM-MIR"
	languageLLVMMIRBodyStr                 = "LLVM-MIR Body"
	languageLogosStr                       = "Logos"
	languageLogFileStr                     = "Log File"
	languageLogtalkStr                     = "Logtalk"
	languageLOLCODEStr                     = "LOLCODE"
	languageLookMLStr                      = "LookML"
	languageLoomScriptStr                  = "LoomScript"
	languageLSLStr                         = "LSL"
	languageLTspiceSymbolStr               = "LTspice Symbol"
	languageLuaStr                         = "Lua"
	languageMakefileStr                    = "Makefile"
	languageMakoStr                        = "Mako"
	languageManStr                         = "Man"
	languageMAQLStr                        = "MAQL"
	languageMarkdownStr                    = "Markdown"
	languageMarkoStr                       = "Marko"
	languageMaskStr                        = "Mask"
	languageMasonStr                       = "Mason"
	languageMathematicaStr                 = "Mathematica"
	languageMatlabStr                      = "Matlab"
	languageMatlabSessionStr               = "Matlab session"
	languageMaxStr                         = "Max"
	languageMaxMSPStr                      = "Max/MSP"
	languageMesonStr                       = "Meson"
	languageMetafontStr                    = "Metafont"
	languageMetapostStr                    = "Metapost"
	languageMIMEStr                        = "MIME"
	languageMiniDStr                       = "MiniD"
	languageMiniScriptStr                  = "MiniScript"
	languageMiniZincStr                    = "MiniZinc"
	languageMirahStr                       = "Mirah"
	languageMLIRStr                        = "MLIR"
	languageModelicaStr                    = "Modelica"
	languageModula2Str                     = "Modula-2"
	languageMoinWikiStr                    = "MoinMoin/Trac Wiki markup"
	languageMonkeyStr                      = "Monkey"
	languageMonkeyCStr                     = "MonkeyC"
	languageMonteStr                       = "Monte"
	languageMOOCodeStr                     = "MOOCode"
	languageMoonScriptStr                  = "MoonScript"
	languageMorrowindScriptStr             = "MorrowindScript"
	languageMoselStr                       = "Mosel"
	languageMozPreprocHashStr              = "mozhashpreproc"
	languageMozPreprocPercentStr           = "mozpercentpreproc"
	languageMQLStr                         = "MQL"
	languageMustacheStr                    = "Mustache"
	languageMscgenStr                      = "Mscgen"
	languageMSDOSSessionStr                = "MSDOS Session"
	languageMuPADStr                       = "MuPAD"
	languageMXMLStr                        = "MXML"
	languageMyghtyStr                      = "Myghty"
	languageMySQLStr                       = "MySQL"
	languageNASMStr                        = "NASM"
	languageNASMObjdumpStr                 = "objdump-nasm"
	languageNCLStr                         = "NCL"
	languageNemerleStr                     = "Nemerle"
	languageNeonStr                        = "Neon"
	languageNesCStr                        = "nesC"
	languageNewLispStr                     = "newLisp"
	languageNewspeakStr                    = "Newspeak"
	languageNginxStr                       = "Nginx"
	languageNginxConfigStr                 = "Nginx configuration file"
	languageNimrodStr                      = "Nimrod"
	languageNitStr                         = "Nit"
	languageNixStr                         = "Nix"
	languageNotmuchStr                     = "Notmuch"
	languageNSISStr                        = "NSIS"
	languageNuStr                          = "Nu"
	languageNumPyStr                       = "NumPy"
	languageNuSMVStr                       = "NuSMV"
	languageObjdumpStr                     = "objdump"
	languageObjectiveCStr                  = "Objective-C"
	languageObjectiveCPPStr                = "Objective-C++"
	languageObjectiveJStr                  = "Objective-J"
	languageOCamlStr                       = "OCaml"
	languageOctaveStr                      = "Octave"
	languageODINStr                        = "ODIN"
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
	languagePerlStr                        = "Perl"
	languagePerl6Str                       = "Perl6"
	languagePHPStr                         = "PHP"
	languagePHTMLStr                       = "PHTML"
	languagePigStr                         = "Pig"
	languagePikeStr                        = "Pike"
	languagePkgConfigStr                   = "PkgConfig"
	languagePLpgSQLStr                     = "PL/pgSQL"
	languagePointlessStr                   = "Pointless"
	languagePonyStr                        = "Pony"
	languagePostgresStr                    = "PostgreSQL SQL dialect"
	languagePostgresConsoleStr             = "PostgreSQL console (psql)"
	languagePOVRayStr                      = "POVRay"
	languagePostScriptStr                  = "PostScript"
	languagePowerShellStr                  = "PowerShell"
	languagePowerShellSessionStr           = "PowerShell Session"
	languagePraatStr                       = "Praat"
	languagePrologStr                      = "Prolog"
	languagePromQLStr                      = "PromQL"
	languagePropertiesJavaStr              = "Properties"
	languageProtocolBufferStr              = "Protocol Buffer"
	languagePsyShPHPStr                    = "PsySH console session for PHP"
	languagePugStr                         = "Pug"
	languagePuppetStr                      = "Puppet"
	languagePureDataStr                    = "Pure Data"
	languagePureScriptStr                  = "PureScript"
	languagePyPyLogStr                     = "PyPy Log"
	languagePythonStr                      = "Python"
	languagePython2Str                     = "Python 2.x"
	languagePython2TracebackStr            = "Python 2.x Traceback"
	languagePythonConsoleStr               = "Python console session"
	languagePythonTracebackStr             = "Python Traceback"
	languageQBasicStr                      = "QBasic"
	languageQMLStr                         = "QML"
	languageQVTOStr                        = "QVTO"
	languageRStr                           = "R"
	languageRacketStr                      = "Racket"
	languageRagelStr                       = "Ragel"
	languageRagelEmbeddedStr               = "Embedded Ragel"
	languageRakuStr                        = "Raku"
	languageRAMLStr                        = "RAML"
	languageRascalStr                      = "Rascal"
	languageRawTokenStr                    = "Raw token data" // nolint:gosec
	languageRConsoleStr                    = "RConsole"
	languageRdStr                          = "Rd"
	languageRDocStr                        = "RDoc"
	languageReadlineConfigStr              = "Readline Config"
	languageREALbasicStr                   = "REALbasic"
	languageReasonMLStr                    = "Reason"
	languageREBOLStr                       = "Rebol"
	languageRecordJarStr                   = "Record Jar"
	languageRedStr                         = "Red"
	languageRedcodeStr                     = "Redcode"
	languageRegistryStr                    = "reg"
	languageRegularExpressionStr           = "Regular Expression"
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
	languageRoffStr                        = "Roff"
	languageRoffManpageStr                 = "Roff Manpage"
	languageRougeStr                       = "Rouge"
	languageRPCStr                         = "RPC"
	languageRPMSpecStr                     = "RPMSpec"
	languageRQLStr                         = "RQL"
	languageRSLStr                         = "RSL"
	languageRubyStr                        = "Ruby"
	languageRubyIRBSessionStr              = "Ruby irb session"
	languageRUNOFFStr                      = "RUNOFF"
	languageRustStr                        = "Rust"
	languageSStr                           = "S"
	languageSageStr                        = "Sage"
	languageSaltStr                        = "Salt"
	languageSaltStackStr                   = "SaltStack"
	languageSARLStr                        = "SARL"
	languageSASStr                         = "SAS"
	languageSassStr                        = "Sass"
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
	languageShellStr                       = "Shell"
	languageShellSessionStr                = "ShellSession"
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
	languageSlurmStr                       = "Slurm"
	languageSmaliStr                       = "Smali"
	languageSmalltalkStr                   = "Smalltalk"
	languageSmartGameFormatStr             = "SmartGameFormat"
	languageSmartyStr                      = "Smarty"
	languageSMIMEStr                       = "S/MIME"
	languageSMLStr                         = "Standard ML"
	languageSmPLStr                        = "SmPL"
	languageSMTStr                         = "SMT"
	languageSnobolStr                      = "Snobol"
	languageSnowballStr                    = "Snowball"
	languageSolidityStr                    = "Solidity"
	languageSourcePawnStr                  = "SourcePawn"
	languageSourcesListStr                 = "Debian Sourcelist"
	languageSPARQLStr                      = "SPARQL"
	languageSplineFontDatabaseStr          = "Spline Font Database"
	languageSQFStr                         = "SQF"
	languageSQLStr                         = "SQL"
	languageSQLPLStr                       = "SQLPL"
	languageSqlite3conStr                  = "sqlite3con"
	languageSquidConfStr                   = "SquidConf"
	languageSquirrelStr                    = "Squirrel"
	languageSRecodeTemplateStr             = "SRecode Template"
	languageSSHConfigStr                   = "SSH Config"
	languageSSPStr                         = "Scalate Server Page"
	languageStanStr                        = "Stan"
	languageStarlarkStr                    = "Starlark"
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
	languageTcshStr                        = "Tcsh"
	languageTcshSessionStr                 = "Tcsh Session"
	languageTeaStr                         = "Tea"
	languageTeraTermStr                    = "Tera Term macro"
	languageTermcapStr                     = "Termcap"
	languageTerminfoStr                    = "Terminfo"
	languageTerraStr                       = "Terra"
	languageTerraformStr                   = "Terraform"
	languageTeXStr                         = "TeX"
	languageTexinfoStr                     = "Texinfo"
	languageTextStr                        = "Text"
	languageTextileStr                     = "Textile"
	languageThriftStr                      = "Thrift"
	languageTiddlerStr                     = "tiddler"
	languageTIProgramStr                   = "TI Program"
	languageTLAStr                         = "TLA"
	languageTNTStr                         = "Typographic Number Theory"
	languageTodotxtStr                     = "Todotxt"
	languageTOMLStr                        = "TOML"
	languageTradingViewStr                 = "TradingView"
	languageTrafficScriptStr               = "TrafficScript"
	languageTransactSQLStr                 = "Transact-SQL"
	languageTSQLStr                        = "TSQL"
	languageTreetopStr                     = "Treetop"
	languageTSVStr                         = "TSV"
	languageTSXStr                         = "TSX"
	languageTuringStr                      = "Turing"
	languageTurtleStr                      = "Turtle"
	languageTwigStr                        = "Twig"
	languageTXLStr                         = "TXL"
	languageTypeLanguageStr                = "Type Language"
	languageTypeScriptStr                  = "TypeScript"
	languageTypoScriptStr                  = "TypoScript"
	languageUcodeStr                       = "ucode"
	languageUniconStr                      = "Unicon"
	languageUnifiedParallelCStr            = "Unified Parallel C"
	languageUnity3DAssetStr                = "Unity3D Asset"
	languageUnixAssemblyStr                = "Unix Assembly"
	languageUnoStr                         = "Uno"
	languageUnrealScriptStr                = "UnrealScript"
	languageUrbiScriptStr                  = "UrbiScript"
	languageUrWebStr                       = "UrWeb"
	languageUSDStr                         = "USD"
	languageVStr                           = "V"
	languageValaStr                        = "Vala"
	languageVBStr                          = "VB"
	languageVBAStr                         = "VBA"
	languageVBNetStr                       = "VB.NET"
	languageVBScriptStr                    = "VBScript"
	languageVCLStr                         = "VCL"
	languageVCLSnippetsStr                 = "VCLSnippets"
	languageVCTreeStatusStr                = "VCTreeStatus"
	languageVelocityStr                    = "Velocity"
	languageVerilogStr                     = "Verilog"
	languageVGLStr                         = "VGL"
	languageVHDLStr                        = "VHDL"
	languageVimHelpFileStr                 = "Vim Help File"
	languageVimLStr                        = "VimL"
	languageVimScriptStr                   = "Vim script"
	languageVimSnippetStr                  = "Vim Snippet"
	languageVisualBasicNet                 = "Visual Basic .NET"
	languageVoltStr                        = "Volt"
	languageVueJSStr                       = "Vue.js"
	languageWavefrontMaterialStr           = "Wavefront Material"
	languageWavefrontObjectStr             = "Wavefront Object"
	languageWdlStr                         = "wdl"
	languageWDTEStr                        = "WDTE"
	languageWDiffStr                       = "WDiff"
	languageWebAssemblyStr                 = "WebAssembly"
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
	languageXCStr                          = "XC"
	languageXComposeStr                    = "XCompose"
	languageXFontDirectoryIndexStr         = "X Font Directory Index"
	languageXMLStr                         = "XML"
	languageXMLPropertyListStr             = "XML Property List"
	languageXojoStr                        = "Xojo"
	languageXorgStr                        = "Xorg"
	languageXPagesStr                      = "XPages"
	languageXPixMapStr                     = "X PixMap"
	languageXProcStr                       = "XProc"
	languageXQueryStr                      = "XQuery"
	languageXSStr                          = "XS"
	languageXSLTStr                        = "XSLT"
	languageXtendStr                       = "Xtend"
	languageXtlangStr                      = "xtlang"
	languageYaccStr                        = "Yacc"
	languageYAMLStr                        = "YAML"
	languageYANGStr                        = "YANG"
	languageYARAStr                        = "YARA"
	languageYASnippetStr                   = "YASnippet"
	languageZAPStr                         = "ZAP"
	languageZeekStr                        = "Zeek"
	languageZenScriptStr                   = "ZenScript"
	languageZephirStr                      = "Zephir"
	languageZigStr                         = "Zig"
	languageZILStr                         = "ZIL"
	languageZimplStr                       = "Zimpl"
)

const (
	languageAMPLChromaStr               = "Ampl"
	languageAssemblyChromaStr           = "GAS"
	languageAutoHotkeyChromaStr         = "autohotkey"
	languageCObjdumpChromaStr           = "c-objdump"
	languageColdfusionCFCChromaStr      = "Coldfusion CFC"
	languageColdfusionHTMLChromaStr     = "Coldfusion HTML"
	languageCppObjdumpChromaStr         = "cpp-objdump"
	languageCUDAChromaStr               = "CUDA"
	languageFSharpChromaStr             = "FSharp"
	languageEmacsLispChromaStr          = "EmacsLisp"
	languageGoHTMLTemplateChromaStr     = "Go HTML Template"
	languageGoTextTemplateChromaStr     = "Go Text Template"
	languageHxmlChromaStr               = "Hxml"
	languageJSXChromaStr                = "react"
	languageJSONLDChromaStr             = "JSON-LD"
	languageLessChromaStr               = "LessCss"
	languageLiquidChromaStr             = "liquid"
	languageMakefileChromaStr           = "Base Makefile"
	languageMarkdownChromaStr           = "markdown"
	languageNewLispChromaStr            = "NewLisp"
	languageNimrodChromaStr             = "Nim"
	languagePython3ChromaStr            = "Python 3"
	languageOocChromaStr                = "Ooc"
	languageOrgChromaStr                = "Org Mode"
	languageRChromaStr                  = "R"
	languageReasonMLChromaStr           = "ReasonML"
	languageREBOLChromaStr              = "REBOL"
	languageRexxChromaStr               = "Rexx"
	languageSYSTEMDChromaStr            = "SYSTEMD"
	languageSystemVerilogChromaStr      = "systemverilog"
	languageTextChromaStr               = "plaintext"
	languageTransactSQLChromaStr        = "Transact-SQL"
	languageTypoScriptHTMLDataChromaStr = "TypoScriptHtmlData"
	languageTypoScriptCSSDataChromaStr  = "TypoScriptCssData"
	languageVBNetChromaStr              = "VB.net"
	languageVerilogChromaStr            = "verilog"
	languageVueJSChromaStr              = "vue"
	languageWebIDLChromaStr             = "Web IDL"
)

// ParseLanguage parses a language from a string. Will return false
// as second parameter, if language could not be parsed.
// nolint:gocyclo
func ParseLanguage(s string) (Language, bool) {
	switch normalizeString(s) {
	case normalizeString(language1CEnterpriseStr):
		return Language1CEnterprise, true
	case normalizeString(language4DStr):
		return Language4D, true
	case normalizeString(languageABNFStr):
		return LanguageABNF, true
	case normalizeString(languageABAPStr):
		return LanguageABAP, true
	case normalizeString(languageAdaStr):
		return LanguageAda, true
	case normalizeString(languageADLStr):
		return LanguageADL, true
	case normalizeString(languageAdobeFontMetricsStr):
		return LanguageAdobeFontMetrics, true
	case normalizeString(languageAdvPLStr):
		return LanguageAdvPL, true
	case normalizeString(languageActionScriptStr):
		return LanguageActionScript, true
	case normalizeString(languageActionScript3Str):
		return LanguageActionScript3, true
	case normalizeString(languageAgdaStr):
		return LanguageAgda, true
	case normalizeString(languageAGSScriptStr):
		return LanguageAGSScript, true
	case normalizeString(languageAheuiStr):
		return LanguageAheui, true
	case normalizeString(languageALStr):
		return LanguageAL, true
	case normalizeString(languageAlloyStr):
		return LanguageAlloy, true
	case normalizeString(languageAlpineAbuildStr):
		return LanguageAlpineAbuild, true
	case normalizeString(languageAltiumDesignerStr):
		return LanguageAltiumDesigner, true
	case normalizeString(languageAmbientTalkStr):
		return LanguageAmbientTalk, true
	case normalizeString(languageAMPLStr):
		return LanguageAMPL, true
	case normalizeString(languageAngelScriptStr):
		return LanguageAngelScript, true
	case normalizeString(languageAngular2Str):
		return LanguageAngular2, true
	case normalizeString(languageAnsibleStr):
		return LanguageAnsible, true
	case normalizeString(languageAntBuildSystemStr):
		return LanguageAntBuildSystem, true
	case normalizeString(languageANTLRStr):
		return LanguageANTLR, true
	case normalizeString(languageApacheConfStr):
		return LanguageApacheConfig, true
	case normalizeString(languageApacheConfigStr):
		return LanguageApacheConfig, true
	case normalizeString(languageApexStr):
		return LanguageApex, true
	case normalizeString(languageAPIBlueprintStr):
		return LanguageAPIBlueprint, true
	case normalizeString(languageAPLStr):
		return LanguageAPL, true
	case normalizeString(languageApolloGuidanceComputerStr):
		return LanguageApolloGuidanceComputer, true
	case normalizeString(languageAppleScriptStr):
		return LanguageAppleScript, true
	case normalizeString(languageArcStr):
		return LanguageArc, true
	case normalizeString(languageArduinoStr):
		return LanguageArduino, true
	case normalizeString(languageArrowStr):
		return LanguageArrow, true
	case normalizeString(languageASCIIDocStr):
		return LanguageASCIIDoc, true
	case normalizeString(languageASLStr):
		return LanguageASL, true
	case normalizeString(languageASN1Str):
		return LanguageASN1, true
	case normalizeString(languageASPClassicStr):
		return LanguageClassicASP, true
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
	case normalizeString(languageATSStr):
		return LanguageATS, true
	case normalizeString(languageAugeasStr):
		return LanguageAugeas, true
	case normalizeString(languageAutoconfStr):
		return LanguageAutoconf, true
	case normalizeString(languageAutoHotkeyStr):
		return LanguageAutoHotkey, true
	case normalizeString(languageAutoItStr):
		return LanguageAutoIt, true
	case normalizeString(languageAvroIDLStr):
		return LanguageAvroIDL, true
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
		return LanguageBatchfile, true
	case normalizeString(languageBBCBasicStr):
		return LanguageBBCBasic, true
	case normalizeString(languageBBCodeStr):
		return LanguageBBCode, true
	case normalizeString(languageBCStr):
		return LanguageBC, true
	case normalizeString(languageBeefStr):
		return LanguageBeef, true
	case normalizeString(languageBefungeStr):
		return LanguageBefunge, true
	case normalizeString(languageBibTeXStr):
		return LanguageBibTeX, true
	case normalizeString(languageBisonStr):
		return LanguageBison, true
	case normalizeString(languageBitBakeStr):
		return LanguageBitBake, true
	case normalizeString(languageBladeStr):
		return LanguageBlade, true
	case normalizeString(languageBladeTemplateStr):
		return LanguageBladeTemplate, true
	case normalizeString(languageBlazorStr):
		return LanguageBlazor, true
	case normalizeString(languageBlitzBasicStr):
		return LanguageBlitzBasic, true
	case normalizeString(languageBlitzMaxStr):
		return LanguageBlitzMax, true
	case normalizeString(languageBluespecStr):
		return LanguageBluespec, true
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
	case normalizeString(languageBrowserslistStr):
		return LanguageBrowserslist, true
	case normalizeString(languageBSTStr):
		return LanguageBST, true
	case normalizeString(languageBUGSStr):
		return LanguageBUGS, true
	case normalizeString(languageCStr):
		return LanguageC, true
	case normalizeString(languageC2hsHaskellStr):
		return LanguageC2hsHaskell, true
	case normalizeString(languageCa65AssemblerStr):
		return LanguageCa65Assembler, true
	case normalizeString(languageCabalConfigStr):
		return LanguageCabalConfig, true
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
	case normalizeString(languageCartoCSSStr):
		return LanguageCartoCSS, true
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
	case normalizeString(languageCharityStr):
		return LanguageCharity, true
	case normalizeString(languageCharmciStr):
		return LanguageCharmci, true
	case normalizeString(languageCheetahStr):
		return LanguageCheetah, true
	case normalizeString(languageChucKStr):
		return LanguageChucK, true
	case normalizeString(languageCirruStr):
		return LanguageCirru, true
	case normalizeString(languageClarionStr):
		return LanguageClarion, true
	case normalizeString(languageClassicASPStr):
		return LanguageClassicASP, true
	case normalizeString(languageClayStr):
		return LanguageClay, true
	case normalizeString(languageCleanStr):
		return LanguageClean, true
	case normalizeString(languageClickStr):
		return LanguageClick, true
	case normalizeString(languageCLIPSStr):
		return LanguageCLIPS, true
	case normalizeString(languageClojureStr):
		return LanguageClojure, true
	case normalizeString(languageClojureScriptStr):
		return LanguageClojureScript, true
	case normalizeString(languageClosureTemplatesStr):
		return LanguageClosureTemplates, true
	case normalizeString(languageCloudFirestoreSecurityRulesStr):
		return LanguageCloudFirestoreSecurityRules, true
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
	case normalizeString(languageCodeQLStr):
		return LanguageCodeQL, true
	case normalizeString(languageCoffeeScriptStr):
		return LanguageCoffeeScript, true
	case normalizeString(languageColdfusionCFCStr):
		return LanguageColdfusionCFC, true
	case normalizeString(languageColdfusionHTMLStr):
		return LanguageColdfusionHTML, true
	case normalizeString(languageCOLLADAStr):
		return LanguageCOLLADA, true
	case normalizeString(languageCommonLispStr):
		return LanguageCommonLisp, true
	case normalizeString(languageCommonWorkflowLanguageStr):
		return LanguageCommonWorkflowLanguage, true
	case normalizeString(languageComponentPascalStr):
		return LanguageComponentPascal, true
	case normalizeString(languageConfigStr):
		return LanguageConfig, true
	case normalizeString(languageCoNLLUStr):
		return LanguageCoNLLU, true
	case normalizeString(languageCoolStr):
		return LanguageCool, true
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
	case normalizeString(languageCreoleStr):
		return LanguageCreole, true
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
	case normalizeString(languageCsoundStr):
		return LanguageCsound, true
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
	case normalizeString(languagecURLConfigStr):
		return LanguagecURLConfig, true
	case normalizeString(languageCVSStr):
		return LanguageCVS, true
	case normalizeString(languageCWebStr):
		return LanguageCWeb, true
	case normalizeString(languageCycriptStr):
		return LanguageCycript, true
	case normalizeString(languageCypherStr):
		return LanguageCypher, true
	case normalizeString(languageCythonStr):
		return LanguageCython, true
	case normalizeString(languageDStr):
		return LanguageD, true
	case normalizeString(languageDafnyStr):
		return LanguageDafny, true
	case normalizeString(languageDarcsPatchStr):
		return LanguageDarcsPatch, true
	case normalizeString(languageDartStr):
		return LanguageDart, true
	case normalizeString(languageDataWeaveStr):
		return LanguageDataWeave, true
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
	case normalizeString(languageDesktopStr):
		return LanguageDesktop, true
	case normalizeString(languageDevicetreeStr):
		return LanguageDevicetree, true
	case normalizeString(languageDGStr):
		return LanguageDG, true
	case normalizeString(languageDhallStr):
		return LanguageDhall, true
	case normalizeString(languageDiffStr):
		return LanguageDiff, true
	case normalizeString(languageDigitalCommandStr):
		return LanguageDigitalCommand, true
	case normalizeString(languageDircolorsStr):
		return LanguageDircolors, true
	case normalizeString(languageDirectX3DFileStr):
		return LanguageDirectX3DFile, true
	case normalizeString(languageDjangoJinjaStr):
		return LanguageDjangoJinja, true
	case normalizeString(languageDMStr):
		return LanguageDM, true
	case normalizeString(languageDNSZoneStr):
		return LanguageDNSZone, true
	case normalizeString(languageDObjdumpStr):
		return LanguageDObjdump, true
	case normalizeString(languageDockerStr):
		return LanguageDocker, true
	case normalizeString(languageDockerfileStr):
		return LanguageDockerfile, true
	case normalizeString(languageDocTeXStr):
		return LanguageDocTeX, true
	case normalizeString(languageDocumentationStr):
		return LanguageDocumentation, true
	case normalizeString(languageDogescriptStr):
		return LanguageDogescript, true
	case normalizeString(languageDTDStr):
		return LanguageDTD, true
	case normalizeString(languageDTraceStr):
		return LanguageDTrace, true
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
	case normalizeString(languageEStr):
		return LanguageE, true
	case normalizeString(languageEagleStr):
		return LanguageEagle, true
	case normalizeString(languageEarlGreyStr):
		return LanguageEarlGrey, true
	case normalizeString(languageEasybuildStr):
		return LanguageEasybuild, true
	case normalizeString(languageEasytrieveStr):
		return LanguageEasytrieve, true
	case normalizeString(languageEBNFStr):
		return LanguageEBNF, true
	case normalizeString(languageECStr):
		return LanguageEC, true
	case normalizeString(languageEcereProjectsStr):
		return LanguageEcereProjects, true
	case normalizeString(languageECLStr):
		return LanguageECL, true
	case normalizeString(languageEclipseStr):
		return LanguageEclipse, true
	case normalizeString(languageEditorConfigStr):
		return LanguageEditorConfig, true
	case normalizeString(languageEdjeDataCollectionStr):
		return LanguageEdjeDataCollection, true
	case normalizeString(languageEdnStr):
		return LanguageEdn, true
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
	case normalizeString(languageEMailStr):
		return LanguageEMail, true
	case normalizeString(languageEmberScriptStr):
		return LanguageEmberScript, true
	case normalizeString(languageEMLStr):
		return LanguageEML, true
	case normalizeString(languageEQStr):
		return LanguageEQ, true
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
	case normalizeString(languageFaustStr):
		return LanguageFaust, true
	case normalizeString(languageFelixStr):
		return LanguageFelix, true
	case normalizeString(languageFennelStr):
		return LanguageFennel, true
	case normalizeString(languageFIGletFontStr):
		return LanguageFIGletFont, true
	case normalizeString(languageFilebenchWMLStr):
		return LanguageFilebenchWML, true
	case normalizeString(languageFilterscriptStr):
		return LanguageFilterscript, true
	case normalizeString(languageFlatlineStr):
		return LanguageFlatline, true
	case normalizeString(languageFloScriptStr):
		return LanguageFloScript, true
	case normalizeString(languageFishStr):
		return LanguageFish, true
	case normalizeString(languageFLUXStr):
		return LanguageFLUX, true
	case normalizeString(languageFontStr):
		return LanguageFont, true
	case normalizeString(languageFormattedStr):
		return LanguageFormatted, true
	case normalizeString(languageForthStr):
		return LanguageForth, true
	case normalizeString(languageFortranStr):
		return LanguageFortran, true
	case normalizeString(languageFortranFixedStr):
		return LanguageFortranFixed, true
	case normalizeString(languageFortranFreeFormStr):
		return LanguageFortranFreeForm, true
	case normalizeString(languageFSharpStr):
		return LanguageFSharp, true
	case normalizeString(languageFoxProStr):
		return LanguageFoxPro, true
	case normalizeString(languageFreefemStr):
		return LanguageFreefem, true
	case normalizeString(langauageFreeMarkerStr):
		return LangauageFreeMarker, true
	case normalizeString(langauageFregeStr):
		return LangauageFrege, true
	case normalizeString(languageFutharkStr):
		return LanguageFuthark, true
	case normalizeString(languageFStarStr):
		return LanguageFStar, true
	case normalizeString(languageFStarLiteralStr):
		return LanguageFStar, true
	case normalizeString(languageGameMakerLanguageStr):
		return LanguageGameMakerLanguage, true
	case normalizeString(languageGAMLStr):
		return LanguageGAML, true
	case normalizeString(languageGAMSStr):
		return LanguageGAMS, true
	case normalizeString(languageGapStr):
		return LanguageGap, true
	case normalizeString(languageGasStr):
		return LanguageGas, true
	case normalizeString(languageGCCMachineDescriptionStr):
		return LanguageGCCMachineDescription, true
	case normalizeString(languageGCodeStr):
		return LanguageGCode, true
	case normalizeString(languageGDBStr):
		return LanguageGDB, true
	case normalizeString(languageGDNativeStr):
		return LanguageGDNative, true
	case normalizeString(languageGDScriptStr):
		return LanguageGDScript, true
	case normalizeString(languageGEDCOMStr):
		return LanguageGEDCOM, true
	case normalizeString(languageGenieStr):
		return LanguageGenie, true
	case normalizeString(languageGenshiStr):
		return LanguageGenshi, true
	case normalizeString(languageGenshiHTMLStr):
		return LanguageGenshiHTML, true
	case normalizeString(languageGenshiTextStr):
		return LanguageGenshiText, true
	case normalizeString(languageGentooEbuildStr):
		return LanguageGentooEbuild, true
	case normalizeString(languageGentooEclassStr):
		return LanguageGentooEclass, true
	case normalizeString(languageGerberImageStr):
		return LanguageGerberImage, true
	case normalizeString(languageGettextCatalogStr):
		return LanguageGettextCatalog, true
	case normalizeString(languageGherkinStr):
		return LanguageGherkin, true
	case normalizeString(languageGitStr):
		return LanguageGit, true
	case normalizeString(languageGitAttributesStr):
		return LanguageGitAttributes, true
	case normalizeString(languageGitConfigStr):
		return LanguageGitConfig, true
	case normalizeString(languageGLSLStr):
		return LanguageGLSL, true
	case normalizeString(languageGlyphStr):
		return LanguageGlyph, true
	case normalizeString(languageGlyphBitmapStr):
		return LanguageGlyphBitmap, true
	case normalizeString(languageGNStr):
		return LanguageGN, true
	case normalizeString(languageGnuplotStr):
		return LanguageGnuplot, true
	case normalizeString(languageGoStr):
		return LanguageGo, true
	case normalizeString(languageGolangStr):
		return LanguageGo, true
	case normalizeString(languageGoloStr):
		return LanguageGolo, true
	case normalizeString(languageGoodDataCLStr):
		return LanguageGoodDataCL, true
	case normalizeString(languageGosuStr):
		return LanguageGosu, true
	case normalizeString(languageGosuTemplateStr):
		return LanguageGosuTemplate, true
	case normalizeString(languageGraceStr):
		return LanguageGrace, true
	case normalizeString(languageGradleStr):
		return LanguageGradle, true
	case normalizeString(languageGradleConfigStr):
		return LanguageGradleConfig, true
	case normalizeString(languageGrammaticalFrameworkStr):
		return LanguageGrammaticalFramework, true
	case normalizeString(languageGraphModelingLanguageStr):
		return LanguageGraphModelingLanguage, true
	case normalizeString(languageGraphQLStr):
		return LanguageGraphQL, true
	case normalizeString(languageGraphvizDOTStr):
		return LanguageGraphvizDOT, true
	case normalizeString(languageGroffStr):
		return LanguageGroff, true
	case normalizeString(languageGroovyStr):
		return LanguageGroovy, true
	case normalizeString(languageGroovyServerPagesStr):
		return LanguageGroovyServerPages, true
	case normalizeString(languageHackStr):
		return LanguageHack, true
	case normalizeString(languageHamlStr):
		return LanguageHaml, true
	case normalizeString(languageHandlebarsStr):
		return LanguageHandlebars, true
	case normalizeString(languageHAProxyStr):
		return LanguageHAProxy, true
	case normalizeString(languageHarbourStr):
		return LanguageHarbour, true
	case normalizeString(languageHaskellStr):
		return LanguageHaskell, true
	case normalizeString(languageHaxeStr):
		return LanguageHaxe, true
	case normalizeString(languageHCLStr):
		return LanguageHCL, true
	case normalizeString(languageHexdumpStr):
		return LanguageHexdump, true
	case normalizeString(languageHiveQLStr):
		return LanguageHiveQL, true
	case normalizeString(languageHLBStr):
		return LanguageHLB, true
	case normalizeString(languageHLSLStr):
		return LanguageHLSL, true
	case normalizeString(languageHolyCStr):
		return LanguageHolyC, true
	case normalizeString(languageHSAILStr):
		return LanguageHSAIL, true
	case normalizeString(languageHspecStr):
		return LanguageHspec, true
	case normalizeString(languageHTMLStr):
		return LanguageHTML, true
	case normalizeString(languageHTMLDjangoStr):
		return LanguageHTMLDjango, true
	case normalizeString(languageHTMLECRStr):
		return LanguageHTMLECR, true
	case normalizeString(languageHTMLEEXStr):
		return LanguageHTMLEEX, true
	case normalizeString(languageHTMLERBStr):
		return LanguageHTMLERB, true
	case normalizeString(languageHTMLPHPStr):
		return LanguageHTMLPHP, true
	case normalizeString(languageHTMLRazorStr):
		return LanguageHTMLRazor, true
	case normalizeString(languageHTTPStr):
		return LanguageHTTP, true
	case normalizeString(languageHxmlStr):
		return LanguageHxml, true
	case normalizeString(languageHyStr):
		return LanguageHy, true
	case normalizeString(languageHybrisStr):
		return LanguageHybris, true
	case normalizeString(languageHyPhyStr):
		return LanguageHyPhy, true
	case normalizeString(languageIconStr):
		return LanguageIcon, true
	case normalizeString(languageIDLStr):
		return LanguageIDL, true
	case normalizeString(languageIdrisStr):
		return LanguageIdris, true
	case normalizeString(languageIgnoreListStr):
		return LanguageIgnoreList, true
	case normalizeString(languageIgorStr):
		return LanguageIgor, true
	case normalizeString(languageIGORProStr):
		return LanguageIGORPro, true
	case normalizeString(languageImageJMacroStr):
		return LanguageImageJMacro, true
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
	case normalizeString(languageInnoSetupStr):
		return LanguageInnoSetup, true
	case normalizeString(languageIoStr):
		return LanguageIo, true
	case normalizeString(languageIokeStr):
		return LanguageIoke, true
	case normalizeString(languageIRCLogsStr):
		return LanguageIRCLogs, true
	case normalizeString(languageIsabelleStr):
		return LanguageIsabelle, true
	case normalizeString(languageIsabelleRootStr):
		return LanguageIsabelleRoot, true
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
	case normalizeString(languageJavaPropertiesStr):
		return LanguageJavaProperties, true
	case normalizeString(languageJavaScriptStr):
		return LanguageJavaScript, true
	case normalizeString(languageJavaScriptERBStr):
		return LanguageJavaScriptERB, true
	case normalizeString(languageJCLStr):
		return LanguageJCL, true
	case normalizeString(languageJFlexStr):
		return LanguageJFlex, true
	case normalizeString(languageJisonStr):
		return LanguageJison, true
	case normalizeString(languageJisonLexStr):
		return LanguageJisonLex, true
	case normalizeString(languageJolieStr):
		return LanguageJolie, true
	case normalizeString(languageJSGFStr):
		return LanguageJSGF, true
	case normalizeString(languageJSONStr):
		return LanguageJSON, true
	case normalizeString(languageJSON5Str):
		return LanguageJSON5, true
	case normalizeString(languageJSONiqStr):
		return LanguageJSONiq, true
	case normalizeString(languageJSONLDStr):
		return LanguageJSONLD, true
	case normalizeString(languageJsonnetStr):
		return LanguageJsonnet, true
	case normalizeString(languageJSONWithCommentsStr):
		return LanguageJSONWithComments, true
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
	case normalizeString(languageJupyterNotebookStr):
		return LanguageJupyterNotebook, true
	case normalizeString(languageJuttleStr):
		return LanguageJuttle, true
	case normalizeString(languageKaitaiStr):
		return LanguageKaitai, true
	case normalizeString(languageKalStr):
		return LanguageKal, true
	case normalizeString(languageKconfigStr):
		return LanguageKconfig, true
	case normalizeString(languageKernelLogStr):
		return LanguageKernelLog, true
	case normalizeString(languageKiCadLayoutStr):
		return LanguageKiCadLayout, true
	case normalizeString(languageKiCadLegacyLayoutStr):
		return LanguageKiCadLegacyLayout, true
	case normalizeString(languageKiCadSchematicStr):
		return LanguageKiCadSchematic, true
	case normalizeString(languageKitStr):
		return LanguageKit, true
	case normalizeString(languageKokaStr):
		return LanguageKoka, true
	case normalizeString(languageKotlinStr):
		return LanguageKotlin, true
	case normalizeString(languageKRLStr):
		return LanguageKRL, true
	case normalizeString(languageLabVIEWStr):
		return LanguageLabVIEW, true
	case normalizeString(languageLaravelTemplateStr):
		return LanguageLaravelTemplate, true
	case normalizeString(languageLarkStr):
		return LanguageLark, true
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
	case normalizeString(languageLexStr):
		return LanguageLex, true
	case normalizeString(languageLFEStr):
		return LanguageLFE, true
	case normalizeString(languageLighttpdStr):
		return LanguageLighttpd, true
	case normalizeString(languageLilyPondStr):
		return LanguageLilyPond, true
	case normalizeString(languageLimboStr):
		return LanguageLimbo, true
	case normalizeString(languageLinkerScriptStr):
		return LanguageLinkerScript, true
	case normalizeString(languageLinuxKernelModuleStr):
		return LanguageLinuxKernelModule, true
	case normalizeString(languageLiquidStr):
		return LanguageLiquid, true
	case normalizeString(languageLiterateAgdaStr):
		return LanguageLiterateAgda, true
	case normalizeString(languageLiterateCoffeeScriptStr):
		return LanguageLiterateCoffeeScript, true
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
	case normalizeString(languageLOLCODEStr):
		return LanguageLOLCODE, true
	case normalizeString(languageLookMLStr):
		return LanguageLookML, true
	case normalizeString(languageLoomScriptStr):
		return LanguageLoomScript, true
	case normalizeString(languageLSLStr):
		return LanguageLSL, true
	case normalizeString(languageLTspiceSymbolStr):
		return LanguageLTspiceSymbol, true
	case normalizeString(languageLuaStr):
		return LanguageLua, true
	case normalizeString(languageMakefileStr):
		return LanguageMakefile, true
	case normalizeString(languageMakoStr):
		return LanguageMako, true
	case normalizeString(languageManStr):
		return LanguageMan, true
	case normalizeString(languageMAQLStr):
		return LanguageMAQL, true
	case normalizeString(languageMarkdownStr):
		return LanguageMarkdown, true
	case normalizeString(languageMarkoStr):
		return LanguageMarko, true
	case normalizeString(languageMaskStr):
		return LanguageMask, true
	case normalizeString(languageMasonStr):
		return LanguageMason, true
	case normalizeString(languageMathematicaStr):
		return LanguageMathematica, true
	case normalizeString(languageMatlabStr):
		return LanguageMatlab, true
	case normalizeString(languageMatlabSessionStr):
		return LanguageMatlabSession, true
	case normalizeString(languageMaxStr):
		return LanguageMax, true
	case normalizeString(languageMaxMSPStr):
		return LanguageMaxMSP, true
	case normalizeString(languageMesonStr):
		return LanguageMeson, true
	case normalizeString(languageMetafontStr):
		return LanguageMetafont, true
	case normalizeString(languageMetapostStr):
		return LanguageMetapost, true
	case normalizeString(languageMIMEStr):
		return LanguageMIME, true
	case normalizeString(languageMiniDStr):
		return LanguageMiniD, true
	case normalizeString(languageMiniScriptStr):
		return LanguageMiniScript, true
	case normalizeString(languageMiniZincStr):
		return LanguageMiniZinc, true
	case normalizeString(languageMirahStr):
		return LanguageMirah, true
	case normalizeString(languageMLIRStr):
		return LanguageMLIR, true
	case normalizeString(languageModelicaStr):
		return LanguageModelica, true
	case normalizeString(languageModula2Str):
		return LanguageModula2, true
	case normalizeString(languageMoinWikiStr):
		return LanguageMoinWiki, true
	case normalizeString(languageMonkeyStr):
		return LanguageMonkey, true
	case normalizeString(languageMonkeyCStr):
		return LanguageMonkeyC, true
	case normalizeString(languageMonteStr):
		return LanguageMonte, true
	case normalizeString(languageMOOCodeStr):
		return LanguageMOOCode, true
	case normalizeString(languageMoonScriptStr):
		return LanguageMoonScript, true
	case normalizeString(languageMorrowindScriptStr):
		return LanguageMorrowindScript, true
	case normalizeString(languageMoselStr):
		return LanguageMosel, true
	case normalizeString(languageMozPreprocHashStr):
		return LanguageMozPreprocHash, true
	case normalizeString(languageMozPreprocPercentStr):
		return LanguageMozPreprocPercent, true
	case normalizeString(languageMQLStr):
		return LanguageMQL, true
	case normalizeString(languageMscgenStr):
		return LanguageMscgen, true
	case normalizeString(languageMSDOSSessionStr):
		return LanguageMSDOSSession, true
	case normalizeString(languageMuPADStr):
		return LanguageMuPAD, true
	case normalizeString(languageMXMLStr):
		return LanguageMXML, true
	case normalizeString(languageMyghtyStr):
		return LanguageMyghty, true
	case normalizeString(languageMySQLStr):
		return LanguageMySQL, true
	case normalizeString(languageMustacheStr):
		return LanguageMustache, true
	case normalizeString(languageNASMStr):
		return LanguageNASM, true
	case normalizeString(languageNCLStr):
		return LanguageNCL, true
	case normalizeString(languageNemerleStr):
		return LanguageNemerle, true
	case normalizeString(languageNeonStr):
		return LanguageNeon, true
	case normalizeString(languageNesCStr):
		return LanguageNesC, true
	case normalizeString(languageNewLispStr):
		return LanguageNewLisp, true
	case normalizeString(languageNewspeakStr):
		return LanguageNewspeak, true
	case normalizeString(languageNginxStr):
		return LanguageNginx, true
	case normalizeString(languageNginxConfigStr):
		return LanguageNginxConfig, true
	case normalizeString(languageNimrodStr):
		return LanguageNimrod, true
	case normalizeString(languageNitStr):
		return LanguageNit, true
	case normalizeString(languageNixStr):
		return LanguageNix, true
	case normalizeString(languageNotmuchStr):
		return LanguageNotmuch, true
	case normalizeString(languageNSISStr):
		return LanguageNSIS, true
	case normalizeString(languageNuStr):
		return LanguageNu, true
	case normalizeString(languageNumPyStr):
		return LanguageNumPy, true
	case normalizeString(languageNuSMVStr):
		return LanguageNuSMV, true
	case normalizeString(languageNASMObjdumpStr):
		return LanguageNASMObjdump, true
	case normalizeString(languageObjdumpStr):
		return LanguageObjdump, true
	case normalizeString(languageObjectiveCStr):
		return LanguageObjectiveC, true
	case normalizeString(languageObjectiveCPPStr):
		return LanguageObjectiveCPP, true
	case normalizeString(languageObjectiveJStr):
		return LanguageObjectiveJ, true
	case normalizeString(languageOCamlStr):
		return LanguageOCaml, true
	case normalizeString(languageOctaveStr):
		return LanguageOctave, true
	case normalizeString(languageODINStr):
		return LanguageODIN, true
	case normalizeString(languageOocStr):
		return LanguageOoc, true
	case normalizeString(languageOpaStr):
		return LanguageOpa, true
	case normalizeString(languageOpenEdgeABLStr):
		return LanguageOpenEdgeABL, true
	case normalizeString(languageOpenSCADStr):
		return LanguageOpenSCAD, true
	case normalizeString(languageOrgStr):
		return LanguageOrg, true
	case normalizeString(languagePacmanConfStr):
		return LanguagePacmanConf, true
	case normalizeString(languagePanStr):
		return LanguagePan, true
	case normalizeString(languageParaSailStr):
		return LanguageParaSail, true
	case normalizeString(languageParrotStr):
		return LanguageParrot, true
	case normalizeString(languagePascalStr):
		return LanguagePascal, true
	case normalizeString(languagePawnStr):
		return LanguagePawn, true
	case normalizeString(languagePEGStr):
		return LanguagePEG, true
	case normalizeString(languagePerlStr):
		return LanguagePerl, true
	case normalizeString(languagePerl6Str):
		return LanguagePerl6, true
	case normalizeString(languagePHPStr):
		return LanguagePHP, true
	case normalizeString(languagePHTMLStr):
		return LanguagePHTML, true
	case normalizeString(languagePigStr):
		return LanguagePig, true
	case normalizeString(languagePikeStr):
		return LanguagePike, true
	case normalizeString(languagePkgConfigStr):
		return LanguagePkgConfig, true
	case normalizeString(languagePLpgSQLStr):
		return LanguagePLpgSQL, true
	case normalizeString(languagePointlessStr):
		return LanguagePointless, true
	case normalizeString(languagePonyStr):
		return LanguagePony, true
	case normalizeString(languagePostgresStr):
		return LanguagePostgres, true
	case normalizeString(languagePostgresConsoleStr):
		return LanguagePostgresConsole, true
	case normalizeString(languagePostScriptStr):
		return LanguagePostScript, true
	case normalizeString(languagePOVRayStr):
		return LanguagePOVRay, true
	case normalizeString(languagePowerShellStr):
		return LanguagePowerShell, true
	case normalizeString(languagePowerShellSessionStr):
		return LanguagePowerShellSession, true
	case normalizeString(languagePraatStr):
		return LanguagePraat, true
	case normalizeString(languagePrologStr):
		return LanguageProlog, true
	case normalizeString(languagePromQLStr):
		return LanguagePromQL, true
	case normalizeString(languagePropertiesJavaStr):
		return LanguagePropertiesJava, true
	case normalizeString(languageProtocolBufferStr):
		return LanguageProtocolBuffer, true
	case normalizeString(languagePsyShPHPStr):
		return LanguagePsyShPHP, true
	case normalizeString(languagePugStr):
		return LanguagePug, true
	case normalizeString(languagePuppetStr):
		return LanguagePuppet, true
	case normalizeString(languagePureDataStr):
		return LanguagePureData, true
	case normalizeString(languagePureScriptStr):
		return LanguagePureScript, true
	case normalizeString(languagePyPyLogStr):
		return LanguagePyPyLog, true
	case normalizeString(languagePythonStr):
		return LanguagePython, true
	case normalizeString(languagePython2Str):
		return LanguagePython2, true
	case normalizeString(languagePython2TracebackStr):
		return LanguagePython2Traceback, true
	case normalizeString(languagePythonConsoleStr):
		return LanguagePythonConsole, true
	case normalizeString(languagePythonTracebackStr):
		return LanguagePythonTraceback, true
	case normalizeString(languageQBasicStr):
		return LanguageQBasic, true
	case normalizeString(languageQMLStr):
		return LanguageQML, true
	case normalizeString(languageQVTOStr):
		return LanguageQVTO, true
	case normalizeString(languageRStr):
		return LanguageR, true
	case normalizeString(languageRacketStr):
		return LanguageRacket, true
	case normalizeString(languageRagelStr):
		return LanguageRagel, true
	case normalizeString(languageRagelEmbeddedStr):
		return LanguageRagelEmbedded, true
	case normalizeString(languageRakuStr):
		return LanguageRaku, true
	case normalizeString(languageRAMLStr):
		return LanguageRAML, true
	case normalizeString(languageRascalStr):
		return LanguageRascal, true
	case normalizeString(languageRawTokenStr):
		return LanguageRawToken, true
	case normalizeString(languageRConsoleStr):
		return LanguageRConsole, true
	case normalizeString(languageRdStr):
		return LanguageRd, true
	case normalizeString(languageRDocStr):
		return LanguageRDoc, true
	case normalizeString(languageReadlineConfigStr):
		return LanguageReadlineConfig, true
	case normalizeString(languageREALbasicStr):
		return LanguageREALbasic, true
	case normalizeString(languageReasonMLStr):
		return LanguageReasonML, true
	case normalizeString(languageREBOLStr):
		return LanguageREBOL, true
	case normalizeString(languageRecordJarStr):
		return LanguageRecordJar, true
	case normalizeString(languageRedStr):
		return LanguageRed, true
	case normalizeString(languageRedcodeStr):
		return LanguageRedcode, true
	case normalizeString(languageRegistryStr):
		return LanguageRegistry, true
	case normalizeString(languageRegularExpressionStr):
		return LanguageRegularExpression, true
	case normalizeString(languageRenderScriptStr):
		return LanguageRenderScript, true
	case normalizeString(languageRenPyStr):
		return LanguageRenPy, true
	case normalizeString(languageReScriptStr):
		return LanguageReScript, true
	case normalizeString(languageResourceBundleStr):
		return LanguageResourceBundle, true
	case normalizeString(languageReStructuredTextStr):
		return LanguageReStructuredText, true
	case normalizeString(languageRexxStr):
		return LanguageRexx, true
	case normalizeString(languageRHTMLStr):
		return LanguageRHTML, true
	case normalizeString(languageRichTextFormatStr):
		return LanguageRichTextFormat, true
	case normalizeString(languageRideStr):
		return LanguageRide, true
	case normalizeString(languageRingStr):
		return LanguageRing, true
	case normalizeString(languageRiotStr):
		return LanguageRiot, true
	case normalizeString(languageRMarkdownStr):
		return LanguageRMarkdown, true
	case normalizeString(languageRNGCompactStr):
		return LanguageRNGCompact, true
	case normalizeString(languageRoboconfGraphStr):
		return LanguageRoboconfGraph, true
	case normalizeString(languageRoboconfInstancesStr):
		return LanguageRoboconfInstances, true
	case normalizeString(languageRobotFrameworkStr):
		return LanguageRobotFramework, true
	case normalizeString(languageRoffStr):
		return LanguageRoff, true
	case normalizeString(languageRoffManpageStr):
		return LanguageRoffManpage, true
	case normalizeString(languageRougeStr):
		return LanguageRouge, true
	case normalizeString(languageRPCStr):
		return LanguageRPC, true
	case normalizeString(languageRPMSpecStr):
		return LanguageRPMSpec, true
	case normalizeString(languageRQLStr):
		return LanguageRQL, true
	case normalizeString(languageRSLStr):
		return LanguageRSL, true
	case normalizeString(languageRubyStr):
		return LanguageRuby, true
	case normalizeString(languageRubyIRBSessionStr):
		return LanguageRubyIRBSession, true
	case normalizeString(languageRUNOFFStr):
		return LanguageRUNOFF, true
	case normalizeString(languageRustStr):
		return LanguageRust, true
	case normalizeString(languageSStr):
		return LanguageS, true
	case normalizeString(languageSageStr):
		return LanguageSage, true
	case normalizeString(languageSaltStr):
		return LanguageSalt, true
	case normalizeString(languageSaltStackStr):
		return LanguageSalt, true
	case normalizeString(languageSARLStr):
		return LanguageSARL, true
	case normalizeString(languageSASStr):
		return LanguageSAS, true
	case normalizeString(languageSassStr):
		return LanguageSass, true
	case normalizeString(languageScalaStr):
		return LanguageScala, true
	case normalizeString(languageScamlStr):
		return LanguageScaml, true
	case normalizeString(languageScdocStr):
		return LanguageScdoc, true
	case normalizeString(languageSchemeStr):
		return LanguageScheme, true
	case normalizeString(languageScilabStr):
		return LanguageScilab, true
	case normalizeString(languageScribeStr):
		return LanguageScribe, true
	case normalizeString(languageSCSSStr):
		return LanguageSCSS, true
	case normalizeString(languageSedStr):
		return LanguageSed, true
	case normalizeString(languageSelfStr):
		return LanguageSelf, true
	case normalizeString(languageSGMLStr):
		return LanguageSGML, true
	case normalizeString(languageShaderLabStr):
		return LanguageShaderLab, true
	case normalizeString(languageShellStr):
		return LanguageShell, true
	case normalizeString(languageShellSessionStr):
		return LanguageShellSession, true
	case normalizeString(languageShenStr):
		return LanguageShen, true
	case normalizeString(languageShExCStr):
		return LanguageShExC, true
	case normalizeString(languageSieveStr):
		return LanguageSieve, true
	case normalizeString(languageSilverStr):
		return LanguageSilver, true
	case normalizeString(languageSimulaStr):
		return LanguageSimula, true
	case normalizeString(languageSingularityStr):
		return LanguageSingularity, true
	case normalizeString(languageSketchDrawingStr):
		return LanguageSketchDrawing, true
	case normalizeString(languageSKILLStr):
		return LanguageSKILL, true
	case normalizeString(languageSlashStr):
		return LanguageSlash, true
	case normalizeString(languageSliceStr):
		return LanguageSlice, true
	case normalizeString(languageSlimStr):
		return LanguageSlim, true
	case normalizeString(languageSlurmStr):
		return LanguageSlurm, true
	case normalizeString(languageSmaliStr):
		return LanguageSmali, true
	case normalizeString(languageSmalltalkStr):
		return LanguageSmalltalk, true
	case normalizeString(languageSmartGameFormatStr):
		return LanguageSmartGameFormat, true
	case normalizeString(languageSmartyStr):
		return LanguageSmarty, true
	case normalizeString(languageSMIMEStr):
		return LanguageSMIME, true
	case normalizeString(languageSMLStr):
		return LanguageSML, true
	case normalizeString(languageSmPLStr):
		return LanguageSmPL, true
	case normalizeString(languageSMTStr):
		return LanguageSMT, true
	case normalizeString(languageSnobolStr):
		return LanguageSnobol, true
	case normalizeString(languageSnowballStr):
		return LanguageSnowball, true
	case normalizeString(languageSolidityStr):
		return LanguageSolidity, true
	case normalizeString(languageSourcePawnStr):
		return LanguageSourcePawn, true
	case normalizeString(languageSourcesListStr):
		return LanguageSourcesList, true
	case normalizeString(languageSPARQLStr):
		return LanguageSPARQL, true
	case normalizeString(languageSplineFontDatabaseStr):
		return LanguageSplineFontDatabase, true
	case normalizeString(languageSQFStr):
		return LanguageSQF, true
	case normalizeString(languageSQLStr):
		return LanguageSQL, true
	case normalizeString(languageSQLPLStr):
		return LanguageSQLPL, true
	case normalizeString(languageSqlite3conStr):
		return LanguageSqlite3con, true
	case normalizeString(languageSquidConfStr):
		return LanguageSquidConf, true
	case normalizeString(languageSquirrelStr):
		return LanguageSquirrel, true
	case normalizeString(languageSRecodeTemplateStr):
		return LanguageSRecodeTemplate, true
	case normalizeString(languageSSHConfigStr):
		return LanguageSSHConfig, true
	case normalizeString(languageSSPStr):
		return LanguageSSP, true
	case normalizeString(languageStanStr):
		return LanguageStan, true
	case normalizeString(languageStarlarkStr):
		return LanguageStarlark, true
	case normalizeString(languageStataStr):
		return LanguageStata, true
	case normalizeString(languageSTONStr):
		return LanguageSTON, true
	case normalizeString(languageStylusStr):
		return LanguageStylus, true
	case normalizeString(languageSublimeTextConfigStr):
		return LanguageSublimeTextConfig, true
	case normalizeString(languageSubRipTextStr):
		return LanguageSubRipText, true
	case normalizeString(languageSugarSSStr):
		return LanguageSugarSS, true
	case normalizeString(languageSuperColliderStr):
		return LanguageSuperCollider, true
	case normalizeString(languageSvelteStr):
		return LanguageSvelte, true
	case normalizeString(languageSVGStr):
		return LanguageSVG, true
	case normalizeString(languageSwiftStr):
		return LanguageSwift, true
	case normalizeString(languageSWIGStr):
		return LanguageSWIG, true
	case normalizeString(languageSYSTEMDStr):
		return LanguageSYSTEMD, true
	case normalizeString(languageSystemVerilogStr):
		return LanguageSystemVerilog, true
	case normalizeString(languageTableGenStr):
		return LanguageTableGen, true
	case normalizeString(languageTADS3Str):
		return LanguageTADS3, true
	case normalizeString(languageTAPStr):
		return LanguageTAP, true
	case normalizeString(languageTASMStr):
		return LanguageTASM, true
	case normalizeString(languageTclStr):
		return LanguageTcl, true
	case normalizeString(languageTcshStr):
		return LanguageTcsh, true
	case normalizeString(languageTcshSessionStr):
		return LanguageTcshSession, true
	case normalizeString(languageTeaStr):
		return LanguageTea, true
	case normalizeString(languageTeraTermStr):
		return LanguageTeraTerm, true
	case normalizeString(languageTermcapStr):
		return LanguageTermcap, true
	case normalizeString(languageTerminfoStr):
		return LanguageTerminfo, true
	case normalizeString(languageTerraStr):
		return LanguageTerra, true
	case normalizeString(languageTerraformStr):
		return LanguageTerraform, true
	case normalizeString(languageTeXStr):
		return LanguageTeX, true
	case normalizeString(languageTexinfoStr):
		return LanguageTexinfo, true
	case normalizeString(languageTextStr):
		return LanguageText, true
	case normalizeString(languageTextileStr):
		return LanguageTextile, true
	case normalizeString(languageThriftStr):
		return LanguageThrift, true
	case normalizeString(languageTiddlerStr):
		return LanguageTiddler, true
	case normalizeString(languageTIProgramStr):
		return LanguageTIProgram, true
	case normalizeString(languageTLAStr):
		return LanguageTLA, true
	case normalizeString(languageTNTStr):
		return LanguageTNT, true
	case normalizeString(languageTodotxtStr):
		return LanguageTodotxt, true
	case normalizeString(languageTOMLStr):
		return LanguageTOML, true
	case normalizeString(languageTradingViewStr):
		return LanguageTradingView, true
	case normalizeString(languageTrafficScriptStr):
		return LanguageTrafficScript, true
	case normalizeString(languageTransactSQLStr):
		return LanguageTransactSQL, true
	case normalizeString(languageTSQLStr):
		return LanguageTransactSQL, true
	case normalizeString(languageTreetopStr):
		return LanguageTreetop, true
	case normalizeString(languageTSVStr):
		return LanguageTSV, true
	case normalizeString(languageTSXStr):
		return LanguageTSX, true
	case normalizeString(languageTuringStr):
		return LanguageTuring, true
	case normalizeString(languageTurtleStr):
		return LanguageTurtle, true
	case normalizeString(languageTwigStr):
		return LanguageTwig, true
	case normalizeString(languageTXLStr):
		return LanguageTXL, true
	case normalizeString(languageTypeLanguageStr):
		return LanguageTypeLanguage, true
	case normalizeString(languageTypeScriptStr):
		return LanguageTypeScript, true
	case normalizeString(languageTypoScriptStr):
		return LanguageTypoScript, true
	case normalizeString(languageUcodeStr):
		return LanguageUcode, true
	case normalizeString(languageUniconStr):
		return LanguageUnicon, true
	case normalizeString(languageUnifiedParallelCStr):
		return LanguageUnifiedParallelC, true
	case normalizeString(languageUnity3DAssetStr):
		return LanguageUnity3DAsset, true
	case normalizeString(languageUnixAssemblyStr):
		return LanguageUnixAssembly, true
	case normalizeString(languageUnoStr):
		return LanguageUno, true
	case normalizeString(languageUnrealScriptStr):
		return LanguageUnrealScript, true
	case normalizeString(languageUrbiScriptStr):
		return LanguageUrbiScript, true
	case normalizeString(languageUrWebStr):
		return LanguageUrWeb, true
	case normalizeString(languageUSDStr):
		return LanguageUSD, true
	case normalizeString(languageVStr):
		return LanguageV, true
	case normalizeString(languageValaStr):
		return LanguageVala, true
	case normalizeString(languageVBStr):
		return LanguageVB, true
	case normalizeString(languageVBAStr):
		return LanguageVBA, true
	case normalizeString(languageVBNetStr):
		return LanguageVBNet, true
	case normalizeString(languageVBScriptStr):
		return LanguageVBScript, true
	case normalizeString(languageVCLStr):
		return LanguageVCL, true
	case normalizeString(languageVCLSnippetsStr):
		return LanguageVCLSnippets, true
	case normalizeString(languageVCTreeStatusStr):
		return LanguageVCTreeStatus, true
	case normalizeString(languageVelocityStr):
		return LanguageVelocity, true
	case normalizeString(languageVerilogStr):
		return LanguageVerilog, true
	case normalizeString(languageVGLStr):
		return LanguageVGL, true
	case normalizeString(languageVHDLStr):
		return LanguageVHDL, true
	case normalizeString(languageVimHelpFileStr):
		return LanguageVimHelpFile, true
	case normalizeString(languageVimLStr):
		return LanguageVimL, true
	case normalizeString(languageVimScriptStr):
		return LanguageVimScript, true
	case normalizeString(languageVimSnippetStr):
		return LanguageVimSnippet, true
	case normalizeString(languageVisualBasicNet):
		return LanguageVBNet, true
	case normalizeString(languageVoltStr):
		return LanguageVolt, true
	case normalizeString(languageVueJSStr):
		return LanguageVueJS, true
	case normalizeString(languageWavefrontMaterialStr):
		return LanguageWavefrontMaterial, true
	case normalizeString(languageWavefrontObjectStr):
		return LanguageWavefrontObject, true
	case normalizeString(languageWdlStr):
		return LanguageWdl, true
	case normalizeString(languageWDTEStr):
		return LanguageWDTE, true
	case normalizeString(languageWDiffStr):
		return LanguageWDiff, true
	case normalizeString(languageWebAssemblyStr):
		return LanguageWebAssembly, true
	case normalizeString(languageWebIDLStr):
		return LanguageWebIDL, true
	case normalizeString(languageWebOntologyLanguageStr):
		return LanguageWebOntologyLanguage, true
	case normalizeString(languageWebVTTStr):
		return LanguageWebVTT, true
	case normalizeString(languageWgetConfigStr):
		return LanguageWgetConfig, true
	case normalizeString(languageWhileyStr):
		return LanguageWhiley, true
	case normalizeString(languageWindowsRegistryEntriesStr):
		return LanguageWindowsRegistryEntries, true
	case normalizeString(languageWispStr):
		return LanguageWisp, true
	case normalizeString(languageWollokStr):
		return LanguageWollok, true
	case normalizeString(languageWowAddonDataStr):
		return LanguageWowAddonData, true
	case normalizeString(languageX10Str):
		return LanguageX10, true
	case normalizeString(languageXAMLStr):
		return LanguageXAML, true
	case normalizeString(languageXBaseStr):
		return LanguageXBase, true
	case normalizeString(languageXBitMapStr):
		return LanguageXBitMap, true
	case normalizeString(languageXCStr):
		return LanguageXC, true
	case normalizeString(languageXComposeStr):
		return LanguageXCompose, true
	case normalizeString(languageXFontDirectoryIndexStr):
		return LanguageXFontDirectoryIndex, true
	case normalizeString(languageXMLStr):
		return LanguageXML, true
	case normalizeString(languageXMLPropertyListStr):
		return LanguageXMLPropertyList, true
	case normalizeString(languageXojoStr):
		return LanguageXojo, true
	case normalizeString(languageXorgStr):
		return LanguageXorg, true
	case normalizeString(languageXPagesStr):
		return LanguageXPages, true
	case normalizeString(languageXPixMapStr):
		return LanguageXPixMap, true
	case normalizeString(languageXProcStr):
		return LanguageXProc, true
	case normalizeString(languageXQueryStr):
		return LanguageXQuery, true
	case normalizeString(languageXSStr):
		return LanguageXS, true
	case normalizeString(languageXSLTStr):
		return LanguageXSLT, true
	case normalizeString(languageXtendStr):
		return LanguageXtend, true
	case normalizeString(languageXtlangStr):
		return LanguageXtlang, true
	case normalizeString(languageYaccStr):
		return LanguageYacc, true
	case normalizeString(languageYAMLStr):
		return LanguageYAML, true
	case normalizeString(languageYANGStr):
		return LanguageYANG, true
	case normalizeString(languageYARAStr):
		return LanguageYARA, true
	case normalizeString(languageYASnippetStr):
		return LanguageYASnippet, true
	case normalizeString(languageZAPStr):
		return LanguageZAP, true
	case normalizeString(languageZeekStr):
		return LanguageZeek, true
	case normalizeString(languageZenScriptStr):
		return LanguageZenScript, true
	case normalizeString(languageZephirStr):
		return LanguageZephir, true
	case normalizeString(languageZigStr):
		return LanguageZig, true
	case normalizeString(languageZILStr):
		return LanguageZIL, true
	case normalizeString(languageZimplStr):
		return LanguageZimpl, true
	default:
		return LanguageUnknown, false
	}
}

// ParseLanguageFromChroma parses a language from a chroma lexer name.
// Will return false as second parameter, if language could not be parsed.
// nolint:gocyclo
func ParseLanguageFromChroma(lexerName string) (Language, bool) {
	switch normalizeString(lexerName) {
	case normalizeString(languageAMPLChromaStr):
		return LanguageAMPL, true
	case normalizeString(languageAssemblyChromaStr):
		return LanguageAssembly, true
	case normalizeString(languageAutoHotkeyChromaStr):
		return LanguageAutoHotkey, true
	case normalizeString(languageCObjdumpChromaStr):
		return LanguageCObjdump, true
	case normalizeString(languageColdfusionCFCChromaStr):
		return LanguageColdfusionCFC, true
	case normalizeString(languageColdfusionHTMLChromaStr):
		return LanguageColdfusionHTML, true
	case normalizeString(languageCppObjdumpChromaStr):
		return LanguageCppObjdump, true
	case normalizeString(languageCUDAChromaStr):
		return LanguageCUDA, true
	case normalizeString(languageEmacsLispChromaStr):
		return LanguageEmacsLisp, true
	case normalizeString(languageGoHTMLTemplateChromaStr):
		return LanguageGo, true
	case normalizeString(languageGoTextTemplateChromaStr):
		return LanguageGo, true
	case normalizeString(languageFSharpChromaStr):
		return LanguageFSharp, true
	case normalizeString(languageHxmlChromaStr):
		return LanguageHxml, true
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
	case normalizeString(languageNewLispChromaStr):
		return LanguageNewLisp, true
	case normalizeString(languageNimrodChromaStr):
		return LanguageNimrod, true
	case normalizeString(languageOocChromaStr):
		return LanguageOoc, true
	case normalizeString(languageOrgChromaStr):
		return LanguageOrg, true
	case normalizeString(languagePython3ChromaStr):
		return LanguagePython, true
	case normalizeString(languageRChromaStr):
		return LanguageS, true
	case normalizeString(languageRexxChromaStr):
		return LanguageRexx, true
	case normalizeString(languageReasonMLChromaStr):
		return LanguageReasonML, true
	case normalizeString(languageREBOLChromaStr):
		return LanguageREBOL, true
	case normalizeString(languageSystemVerilogChromaStr):
		return LanguageSystemVerilog, true
	case normalizeString(languageTextChromaStr):
		return LanguageText, true
	case normalizeString(languageTransactSQLChromaStr):
		return LanguageTransactSQL, true
	case normalizeString(languageTypoScriptCSSDataChromaStr):
		return LanguageTypoScript, true
	case normalizeString(languageTypoScriptHTMLDataChromaStr):
		return LanguageTypoScript, true
	case normalizeString(languageVBNetChromaStr):
		return LanguageVBNet, true
	case normalizeString(languageVerilogChromaStr):
		return LanguageVerilog, true
	case normalizeString(languageVueJSChromaStr):
		return LanguageVueJS, true
	case normalizeString(languageWebIDLChromaStr):
		return LanguageWebIDL, true
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
	case Language1CEnterprise:
		return language1CEnterpriseStr
	case Language4D:
		return language4DStr
	case LanguageABAP:
		return languageABAPStr
	case LanguageABNF:
		return languageABNFStr
	case LanguageAda:
		return languageAdaStr
	case LanguageADL:
		return languageADLStr
	case LanguageAdobeFontMetrics:
		return languageAdobeFontMetricsStr
	case LanguageAdvPL:
		return languageAdvPLStr
	case LanguageActionScript:
		return languageActionScriptStr
	case LanguageActionScript3:
		return languageActionScript3Str
	case LanguageAgda:
		return languageAgdaStr
	case LanguageAGSScript:
		return languageAGSScriptStr
	case LanguageAheui:
		return languageAheuiStr
	case LanguageAL:
		return languageALStr
	case LanguageAlloy:
		return languageAlloyStr
	case LanguageAlpineAbuild:
		return languageAlpineAbuildStr
	case LanguageAltiumDesigner:
		return languageAltiumDesignerStr
	case LanguageAmbientTalk:
		return languageAmbientTalkStr
	case LanguageAMPL:
		return languageAMPLStr
	case LanguageAngelScript:
		return languageAngelScriptStr
	case LanguageAngular2:
		return languageAngular2Str
	case LanguageAnsible:
		return languageAnsibleStr
	case LanguageAntBuildSystem:
		return languageAntBuildSystemStr
	case LanguageANTLR:
		return languageANTLRStr
	case LanguageApacheConfig:
		return languageApacheConfigStr
	case LanguageApex:
		return languageApexStr
	case LanguageAPIBlueprint:
		return languageAPIBlueprintStr
	case LanguageAPL:
		return languageAPLStr
	case LanguageApolloGuidanceComputer:
		return languageApolloGuidanceComputerStr
	case LanguageAppleScript:
		return languageAppleScriptStr
	case LanguageArc:
		return languageArcStr
	case LanguageArduino:
		return languageArduinoStr
	case LanguageArrow:
		return languageArrowStr
	case LanguageASCIIDoc:
		return languageASCIIDocStr
	case LanguageASL:
		return languageASLStr
	case LanguageASN1:
		return languageASN1Str
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
	case LanguageATS:
		return languageATSStr
	case LanguageAugeas:
		return languageAugeasStr
	case LanguageAutoconf:
		return languageAutoconfStr
	case LanguageAutoHotkey:
		return languageAutoHotkeyStr
	case LanguageAutoIt:
		return languageAutoItStr
	case LanguageAvroIDL:
		return languageAvroIDLStr
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
	case LanguageBBCBasic:
		return languageBBCBasicStr
	case LanguageBBCode:
		return languageBBCodeStr
	case LanguageBC:
		return languageBCStr
	case LanguageBeef:
		return languageBeefStr
	case LanguageBefunge:
		return languageBefungeStr
	case LanguageBibTeX:
		return languageBibTeXStr
	case LanguageBison:
		return languageBisonStr
	case LanguageBitBake:
		return languageBitBakeStr
	case LanguageBlade:
		return languageBladeStr
	case LanguageBladeTemplate:
		return languageBladeTemplateStr
	case LanguageBlazor:
		return languageBlazorStr
	case LanguageBlitzBasic:
		return languageBlitzBasicStr
	case LanguageBlitzMax:
		return languageBlitzMaxStr
	case LanguageBluespec:
		return languageBluespecStr
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
	case LanguageBrowserslist:
		return languageBrowserslistStr
	case LanguageBST:
		return languageBSTStr
	case LanguageBUGS:
		return languageBUGSStr
	case LanguageC:
		return languageCStr
	case LanguageC2hsHaskell:
		return languageC2hsHaskellStr
	case LanguageCa65Assembler:
		return languageCa65AssemblerStr
	case LanguageCabalConfig:
		return languageCabalConfigStr
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
	case LanguageCartoCSS:
		return languageCartoCSSStr
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
	case LanguageCharity:
		return languageCharityStr
	case LanguageCharmci:
		return languageCharmciStr
	case LanguageCheetah:
		return languageCheetahStr
	case LanguageChucK:
		return languageChucKStr
	case LanguageCirru:
		return languageCirruStr
	case LanguageClarion:
		return languageClarionStr
	case LanguageClassicASP:
		return languageClassicASPStr
	case LanguageClay:
		return languageClayStr
	case LanguageClean:
		return languageCleanStr
	case LanguageClick:
		return languageClickStr
	case LanguageCLIPS:
		return languageCLIPSStr
	case LanguageClojure:
		return languageClojureStr
	case LanguageClojureScript:
		return languageClojureScriptStr
	case LanguageClosureTemplates:
		return languageClosureTemplatesStr
	case LanguageCloudFirestoreSecurityRules:
		return languageCloudFirestoreSecurityRulesStr
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
	case LanguageCodeQL:
		return languageCodeQLStr
	case LanguageCoffeeScript:
		return languageCoffeeScriptStr
	case LanguageColdfusionCFC:
		return languageColdfusionCFCStr
	case LanguageColdfusionHTML:
		return languageColdfusionHTMLStr
	case LanguageCOLLADA:
		return languageCOLLADAStr
	case LanguageCommonLisp:
		return languageCommonLispStr
	case LanguageCommonWorkflowLanguage:
		return languageCommonWorkflowLanguageStr
	case LanguageComponentPascal:
		return languageComponentPascalStr
	case LanguageConfig:
		return languageConfigStr
	case LanguageCoNLLU:
		return languageCoNLLUStr
	case LanguageCool:
		return languageCoolStr
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
	case LanguageCreole:
		return languageCreoleStr
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
	case LanguageCsound:
		return languageCsoundStr
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
	case LanguagecURLConfig:
		return languagecURLConfigStr
	case LanguageCVS:
		return languageCVSStr
	case LanguageCWeb:
		return languageCWebStr
	case LanguageCycript:
		return languageCycriptStr
	case LanguageCypher:
		return languageCypherStr
	case LanguageCython:
		return languageCythonStr
	case LanguageD:
		return languageDStr
	case LanguageDafny:
		return languageDafnyStr
	case LanguageDarcsPatch:
		return languageDarcsPatchStr
	case LanguageDart:
		return languageDartStr
	case LanguageDataWeave:
		return languageDataWeaveStr
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
	case LanguageDesktop:
		return languageDesktopStr
	case LanguageDevicetree:
		return languageDevicetreeStr
	case LanguageDG:
		return languageDGStr
	case LanguageDhall:
		return languageDhallStr
	case LanguageDiff:
		return languageDiffStr
	case LanguageDigitalCommand:
		return languageDigitalCommandStr
	case LanguageDircolors:
		return languageDircolorsStr
	case LanguageDirectX3DFile:
		return languageDirectX3DFileStr
	case LanguageDjangoJinja:
		return languageDjangoJinjaStr
	case LanguageDM:
		return languageDMStr
	case LanguageDNSZone:
		return languageDNSZoneStr
	case LanguageDObjdump:
		return languageDObjdumpStr
	case LanguageDocker:
		return languageDockerStr
	case LanguageDockerfile:
		return languageDockerfileStr
	case LanguageDocTeX:
		return languageDocTeXStr
	case LanguageDocumentation:
		return languageDocumentationStr
	case LanguageDogescript:
		return languageDogescriptStr
	case LanguageDTD:
		return languageDTDStr
	case LanguageDTrace:
		return languageDTraceStr
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
	case LanguageE:
		return languageEStr
	case LanguageEagle:
		return languageEagleStr
	case LanguageEarlGrey:
		return languageEarlGreyStr
	case LanguageEasybuild:
		return languageEasybuildStr
	case LanguageEasytrieve:
		return languageEasytrieveStr
	case LanguageEBNF:
		return languageEBNFStr
	case LanguageEC:
		return languageECStr
	case LanguageEcereProjects:
		return languageEcereProjectsStr
	case LanguageECL:
		return languageECLStr
	case LanguageEclipse:
		return languageEclipseStr
	case LanguageEditorConfig:
		return languageEditorConfigStr
	case LanguageEdjeDataCollection:
		return languageEdjeDataCollectionStr
	case LanguageEdn:
		return languageEdnStr
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
	case LanguageEmberScript:
		return languageEmberScriptStr
	case LanguageEML:
		return languageEMLStr
	case LanguageEQ:
		return languageEQStr
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
	case LanguageFaust:
		return languageFaustStr
	case LanguageFelix:
		return languageFelixStr
	case LanguageFennel:
		return languageFennelStr
	case LanguageFIGletFont:
		return languageFIGletFontStr
	case LanguageFilebenchWML:
		return languageFilebenchWMLStr
	case LanguageFilterscript:
		return languageFilterscriptStr
	case LanguageFlatline:
		return languageFlatlineStr
	case LanguageFloScript:
		return languageFloScriptStr
	case LanguageFish:
		return languageFishStr
	case LanguageFLUX:
		return languageFLUXStr
	case LanguageFont:
		return languageFontStr
	case LanguageFormatted:
		return languageFormattedStr
	case LanguageForth:
		return languageForthStr
	case LanguageFortran:
		return languageFortranStr
	case LanguageFortranFixed:
		return languageFortranFixedStr
	case LanguageFortranFreeForm:
		return languageFortranFreeFormStr
	case LanguageFSharp:
		return languageFSharpStr
	case LanguageFoxPro:
		return languageFoxProStr
	case LanguageFreefem:
		return languageFreefemStr
	case LangauageFreeMarker:
		return langauageFreeMarkerStr
	case LangauageFrege:
		return langauageFregeStr
	case LanguageFuthark:
		return languageFutharkStr
	case LanguageFStar:
		return languageFStarStr
	case LanguageGameMakerLanguage:
		return languageGameMakerLanguageStr
	case LanguageGAML:
		return languageGAMLStr
	case LanguageGAMS:
		return languageGAMSStr
	case LanguageGap:
		return languageGapStr
	case LanguageGas:
		return languageGasStr
	case LanguageGCCMachineDescription:
		return languageGCCMachineDescriptionStr
	case LanguageGCode:
		return languageGCodeStr
	case LanguageGDB:
		return languageGDBStr
	case LanguageGDScript:
		return languageGDScriptStr
	case LanguageGEDCOM:
		return languageGEDCOMStr
	case LanguageGenie:
		return languageGenieStr
	case LanguageGenshi:
		return languageGenshiStr
	case LanguageGenshiHTML:
		return languageGenshiHTMLStr
	case LanguageGenshiText:
		return languageGenshiTextStr
	case LanguageGentooEbuild:
		return languageGentooEbuildStr
	case LanguageGentooEclass:
		return languageGentooEclassStr
	case LanguageGerberImage:
		return languageGerberImageStr
	case LanguageGettextCatalog:
		return languageGettextCatalogStr
	case LanguageGherkin:
		return languageGherkinStr
	case LanguageGDNative:
		return languageGDNativeStr
	case LanguageGit:
		return languageGitStr
	case LanguageGitAttributes:
		return languageGitAttributesStr
	case LanguageGitConfig:
		return languageGitConfigStr
	case LanguageGLSL:
		return languageGLSLStr
	case LanguageGlyph:
		return languageGlyphStr
	case LanguageGlyphBitmap:
		return languageGlyphBitmapStr
	case LanguageGN:
		return languageGNStr
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
	case LanguageGrace:
		return languageGraceStr
	case LanguageGradle:
		return languageGradleStr
	case LanguageGradleConfig:
		return languageGradleConfigStr
	case LanguageGrammaticalFramework:
		return languageGrammaticalFrameworkStr
	case LanguageGraphModelingLanguage:
		return languageGraphModelingLanguageStr
	case LanguageGraphQL:
		return languageGraphQLStr
	case LanguageGraphvizDOT:
		return languageGraphvizDOTStr
	case LanguageGroff:
		return languageGroffStr
	case LanguageGroovy:
		return languageGroovyStr
	case LanguageGroovyServerPages:
		return languageGroovyServerPagesStr
	case LanguageHack:
		return languageHackStr
	case LanguageHaml:
		return languageHamlStr
	case LanguageHandlebars:
		return languageHandlebarsStr
	case LanguageHAProxy:
		return languageHAProxyStr
	case LanguageHarbour:
		return languageHarbourStr
	case LanguageHaskell:
		return languageHaskellStr
	case LanguageHaxe:
		return languageHaxeStr
	case LanguageHCL:
		return languageHCLStr
	case LanguageHexdump:
		return languageHexdumpStr
	case LanguageHiveQL:
		return languageHiveQLStr
	case LanguageHLB:
		return languageHLBStr
	case LanguageHLSL:
		return languageHLSLStr
	case LanguageHolyC:
		return languageHolyCStr
	case LanguageHSAIL:
		return languageHSAILStr
	case LanguageHspec:
		return languageHspecStr
	case LanguageHTML:
		return languageHTMLStr
	case LanguageHTMLDjango:
		return languageHTMLDjangoStr
	case LanguageHTMLECR:
		return languageHTMLECRStr
	case LanguageHTMLEEX:
		return languageHTMLEEXStr
	case LanguageHTMLERB:
		return languageHTMLERBStr
	case LanguageHTMLPHP:
		return languageHTMLPHPStr
	case LanguageHTMLRazor:
		return languageHTMLRazorStr
	case LanguageHTTP:
		return languageHTTPStr
	case LanguageHxml:
		return languageHxmlStr
	case LanguageHy:
		return languageHyStr
	case LanguageHybris:
		return languageHybrisStr
	case LanguageHyPhy:
		return languageHyPhyStr
	case LanguageIcon:
		return languageIconStr
	case LanguageIDL:
		return languageIDLStr
	case LanguageIdris:
		return languageIdrisStr
	case LanguageIgnoreList:
		return languageIgnoreListStr
	case LanguageIgor:
		return languageIgorStr
	case LanguageIGORPro:
		return languageIGORProStr
	case LanguageImageJMacro:
		return languageImageJMacroStr
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
	case LanguageInnoSetup:
		return languageInnoSetupStr
	case LanguageIo:
		return languageIoStr
	case LanguageIoke:
		return languageIokeStr
	case LanguageIRCLogs:
		return languageIRCLogsStr
	case LanguageIsabelle:
		return languageIsabelleStr
	case LanguageIsabelleRoot:
		return languageIsabelleRootStr
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
	case LanguageJavaProperties:
		return languageJavaPropertiesStr
	case LanguageJavaScript:
		return languageJavaScriptStr
	case LanguageJavaScriptERB:
		return languageJavaScriptERBStr
	case LanguageJCL:
		return languageJCLStr
	case LanguageJFlex:
		return languageJFlexStr
	case LanguageJison:
		return languageJisonStr
	case LanguageJisonLex:
		return languageJisonLexStr
	case LanguageJolie:
		return languageJolieStr
	case LanguageJSGF:
		return languageJSGFStr
	case LanguageJSON:
		return languageJSONStr
	case LanguageJSON5:
		return languageJSON5Str
	case LanguageJSONiq:
		return languageJSONiqStr
	case LanguageJSONLD:
		return languageJSONLDStr
	case LanguageJsonnet:
		return languageJsonnetStr
	case LanguageJSONWithComments:
		return languageJSONWithCommentsStr
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
	case LanguageJupyterNotebook:
		return languageJupyterNotebookStr
	case LanguageJuttle:
		return languageJuttleStr
	case LanguageKaitai:
		return languageKaitaiStr
	case LanguageKal:
		return languageKalStr
	case LanguageKconfig:
		return languageKconfigStr
	case LanguageKernelLog:
		return languageKernelLogStr
	case LanguageKiCadLayout:
		return languageKiCadLayoutStr
	case LanguageKiCadLegacyLayout:
		return languageKiCadLegacyLayoutStr
	case LanguageKiCadSchematic:
		return languageKiCadSchematicStr
	case LanguageKit:
		return languageKitStr
	case LanguageKoka:
		return languageKokaStr
	case LanguageKotlin:
		return languageKotlinStr
	case LanguageKRL:
		return languageKRLStr
	case LanguageLabVIEW:
		return languageLabVIEWStr
	case LanguageLaravelTemplate:
		return languageLaravelTemplateStr
	case LanguageLark:
		return languageLarkStr
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
	case LanguageLex:
		return languageLexStr
	case LanguageLFE:
		return languageLFEStr
	case LanguageLighttpd:
		return languageLighttpdStr
	case LanguageLilyPond:
		return languageLilyPondStr
	case LanguageLimbo:
		return languageLimboStr
	case LanguageLinkerScript:
		return languageLinkerScriptStr
	case LanguageLinuxKernelModule:
		return languageLinuxKernelModuleStr
	case LanguageLiquid:
		return languageLiquidStr
	case LanguageLiterateAgda:
		return languageLiterateAgdaStr
	case LanguageLiterateCoffeeScript:
		return languageLiterateCoffeeScriptStr
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
	case LanguageLOLCODE:
		return languageLOLCODEStr
	case LanguageLookML:
		return languageLookMLStr
	case LanguageLoomScript:
		return languageLoomScriptStr
	case LanguageLSL:
		return languageLSLStr
	case LanguageLTspiceSymbol:
		return languageLTspiceSymbolStr
	case LanguageLua:
		return languageLuaStr
	case LanguageMakefile:
		return languageMakefileStr
	case LanguageMako:
		return languageMakoStr
	case LanguageMan:
		return languageManStr
	case LanguageMAQL:
		return languageMAQLStr
	case LanguageMarkdown:
		return languageMarkdownStr
	case LanguageMarko:
		return languageMarkoStr
	case LanguageMask:
		return languageMaskStr
	case LanguageMason:
		return languageMasonStr
	case LanguageMathematica:
		return languageMathematicaStr
	case LanguageMatlab:
		return languageMatlabStr
	case LanguageMatlabSession:
		return languageMatlabSessionStr
	case LanguageMax:
		return languageMaxStr
	case LanguageMaxMSP:
		return languageMaxMSPStr
	case LanguageMeson:
		return languageMesonStr
	case LanguageMetafont:
		return languageMetafontStr
	case LanguageMetapost:
		return languageMetapostStr
	case LanguageMIME:
		return languageMIMEStr
	case LanguageMiniD:
		return languageMiniDStr
	case LanguageMiniScript:
		return languageMiniScriptStr
	case LanguageMiniZinc:
		return languageMiniZincStr
	case LanguageMirah:
		return languageMirahStr
	case LanguageMLIR:
		return languageMLIRStr
	case LanguageModelica:
		return languageModelicaStr
	case LanguageModula2:
		return languageModula2Str
	case LanguageMoinWiki:
		return languageMoinWikiStr
	case LanguageMonkey:
		return languageMonkeyStr
	case LanguageMonkeyC:
		return languageMonkeyCStr
	case LanguageMonte:
		return languageMonteStr
	case LanguageMOOCode:
		return languageMOOCodeStr
	case LanguageMoonScript:
		return languageMoonScriptStr
	case LanguageMorrowindScript:
		return languageMorrowindScriptStr
	case LanguageMosel:
		return languageMoselStr
	case LanguageMozPreprocHash:
		return languageMozPreprocHashStr
	case LanguageMozPreprocPercent:
		return languageMozPreprocPercentStr
	case LanguageMQL:
		return languageMQLStr
	case LanguageMscgen:
		return languageMscgenStr
	case LanguageMSDOSSession:
		return languageMSDOSSessionStr
	case LanguageMuPAD:
		return languageMuPADStr
	case LanguageMXML:
		return languageMXMLStr
	case LanguageMyghty:
		return languageMyghtyStr
	case LanguageMySQL:
		return languageMySQLStr
	case LanguageMustache:
		return languageMustacheStr
	case LanguageNASM:
		return languageNASMStr
	case LanguageNCL:
		return languageNCLStr
	case LanguageNemerle:
		return languageNemerleStr
	case LanguageNeon:
		return languageNeonStr
	case LanguageNesC:
		return languageNesCStr
	case LanguageNewLisp:
		return languageNewLispStr
	case LanguageNewspeak:
		return languageNewspeakStr
	case LanguageNginx:
		return languageNginxStr
	case LanguageNginxConfig:
		return languageNginxConfigStr
	case LanguageNimrod:
		return languageNimrodStr
	case LanguageNit:
		return languageNitStr
	case LanguageNix:
		return languageNixStr
	case LanguageNotmuch:
		return languageNotmuchStr
	case LanguageNSIS:
		return languageNSISStr
	case LanguageNu:
		return languageNuStr
	case LanguageNumPy:
		return languageNumPyStr
	case LanguageNuSMV:
		return languageNuSMVStr
	case LanguageNASMObjdump:
		return languageNASMObjdumpStr
	case LanguageObjdump:
		return languageObjdumpStr
	case LanguageObjectiveC:
		return languageObjectiveCStr
	case LanguageObjectiveCPP:
		return languageObjectiveCPPStr
	case LanguageObjectiveJ:
		return languageObjectiveJStr
	case LanguageOCaml:
		return languageOCamlStr
	case LanguageOctave:
		return languageOctaveStr
	case LanguageODIN:
		return languageODINStr
	case LanguageOoc:
		return languageOocStr
	case LanguageOpa:
		return languageOpaStr
	case LanguageOpenEdgeABL:
		return languageOpenEdgeABLStr
	case LanguageOpenSCAD:
		return languageOpenSCADStr
	case LanguageOrg:
		return languageOrgStr
	case LanguagePacmanConf:
		return languagePacmanConfStr
	case LanguagePan:
		return languagePanStr
	case LanguageParaSail:
		return languageParaSailStr
	case LanguageParrot:
		return languageParrotStr
	case LanguagePascal:
		return languagePascalStr
	case LanguagePawn:
		return languagePawnStr
	case LanguagePEG:
		return languagePEGStr
	case LanguagePerl:
		return languagePerlStr
	case LanguagePerl6:
		return languagePerl6Str
	case LanguagePHTML:
		return languagePHTMLStr
	case LanguagePHP:
		return languagePHPStr
	case LanguagePig:
		return languagePigStr
	case LanguagePike:
		return languagePikeStr
	case LanguagePkgConfig:
		return languagePkgConfigStr
	case LanguagePLpgSQL:
		return languagePLpgSQLStr
	case LanguagePointless:
		return languagePointlessStr
	case LanguagePony:
		return languagePonyStr
	case LanguagePostgres:
		return languagePostgresStr
	case LanguagePostgresConsole:
		return languagePostgresConsoleStr
	case LanguagePostScript:
		return languagePostScriptStr
	case LanguagePOVRay:
		return languagePOVRayStr
	case LanguagePowerShell:
		return languagePowerShellStr
	case LanguagePowerShellSession:
		return languagePowerShellSessionStr
	case LanguagePraat:
		return languagePraatStr
	case LanguageProlog:
		return languagePrologStr
	case LanguagePromQL:
		return languagePromQLStr
	case LanguagePropertiesJava:
		return languagePropertiesJavaStr
	case LanguageProtocolBuffer:
		return languageProtocolBufferStr
	case LanguagePsyShPHP:
		return languagePsyShPHPStr
	case LanguagePug:
		return languagePugStr
	case LanguagePuppet:
		return languagePuppetStr
	case LanguagePureData:
		return languagePureDataStr
	case LanguagePureScript:
		return languagePureScriptStr
	case LanguagePyPyLog:
		return languagePyPyLogStr
	case LanguagePython:
		return languagePythonStr
	case LanguagePython2:
		return languagePython2Str
	case LanguagePython2Traceback:
		return languagePython2TracebackStr
	case LanguagePythonConsole:
		return languagePythonConsoleStr
	case LanguagePythonTraceback:
		return languagePythonTracebackStr
	case LanguageQBasic:
		return languageQBasicStr
	case LanguageQML:
		return languageQMLStr
	case LanguageQVTO:
		return languageQVTOStr
	case LanguageR:
		return languageRStr
	case LanguageRacket:
		return languageRacketStr
	case LanguageRagel:
		return languageRagelStr
	case LanguageRagelEmbedded:
		return languageRagelEmbeddedStr
	case LanguageRaku:
		return languageRakuStr
	case LanguageRAML:
		return languageRAMLStr
	case LanguageRascal:
		return languageRascalStr
	case LanguageRawToken:
		return languageRawTokenStr
	case LanguageRConsole:
		return languageRConsoleStr
	case LanguageRd:
		return languageRdStr
	case LanguageRDoc:
		return languageRDocStr
	case LanguageReadlineConfig:
		return languageReadlineConfigStr
	case LanguageREALbasic:
		return languageREALbasicStr
	case LanguageReasonML:
		return languageReasonMLStr
	case LanguageREBOL:
		return languageREBOLStr
	case LanguageRecordJar:
		return languageRecordJarStr
	case LanguageRed:
		return languageRedStr
	case LanguageRedcode:
		return languageRedcodeStr
	case LanguageRegistry:
		return languageRegistryStr
	case LanguageRegularExpression:
		return languageRegularExpressionStr
	case LanguageRenderScript:
		return languageRenderScriptStr
	case LanguageRenPy:
		return languageRenPyStr
	case LanguageReScript:
		return languageReScriptStr
	case LanguageResourceBundle:
		return languageResourceBundleStr
	case LanguageReStructuredText:
		return languageReStructuredTextStr
	case LanguageRexx:
		return languageRexxStr
	case LanguageRHTML:
		return languageRHTMLStr
	case LanguageRichTextFormat:
		return languageRichTextFormatStr
	case LanguageRide:
		return languageRideStr
	case LanguageRing:
		return languageRingStr
	case LanguageRiot:
		return languageRiotStr
	case LanguageRMarkdown:
		return languageRMarkdownStr
	case LanguageRNGCompact:
		return languageRNGCompactStr
	case LanguageRoboconfGraph:
		return languageRoboconfGraphStr
	case LanguageRoboconfInstances:
		return languageRoboconfInstancesStr
	case LanguageRobotFramework:
		return languageRobotFrameworkStr
	case LanguageRoff:
		return languageRoffStr
	case LanguageRoffManpage:
		return languageRoffManpageStr
	case LanguageRouge:
		return languageRougeStr
	case LanguageRPC:
		return languageRPCStr
	case LanguageRPMSpec:
		return languageRPMSpecStr
	case LanguageRQL:
		return languageRQLStr
	case LanguageRSL:
		return languageRSLStr
	case LanguageRuby:
		return languageRubyStr
	case LanguageRubyIRBSession:
		return languageRubyIRBSessionStr
	case LanguageRUNOFF:
		return languageRUNOFFStr
	case LanguageRust:
		return languageRustStr
	case LanguageS:
		return languageSStr
	case LanguageSage:
		return languageSageStr
	case LanguageSalt:
		return languageSaltStr
	case LanguageSARL:
		return languageSARLStr
	case LanguageSAS:
		return languageSASStr
	case LanguageSass:
		return languageSassStr
	case LanguageScala:
		return languageScalaStr
	case LanguageScaml:
		return languageScamlStr
	case LanguageScdoc:
		return languageScdocStr
	case LanguageScheme:
		return languageSchemeStr
	case LanguageScilab:
		return languageScilabStr
	case LanguageScribe:
		return languageScribeStr
	case LanguageSCSS:
		return languageSCSSStr
	case LanguageSed:
		return languageSedStr
	case LanguageSelf:
		return languageSelfStr
	case LanguageSGML:
		return languageSGMLStr
	case LanguageShaderLab:
		return languageShaderLabStr
	case LanguageShell:
		return languageShellStr
	case LanguageShellSession:
		return languageShellSessionStr
	case LanguageShen:
		return languageShenStr
	case LanguageShExC:
		return languageShExCStr
	case LanguageSingularity:
		return languageSingularityStr
	case LanguageSieve:
		return languageSieveStr
	case LanguageSilver:
		return languageSilverStr
	case LanguageSimula:
		return languageSimulaStr
	case LanguageSketchDrawing:
		return languageSketchDrawingStr
	case LanguageSKILL:
		return languageSKILLStr
	case LanguageSlash:
		return languageSlashStr
	case LanguageSlice:
		return languageSliceStr
	case LanguageSlim:
		return languageSlimStr
	case LanguageSlurm:
		return languageSlurmStr
	case LanguageSmali:
		return languageSmaliStr
	case LanguageSmalltalk:
		return languageSmalltalkStr
	case LanguageSmartGameFormat:
		return languageSmartGameFormatStr
	case LanguageSmarty:
		return languageSmartyStr
	case LanguageSMIME:
		return languageSMIMEStr
	case LanguageSML:
		return languageSMLStr
	case LanguageSmPL:
		return languageSmPLStr
	case LanguageSMT:
		return languageSMTStr
	case LanguageSnobol:
		return languageSnobolStr
	case LanguageSnowball:
		return languageSnowballStr
	case LanguageSolidity:
		return languageSolidityStr
	case LanguageSourcePawn:
		return languageSourcePawnStr
	case LanguageSourcesList:
		return languageSourcesListStr
	case LanguageSPARQL:
		return languageSPARQLStr
	case LanguageSplineFontDatabase:
		return languageSplineFontDatabaseStr
	case LanguageSQF:
		return languageSQFStr
	case LanguageSQL:
		return languageSQLStr
	case LanguageSQLPL:
		return languageSQLPLStr
	case LanguageSqlite3con:
		return languageSqlite3conStr
	case LanguageSquidConf:
		return languageSquidConfStr
	case LanguageSquirrel:
		return languageSquirrelStr
	case LanguageSRecodeTemplate:
		return languageSRecodeTemplateStr
	case LanguageSSHConfig:
		return languageSSHConfigStr
	case LanguageSSP:
		return languageSSPStr
	case LanguageStan:
		return languageStanStr
	case LanguageStarlark:
		return languageStarlarkStr
	case LanguageStata:
		return languageStataStr
	case LanguageSTON:
		return languageSTONStr
	case LanguageStylus:
		return languageStylusStr
	case LanguageSublimeTextConfig:
		return languageSublimeTextConfigStr
	case LanguageSubRipText:
		return languageSubRipTextStr
	case LanguageSugarSS:
		return languageSugarSSStr
	case LanguageSuperCollider:
		return languageSuperColliderStr
	case LanguageSvelte:
		return languageSvelteStr
	case LanguageSVG:
		return languageSVGStr
	case LanguageSwift:
		return languageSwiftStr
	case LanguageSWIG:
		return languageSWIGStr
	case LanguageSYSTEMD:
		return languageSYSTEMDStr
	case LanguageSystemVerilog:
		return languageSystemVerilogStr
	case LanguageTableGen:
		return languageTableGenStr
	case LanguageTADS3:
		return languageTADS3Str
	case LanguageTAP:
		return languageTAPStr
	case LanguageTASM:
		return languageTASMStr
	case LanguageTcl:
		return languageTclStr
	case LanguageTcsh:
		return languageTcshStr
	case LanguageTcshSession:
		return languageTcshSessionStr
	case LanguageTea:
		return languageTeaStr
	case LanguageTeraTerm:
		return languageTeraTermStr
	case LanguageTermcap:
		return languageTermcapStr
	case LanguageTerminfo:
		return languageTerminfoStr
	case LanguageTerra:
		return languageTerraStr
	case LanguageTerraform:
		return languageTerraformStr
	case LanguageTeX:
		return languageTeXStr
	case LanguageTexinfo:
		return languageTexinfoStr
	case LanguageText:
		return languageTextStr
	case LanguageTextile:
		return languageTextileStr
	case LanguageThrift:
		return languageThriftStr
	case LanguageTiddler:
		return languageTiddlerStr
	case LanguageTIProgram:
		return languageTIProgramStr
	case LanguageTLA:
		return languageTLAStr
	case LanguageTNT:
		return languageTNTStr
	case LanguageTodotxt:
		return languageTodotxtStr
	case LanguageTOML:
		return languageTOMLStr
	case LanguageTradingView:
		return languageTradingViewStr
	case LanguageTrafficScript:
		return languageTrafficScriptStr
	case LanguageTransactSQL:
		return languageTSQLStr
	case LanguageTreetop:
		return languageTreetopStr
	case LanguageTSV:
		return languageTSVStr
	case LanguageTSX:
		return languageTSXStr
	case LanguageTuring:
		return languageTuringStr
	case LanguageTurtle:
		return languageTurtleStr
	case LanguageTwig:
		return languageTwigStr
	case LanguageTXL:
		return languageTXLStr
	case LanguageTypeLanguage:
		return languageTypeLanguageStr
	case LanguageTypeScript:
		return languageTypeScriptStr
	case LanguageTypoScript:
		return languageTypoScriptStr
	case LanguageUcode:
		return languageUcodeStr
	case LanguageUnicon:
		return languageUniconStr
	case LanguageUnifiedParallelC:
		return languageUnifiedParallelCStr
	case LanguageUnity3DAsset:
		return languageUnity3DAssetStr
	case LanguageUnixAssembly:
		return languageUnixAssemblyStr
	case LanguageUno:
		return languageUnoStr
	case LanguageUnrealScript:
		return languageUnrealScriptStr
	case LanguageUrbiScript:
		return languageUrbiScriptStr
	case LanguageUrWeb:
		return languageUrWebStr
	case LanguageUSD:
		return languageUSDStr
	case LanguageV:
		return languageVStr
	case LanguageVala:
		return languageValaStr
	case LanguageVB:
		return languageVBStr
	case LanguageVBA:
		return languageVBAStr
	case LanguageVBNet:
		return languageVBNetStr
	case LanguageVBScript:
		return languageVBScriptStr
	case LanguageVCL:
		return languageVCLStr
	case LanguageVCLSnippets:
		return languageVCLSnippetsStr
	case LanguageVCTreeStatus:
		return languageVCTreeStatusStr
	case LanguageVelocity:
		return languageVelocityStr
	case LanguageVGL:
		return languageVGLStr
	case LanguageVerilog:
		return languageVerilogStr
	case LanguageVHDL:
		return languageVHDLStr
	case LanguageVimHelpFile:
		return languageVimHelpFileStr
	case LanguageVimL:
		return languageVimLStr
	case LanguageVimScript:
		return languageVimScriptStr
	case LanguageVimSnippet:
		return languageVimSnippetStr
	case LanguageVolt:
		return languageVoltStr
	case LanguageVueJS:
		return languageVueJSStr
	case LanguageWavefrontMaterial:
		return languageWavefrontMaterialStr
	case LanguageWavefrontObject:
		return languageWavefrontObjectStr
	case LanguageWdl:
		return languageWdlStr
	case LanguageWDTE:
		return languageWDTEStr
	case LanguageWDiff:
		return languageWDiffStr
	case LanguageWebAssembly:
		return languageWebAssemblyStr
	case LanguageWebIDL:
		return languageWebIDLStr
	case LanguageWebOntologyLanguage:
		return languageWebOntologyLanguageStr
	case LanguageWebVTT:
		return languageWebVTTStr
	case LanguageWgetConfig:
		return languageWgetConfigStr
	case LanguageWhiley:
		return languageWhileyStr
	case LanguageWindowsRegistryEntries:
		return languageWindowsRegistryEntriesStr
	case LanguageWisp:
		return languageWispStr
	case LanguageWollok:
		return languageWollokStr
	case LanguageWowAddonData:
		return languageWowAddonDataStr
	case LanguageX10:
		return languageX10Str
	case LanguageXAML:
		return languageXAMLStr
	case LanguageXBase:
		return languageXBaseStr
	case LanguageXBitMap:
		return languageXBitMapStr
	case LanguageXC:
		return languageXCStr
	case LanguageXCompose:
		return languageXComposeStr
	case LanguageXFontDirectoryIndex:
		return languageXFontDirectoryIndexStr
	case LanguageXML:
		return languageXMLStr
	case LanguageXMLPropertyList:
		return languageXMLPropertyListStr
	case LanguageXojo:
		return languageXojoStr
	case LanguageXorg:
		return languageXorgStr
	case LanguageXPages:
		return languageXPagesStr
	case LanguageXPixMap:
		return languageXPixMapStr
	case LanguageXProc:
		return languageXProcStr
	case LanguageXQuery:
		return languageXQueryStr
	case LanguageXS:
		return languageXSStr
	case LanguageXSLT:
		return languageXSLTStr
	case LanguageXtend:
		return languageXtendStr
	case LanguageXtlang:
		return languageXtlangStr
	case LanguageYacc:
		return languageYaccStr
	case LanguageYAML:
		return languageYAMLStr
	case LanguageYANG:
		return languageYANGStr
	case LanguageYARA:
		return languageYARAStr
	case LanguageYASnippet:
		return languageYASnippetStr
	case LanguageZAP:
		return languageZAPStr
	case LanguageZeek:
		return languageZeekStr
	case LanguageZenScript:
		return languageZenScriptStr
	case LanguageZephir:
		return languageZephirStr
	case LanguageZig:
		return languageZigStr
	case LanguageZIL:
		return languageZILStr
	case LanguageZimpl:
		return languageZimplStr

	default:
		return languageUnkownStr
	}
}

// StringChroma returns the corresponding chroma lexer name.
// nolint:gocyclo
func (l Language) StringChroma() string {
	switch l {
	case LanguageAMPL:
		return languageAMPLChromaStr
	case LanguageApacheConfig:
		return languageApacheConfStr
	case LanguageAssembly:
		return languageAssemblyChromaStr
	case LanguageAutoHotkey:
		return languageAutoHotkeyChromaStr
	case LanguageCObjdump:
		return languageCObjdumpChromaStr
	case LanguageColdfusionCFC:
		return languageColdfusionCFCChromaStr
	case LanguageColdfusionHTML:
		return languageColdfusionHTMLChromaStr
	case LanguageCppObjdump:
		return languageCppObjdumpChromaStr
	case LanguageCUDA:
		return languageCUDAChromaStr
	case LanguageEmacsLisp:
		return languageEmacsLispChromaStr
	case LanguageFSharp:
		return languageFSharpChromaStr
	case LanguageFStar:
		return languageFStarLiteralStr
	case LanguageHxml:
		return languageHxmlChromaStr
	case LanguageJSONLD:
		return languageJSONLDChromaStr
	case LanguageJSX:
		return languageJSXChromaStr
	case LanguageLess:
		return languageLessChromaStr
	case LanguageLiquid:
		return languageLiquidChromaStr
	case LanguageMakefile:
		return languageMakefileChromaStr
	case LanguageMarkdown:
		return languageMarkdownChromaStr
	case LanguageNewLisp:
		return languageNewLispChromaStr
	case LanguageNimrod:
		return languageNimrodChromaStr
	case LanguageOoc:
		return languageOocChromaStr
	case LanguageOrg:
		return languageOrgChromaStr
	case LanguageReasonML:
		return languageReasonMLChromaStr
	case LanguageREBOL:
		return languageREBOLChromaStr
	case LanguageRexx:
		return languageRexxChromaStr
	case LanguageS:
		return languageRChromaStr
	case LanguageSYSTEMD:
		return languageSYSTEMDChromaStr
	case LanguageSystemVerilog:
		return languageSystemVerilogChromaStr
	case LanguageText:
		return languageTextChromaStr
	case LanguageTransactSQL:
		return languageTransactSQLChromaStr
	case LanguageVBNet:
		return languageVBNetChromaStr
	case LanguageVerilog:
		return languageVerilogChromaStr
	case LanguageVueJS:
		return languageVueJSChromaStr
	case LanguageWebIDL:
		return languageWebIDLChromaStr
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
