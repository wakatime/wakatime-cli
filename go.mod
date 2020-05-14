module github.com/wakatime/wakatime-cli

go 1.15

require (
	github.com/Azure/go-ntlmssp v0.0.0-20200615164410-66371956d46c
	github.com/PuerkitoBio/goquery v1.6.0 // indirect
	github.com/alecthomas/chroma v0.8.1
	github.com/armon/consul-api v0.0.0-20180202201655-eb2c6b5be1b6 // indirect
	github.com/certifi/gocertifi v0.0.0-20200922220541-2c3bb06c6054
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964
	github.com/dlclark/regexp2 v1.4.0
	github.com/matishsiao/goInfo v0.0.0-20200404012835-b5f882ee2288
	github.com/mattn/go-sqlite3 v1.14.4
	github.com/mitchellh/go-homedir v1.1.0
	github.com/slongfield/pyfmt v0.0.0-20180124071345-020a7cb18bca
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.1.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/ugorji/go v1.1.13 // indirect
	github.com/xordataexchange/crypt v0.0.3-0.20170626215501-b2862e3d0a77 // indirect
	github.com/yookoala/realpath v1.0.0
	golang.org/x/net v0.0.0-20201031054903-ff519b6c9102 // indirect
	golang.org/x/sys v0.0.0-20201101102859-da207088b7d1 // indirect
	golang.org/x/text v0.3.4 // indirect
	gopkg.in/ini.v1 v1.62.0
)

replace github.com/alecthomas/chroma => github.com/wakatime/chroma v0.8.1-wakatime.1
