source = ["./build/wakatime-cli-darwin-amd64", "./build/wakatime-cli-darwin-arm64"]
bundle_id = "com.wakatime.wakatime-cli"

apple_id {
  username = "alan@wakatime.com"
  password = "@env:AC_PASSWORD"
}

sign {
  application_identity = "Developer ID Application: WAKATIME, LLC"
}

zip {
  output_path = "wakatime-cli-darwin.zip"
}
