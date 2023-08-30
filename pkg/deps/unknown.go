package deps

import (
	"path/filepath"
	"strings"
)

// nolint:gochecknoglobals
var filesUnknown = map[string]struct {
	exact      bool
	dependency string
}{
	"bower": {false, "bower"},
	"grunt": {false, "grunt"},
}

// ParserUnknown is a dependency parser for unknown parser.
// It is not thread safe.
type ParserUnknown struct {
	Output []string
}

// Parse parses dependencies from any file content via ReadCloser using the chroma golang lexer.
func (p *ParserUnknown) Parse(fp string) ([]string, error) {
	p.init()
	defer p.init()

	filename := filepath.Base(fp)

	for k, f := range filesUnknown {
		if f.exact && k == filename {
			p.Output = append(p.Output, f.dependency)
			continue
		}

		if !f.exact && strings.Contains(strings.ToLower(filename), k) {
			p.Output = append(p.Output, f.dependency)
		}
	}

	return p.Output, nil
}

func (p *ParserUnknown) init() {
	p.Output = nil
}
