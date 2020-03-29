package utils

import (
	"bufio"
	"log"
	"os"
)

// ReadFile Reads a file and return an array of lines
func ReadFile(p string) ([]string, error) {
	file, err := os.Open(p)
	if err != nil {
		log.Printf("Error while opening file '%s' (%s)", p, err)
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	file.Close()

	return lines, nil
}
