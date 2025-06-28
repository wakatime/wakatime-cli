package file

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/wakatime/wakatime-cli/pkg/log"
)

// MaxFileSizeSupported is the max number of bytes we will ever read from a file. Files
// larger than this in bytes will use only the first MaxFileSizeSupported bytes when detecting
// language, dependencies, and their line count will be nil. Default is 5 MB.
const MaxFileSizeSupported = 5 * 1024 * 1024

// ReadHead returns the first max bytes of a file as a byte array.
func ReadHead(ctx context.Context, filepath string, max int) ([]byte, error) {
	logger := log.Extract(ctx)

	if max < 1 || max > MaxFileSizeSupported {
		max = MaxFileSizeSupported
	}

	f, err := OpenNoLock(filepath) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %s", filepath, err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			logger.Debugf("failed to close file: %s", err)
		}
	}()

	buf := make([]byte, max)

	c, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return buf[:c], nil
}

// ReadLines reads a file until max number of lines and return an array of lines.
func ReadLines(ctx context.Context, fp string, max int) ([]string, error) {
	if fp == "" {
		return nil, errors.New("filepath cannot be empty")
	}

	logger := log.Extract(ctx)

	file, err := OpenNoLock(fp) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %s", fp, err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			logger.Debugf("failed to close file '%s': %s", file.Name(), err)
		}
	}()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var (
		lines []string
		i     = 0
	)

	for scanner.Scan() {
		i++

		if i > max {
			break
		}

		lines = append(lines, scanner.Text())
	}

	return lines, nil
}

// CountLines counts the number of lines in a file.
func CountLines(ctx context.Context, fp string) (int, error) {
	if fp == "" {
		return 0, errors.New("filepath cannot be empty")
	}

	f, err := OpenNoLock(fp) // nolint:gosec
	if err != nil {
		return 0, fmt.Errorf("failed to open file %q: %s", fp, err)
	}

	logger := log.Extract(ctx)

	defer func() {
		if err := f.Close(); err != nil {
			logger.Debugf("failed to close file '%s': %s", f.Name(), err)
		}
	}()

	reader := io.LimitReader(f, MaxFileSizeSupported)

	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	var count int
	for scanner.Scan() {
		count++
	}

	if err := scanner.Err(); err != nil {
		return count, fmt.Errorf("failed to read file %q: %w", fp, err)
	}

	return count, nil
}
