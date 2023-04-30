# WakaTime CLI

[![Tests](https://img.shields.io/github/actions/workflow/status/wakatime/wakatime-cli/on_push.yml?branch=develop&label=tests)](https://github.com/wakatime/wakatime-cli/actions)
[![Coverage](https://img.shields.io/codecov/c/gh/wakatime/wakatime-cli/develop)](https://codecov.io/gh/wakatime/wakatime-cli)
[![wakatime](https://wakatime.com/badge/github/wakatime/wakatime-cli.svg)](https://wakatime.com)



The WakaTime CLI is a powerful tool that allows developers to track their coding activity across multiple text editors and IDEs. It is used by all WakaTime text editor plugins, and can be installed on a variety of operating systems.

To use the WakaTime CLI, you'll need to first create an account with WakaTime at http://wakatime.com. Once you have an account, you can follow the instructions at [editors](http://wakatime.com/editors) to install the WakaTime plugin for your text editor or IDE.

The WakaTime CLI provides a wide range of options for tracking your coding activity, including the ability to track activity for specific files or projects, generate reports, and view detailed analytics on your coding habits. You can also customize the CLI to suit your specific needs, such as setting up custom keybindings or integrating with other tools in your development workflow.

One important feature of the WakaTime CLI is its ability to integrate with popular continuous integration and deployment tools, such as Jenkins, Travis CI, and CircleCI. This allows you to track your coding activity and monitor your productivity as part of your automated build and release processes.

To ensure that your code is performing optimally, it's important to test it thoroughly. WakaTime provides comprehensive testing coverage for its CLI, ensuring that you can rely on accurate data when tracking your coding activity. You can find more information about the testing coverage for WakaTime at https://wakatime.com/developers#tests.

Overall, the WakaTime CLI is a powerful and flexible tool that can help developers better understand their coding habits and improve their productivity. By tracking your coding activity with WakaTime, you can gain valuable insights into your development process and make more informed decisions about how to optimize your workflow.



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
