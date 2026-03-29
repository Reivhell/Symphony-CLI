# Blueprint specification (`template.yaml`)

Symphony blueprints are YAML files named `template.yaml` (or resolved via inheritance). This document describes each supported field and the external **plugin** protocol.

---

## Top-level fields

### `schema_version` (string, required)

Document format version. Use **`2`**.

**Example:** `schema_version: "2"`

---

### `name` (string, required)

Human-readable template name.

---

### `version` (string, required)

Semantic version of the template (not the CLI).

---

### `author` (string, optional)

Author or maintainer identifier (name, org, or handle).

---

### `description` (string, optional)

Short description shown in tooling and metadata.

---

### `min_symphony_version` (string, optional)

Minimum Symphony CLI version required to run this blueprint. If the running CLI is older, generation should fail early.

**Example:** `min_symphony_version: "0.1.0"`

---

### `tags` (list of strings, optional)

Free-form labels for search or categorization.

**Example:**

```yaml
tags: [go, api, grpc]
```

---

### `extends` (string, optional)

Path or reference to a parent blueprint. Child values override parent; lists such as `plugins` and `actions` are merged according to Symphony’s inheritance rules.

---

### `validations` (list, optional)

Rules applied after prompts. Each entry:

| Field | Type | Meaning |
|--------|------|---------|
| `field` | string | Prompt `id` to validate |
| `rule` | string | e.g. `required`, `regex`, `min_length` |
| `pattern` | string | Regex pattern when `rule` is `regex` |
| `message` | string | Error message on failure |

---

### `prompts` (list, required for interactive flows)

Each prompt:

| Field | Type | Meaning |
|--------|------|---------|
| `id` | string | Key stored in the engine context |
| `question` | string | Text shown to the user |
| `type` | string | `input`, `select`, `multiselect`, `confirm` |
| `options` | list | Choices for `select` / `multiselect` |
| `default` | any | Default value |
| `depends_on` | string | Optional expression controlling visibility |

---

### `actions` (list, required)

Steps executed in order.

| `type` | Required fields | Behavior |
|--------|-----------------|----------|
| `render` | `source`, `target` | Copy/transform a file from template dir to output |
| `shell` | `command` | Run a shell command (often after files exist) |
| `ast-inject` | `target`, `content`, `anchor` / injector strategy | Inject generated content into an existing file |

Optional on any action:

- `if` — expression evaluated against prompt values; if false, the action is skipped.

Paths in `source` / `target` may use template syntax where supported by the engine.

---

### `completion_message` (string, optional)

Message shown after successful generation (may be templated depending on engine support).

---

### `plugins` (list, optional)

Registers **external renderer** programs for files that should not use the built-in `text/template` pipeline.

Each plugin:

| Field | Type | Meaning |
|--------|------|---------|
| `name` | string | Logical name for errors/logging |
| `executable` | string | Path to the binary, absolute or relative to the template root |
| `handles` | list of strings | Glob patterns matched against the **source file base name** (e.g. `*.proto`) |

**Precedence:** The first plugin whose pattern matches wins. If none match, Symphony uses normal rules (e.g. `.tmpl` files go through `text/template`).

#### Plugin I/O protocol (JSON)

**stdin (request):**

```json
{
  "context": { "...": "prompt values and OUTPUT_DIR, SOURCE_DIR" },
  "file_content": "raw file body",
  "source_path": "/absolute/path/to/source",
  "target_path": "/absolute/path/to/target"
}
```

**stdout (response):**

```json
{
  "rendered_content": "final file body",
  "error": "optional error message"
}
```

Execution timeout: **30 seconds**. The executable must not contain `..` path segments; it must exist and be a regular file with execute permission (on Unix).

---

## See also

- [Contributing](contributing.md)
- [README](../README.md)
