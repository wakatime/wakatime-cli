# wakatime-cli

Command line interface to [WakaTime](https://wakatime.com) used by all WakaTime [text editor plugins](https://wakatime.com/editors).

Go to <http://wakatime.com/editors> to install the plugin for your text editor or IDE.

### Contributing

* you first need to setup Go in your machine https://golang.org/doc/install#install
* check if you have make installed.
* `make test` - run tests
* `make lint` - run linter (must be done before any push)
* `make build-darwin -B` - compiles for macOS. For reference take a look at `Makefile` file
* `./build/darwin/amd64/wakatime-cli --help` - run help
* all branches should be named `feature/something` for example `feature/config-read` and we’re working direct into the main repo.
* Clone the repo following this path `…/github.com/wakatime/wakatime-cli` for example `~/github.com/wakatime/wakatime-cli`
