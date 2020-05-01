package arguments_test

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wakatime/wakatime-cli/cmd"
	"github.com/wakatime/wakatime-cli/lib/arguments"
	"github.com/wakatime/wakatime-cli/lib/configs"
)

func TestArgument_EntityMissingShouldPanic(t *testing.T) {
	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	assert.Panics(t, func() { cmd.Execute() })
}

func TestArgument_EntityMissingShouldBeEqualFileArgument(t *testing.T) {
	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--file", "/etc/file.txt",
	})

	assert.Equal(t, a.ObsoleteArgs.File, a.Entity.Entity)
}

func TestArgument_TimeMissingShouldBeTimestampNow(t *testing.T) {
	expected := math.MinInt64

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.NotEqual(t, expected, a.Time)
}

func TestArgument_TimeShouldBeTimestampNow(t *testing.T) {
	expected := time.Now().Unix()

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--time", strconv.FormatInt(expected, 10),
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Time)
}

func TestArgument_HostnameMissingShouldPass(t *testing.T) {
	expected := "hostname"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--key", "cc3939f6-000b-4756-ba33-966e29e66485",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Hostname)
}

func TestArgument_HostnameShouldPass(t *testing.T) {
	expected := "pc.local"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--key", "cc3939f6-000b-4756-ba33-966e29e66485",
		"--hostname", expected,
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Hostname)
}

func TestArgument_HostnameFailingShouldPass(t *testing.T) {
	expected := ""

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMockFail{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--key", "cc3939f6-000b-4756-ba33-966e29e66485",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Hostname)
}

func TestArgument_KeyShouldBeValidGuid(t *testing.T) {
	expected := "4b5fcb49-4f9e-408f-bf17-dae28f0eccd1"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--key", "4b5fcb49-4f9e-408f-bf17-dae28f0eccd1",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Key)
}

func TestArgument_KeyMissingShouldPass(t *testing.T) {
	expected := "167fc48a-fe6f-4893-8621-90dc1489fbe4"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Key)
}

func TestArgument_EntityMissingAndSyncOfflineActivityShouldPass(t *testing.T) {
	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--sync-offline-activity", "50",
	})

	cmd.Execute()

	assert.NoError(t, nil)
}

func TestArgument_EntityMissingAndTodayArgumentShouldPass(t *testing.T) {
	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--today",
	})

	cmd.Execute()

	assert.NoError(t, nil)
}

func TestArgument_SyncOfflineActivityLessThanZeroShouldPanic(t *testing.T) {
	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--sync-offline-activity", "-1",
	})

	assert.Panics(t, func() { cmd.Execute() })
}

func TestArgument_LanguageShouldPass(t *testing.T) {
	expected := "golang"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--language", expected,
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Language)
}

func TestArgument_LanguageMissingShouldBeEqualAlternateLanguageArgument(t *testing.T) {
	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"alternate-language", "python",
	})

	cmd.Execute()

	assert.Equal(t, a.AlternateLanguage, a.Language)
}

func TestArgument_IgnoredPatternShouldPass(t *testing.T) {
	expected := []string{"file1.txt", "*.log", ".data", "*.inf", "log*.log"}

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Exclude.Exclude)
}

func TestArgument_IgnoredPatternFailingShouldPass(t *testing.T) {
	expected := []string{}

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMockFail{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Exclude.Exclude)
}

func TestArgument_IncludeOnlyWithProjectFileShouldPass(t *testing.T) {
	expected := true

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--include-only-with-project-file",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Include.IncludeOnlyWithProjectFile)
}

func TestArgument_IncludeOnlyWithProjectFileMissingShouldPass(t *testing.T) {
	expected := true

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Include.IncludeOnlyWithProjectFile)
}

func TestArgument_IncludeOnlyWithProjectFileFailingShouldPass(t *testing.T) {
	expected := false

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMockFail{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Include.IncludeOnlyWithProjectFile)
}

func TestArgument_IncludedPatternShouldPass(t *testing.T) {
	expected := []string{".waka", "*.abc"}

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Include.Include)
}

func TestArgument_ExcludeUnknownProjectMisingShouldPass(t *testing.T) {
	expected := true

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Exclude.ExcludeUnknownProject)
}

func TestArgument_ExcludeUnknownProjectShouldPass(t *testing.T) {
	expected := true

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--exclude-unknown-project",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Exclude.ExcludeUnknownProject)
}

func TestArgument_ExcludeUnknownProjectFailingShouldPass(t *testing.T) {
	expected := false

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMockFail{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Exclude.ExcludeUnknownProject)
}

func TestArgument_HideFileNamesShouldAppendToArray(t *testing.T) {
	expected := ".*"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--hide-file-names",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Obfuscate.HiddenFileNames[0])
}

func TestArgument_HideFileNamesShouldPass(t *testing.T) {
	expected := "hide-file-names"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Obfuscate.HiddenFileNames[0])
}

func TestArgument_HideProjectNamesShouldAppendToArray(t *testing.T) {
	expected := ".*"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--hide-project-names",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Obfuscate.HiddenProjectNames[0])
}

func TestArgument_HideProjectNamesShouldPass(t *testing.T) {
	expected := "hide-project-names"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Obfuscate.HiddenProjectNames[0])
}

func TestArgument_HideBranchNamesShouldAppendToArray(t *testing.T) {
	expected := ".*"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--hide-branch-names",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Obfuscate.HiddenBranchNames[0])
}

func TestArgument_HideBranchNamesShouldPass(t *testing.T) {
	expected := "hide-branch-names"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Obfuscate.HiddenBranchNames[0])
}

func TestArgument_OfflineShouldPass(t *testing.T) {
	expected := true

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.DisableOffline)
}

func TestArgument_OfflineMissingShouldPass(t *testing.T) {
	expected := true

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--disable-offline",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.DisableOffline)
}

func TestArgument_OfflineFailingShouldPass(t *testing.T) {
	expected := false

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMockFail{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--disable-offline",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.DisableOffline)
}

func TestArgument_ProxyShouldPass(t *testing.T) {
	expected := "https://waka:time@domain.be:8080"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--proxy", expected,
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Proxy.Address)
}

func TestArgument_ProxyMissingShouldPass(t *testing.T) {
	expected := "https://waka:time@domain.be:8080"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Proxy.Address)
}

func TestArgument_ProxyFailingShouldPass(t *testing.T) {
	expected := ""

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMockFail{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Proxy.Address)
}

func TestArgument_ProxyInvalidShouldPanic(t *testing.T) {
	expected := "ht://user:pass@me.com"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--proxy", expected,
	})

	assert.Panics(t, func() { cmd.Execute() })
}

func TestArgument_NoSslVerifyShouldPass(t *testing.T) {
	expected := true

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Proxy.NoSslVerify)
}

func TestArgument_NoSslVerifyFailingShouldPass(t *testing.T) {
	expected := false

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMockFail{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Proxy.NoSslVerify)
}

func TestArgument_SslCertsFileShouldPass(t *testing.T) {
	expected := "ssl_certs_file"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Proxy.SslCertsFile)
}

func TestArgument_SslCertsFileFailingShouldPass(t *testing.T) {
	expected := ""

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMockFail{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Proxy.SslCertsFile)
}

func TestArgument_VerboseMissingShouldPass(t *testing.T) {
	expected := true

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Verbose)
}

func TestArgument_VerboseShouldPass(t *testing.T) {
	expected := true

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--verbose",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Verbose)
}

func TestArgument_LogFileShouldPass(t *testing.T) {
	expected := "~/folder/.wakatime.log"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--log-file", expected,
	})

	cmd.Execute()

	assert.Equal(t, expected, a.LogFile)
}

func TestArgument_LogFileMissingShouldUseLegacyLogFileArgument(t *testing.T) {
	expected := "~/folder/.wakatime.log"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--logfile", expected,
	})

	cmd.Execute()

	assert.Equal(t, expected, a.LogFile)
}

func TestArgument_LogFileMissingShouldPass(t *testing.T) {
	expected := "~/f/w/.wakatime.log"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.LogFile)
}

func TestArgument_ApiUrlShouldPass(t *testing.T) {
	expected := "https://api.wakatime.com/v1"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--api-url", expected,
	})

	cmd.Execute()

	assert.Equal(t, expected, a.APIURL)
}

func TestArgument_ApiUrlMissingShouldUseLegacyApiUrlArgument(t *testing.T) {
	expected := "https://api.wakatime.com/v1"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--apiurl", expected,
	})

	cmd.Execute()

	assert.Equal(t, expected, a.APIURL)
}

func TestArgument_ApiUrlMissingShouldPass(t *testing.T) {
	expected := "https://proxy.wakatime.com/api/v1"

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.APIURL)
}

func TestArgument_ApiUrlFailingShouldPass(t *testing.T) {
	expected := ""

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMockFail{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.APIURL)
}

func TestArgument_TimeoutShouldPass(t *testing.T) {
	expected := 100

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
		"--timeout", "100",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Timeout)
}

func TestArgument_TimeoutMissingShouldPass(t *testing.T) {
	expected := 30

	a := arguments.NewArguments()
	cfg := configs.ConfigFileMock{}
	cmd := cmd.NewRootCmd(a, cfg)

	cmd.SetArgs([]string{
		"--entity", "/etc/file.txt",
	})

	cmd.Execute()

	assert.Equal(t, expected, a.Timeout)
}
