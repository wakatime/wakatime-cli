package system

import (
	"fmt"
	"strings"

	"github.com/matishsiao/goInfo"
)

//GetOSInfo Retrieve OS' info
func GetOSInfo() string {
	info := goInfo.GetInfo()
	return fmt.Sprintf("%s-%s-%s", strings.Title(info.GoOS), info.Core, info.Platform)
}
