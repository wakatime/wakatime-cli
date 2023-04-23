package cmd

import (
	"fmt"

	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/version"

	"github.com/spf13/viper"
)
// runVersion displays the version of the wakatime-cli tool.

func runVersion(v *viper.Viper) (exitCode int, err error) {
	if v.GetBool("verbose") {
		fmt.Printf(
			"wakatime-cli\n  Version: %s\n  Commit: %s\n  Built: %s\n  OS/Arch: %s/%s\n",
			version.Version,
			version.Commit,
			version.BuildDate,
			version.OS,
			version.Arch,
		)

		exitCode = exitcode.Success
	} else {
		fmt.Println(version.Version)
                exitCode = exitcode.Success
	}

	return exitCode, nil
}
