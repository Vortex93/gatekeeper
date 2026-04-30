# Contributions

## Development Setup

- Install Go `1.21.3` or newer
- Clone the repository
- Run `go test ./...` before opening a pull request

Optional local task shortcuts:

```bash
task build
task test
task race
task check
```

## Pull Requests

- Keep changes focused and easy to review
- Add or update tests when behavior changes
- Update `README.md` when the public API or behavior changes
- Verify `go test ./...` and `go test -race ./...` pass before submitting

## Releases

Maintainers can cut a release with:

```bash
task release VERSION=0.1.8
```

That task runs tests, creates the tag, and pushes it to GitHub. The release workflow then publishes source archives and checksums.
