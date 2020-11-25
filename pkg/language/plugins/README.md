# Plugin language support

Languages are spelled differently across the different wakatime plugins. To be able to parse language representations, which differ from the standard ones, mapping json files can be added to this folder. Mapping in `default.json` is used by default.

The filename should match the plugin name, like it is being provided via `--plugin` param of the wakatime cli. Lowercase spellings are accepted.

## Render mapping files

The json mappings have to be converted to go code via generate script. This is also performed by default before running the tests via `make test`.

```bash
make generate
```
