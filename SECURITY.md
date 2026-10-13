# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| older   | :x:                |

Only the latest tagged release receives security fixes. Please update before reporting.

## Reporting a Vulnerability

Do **not** open a public issue for security vulnerabilities.

Use [GitHub's private vulnerability reporting](https://github.com/LarsArtmann/golangci-lint-auto-configure/security/advisories/new) to report privately. Include:

1. A description of the vulnerability and its impact
2. Steps to reproduce (a minimal config or command is ideal)
3. Affected version(s) (`golangci-lint-auto-configure --version`)

You can expect an initial response within 7 days. Fixes are released as soon as practical, and reporters are credited unless they prefer otherwise.

## Security Practices

- **Dependency scanning:** [`govulncheck`](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) runs in CI (`govulncheck ./...`) and fails the build on known vulnerabilities reachable from the module graph.
- **Secret scanning:** `gitleaks` runs as a weekly scheduled workflow over the full git history, so credentials committed directly to a branch are caught even if they never appear in a PR diff.
- **Least-privilege CI:** workflows declare explicit GitHub token `permissions` instead of relying on defaults. CI, lint, and scan workflows are read-only; only release/image-publish/watchdog workflows hold the specific write scopes they need (`contents`+`packages`+`id-token` for releases, `packages` for image backfill, `issues` for the CI watchdog).
- **Local-only operation:** the tool parses and writes golangci-lint configs on disk. It makes no third-party network calls, collects no telemetry, and keeps its audit ledger under the local OS cache directory.
- **No secrets in output:** reports (HTML/JSON/SARIF) summarize lint findings about config structure — they do not embed file contents or credentials.

## Scope

This tool reads and writes golangci-lint configuration files locally. It does not run as a service, does not make network calls to third parties, and does not transmit any data. Issues of particular interest:

- Config injection or code execution through malicious `.golangci.yml` content
- Path traversal in config discovery, report generation, or the audit ledger
- Credential exposure through report output (HTML/JSON/SARIF)
