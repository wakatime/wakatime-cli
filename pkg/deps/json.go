package deps

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers/j"
)

// nolint: gochecknoglobals
var filesJSON = map[string]struct {
	exact      bool
	dependency string
}{
	"bower.json":     {true, "bower"},
	"component.json": {true, "bower"},
	"package.json":   {true, "npm"},
}

// StateJSON is a token parsing state.
type StateJSON int

const (
	// StateJSONUnknown represents a unknown token parsing state.
	StateJSONUnknown StateJSON = iota
	// StateJSONDependencies means we are in dependencies section during token parsing.
	StateJSONDependencies
)

// ParserJSON is a dependency parser for JSON parser.
// It is not thread safe.
type ParserJSON struct {
	Level  int
	Output []string
	State  StateJSON
}

// Parse parses dependencies from JSON file content using the chroma JSON lexer.
func (p *ParserJSON) Parse(filepath string) ([]string, error) {
	reader, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %s", filepath, err)
	}

	defer reader.Close()

	p.init()
	defer p.init()

	// detect dependencies via filename
	p.processFilename(filepath)

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %s", err)
	}

	iter, err := j.JSON.Tokenise(nil, string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize file content: %s", err)
	}

	for _, token := range iter.Tokens() {
		p.processToken(token)
	}

	return p.Output, nil
}

func (p *ParserJSON) append(dep string) {
	p.Output = append(p.Output, strings.Trim(dep, `"' `))
}

func (p *ParserJSON) init() {
	p.Level = 0
	p.Output = nil
	p.State = StateJSONUnknown
}

func (p *ParserJSON) processToken(token chroma.Token) {
	switch token.Type {
	case chroma.NameTag:
		p.processNameTag(token.Value)
	case chroma.Punctuation:
		p.processPunctuation(token.Value)
	}
}

func (p *ParserJSON) processFilename(fp string) {
	filename := filepath.Base(fp)

	for k, f := range filesJSON {
		if f.exact && k == filename {
			p.Output = append(p.Output, f.dependency)
			continue
		}

		if !f.exact && strings.Contains(strings.ToLower(filename), k) {
			p.Output = append(p.Output, f.dependency)
		}
	}
}

func (p *ParserJSON) processNameTag(value string) {
	trimmed := strings.Trim(value, `"'`)

	if trimmed == "dependencies" || trimmed == "devDependencies" {
		p.State = StateJSONDependencies
		return
	}

	if p.State == StateJSONDependencies && p.Level == 2 {
		p.append(value)
	}
}

func (p *ParserJSON) processPunctuation(value string) {
	switch value {
	case "{":
		p.Level++
	case "}":
		p.Level--
		if p.State == StateJSONDependencies && p.Level <= 1 {
			p.State = StateJSONUnknown
		}
	}
}
