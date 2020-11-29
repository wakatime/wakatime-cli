package language

func priority(lang string) (float32, bool) {
	prios := map[string]float32{
		"FSharp": 0.01,
		"Perl":   0.01,
		// Higher priority than the TypoScriptLexer, as TypeScript is far more
		// common these days
		"TypeScript": 0.5,
	}

	p, ok := prios[lang]

	return p, ok
}
