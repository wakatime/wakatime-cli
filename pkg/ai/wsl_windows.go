package ai

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func discoverWSLHomes(ctx context.Context) []string {
	if strings.EqualFold(os.Getenv("WAKATIME_WSL"), "off") {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	// Enumerate running distributions only: probing a stopped UNC share boots it.
	root := os.Getenv("SystemRoot")
	if !filepath.IsAbs(root) {
		return nil
	}

	executable := filepath.Join(root, "System32", "wsl.exe")

	raw, err := exec.CommandContext(ctx, executable, "--list", "--quiet", "--running").Output() // nolint:gosec
	if err != nil {
		return nil
	}

	var homes []string

	for _, distro := range parseWSLDistros(raw) {
		if ctx.Err() != nil {
			break
		}

		root := `\\wsl$\` + distro

		entries, err := os.ReadDir(filepath.Join(root, "home"))
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					homes = append(homes, filepath.Join(root, "home", entry.Name()))
				}
			}
		}

		if info, err := os.Stat(filepath.Join(root, "root")); err == nil && info.IsDir() {
			homes = append(homes, filepath.Join(root, "root"))
		}
	}

	return homes
}
