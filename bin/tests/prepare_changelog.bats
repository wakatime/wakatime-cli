#!/usr/bin/env bats

load 'libs/bats-assert/load'

@test "changelog for develop" {
    changelog=$(cat ./bin/tests/data/changelog_develop.txt)

    run ./bin/prepare_changelog.sh "develop" "${changelog}"
    assert_success
    assert_line -n 0 "::set-output name=changelog::'%0A'8bb1d12 Break single quote for replace string'%0A'0b138bb Ensure error response parsing for 4xx and 5xx heartbeat response errors"
    assert_line -n 1 "::set-output name=slack::*Changelog*\n<https://github.com/wakatime/wakatime-cli/commit/8bb1d12|8bb1d12> Break single quote for replace string\n<https://github.com/wakatime/wakatime-cli/commit/0b138bb|0b138bb> Ensure error response parsing for 4xx and 5xx heartbeat response errors\n"
}

@test "changelog for release" {
    changelog=$(cat ./bin/tests/data/changelog_release.txt)

    run ./bin/prepare_changelog.sh "release" "${changelog}"
    assert_success
    assert_line -n 0 "::set-output name=changelog::- Add sync offline'%0A'- Fix x509"
    assert_line -n 1 "::set-output name=slack::*Changelog*\n- Add sync offline\n- Fix x509\n"
}

@test "incorrect changelog for release should exit earlier" {
    changelog=$(cat ./bin/tests/data/changelog_release_incorrect.txt)

    run ./bin/prepare_changelog.sh "release" "${changelog}"
    assert_failure
}
