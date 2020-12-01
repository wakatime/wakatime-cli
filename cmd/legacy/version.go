package legacy

import (
	"fmt"

	"github.com/wakatime/wakatime-cli/pkg/version"
)

func runVersion(verbose bool) {
	if verbose {
		fmt.Printf(
			"wakatime-cli\n  Version: %s\n  Commit: %s\n  Built: %s\n  OS/Arch: %s/%s\n",
			version.Version,
			version.Commit,
			version.BuildDate,
			version.OS,
			version.Arch,
		)

		return
	}

	fmt.Println(version.Version)
}
