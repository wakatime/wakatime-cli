package utils

import (
	"bufio"
	"log"
	"os"
)

// ReadFile Reads a file and return an array of lines
func ReadFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Error while opening file '%s' (%s)", path, err)
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
