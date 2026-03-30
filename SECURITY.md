# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | ✅ Yes    |

## Reporting a Vulnerability

Symphony CLI processes templates from external sources and can execute shell
commands via post-scaffold hooks. Security is a top priority.

**Do NOT create a public GitHub Issue for security vulnerabilities.**

Please report security vulnerabilities privately using **GitHub Security Advisories**:

1. Go to https://github.com/Reivhell/Symphony-CLI/security/advisories
2. Click "New draft security advisory"
3. Describe the vulnerability, steps to reproduce, and potential impact

We commit to responding within **48 hours** and releasing a patch within
**7 days** for confirmed vulnerabilities.

## Security Scope

**In-scope (please report):**
- Expression evaluator injection (`pkg/expr`) — template `if:` conditions
- Path traversal in file writer (`internal/engine/writer.go`)
- Malicious post-scaffold hook execution
- Remote template fetching vulnerabilities
- Arbitrary code execution via template rendering

**Out-of-scope:**
- Security of third-party templates (responsibility of template authors)
- Vulnerabilities in dependencies (report to respective maintainers)

## Security Hardening Notes for Template Authors

Symphony executes shell commands defined in template `hooks`. Users are
encouraged to always review `template.yaml` — especially the `actions` section
with `type: shell` — before running `symphony gen` on untrusted templates.
