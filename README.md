# WakaTime CLI

[![Tests](https://img.shields.io/github/actions/workflow/status/wakatime/wakatime-cli/on_push.yml?branch=develop&label=tests)](https://github.com/wakatime/wakatime-cli/actions)
[![Coverage](https://img.shields.io/codecov/c/gh/wakatime/wakatime-cli/develop)](https://codecov.io/gh/wakatime/wakatime-cli)
[![wakatime](https://wakatime.com/badge/github/wakatime/wakatime-cli.svg)](https://wakatime.com)

Command line interface to [WakaTime][wakatime] used by all WakaTime [text editor plugins][editors].

Go to [http://wakatime.com/editors][editors] to install the plugin for your text editor or IDE.

## Usage

Normally you don't need to use wakatime-cli directly unless you're building a new WakaTime plugin.
If you're building a plugin using the [WakaTime API][api], follow the [Creating a Plugin][creating-plugin] guide.

WakaTime plugins and wakatime-cli share a common [INI][ini] config file:

`$WAKATIME_HOME/.wakatime.cfg`

See [Usage][usage] or the [WakaTime FAQ][faq] for more details.

## Contributing

Pull requests and issues are welcome!
See [Contributing][contributing] for more details.

## Troubleshooting

See [Troubleshooting][troubleshooting] for more details.

Many thanks to all [contributors][authors]!

Made with :heart: by the WakaTime Team.

[wakatime]: http://wakatime.com
[editors]: http://wakatime.com/editors
[api]: https://wakatime.com/developers/
[creating-plugin]: https://wakatime.com/help/misc/creating-plugin
[ini]: http://en.wikipedia.org/wiki/INI_file
[faq]: https://wakatime.com/faq
[usage]: USAGE.md
[contributing]: CONTRIBUTING.md
[troubleshooting]: TROUBLESHOOTING.md
[authors]: AUTHORS
