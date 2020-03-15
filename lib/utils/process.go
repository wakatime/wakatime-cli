package utils

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

//Popen Patched Popen to prevent opening cmd window on Windows platform.
func Popen(args []string, kwargs []string) (string, error) {
	cmd := exec.Command(strings.Join(args, " "), kwargs...)

	if strings.ToLower(runtime.GOOS) == "windows" {
		//Find a way to set startupinfo as Python does
	}

	cmd.Env = os.Environ()
	//Need to set the LANG env var is it's not present on kwargs array

	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(stdout), nil
}
