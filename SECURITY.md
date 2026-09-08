# Security Policy

## Supported Versions

| Version  | Supported          |
| -------- | ------------------ |
| latest   | :white_check_mark: |
| older    | :x:                |

Only the latest tagged release receives security fixes. Please update before reporting.

## Reporting a Vulnerability

Do **not** open a public issue for security vulnerabilities.

Use [GitHub's private vulnerability reporting](https://github.com/LarsArtmann/golangci-lint-auto-configure/security/advisories/new) to report privately. Include:

1. A description of the vulnerability and its impact
2. Steps to reproduce (a minimal config or command is ideal)
3. Affected version(s) (`golangci-lint-auto-configure --version`)

You can expect an initial response within 7 days. Fixes are released as soon as practical, and reporters are credited unless they prefer otherwise.

## Scope

This tool reads and writes golangci-lint configuration files locally. It does not run as a service, does not make network calls to third parties, and does not transmit any data. Issues of particular interest:

- Config injection or code execution through malicious `.golangci.yml` content
- Path traversal in config discovery, report generation, or the audit ledger
- Credential exposure through report output (HTML/JSON/SARIF)
