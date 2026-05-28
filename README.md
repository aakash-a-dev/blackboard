# mockapi

Spin up a live mock HTTP server from a YAML file. No code, no account, no runtime.

```bash
mockapi serve api.yaml
```

## Install

**Linux / macOS** — one-liner, no Go required:
```bash
curl -sSL https://raw.githubusercontent.com/aakash-a-dev/blackboard/main/blackboard-cli/scripts/install.sh | bash
```

**Windows** — run in PowerShell (adds to PATH automatically):
```powershell
irm https://raw.githubusercontent.com/aakash-a-dev/blackboard/main/blackboard-cli/scripts/install.ps1 | iex
```

**Go toolchain** (any OS):
```bash
go install github.com/aakash-a-dev/blackboard@latest
```

**Manual** — download a pre-built binary from [GitHub Releases](https://github.com/aakash-a-dev/blackboard/releases) and place it anywhere in your `PATH`.

## Quickstart

```bash
mockapi init          # scaffold api.yaml in the current directory
mockapi serve api.yaml
```

## Commands

| Command | Description |
|---|---|
| `mockapi serve <file>` | Start the mock server |
| `mockapi validate <file>` | Validate YAML, exit |
| `mockapi init` | Scaffold a starter api.yaml |

## `serve` flags

| Flag | Default | Description |
|---|---|---|
| `--port` | `4000` | Port to listen on |
| `--host` | `127.0.0.1` | Interface to bind |
| `--watch` | `true` | Hot-reload on file save |
| `--latency` | `0` | Global latency in ms |
| `--cors` | `true` | Permissive CORS headers |
| `--log` | `pretty` | `pretty` / `json` / `silent` |

## YAML reference

```yaml
version: "1"
info:
  title: My API
  base_path: /api/v1   # optional prefix

endpoints:
  - path: /users
    method: GET
    response:
      status: 200
      body:
        - id: "{{uuid}}"
          name: "{{fullName}}"
          email: "{{email}}"
      count: 10          # repeat body item N times

  - path: /users/:id
    method: GET
    response:
      status: 200
      body:
        id: "{{param.id}}"
        name: "{{fullName}}"

  - path: /users
    method: POST
    request:
      required_fields: [name, email]
      body_schema:
        email:
          type: string
          format: email
    response:
      status: 201
      body:
        id: "{{uuid}}"

  # Error simulation
  - path: /payments
    method: POST
    behavior:
      error_rate: 0.15
      error_status: 503
      error_body:
        message: "Service unavailable"
    response:
      status: 200
      body:
        ok: true

  # Artificial latency
  - path: /reports
    method: GET
    behavior:
      latency_ms: 800
    response:
      status: 200
      body:
        data: "{{sentence}}"

  # Pagination
  - path: /posts
    method: GET
    response:
      status: 200
      paginated: true
      body:
        - id: "{{uuid}}"
          title: "{{sentence}}"
      total: 50

  # Conditional variants
  - path: /products
    method: GET
    variants:
      - condition:
          query:
            category: "electronics"
        response:
          status: 200
          body:
            - name: "Laptop"
          count: 5
      - default: true
        response:
          status: 200
          body: []

  # Custom headers
  - path: /auth/token
    method: POST
    response:
      status: 200
      headers:
        X-Request-Id: "{{uuid}}"
        Cache-Control: "no-store"
      body:
        token: "{{uuid}}"
        expires_in: 3600
```

## Fake data tokens

| Token | Example |
|---|---|
| `{{uuid}}` | `a3f2b1c0-...` |
| `{{fullName}}` | `Jane Appleton` |
| `{{firstName}}` | `Jane` |
| `{{lastName}}` | `Appleton` |
| `{{email}}` | `jane@example.com` |
| `{{phone}}` | `+1-555-0132` |
| `{{isoDate}}` | `2026-03-15T10:22:00Z` |
| `{{date}}` | `2026-03-15` |
| `{{int(1,100)}}` | `42` |
| `{{float(0,1)}}` | `0.73` |
| `{{bool}}` | `true` |
| `{{sentence}}` | `The quick brown fox` |
| `{{word}}` | `laptop` |
| `{{url}}` | `https://example.com/foo` |
| `{{param.id}}` | URL path segment `:id` |
| `{{query.field}}` | Query string value |
| `{{body.field}}` | Request body JSON field |

## Contributing

1. Fork and clone the repo
2. `go mod tidy`
3. `go run . serve testdata/api.yaml` to smoke-test locally
4. Add tests under `internal/*/` alongside the package they cover
5. Open a PR — keep it focused, one concern per PR

All decisions that shaped the v1 design live in [doc/decisions.md](../doc/decisions.md).
