package deps

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var htmlDjangoPlaceholderRegex = regexp.MustCompile(`(?i)\{\{[^\}]+\}\}[/\\]?`)

// StateHTML is a token parsing state.
type StateHTML int

const (
	// StateHTMLUnknown represents an unknown token parsing state.
	StateHTMLUnknown StateHTML = iota
	// StateHTMLTag means we are inside an html tag during token parsing.
	StateHTMLTag
)

// ParserHTML is a dependency parser for the HTML markup language.
// It is not thread safe.
type ParserHTML struct {
	CurrentAttribute string
	CurrentTag       string
	State            StateHTML
	Output           []string
}

// Parse parses dependencies from HTML file content via ReadCloser using the chroma HTML lexer.
func (p *ParserHTML) Parse(filepath string) ([]string, error) {
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

	iter, err := lexers.HTML.Tokenise(nil, string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize file content: %s", err)
	}

	for _, token := range iter.Tokens() {
		p.processToken(token)
	}

	return p.Output, nil
}

func (p *ParserHTML) append(dep string) {
	// filter placeholders
	dep = htmlDjangoPlaceholderRegex.ReplaceAllString(dep, "")

	// trim whitespaces
	dep = strings.TrimSpace(dep)

	p.Output = append(p.Output, dep)
}

func (p *ParserHTML) init() {
	p.CurrentAttribute = ""
	p.CurrentTag = ""
	p.State = StateHTMLUnknown
	p.Output = []string{}
}

func (p *ParserHTML) processToken(token chroma.Token) {
	switch token.Type {
	case chroma.Punctuation:
		p.processPunctuation(token.Value)
	case chroma.NameTag:
		p.processNameTag(token.Value)
	case chroma.NameAttribute:
		p.processNameAttribute(token.Value)
	case chroma.LiteralString:
		p.processLiteralString(token.Value)
	}
}

func (p *ParserHTML) processPunctuation(value string) {
	switch value {
	case "<":
		p.State = StateHTMLTag
		p.CurrentAttribute = ""
		p.CurrentTag = ""
	case ">", "/":
		p.State = StateHTMLUnknown
		p.CurrentAttribute = ""
		p.CurrentTag = ""
	}
}

func (p *ParserHTML) processNameTag(value string) {
	if p.State == StateHTMLTag {
		p.CurrentTag = value
	}

	p.CurrentAttribute = ""
}

func (p *ParserHTML) processNameAttribute(value string) {
	if p.State == StateHTMLTag && p.CurrentTag != "" {
		p.CurrentAttribute = value
	}
}

func (p *ParserHTML) processLiteralString(value string) {
	if p.State == StateHTMLTag && p.CurrentTag == "script" && p.CurrentAttribute == "src" {
		p.append(value)
	}
}
