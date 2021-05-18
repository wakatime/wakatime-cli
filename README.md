# WakaTime CLI

![Tests master](https://img.shields.io/github/workflow/status/wakatime/wakatime-cli/Create%20Release/master?label=%20tests) ![Build master](https://img.shields.io/github/workflow/status/wakatime/wakatime-cli/Build%20and%20upload%20release%20assets) [![Coverage Status](https://coveralls.io/repos/github/wakatime/wakatime-cli/badge.svg?branch=master)](https://coveralls.io/github/wakatime/wakatime-cli?branch=master)

Command line interface to [WakaTime](https://wakatime.com) used by all WakaTime [text editor plugins](https://wakatime.com/editors).

Go to <http://wakatime.com/editors> to install the plugin for your text editor or IDE.

## Usage

If you are building a plugin using the [WakaTime API](https://wakatime.com/developers/) then follow the [Creating a Plugin](https://wakatime.com/help/misc/creating-plugin) guide.

Some more usage information is available in the [FAQ](https://wakatime.com/faq).

## Configuring

Options can be passed via command line, or set in the `$WAKATIME_HOME/.wakatime.cfg` config file. Command line arguments take precedence over config file settings. The `$WAKATIME_HOME/.wakatime.cfg` file is in [INI](http://en.wikipedia.org/wiki/INI_file) format. See [Configuring](CONFIGURING.md) for more details.

## Contributing

Pull requests, issues and comments are welcome! See [Contributing](CONTRIBUTING.md) for more details.

Many thanks to all [contributors](AUTHORS)!

Made with :heart: by WakaTime Team.
