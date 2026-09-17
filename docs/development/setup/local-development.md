# Local Development

## Clone and Build

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
```

Run it directly:

```bash
./openframe --version
./openframe --help
```

## Running Without a Full Build

For quick iteration, use `go run` instead of building a binary each time:

```bash
go run . cluster list
go run . app status --output json
```

## Injecting Version Metadata

The root command's version string (`Version`, `Commit`, `Date`) is populated via linker flags at release time, and falls back to Go's embedded VCS info (`debug.ReadBuildInfo()`) on plain dev builds — so `go build -o openframe .` still shows a real commit hash instead of `"none"`. To simulate a release build:

```bash
go build -ldflags "-X github.com/flamingo-stack/openframe-cli/cmd.version=v1.2.3 \
  -X github.com/flamingo-stack/openframe-cli/cmd.commit=abc1234 \
  -X github.com/flamingo-stack/openframe-cli/cmd.date=2025-01-01T00:00:00Z" \
  -o openframe .
```

## Running Against a Real Local Cluster

Most cluster-creation code paths shell out to Docker/k3d/Helm/Terraform via the `CommandExecutor` abstraction. To exercise them for real (not via mocks), you need Docker running locally:

```bash
go build -o openframe .
./openframe prerequisites check
./openframe cluster create dev-test --skip-wizard --nodes 1
./openframe app install dev-test --non-interactive
./openframe cluster delete dev-test --force
```

> **Warning:** These commands create real Docker containers (`k3d`/k3s nodes) and install real Helm releases. Always clean up with `cluster delete` when you're done testing.

## Debug Configuration

### Verbose Output

Pass `--verbose` to any command to see timestamped debug logs, including the exact external commands being executed (with secrets redacted via `internal/shared/redact`):

```bash
./openframe cluster create dev-test --verbose
```

### Plain / Non-Interactive Mode

When debugging in a terminal that doesn't handle spinners well (or when piping output to a file), use `--plain`:

```bash
./openframe app status --plain
```

### Debugging in an IDE

Both VS Code and GoLand can run/debug `main.go` directly with custom arguments. For example, in VS Code's `launch.json`:

```json
{
  "name": "Debug openframe cluster list",
  "type": "go",
  "request": "launch",
  "mode": "auto",
  "program": "${workspaceFolder}",
  "args": ["cluster", "list", "--verbose"]
}
```

## Using the Mock Executor Instead of Real Tools

Much of the business logic depends on `internal/shared/executor.CommandExecutor`, which has a `MockCommandExecutor` test double. When developing new features, you can wire a `MockCommandExecutor` to simulate `docker`/`k3d`/`helm`/`terraform` output without needing those tools installed — see `tests/testutil/setup.go`'s `CreateStandardTestFlags` and `NewTestMockExecutor` for the pattern used throughout the test suite.

## Hot Reload

There is no dedicated hot-reload/watch tool bundled with this CLI project (it's a compiled Go binary, not a long-running server). The fastest iteration loop is:

```bash
go run . <command> <args>
```

which recompiles and executes in one step, or a lightweight file-watcher of your choice (e.g. `entr`, `air`) wrapping `go build`.

## Next Steps

See the [Architecture Overview](../architecture/README.md) to understand how commands, services, and providers fit together, and [Testing](../testing/README.md) for how to validate your changes.
