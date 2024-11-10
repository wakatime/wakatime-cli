package deps

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/file"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var cppExcludeRegex = regexp.MustCompile(`(?i)^(stdio\.h|iostream|stdlib\.h|string\.h|time\.h)$`)

// StateCPP is a token parsing state.
type StateCPP int

const (
	// StateCPPUnknown represents a unknown token parsing state.
	StateCPPUnknown StateCPP = iota
	// StateCPPImport means we are in import section during token parsing.
	StateCPPImport
)

// ParserCPP is a dependency parser for the C++ programming language.
// It is not thread safe.
type ParserCPP struct {
	State  StateCPP
	Output []string
}

// Parse parses dependencies from C++ file content using the C lexer.
func (p *ParserCPP) Parse(ctx context.Context, filepath string) ([]string, error) {
	logger := log.Extract(ctx)

	reader, err := file.OpenNoLock(filepath) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %s", filepath, err)
	}

	defer func() {
		if err := reader.Close(); err != nil {
			logger.Debugf("failed to close file: %s", err)
		}
	}()

	p.init()
	defer p.init()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %s", err)
	}

	l := lexers.Get(heartbeat.LanguageCPP.String())
	if l == nil {
		return nil, fmt.Errorf("failed to get lexer for %s", heartbeat.LanguageCPP.String())
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

func (p *ParserCPP) append(dep string) {
	// only consider first part of an import path
	dep = strings.Split(dep, "/")[0]

	if len(dep) == 0 {
		return
	}

	dep = strings.TrimSpace(dep)

	if cppExcludeRegex.MatchString(dep) {
		return
	}

	// trim extension
	dep = strings.TrimSuffix(dep, ".h")

	p.Output = append(p.Output, dep)
}

func (p *ParserCPP) init() {
	p.Output = nil
	p.State = StateCPPUnknown
}

func (p *ParserCPP) processToken(token chroma.Token) {
	switch token.Type {
	case chroma.CommentPreproc:
		p.processCommentPreproc(token.Value)
	case chroma.CommentPreprocFile:
		p.processCommentPreprocFile(token.Value)
	}
}

func (p *ParserCPP) processCommentPreproc(value string) {
	if strings.HasPrefix(strings.TrimSpace(value), "include") {
		p.State = StateCPPImport
	}
}

func (p *ParserCPP) processCommentPreprocFile(value string) {
	if p.State != StateCPPImport {
		return
	}

	if value != "\n" && value != "#" {
		value = strings.Trim(value, `"<> `)
		p.append(value)
	}

	p.State = StateCPPUnknown
}
