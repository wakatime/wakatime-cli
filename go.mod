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
	github.com/sirupsen/logrus v1.8.1
	github.com/slongfield/pyfmt v0.0.0-20180124071345-020a7cb18bca
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.1.0
	github.com/spf13/viper v1.8.0
	github.com/stretchr/testify v1.7.0
	github.com/yookoala/realpath v1.0.0
	go.etcd.io/bbolt v1.3.5
	gopkg.in/ini.v1 v1.66.2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/juju/errors v0.0.0-20210818161939-5560c4c073ff // indirect
	github.com/juju/retry v0.0.0-20210818141810-5526f6f6ff07 // indirect
	github.com/juju/testing v0.0.0-20210302031854-2c7ee8570c07 // indirect
	github.com/juju/version v0.0.0-20210303051006-2015802527a8 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/pelletier/go-toml v1.9.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	golang.org/x/text v0.3.5 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/alecthomas/chroma => github.com/wakatime/chroma v0.8.2-wakatime.7

replace github.com/matishsiao/goInfo => github.com/wakatime/goInfo v0.1.0-wakatime.6
