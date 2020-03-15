package project

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/wakatime/wakatime-cli/lib/utils"
)

// ProjectFile Information from a .wakatime-project file about the project for
// a given file. First line of .wakatime-project sets the project
// name. Second line sets the current branch name.
type ProjectFile struct {
	Entity string
}

// Process Process
func (p ProjectFile) Process() (*string, *string) {
	projectFile := utils.FindProjectFile(p.Entity)
	if projectFile != nil {
		file, err := os.Open(*projectFile)
		if err != nil {
			log.Printf("Error while opening file '%s' (%s)", projectFile, err)
		}

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		var lines []string

		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		file.Close()

		var name *string
		var branch *string

		if utils.Isset(lines, 0) {
			*name = strings.TrimSpace(lines[0])
		}
		if utils.Isset(lines, 1) {
			*branch = strings.TrimSpace(lines[1])
		}

		return name, branch
	}

	return nil, nil
}
