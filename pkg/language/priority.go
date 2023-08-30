package language

func priority(lang string) (float32, bool) {
	prios := map[string]float32{
		"FSharp": 0.01,
		// Higher priority than the ca 65 assembler and ArmAsm
		"GAS": 0.1,
		// Higher priority than the ca Inform 6
		"INI": 0.1,
		// TASM uses the same file endings, but TASM is not as common as NASM, so we prioritize NASM higher by default.
		"NASM": 0.1,
		"Perl": 0.01,
		// Higher priority than Rebol
		"R": 0.1,
		// Higher priority than the TypoScript, as TypeScript is far more
		// common these days
		"TypeScript": 0.5,
	}

	p, ok := prios[lang]

	return p, ok
}
