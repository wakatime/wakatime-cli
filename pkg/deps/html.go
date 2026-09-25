package deps

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/file"

	"golang.org/x/net/html"
)

var htmlDjangoPlaceholderRegex = regexp.MustCompile(`(?i)\{\{[^\}]+\}\}[/\\]?`)

// ParserHTML is a dependency parser for the HTML markup language.
type ParserHTML struct{}

// Parse parses script sources from HTML file content.
func (*ParserHTML) Parse(ctx context.Context, filepath string) ([]string, error) {
	head, err := file.ReadHead(ctx, filepath, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read: %s", err)
	}

	output := []string{}
	// Tokenize markup without lexing inline JavaScript or CSS. ReadHead can end
	// inside raw text, which must remain inexpensive even without a closing tag.
	z := html.NewTokenizer(bytes.NewReader(head))

	for {
		switch z.Next() {
		case html.TextToken, html.EndTagToken, html.CommentToken, html.DoctypeToken:
			continue
		case html.ErrorToken:
			if err := z.Err(); err != io.EOF {
				return nil, fmt.Errorf("failed to tokenize file content: %s", err)
			}

			return output, nil
		case html.StartTagToken, html.SelfClosingTagToken:
			name, more := z.TagName()
			if string(name) != "script" {
				continue
			}

			for more {
				var key, value []byte

				key, value, more = z.TagAttr()
				if string(key) == "src" {
					dep := htmlDjangoPlaceholderRegex.ReplaceAllString(string(value), "")
					output = append(output, strings.TrimSpace(dep))
				}
			}
		}
	}
}
