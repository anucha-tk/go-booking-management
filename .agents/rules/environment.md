---
trigger: always_on
glob: "**/*.go"
description: "Setup environment for Go project on macOS"
---

# Environment Configuration

Before running any Go-related commands (go, task, golangci-lint), ensure the following paths are added to the PATH environment variable:

- `/opt/homebrew/bin` (Homebrew, Go, Task)
- `/Users/anucha-tk/go/bin` (Go binaries, golangci-lint)

## Mandatory Command Prefix
Always prepend the PATH update to your commands:
`export PATH="/opt/homebrew/bin:/Users/anucha-tk/go/bin:$PATH"`

## Go Settings
- GOROOT: Automatically detected from `/opt/homebrew/bin/go`
- GOPATH: `/Users/anucha-tk/go`
- GO111MODULE: `on`
