module github.com/wakatime/wakatime-cli

go 1.16

require (
	github.com/Azure/go-ntlmssp v0.0.0-20200615164410-66371956d46c
	github.com/alecthomas/chroma v0.8.2
	github.com/certifi/gocertifi v0.0.0-20200922220541-2c3bb06c6054
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964
	github.com/dlclark/regexp2 v1.4.0
	github.com/matishsiao/goInfo v0.0.0-20200404012835-b5f882ee2288
	github.com/mitchellh/go-homedir v1.1.0
	github.com/sirupsen/logrus v1.8.1
	github.com/slongfield/pyfmt v0.0.0-20180124071345-020a7cb18bca
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/yookoala/realpath v1.0.0
	go.etcd.io/bbolt v1.3.5
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/text v0.3.4 // indirect
	gopkg.in/ini.v1 v1.62.0
)

replace github.com/alecthomas/chroma => github.com/wakatime/chroma v0.8.2-wakatime.7

replace github.com/spf13/viper => github.com/wakatime/viper v1.7.1-0.20210127133619-3adac62bd70f

replace github.com/matishsiao/goInfo => github.com/wakatime/goInfo v0.1.0-wakatime.2
