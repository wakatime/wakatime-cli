# Configuring

Here's an example config file with all available options:

```ini
[settings]
debug = false
api_key = your-api-key
api_url = https://new-api-url.com
hide_file_names = false
hide_project_names = false
hide_branch_names =
exclude =
    ^COMMIT_EDITMSG$
    ^TAG_EDITMSG$
    ^/var/(?!www/).*
    ^/etc/
include =
    .*
include_only_with_project_file = false
exclude_unknown_project = false
status_bar_enabled = true
status_bar_coding_activity = true
offline = true
proxy = https://user:pass@localhost:8080
no_ssl_verify = false
ssl_certs_file =
timeout = 30
hostname = machinename
[projectmap]
projects/foo = new project name
^/home/user/projects/bar(\d+)/ = project{0}
[git]
submodules_disabled = false
```

## Settings Section

| option                         | description | allowed values |
| ---                            | ---         | ---            |
| debug                          | Turns on debug messages in log file. | true;false |
| api_key                        | Your wakatime api key. | _key_ |
| api_url                        | Heartbeats api url. For debugging with a local server. | _url_ |
| hide_file_names                | Obfuscate filenames. Will not send file names to api. | true;false;regex list |
| hide_project_names             | Obfuscate project names. When a project folder is detected instead of using the folder name as the project, a `.wakatime-project file` is created with a random project name. | true;false;regex list |
| hide_branch_names              | Obfuscate branch names. Will not send revision control branch names to api. | true;false;regex list |
| exclude                        | Filename patterns to exclude from logging. POSIX regex syntax. | true;false;regex list |
| include                        | Filename patterns to log. When used in combination with `exclude`, files matching `include` will still be logged. POSIX regex syntax | true;false;regex list |
| include_only_with_project_file | Disables tracking folders unless they contain a `.wakatime-project file`. | true;false |
| exclude_unknown_project        | When set, any activity where the project cannot be detected will be ignored. | true;false |
| status_bar_enabled             | Turns on wakatime status bar for certain editors. | true;false |
| status_bar_coding_activity     | Prints today's coding activity. | true;false |
| offline                        | Enables offline mode. All activity and logged time will be queued. | true;false |
| proxy                          | Optional proxy configuration. Supports HTTPS, SOCKS and NTLM proxies. | `https://user:pass@host:port` or `socks5://user:pass@host:port` or `domain\\user:pass` |
| no_ssl_verify                  | Disables SSL certificate verification for HTTPS requests. By default, SSL certificates are verified. | true;false |
| ssl_certs_file                 | Override the bundled CA certs file. By default, uses system ca certs. | _filepath_ |
| timeout                        | Number of seconds to wait when sending heartbeats to api. Defaults to 60 seconds. | _integer_ |
| hostname                       | Optional name of local machine. Defaults to local machine name read from system. | _machinename_ |

## Project Map Section

A key value pair list separated by new line.

```ini
[projectmap]
projects/foo = new project name
^/home/user/projects/bar(\d+)/ = project{0}
```

## Git Section

| option                         | description | allowed values |
| ---                            | ---         | ---            |
| submodules_disabled            | It will be matched against the submodule path and if matching, will skip it. | true;false;regex list |

For commonly used configuration options, see examples in the [FAQ](https://wakatime.com/faq).