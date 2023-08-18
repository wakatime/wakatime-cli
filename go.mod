module github.com/wakatime/wakatime-cli

go 1.21

require (
	github.com/Azure/go-ntlmssp v0.0.0-20221128193559-754e69321358
	github.com/alecthomas/chroma/v2 v2.8.0
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964
	github.com/dlclark/regexp2 v1.10.0
	github.com/gandarez/go-olson-timezone v0.1.0
	github.com/gandarez/go-realpath v1.0.0
	github.com/juju/mutex v0.0.0-20180619145857-d21b13acf4bf
	github.com/kevinburke/ssh_config v1.2.1-0.20220605204831-a56e914e7283
	github.com/matishsiao/goInfo v0.0.0-20210923090445-da2e3fa8d45f
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/sftp v1.13.6
	github.com/sirupsen/logrus v1.9.3
	github.com/slongfield/pyfmt v0.0.0-20220222012616-ea85ff4c361f
	github.com/spf13/cobra v1.7.0
	github.com/spf13/jwalterweatherman v1.1.0
	github.com/spf13/viper v1.16.0
	github.com/stretchr/testify v1.8.4
	go.etcd.io/bbolt v1.3.7
	golang.org/x/crypto v0.12.0
	golang.org/x/net v0.14.0
	golang.org/x/text v0.12.0
	gopkg.in/ini.v1 v1.67.0
)

require (
	github.com/alecthomas/colour v0.1.0 // indirect
	github.com/alecthomas/repr v0.2.0 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
)

require (
	github.com/alecthomas/assert v1.0.0
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/juju/errors v1.0.0 // indirect
	github.com/juju/retry v0.0.0-20210818141810-5526f6f6ff07 // indirect
	github.com/juju/testing v0.0.0-20211215003918-77eb13d6cad2 // indirect
	github.com/juju/version v0.0.0-20210303051006-2015802527a8 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.0.9 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/yookoala/realpath v1.0.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/alecthomas/chroma/v2 => github.com/gandarez/chroma/v2 v2.8.0-wakatime.1

replace github.com/matishsiao/goInfo => github.com/wakatime/goInfo v0.1.0-wakatime.8
