# Troubleshooting

First, read [How to debug WakaTime plugins][faq debug plugins].

Set `debug=true` in your `~/.wakatime.cfg` file to enable verbose logging.

The common wakatime-cli program logs to your user `$HOME` directory `~/.wakatime/wakatime.log`.

Each plugin also has its own log file:

* **Atom** writes errors to the developer console (View -> Developer -> Toggle Developer Tools)
* **Brackets** errors go to the developer console (Debug -> Show Developer Tools)
* **Cloud9** logs to the browser console (View -> Developer -> JavaScript Console)
* **Coda** logs to `/var/log/system.log` so use `sudo tail -f /var/log/system.log` in Terminal to watch Coda 2 logs
* **Eclipse** logs can be found in the Eclipse `Error Log` (Window -> Show View -> Error Log)
* **Emacs** messages go to the *messages* buffer window
* **Jetbrains IDEs (IntelliJ IDEA, PyCharm, RubyMine, PhpStorm, AppCode, AndroidStudio, WebStorm)** log to `idea.log` ([locating IDE log files][locating IDE log files])
* **Komodo** logs are written to `pystderr.log` (Help -> Troubleshooting -> View Log File)
* **Netbeans** logs to it's own log file (View -> IDE Log)
* **Notepad++** errors go to `AppData\Roaming\Notepad++\plugins\config\WakaTime.log` (this file is only created when an error occurs)
* **Sublime** Text logs to the Sublime Console (View -> Show Console)
* **TextMate** logs to stderr so run TextMate from Terminal to see any errors ([enable logging](https://github.com/textmate/textmate/wiki/Enable-Logging))
* **Vim** errors get displayed in the status line or inline (use `:redraw!` to clear inline errors)
* **Visual Studio** logs to the Output window, but uncaught exceptions go to ActivityLog.xml ([more info...](http://blogs.msdn.com/b/visualstudio/archive/2010/02/24/troubleshooting-with-the-activity-log.aspx))
* **VS Code** logs to the developer console (Help -> Toggle Developer Tools)
* **Xcode** type `sudo tail -f /var/log/system.log` in a Terminal to view Xcode errors

Check the [Plugin Status Page](https://wakatime.com/plugins/status) to see when the API last heard from each of your WakaTime plugins.

Useful API Endpoints for debugging:

* [List of your Plugins and when they were last heard from](https://wakatime.com/api/v1/users/current/user_agents)
* [List of your Machines and ther IPs](https://wakatime.com/api/v1/users/current/machine_names)

Useful Resources:

* [More Troubleshooting Info][faq debug plugins]
* [Official API Docs][api docs]

[faq debug plugins]: https://wakatime.com/faq#debug-plugins
[api docs]: https://wakatime.com/developers/
[locating IDE log files]: https://intellij-support.jetbrains.com/hc/en-us/articles/207241085-Locating-IDE-log-files

## SSH configuration

If you’re connected to a remote host using the [ssh extension](https://code.visualstudio.com/docs/remote/ssh) you might want to force WakaTime to run locally, for example when the server you connect to is shared among multiple people. Please follow [this guide](https://code.visualstudio.com/docs/remote/ssh#_advanced-forcing-an-extension-to-run-locally-remotely).
