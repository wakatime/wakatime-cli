package deps

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var phpExcludeRegex = regexp.MustCompile(`(?i)(^app|app\.php)$`)

// StatePHP is a token parsing state.
type StatePHP int

const (
	// StatePHPUnknown represents a unknown php token parsing state.
	StatePHPUnknown StatePHP = iota
	// StatePHPUse represents php token parsing state use.
	StatePHPUse
	// StatePHPUseFunction represents php token parsing state use function.
	StatePHPUseFunction
	// StatePHPInclude represents php token parsing state include.
	StatePHPInclude
	// StatePHPAs represents php token parsing state as.
	StatePHPAs
)

// ParserPHP is a dependency parser for the php programming language.
// It is not thread safe.
type ParserPHP struct {
	State  StatePHP
	Output []string
}

// Parse parses dependencies from PHP file content using the chroma PHP lexer.
func (p *ParserPHP) Parse(filepath string) ([]string, error) {
	reader, err := os.Open(filepath) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %s", filepath, err)
	}

	defer func() {
		if err := reader.Close(); err != nil {
			log.Debugf("failed to close file: %s", err)
		}
	}()

	p.init()
	defer p.init()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %s", err)
	}

	l := lexers.Get(heartbeat.LanguagePHP.String())
	if l == nil {
		return nil, fmt.Errorf("failed to get lexer for %s", heartbeat.LanguagePHP.String())
	}

	iter, err := l.Tokenise(nil, string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize file content: %s", err)
	}

	for _, token := range iter.Tokens() {
		p.processToken(token)
	}

	return p.Output, nil
}

func (p *ParserPHP) append(dep string) {
	dep = strings.TrimSpace(dep)

	if len(dep) == 0 {
		return
	}

	if phpExcludeRegex.MatchString(dep) {
		return
	}

	p.Output = append(p.Output, dep)
}

func (p *ParserPHP) appendTruncate(dep string) {
	p.append(strings.Split(dep, `\`)[0])
}

func (p *ParserPHP) init() {
	p.State = StatePHPUnknown
	p.Output = []string{}
}

func (p *ParserPHP) processToken(token chroma.Token) {
	switch token.Type {
	case chroma.Keyword:
		p.processKeyword(token.Value)
	case chroma.NameFunction:
		p.processNameFunction(token.Value)
	case chroma.LiteralStringSingle:
		p.processLiteralStringSingle(token.Value)
	case chroma.LiteralStringDouble:
		p.processLiteralStringDouble(token.Value)
	case chroma.NameOther:
		p.processNameOther(token.Value)
	case chroma.Punctuation:
		p.processPunctuation(token.Value)
	case chroma.Text, chroma.Operator:
	default:
		p.State = StatePHPUnknown
	}
}

func (p *ParserPHP) processKeyword(value string) {
	switch {
	case value == "include" || value == "include_once" || value == "require" || value == "require_once":
		p.State = StatePHPInclude
	case value == "use":
		p.State = StatePHPUse
	case value == "as":
		p.State = StatePHPAs
	case p.State == StatePHPUse && value == "function":
		p.State = StatePHPUseFunction
	default:
		p.State = StatePHPUnknown
	}
}

func (p *ParserPHP) processNameFunction(value string) {
	if p.State == StatePHPUseFunction {
		p.appendTruncate(value)
		p.State = StatePHPUse
	}
}

func (p *ParserPHP) processLiteralStringSingle(value string) {
	if p.State == StatePHPInclude && value != `"` && value != `'` {
		p.append(strings.TrimSpace(value))
		p.State = StatePHPUnknown
	}
}

func (p *ParserPHP) processLiteralStringDouble(value string) {
	if p.State == StatePHPInclude && value != `"` && value != `'` {
		p.append(strings.TrimSpace("'" + value + "'"))
		p.State = StatePHPUnknown
	}
}

func (p *ParserPHP) processNameOther(value string) {
	if p.State == StatePHPUse {
		p.appendTruncate(value)
	}
}

func (p *ParserPHP) processPunctuation(value string) {
	switch {
	case value == "(" || value == ")":
	case (p.State == StatePHPUse || p.State == StatePHPAs) && value == ",":
		p.State = StatePHPUse
	default:
		p.State = StatePHPUnknown
	}
}
