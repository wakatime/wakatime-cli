#!/bin/bash

set -e

if [[ $# -ne 2 ]]; then
    echo 'incorrect number of arguments'
    exit 1
fi

# Read arguments
branch=$1
changelog=$2
slack=

clean_up() {
    changelog="${changelog//\`/}"
    changelog="${changelog//\'/}"
    changelog="${changelog//\"/}"
}

replace_for_release() {
    changelog="${changelog//'%'/'%25'}"
    changelog="${changelog//$'\n'/'%0A'}"
    changelog="${changelog//$'\r'/'%0D'}"
}

slack_output_for_develop() {
    local IFS=$'\n' # make newlines the only separator
    local temp=
    for j in ${changelog}
    do
        temp="${temp}$(echo "$j" | awk '{printf "<https://github.com/wakatime/wakatime-cli/commit/"$1"|"$1">";$1=""; print $0 }')\n"
    done

    slack="*Changelog*\n${temp}"
}

slack_output_for_release() {
    local IFS=$'\n' # make newlines the only separator
    local temp=
    for j in ${changelog}
    do
        temp="${temp}${j}\n"
    done

    slack="*Changelog*\n${temp}"
}

parse_for_develop() {
    changelog=$(awk 'f;/## Changelog/{f=1}' <<< "$changelog")
}

parse_for_release() {
    changelog=$(awk 'f;/Changelog:/{f=1}' <<< "$changelog")
}

case $branch in
    develop) 
        parse_for_develop
        clean_up
        slack_output_for_develop
        replace_for_release
        ;;
    release)
        parse_for_release
        [ -z "$changelog" ] && exit 1
        clean_up
        slack_output_for_release
        replace_for_release
        ;;
    *) exit 1 ;;
esac

echo "::set-output name=changelog::${changelog}"
echo "::set-output name=slack::${slack}"
