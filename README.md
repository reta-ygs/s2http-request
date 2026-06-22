# s2http-request

A **Simple and Structured HTTP Request** dispatching tool.

## Table of Contents

- [Overview](#overview)
- [Features]
- [Request Definition Format]
- [Built-in Functions]
- [Variable Overrides]
- [Usage]
- [Output Format]
- [Directory Structure]
- [License]
- [Contributing]

## Overview

s2http-request is a versatile and lightweight HTTP request dispatching tool designed for various testing and automation scenarios. It allows users to define complex HTTP requests using pure JSON/YAML/JSONL, eliminating the need for domain-specific languages (DSLs) or templating engines. This approach ensures clarity, simplicity, and ease of integration with existing workflows.

## Features

- 🚀 **Lightweight & Portable:** Distributed as a single, self-contained binary.
- 📝 **Declarative Configuration:** Define requests using intuitive JSON/YAML/JSONL files.
- 🔧 **Dynamic Value Generation:** Leverage powerful built-in functions for on-the-fly data manipulation.
- 📊 **Comprehensive Output:** Get detailed response information for thorough analysis.
- ⚡ **High Performance:** Built with Go for speed and efficiency.
- 📋 **JSON Schema Support:** Benefit from IDE auto-completion and validation for request definitions.

## Request Definition Format

### Basic Request Definition

```json
{
  "method": "GET",
  "path": "/",
  "query": {
    "q": "hello",
    "r": "",
    "s": null,
    "t": {
      "$var": "malicious_key"
    }
  },
  "headers": {
    "Content-Type": "application/www-form-urlencoded"
  },
  "params": {
    "id": "1",
    "name": {
      "$url_encode": "John Doe"
    }
  },
  "variables": {
    "malicious_key": {
      "$concat": [
        {"$var": "v"},
        {"$var": "k"}
      ]
    },
    "v": "1",
    "k": "2"
  }
}
```

### JSONL Request Definition

```jsonl
{"method": "GET", "path": "/api/users"}
{"method": "GET", "path": "/api/users", "query": {"id": 1}}
{"method": "GET", "path": "/api/users", "query": {"id": 2}}
{"method": "POST", "path": "/api/users", "body": {"name": "John", "email": "john@example.com"}}
```

In JSONL format, each line is a separate JSON object. The tool will parse and execute the first valid JSON object found in the file.

### Array-based Parameter Definition

```json
{
  "method": "GET",
  "path": "/",
  "query": [
    {
      "key": "q",
      "value": "v"
    },
    {
      "key": "q",
      "value": "v2"
    },
    {
      "key": {
        "$join": [
          ",",
          {"$var": "v"},
          {"$var": "k"}
        ]
      },
      "value": "v"
    }
  ],
  "headers": [
    {
      "key": "Content-Type",
      "value": "multipart/form-data"
    }
  ],
  "params": [
    {
      "key": "p",
      "value": "v"
    },
    {
      "key": "p",
      "value": "v2"
    },
    {
      "key": {
        "$join": [
          ",",
          {"$var": "v"},
          {"$var": "k"}
        ]
      },
      "value": "v"
    }
  ],
  "variables": {
    "v": {
      "$random": [10]
    },
    "k": "2"
  }
}
```

## Built-in Functions

The tool provides a set of built-in functions for dynamic value generation.

### Variable Operations
- `$var`: Reference a variable
- `$concat`: Concatenate strings
- `$join`: Join array elements

### Encoding/Decoding
- `$url_encode`: URL encode a string
- `$url_decode`: URL decode a string
- `$base64_encode`: Base64 encode a string
- `$base64_decode`: Base64 decode a string
- `$html_encode`: HTML encode a string
- `$html_decode`: HTML decode a string

### Random Generation
- `$random`: Generate a random number (0 to N-1)
- `$random_string`: Generate a random string
- `$uuid`: Generate a UUID

### Time Functions
- `$timestamp`: Current Unix timestamp
- `$date`: Current date (YYYY-MM-DD)
- `$time`: Current time (HH:MM:SS)

### File Operations
- `$file`: Read file content as string (relative paths only)

### Array Operations
- `$concat_arrays`: Concatenate multiple arrays

## Variable Overrides

You can override variables defined in request files using the `--var` command-line flag. This is particularly useful for:

- Testing with different parameter values without modifying request files
- Environment-specific configurations
- Dynamic testing scenarios

### Variable Override Syntax

```bash
# Override single variable
s2req --var key=value request.yaml

# Override multiple variables
s2req --var user_id=123 --var type=admin --var limit=50 request.yaml

# Override with JSON values (arrays, objects, booleans, numbers)
s2req --var 'ids=[1,2,3]' request.yaml
s2req --var 'config={"enabled":true,"timeout":30}' request.yaml
s2req --var 'active=true' request.yaml
s2req --var 'count=42' request.yaml
```

### Variable Override Priority

CLI variables take precedence over file-defined variables:

```yaml
# request.yaml
method: GET
path:
  $concat:
    - "/api/users/"
    - $var: user_id
variables:
  user_id: "default_user"
```

```bash
# This will override the file variable
s2req --var user_id=123 request.yaml
# Result: GET /api/users/123

# Without override
s2req request.yaml  
# Result: GET /api/users/default_user
```

### Supported Value Types

The `--var` flag automatically detects and parses different value types:

- **Strings**: `--var name=John` → `"John"`
- **Numbers**: `--var count=42` → `42`, `--var rate=3.14` → `3.14`
- **Booleans**: `--var enabled=true` → `true`, `--var disabled=false` → `false`
- **JSON Arrays**: `--var 'items=[1,2,3]'` → `[1,2,3]`
- **JSON Objects**: `--var 'config={"key":"value"}'` → `{"key":"value"}`

## Usage

### Installation

```bash
# Install the latest version
go install github.com/secureta/s2http-request/cmd/s2req@latest
go install github.com/secureta/s2http-request/cmd/s2req-schema@latest

# Or, clone the repository and install from the local directory
git clone https://github.com/secureta/s2http-request.git
cd s2http-request
go install ./cmd/s2req
go install ./cmd/s2req-schema
```

The `install` command builds the `s2req` binary and installs it into your `$GOPATH/bin` or `$GOBIN` directory.

Alternatively, you can download a pre-built binary from the [Releases](https://github.com/your-username/s2http-request/releases) page.

```bash
# Example for Linux
curl -L https://github.com/your-username/s2http-request/releases/latest/download/s2req-linux-amd64 -o s2req
chmod +x s2req
# You can move it to a directory in your PATH, e.g., /usr/local/bin
```

### JSON Schema Generation

```bash
# Generate JSON Schema
go tool mage schema

# IDE auto-completion settings (VS Code example):
# Add to .vscode/settings.json:
{
  "json.schemas": [
    {
      "fileMatch": ["**/requests/*.json", "**/examples/*.json"],
      "url": "./request-schema.json"
    }
  ]
}
```

### Basic Usage

```bash
# Execute request from a JSON file
s2req request.json

# Execute request from a YAML file
s2req request.yaml

# Execute request from a JSONL file
s2req request.jsonl

# Execute multiple request files
s2req requests/*.json

# Read from standard input (JSON, JSONL, or YAML format)
cat request.json | s2req
echo '{"method": "GET", "path": "/"}' | s2req
s2req < request.yaml

# Use "-" to explicitly read from stdin
cat request.jsonl | s2req -

# Verbose output mode
s2req --verbose request.json

# Save results to a file
s2req --output results.json request.json
```

### Configuration Options

```bash
# Specify target host
s2req --host https://example.com request.json

# Set timeout
s2req --timeout 30 request.json

# Set retry count
s2req --retry 3 request.json

# Use proxy
s2req --proxy https://proxy.example.com:8080 request.json

# Override variables from command line
s2req --var user_id=123 --var type=admin request.yaml

# Override with JSON values
s2req --var 'ids=[1,2,3]' --var 'config={"enabled":true}' request.yaml
```

## Output Format

```json
{
  "request": {
    "method": "GET",
    "url": "https://example.com/?q=hello&t=12",
    "headers": {
      "Content-Type": "application/www-form-urlencoded"
    },
    "body": "id=1&name=John%20Doe"
  },
  "response": {
    "status_code": 200,
    "headers": {
      "Content-Type": "text/html",
      "Content-Length": "1234"
    },
    "body": "...",
    "time": {
      "total": 0.123,
      "dns": 0.001,
      "connect": 0.010,
      "ssl": 0.050,
      "send": 0.001,
      "wait": 0.060,
      "receive": 0.001
    }
  },
  "metadata": {
    "timestamp": "2025-06-09T04:44:46.758Z",
    "variables": {
      "v": "1",
      "k": "2",
      "malicious_key": "12"
    }
  }
}
```

## Directory Structure

```
s2http-request/
├── examples/
├── internal/
│   ├── config/
│   ├── http/
│   └── parser/
└── pkg/
    └── functions/
```

## Nuclei Template Conversion

`s2req convert-nuclei` analyzes a Nuclei HTTP template and emits a JSON conversion report.
The initial converter intentionally supports only a safe HTTP request subset and records unsupported scanner semantics explicitly.

```bash
s2req convert-nuclei -output report.json template.yaml
```

The report contains:

- `template_id` and `info` metadata from the Nuclei template
- converted s2http `requests`
- `unsupported` entries for features that need a later runner/evaluator layer, such as non-HTTP protocols, `flow`, `raw`, `matchers`, and `extractors`

The converter keeps URL query parameters separate from `path` and emits query parameters as ordered key-value pairs so duplicate keys can be preserved by later request-writing improvements.

## Development

Quality checks are shared between GitHub Actions and local git hooks via `scripts/quality.sh`.

```bash
# Run the same checks used by the pre-commit hook
scripts/quality.sh pre-commit

# Run CI-equivalent checks locally
scripts/quality.sh ci

# Install the repository pre-commit hook
scripts/install-git-hooks.sh
```

The pre-commit hook runs fast checks: `gofmt`, `go vet`, baseline `golangci-lint`, and `go test ./...`. CI also runs race-enabled tests, gosec SAST, build checks, and CodeQL.

## License

MIT License

## Contributing

Pull requests and issue reports are welcome.
