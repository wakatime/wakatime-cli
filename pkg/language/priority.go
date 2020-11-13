package language

func priority(lang string) (float32, bool) {
	prios := map[string]float32{
		"FSharp":     0.01,
		"Perl":       0.01,
		"TypeScript": 0.01,
	}

	p, ok := prios[lang]

	return p, ok
}
