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

var cExcludeRegex = regexp.MustCompile(`(?i)^(stdio\.h|stdlib\.h|string\.h|time\.h)$`)

// StateC is a token parsing state.
type StateC int

const (
	// StateCUnknown represents a unknown token parsing state.
	StateCUnknown StateC = iota
	// StateCImport means we are in import section during token parsing.
	StateCImport
)

// ParserC is a dependency parser for the c programming language.
// It is not thread safe.
type ParserC struct {
	State  StateC
	Output []string
}

// Parse parses dependencies from C file content using the C lexer.
func (p *ParserC) Parse(filepath string) ([]string, error) {
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

	l := lexers.Get(heartbeat.LanguageC.String())
	if l == nil {
		return nil, fmt.Errorf("failed to get lexer for %s", heartbeat.LanguageC.String())
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

func (p *ParserC) append(dep string) {
	// only consider first part of an import path
	dep = strings.Split(dep, "/")[0]

	if len(dep) == 0 {
		return
	}

	dep = strings.TrimSpace(dep)

	if cExcludeRegex.MatchString(dep) {
		return
	}

	// trim extension
	dep = strings.TrimSuffix(dep, ".h")

	p.Output = append(p.Output, dep)
}

func (p *ParserC) init() {
	p.Output = nil
	p.State = StateCUnknown
}

func (p *ParserC) processToken(token chroma.Token) {
	switch token.Type {
	case chroma.CommentPreproc:
		p.processCommentPreproc(token.Value)
	case chroma.CommentPreprocFile:
		p.processCommentPreprocFile(token.Value)
	}
}

func (p *ParserC) processCommentPreproc(value string) {
	if strings.HasPrefix(strings.TrimSpace(value), "include") {
		p.State = StateCImport
	}
}

func (p *ParserC) processCommentPreprocFile(value string) {
	if p.State != StateCImport {
		return
	}

	if value != "\n" && value != "#" {
		value = strings.Trim(value, `"<> `)
		p.append(value)
	}

	p.State = StateCUnknown
}
