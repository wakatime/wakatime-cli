package legacy

import (
	"fmt"

	"github.com/alanhamlett/wakatime-cli/pkg/version"
)

func runVersion() {
	fmt.Println(version.Version)
}
