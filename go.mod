module github.com/wakatime/wakatime-cli

go 1.17

require (
	github.com/Azure/go-ntlmssp v0.0.0-20200615164410-66371956d46c
	github.com/alecthomas/chroma v0.8.2
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964
	github.com/dlclark/regexp2 v1.4.0
	github.com/gandarez/go-olson-timezone v0.1.0
	github.com/juju/mutex v0.0.0-20180619145857-d21b13acf4bf
	github.com/matishsiao/goInfo v0.0.0-20200404012835-b5f882ee2288
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/sftp v1.13.4
	github.com/sirupsen/logrus v1.8.1
	github.com/slongfield/pyfmt v0.0.0-20180124071345-020a7cb18bca
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.1.0
	github.com/spf13/viper v1.8.0
	github.com/stretchr/testify v1.7.0
	github.com/yookoala/realpath v1.0.0
	go.etcd.io/bbolt v1.3.5
	golang.org/x/crypto v0.0.0-20211215153901-e495a2d5b3d3
	gopkg.in/ini.v1 v1.66.2
)

require (
	github.com/juju/errors v0.0.0-20210818161939-5560c4c073ff // indirect
	github.com/juju/retry v0.0.0-20210818141810-5526f6f6ff07 // indirect
	github.com/juju/version v0.0.0-20210303051006-2015802527a8 // indirect
)

replace github.com/alecthomas/chroma => github.com/wakatime/chroma v0.8.2-wakatime.7

replace github.com/matishsiao/goInfo => github.com/wakatime/goInfo v0.1.0-wakatime.6
