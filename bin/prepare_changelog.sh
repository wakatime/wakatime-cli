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
    for j in $(echo "${changelog}")
    do
        hash=${j:0:7}
        link="<https://github.com/wakatime/wakatime-cli/commit/${hash}|${hash}>"
        temp="${temp}$(echo "$j" | awk '{printf "<https://github.com/wakatime/wakatime-cli/commit/"$1"|"$1">";$1=""; print $0 }')\n"
    done

    slack=$(echo "*Changelog*\n${temp}")
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
        slack_output_for_develop
        clean_up
        replace_for_release
        ;;
    release)
        parse_for_release
        clean_up
        slack=$(echo "*Changelog*\n${changelog}")
        replace_for_release
        ;;
    *) exit 1 ;;
esac

echo "::set-output name=changelog::${changelog}"
echo "::set-output name=slack::${slack}"
