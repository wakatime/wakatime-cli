package lexer

import (
	"fmt"

	"github.com/alecthomas/chroma/v2"
	l "github.com/alecthomas/chroma/v2/lexers"
)

// Lexer is an interface that can be implemented by lexers to register them.
type Lexer interface {
	Lexer() chroma.Lexer
	Name() string
}

// RegisterAll registers all custom lexers.
func RegisterAll() error {
	var lexers = []Lexer{
		ADL{},
		AMPL{},
		ActionScript3{},
		Agda{},
		Aheui{},
		Alloy{},
		AmbientTalk{},
		Arrow{},
		AspectJ{},
		AspxCSharp{},
		AspxVBNet{},
		Astro{},
		Asymptote{},
		Augeas{},
		BARE{},
		BBCBasic{},
		BBCode{},
		BC{},
		BST{},
		BUGS{},
		Befunge{},
		Blazor{},
		BlitzMax{},
		Boa{},
		Boo{},
		Boogie{},
		Brainfuck{},
		CADL{},
		CAmkES{},
		CBMBasicV2{},
		COBOLFree{},
		CObjdump{},
		CPSA{},
		CUDA{},
		Ca65Assembler{},
		CapDL{},
		Charmci{},
		Cirru{},
		Clay{},
		Clean{},
		ClojureScript{},
		ColdfusionCFC{},
		ColdfusionHTML{},
		ComponentPascal{},
		Coq{},
		CppObjdump{},
		Crmsh{},
		Croc{},
		Crontab{},
		Cryptol{},
		CsoundDocument{},
		CsoundOrchestra{},
		CsoundScore{},
		Cypher{},
		DASM16{},
		DG{},
		DObjdump{},
		DarcsPatch{},
		DebianControlFile{},
		Delphi{},
		Devicetree{},
		Duel{},
		DylanLID{},
		DylanSession{},
		EC{},
		ECL{},
		EMail{},
		ERB{},
		EarlGrey{},
		Easytrieve{},
		Eiffel{},
		ElixirIexSsession{},
		ErlangErlSession{},
		Evoque{},
		Execline{},
		Ezhil{},
		FSharp{},
		FStar{},
		Fancy{},
		Fantom{},
		Felix{},
		Flatline{},
		FloScript{},
		Forth{},
		FoxPro{},
		Freefem{},
		Gap{},
		Gas{},
		GettextCatalog{},
		Golo{},
		GoodDataCL{},
		Gosu{},
		GosuTemplate{},
		Groff{},
		HSAIL{},
		HTML{},
		HTTP{},
		Haml{},
		Hspec{},
		Hxml{},
		Hy{},
		Hybris{},
		IDL{},
		INI{},
		IRCLogs{},
		Icon{},
		IDA{},
		Inform6{},
		Inform6Template{},
		Inform7{},
		Ioke{},
		Isabelle{},
		JAGS{},
		JCL{},
		JSGF{},
		JSONLD{},
		JSP{},
		Jasmin{},
		JuliaConsole{},
		Juttle{},
		Kal{},
		Kconfig{},
		KernelLog{},
		Koka{},
		LLVMMIR{},
		LLVMMIRBODY{},
		LSL{},
		Lasso{},
		Lean{},
		Less{},
		Limbo{},
		Liquid{},
		LiterateAgda{},
		LiterateCryptol{},
		LiterateHaskell{},
		LiterateIdris{},
		LiveScript{},
		Logos{},
		Logtalk{},
		MAQL{},
		MIME{},
		MOOCode{},
		MQL{},
		MSDOSSession{},
		MXML{},
		Makefile{},
		Marko{},
		Mask{},
		Matlab{},
		MatlabSession{},
		MiniD{},
		MiniScript{},
		Modelica{},
		Modula2{},
		Mojo{},
		Monkey{},
		Monte{},
		MoonScript{},
		Mosel{},
		MozPreprocHash{},
		MozPreprocPercent{},
		Mscgen{},
		MuPAD{},
		Mustache{},
		NASM{},
		NASMObjdump{},
		NCL{},
		NSIS{},
		Nemerle{},
		NesC{},
		NewLisp{},
		Nit{},
		Notmuch{},
		Nushell{},
		NuSMV{},
		NumPy{},
		Objdump{},
		ObjectiveC{},
		ObjectiveCPP{},
		ObjectiveJ{},
		Ooc{},
		Opa{},
		OpenEdgeABL{},
		PEG{},
		POVRay{},
		Pan{},
		ParaSail{},
		Pawn{},
		Perl{},
		Perl6{},
		Pike{},
		Pointless{},
		PostgresConsole{},
		PowerShellSession{},
		Praat{},
		Processing{},
		Prolog{},
		PsyShPHP{},
		Pug{},
		PyPyLog{},
		Python{},
		Python2{},
		Python2Traceback{},
		PythonConsole{},
		PythonTraceback{},
		QBasic{},
		QVTO{},
		R{},
		RConsole{},
		REBOL{},
		RHTML{},
		RNGCompact{},
		RPMSpec{},
		RQL{},
		RSL{},
		RagelEmbedded{},
		RawToken{},
		Razor{},
		Rd{},
		ReScript{},
		Red{},
		Redcode{},
		ResourceBundle{},
		Ride{},
		RoboconfGraph{},
		RoboconfInstances{},
		RobotFramework{},
		RubyIRBSession{},
		SARL{},
		SSP{},
		SWIG{},
		Scaml{},
		Scdoc{},
		ShExC{},
		Shen{},
		Silver{},
		Singularity{},
		SketchDrawing{},
		Slash{},
		Slim{},
		Slint{},
		Slurm{},
		Smali{},
		SmartGameFormat{},
		Snowball{},
		SourcesList{},
		Sqlite3con{},
		Stan{},
		Stata{},
		SublimeTextConfig{},
		SuperCollider{},
		TADS3{},
		TAP{},
		TASM{},
		TNT{},
		TcshSession{},
		Tea{},
		TeraTerm{},
		Tiddler{},
		Todotxt{},
		TrafficScript{},
		TransactSQL{},
		Treetop{},
		Turtle{},
		USD{},
		Ucode{},
		Unicon{},
		UrbiScript{},
		VBNet{},
		VBScript{},
		VCL{},
		VCLSnippets{},
		VCTreeStatus{},
		VGL{},
		Velocity{},
		Verilog{},
		WDiff{},
		WebIDL{},
		X10{},
		XAML{},
		XML{},
		XQuery{},
		XSLT{},
		Xtend{},
		Xtlang{},
		Zeek{},
		Zephir{},
	}

	for _, lexer := range lexers {
		found := lexer.Lexer()
		if found == nil {
			return fmt.Errorf("%q lexer not found", lexer.Name())
		}

		_ = l.Register(lexer.Lexer())
	}

	return nil
}
