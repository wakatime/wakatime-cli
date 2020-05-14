package legacy

import (
	"fmt"

	"github.com/wakatime/wakatime-cli/pkg/version"
)

func runVersion() {
	fmt.Println(version.Version)
}
