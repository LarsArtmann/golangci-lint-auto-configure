<!-- Form-ready body for https://github.com/golangci/golangci-lint/issues/new?template=bug_report.yml
     (the API path is closed: the no-blank-issue bot auto-closes label-less issues and
     external accounts cannot add the form's `bug` label — see the accompanying draft
     2026-11-02_golangci-lint-v2.15-schema-drift.md for the voice-checked freeform version
     that golangci-lint#6901 shipped as before the bot closed it).
     Paste each section into the matching form field. -->

### Welcome

- [x] Yes, I'm using a binary release within 2 latest [releases](https://github.com/golangci/golangci-lint/releases). Only such installations are supported.
- [x] Yes, I've searched similar [issues on GitHub](https://github.com/golangci/golangci-lint/issues) and didn't find any.
- [x] Yes, I've read the `typecheck` section of the [FAQ](https://golangci-lint.run/docs/welcome/faq/#why-do-you-have-typecheck-errors).
- [x] Yes, I've tried with the standalone [linter](https://golangci-lint.run/docs/linters/) if available (e.g., gocritic, go vet, etc.).
- [x] I agree to follow this project's [Code of Conduct](https://github.com/golangci/golangci-lint?tab=coc-ov-file)

### How did you install golangci-lint?

Nix

### Description of the problem

The JSON schema committed at tag v2.15.0 (`jsonschema/golangci.jsonschema.json`) documents the `exhaustive` settings under key names that the v2.15.0 binary itself rejects via `golangci-lint config verify`.

Schema, pinned at the tag — https://github.com/golangci/golangci-lint/blob/v2.15.0/jsonschema/golangci.jsonschema.json (`exhaustiveSettings`):

```json
"explicit-exhaustive-switch": { "type": "boolean", "default": false },
"explicit-exhaustive-map": { "type": "boolean", "default": false },
"ignore-enum-members": { "type": "string" },
"ignore-enum-types": { "type": "string" },
"check": { "items": { "examples": ["switch", "map"] } }
```

`config verify` with the v2.15.0 binary rejects a config using exactly those documented names:

```text
jsonschema: "linters.settings.exhaustive" does not validate with ...: additional properties 'explicit-exhaustive-map', 'explicit-exhaustive-switch', 'ignore-enum-members', 'ignore-enum-types' not allowed
The command is terminated due to an error: the configuration contains invalid elements
```

The names the same binary accepts (found by bisecting the error messages):

```yaml
linters:
  settings:
    exhaustive:
      check: [switch, mapliteral]
      explicit-exhaustive: false
      ignore-constants: [Example.+]
      ignore-types: [Example.+]
```

So `check`'s documented example value `map` is not in the binary's enum (`switch`, `mapliteral`, `typeswitch`), the two `explicit-exhaustive-{switch,map}` booleans are replaced by one `explicit-exhaustive`, and `ignore-enum-{members,types}` (strings) are replaced by `ignore-{constants,types}` (arrays).

Impact: anything validating a config against the published schema (the file's `$id` is `https://json.schemastore.org/golangci-lint.json`, so IDE completion/validation consumes it) will happily produce configs that `config verify` then rejects. We hit this on 15 fleet configs during a convergence pass — every one carried the documented names, none carried the enforced ones.

Proposal: regenerate `jsonschema/golangci.jsonschema.json` from the v2.15 binary's validation schema so the committed file matches what `config verify` enforces; if the old names were deliberately dropped mid-cycle, a rename note in the v2.15 changelog entry for `exhaustive` would save users the error-message bisect.

> [!NOTE]
> This report was drafted by GLM (glm-5.3) via Crush from an AI-run investigation, not at my request. The reproduction below was verified against the v2.15.0 release binary before filing.

### Version of golangci-lint

golangci-lint has version 2.15.0 built with go1.27.3 from v2.15.0 on 1970-01-01T00:00:00Z

### Configuration

<details>

```yaml
linters:
  settings:
    exhaustive:
      check:
        - switch
        - map
      explicit-exhaustive-switch: false
      explicit-exhaustive-map: false
      ignore-enum-members: Example.+
      ignore-enum-types: Example.+
```

</details>

### Go environment

<details>

```console
# paste output here
```

</details>

### Verbose output of running

<details>

```console
$ golangci-lint config verify
jsonschema: "linters.settings.exhaustive" does not validate with "/properties/linters/properties/settings/properties/exhaustive/$ref/additionalProperties": additional properties 'ignore-enum-types', 'explicit-exhaustive-map', 'explicit-exhaustive-switch', 'ignore-enum-members' not allowed
The command is terminated due to an error: the configuration contains invalid elements
```

</details>

### A minimal reproducible example or link to a public repository

<details>

```console
# config.yaml uses the setting names documented in the v2.15.0 jsonschema:
$ golangci-lint config verify --config config.yaml
# -> rejected (see verbose output above)

# renaming to the enforced names makes the same binary accept the config:
#   check: [switch, mapliteral]          (map -> mapliteral)
#   explicit-exhaustive: false           (replaces explicit-exhaustive-{switch,map})
#   ignore-constants: [Example.+]        (replaces ignore-enum-members, string -> array)
#   ignore-types: [Example.+]            (replaces ignore-enum-types, string -> array)
$ golangci-lint config verify --config fixed.yaml
# -> passes
```

</details>

### Validation

- [x] Yes, I've included all the information above (example, version, config, etc.).
