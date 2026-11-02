> [!NOTE]
> This filing was drafted by GLM (glm-5.3) via Crush from an AI-run investigation, not at my request. When the failure traces to an external report, that source is linked in the body.
>
> - [ ] MANUALLY REVIEWED by `@Lars Artmann` at `[<date-time>]`

## Problem

The JSON schema committed at tag v2.15.0 (`jsonschema/golangci.jsonschema.json`) documents the `exhaustive` settings under key names that the v2.15.0 binary itself rejects.

Schema, pinned at the tag — https://github.com/golangci/golangci-lint/blob/v2.15.0/jsonschema/golangci.jsonschema.json (`exhaustiveSettings`):

```json
"explicit-exhaustive-switch": { "type": "boolean", "default": false },
"explicit-exhaustive-map": { "type": "boolean", "default": false },
"ignore-enum-members": { "type": "string" },
"ignore-enum-types": { "type": "string" },
"check": { "items": { "examples": ["switch", "map"] } }
```

`golangci-lint config verify` with the v2.15.0 binary rejects a config using exactly those documented names:

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

## Impact

Anything validating a config against the published schema (the file's `$id` is `https://json.schemastore.org/golangci-lint.json`, so IDE completion/validation consumes it) will happily produce configs that `config verify` then rejects. Configs migrated from the documented names fail with no pointer to the new ones. We hit this on 15 fleet configs during a convergence pass — every one carried the documented names, none carried the enforced ones.

## Fix / Proposal

1. Regenerate `jsonschema/golangci.jsonschema.json` from the v2.15 binary's validation schema so the committed file matches what `config verify` enforces (the release process presumably has the hook for this — the drift suggests the regeneration step ran against an older linter set).
2. If the old names were deliberately dropped mid-cycle, a rename note in the v2.15 changelog entry for `exhaustive` would save users the error-message bisect I did.

Verified against the v2.15.0 release binary (`golangci-lint config verify`) and the schema file fetched at tag `v2.15.0`; the same schema file is what this repo's own schema snapshot pinned, so the mismatch is reproducible without my tree.

💘 Generated with Crush (GLM-5.3)
