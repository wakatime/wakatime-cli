#!/bin/bash

set -e

# ensure existence of release folder
if ! [ -d "./release" ]; then
    mkdir ./release
fi

# ensure zip is installed
if [ "$(which zip)" = "" ]; then
    apt-get update && apt-get install -y zip
fi

# add execution permission
chmod 750 ./build/wakatime-cli-freebsd-386
chmod 750 ./build/wakatime-cli-freebsd-amd64
chmod 750 ./build/wakatime-cli-freebsd-arm
chmod 750 ./build/wakatime-cli-linux-386
chmod 750 ./build/wakatime-cli-linux-amd64
chmod 750 ./build/wakatime-cli-linux-arm
chmod 750 ./build/wakatime-cli-linux-arm64
chmod 750 ./build/wakatime-cli-netbsd-386
chmod 750 ./build/wakatime-cli-netbsd-amd64
chmod 750 ./build/wakatime-cli-netbsd-arm
chmod 750 ./build/wakatime-cli-openbsd-386
chmod 750 ./build/wakatime-cli-openbsd-amd64
chmod 750 ./build/wakatime-cli-openbsd-arm
chmod 750 ./build/wakatime-cli-openbsd-arm64
chmod 750 ./build/wakatime-cli-windows-386.exe
chmod 750 ./build/wakatime-cli-windows-amd64.exe

# create archives
tar -czf ./release/wakatime-cli-freebsd-386.tar.gz -C ./build/ wakatime-cli-freebsd-386
tar -czf ./release/wakatime-cli-freebsd-amd64.tar.gz -C ./build/ wakatime-cli-freebsd-amd64
tar -czf ./release/wakatime-cli-freebsd-arm.tar.gz -C ./build/ wakatime-cli-freebsd-arm
tar -czf ./release/wakatime-cli-linux-386.tar.gz -C ./build/ wakatime-cli-linux-386
tar -czf ./release/wakatime-cli-linux-amd64.tar.gz -C ./build/ wakatime-cli-linux-amd64
tar -czf ./release/wakatime-cli-linux-arm.tar.gz -C ./build/ wakatime-cli-linux-arm
tar -czf ./release/wakatime-cli-linux-arm64.tar.gz -C ./build/ wakatime-cli-linux-arm64
tar -czf ./release/wakatime-cli-netbsd-386.tar.gz -C ./build/ wakatime-cli-netbsd-386
tar -czf ./release/wakatime-cli-netbsd-amd64.tar.gz -C ./build/ wakatime-cli-netbsd-amd64
tar -czf ./release/wakatime-cli-netbsd-arm.tar.gz -C ./build/ wakatime-cli-netbsd-arm
tar -czf ./release/wakatime-cli-openbsd-386.tar.gz -C ./build/ wakatime-cli-openbsd-386
tar -czf ./release/wakatime-cli-openbsd-amd64.tar.gz -C ./build/ wakatime-cli-openbsd-amd64
tar -czf ./release/wakatime-cli-openbsd-arm.tar.gz -C ./build/ wakatime-cli-openbsd-arm
tar -czf ./release/wakatime-cli-openbsd-arm64.tar.gz -C ./build/ wakatime-cli-openbsd-arm64
zip -j ./release/wakatime-cli-windows-386.zip ./build/wakatime-cli-windows-386.exe
zip -j ./release/wakatime-cli-windows-amd64.zip ./build/wakatime-cli-windows-amd64.exe

# handle apple binaries
unzip ./build/wakatime-cli-darwin.zip
chmod 750 ./build/wakatime-cli-darwin-amd64
chmod 750 ./build/wakatime-cli-darwin-arm64
zip -j ./release/wakatime-cli-darwin-amd64.zip ./build/wakatime-cli-darwin-amd64
zip -j ./release/wakatime-cli-darwin-arm64.zip ./build/wakatime-cli-darwin-arm64

# calculate checksums
for file in  ./release/*; do
	checksum=$(sha256sum ${file} | cut -d' ' -f1)
	filename=$(echo ${file} | rev | cut -d/ -f1 | rev)
	echo "${checksum} ${filename}" >> ./release/checksums_sha256.txt
done
