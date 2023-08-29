package lexer

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var (
	easytrieveAnalyserCommetLineRe  = regexp.MustCompile(`^\s*\*`)
	easytrieveAnalyserMacroHeaderRe = regexp.MustCompile(`\s*MACRO`)
)

// Easytrieve lexer.
type Easytrieve struct{}

// Lexer returns the lexer.
// nolint: gocyclo
func (l Easytrieve) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"easytrieve"},
			Filenames: []string{"*.ezt", "*.mac"},
			MimeTypes: []string{"text/x-easytrieve"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// Perform a structural analysis for basic Easytrieve constructs.
		var (
			result           float32
			hasEndProc       bool
			hasHeaderComment bool
			hasFile          bool
			hasJob           bool
			hasProc          bool
			hasParm          bool
			hasReport        bool
		)

		lines := strings.Split(text, "\n")

		// Remove possible empty lines and header comments.
		for range lines {
			if len(lines) == 0 {
				break
			}

			if len(strings.TrimSpace(lines[0])) > 0 && !easytrieveAnalyserCommetLineRe.MatchString(lines[0]) {
				break
			}

			if easytrieveAnalyserCommetLineRe.MatchString(text) {
				hasHeaderComment = true
			}

			lines = lines[1:]
		}

		if len(lines) > 0 && easytrieveAnalyserMacroHeaderRe.MatchString(lines[0]) {
			// Looks like an Easytrieve macro.
			result += 0.4

			if hasHeaderComment {
				result += 0.4
			}

			return result
		}

		// Scan the source for lines starting with indicators.
		for _, line := range lines {
			splitted := strings.Fields(line)

			if len(splitted) < 2 {
				continue
			}

			if !hasReport && !hasJob && !hasFile && !hasParm && splitted[0] == "PARM" {
				hasParm = true
			}

			if !hasReport && !hasJob && !hasFile && splitted[0] == "FILE" {
				hasFile = true
			}

			if !hasReport && !hasJob && splitted[0] == "JOB" {
				hasJob = true
			}

			if !hasReport && splitted[0] == "PROC" {
				hasProc = true
				continue
			}

			if !hasReport && splitted[0] == "END-PROC" {
				hasEndProc = true
				continue
			}

			if !hasReport && splitted[0] == "REPORT" {
				hasReport = true
			}
		}

		// Weight the findings.
		if hasJob && hasProc == hasEndProc && hasHeaderComment {
			result += 0.1
		}

		if hasJob && hasProc == hasEndProc && hasParm && hasProc {
			// Found PARM, JOB and PROC/END-PROC:
			// pretty sure this is Easytrieve.
			result += 0.8

			return result
		}

		if hasJob && hasProc == hasEndProc && hasParm && !hasProc {
			// Found PARAM and JOB: probably this is Easytrieve.
			result += 0.5

			return result
		}

		if hasJob && hasProc == hasEndProc && !hasParm {
			// Found JOB and possibly other keywords: might be Easytrieve.
			result += 0.11
		}

		if hasJob && hasProc == hasEndProc && !hasParm && hasFile {
			result += 0.01
		}

		if hasJob && hasProc == hasEndProc && !hasParm && hasReport {
			result += 0.01
		}

		return result
	})

	return lexer
}

// Name returns the name of the lexer.
func (Easytrieve) Name() string {
	return heartbeat.LanguageEasytrieve.StringChroma()
}
