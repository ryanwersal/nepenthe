# Nepenthe

Selective forgetfulness for macOS Time Machine.

Nepenthe finds high-churn and regenerable directories — `node_modules`, `.build`, `DerivedData`, and more — and excludes them from Time Machine backups. The result: faster backups and less wasted storage.

## Install

### Homebrew

```sh
brew install ryanwersal/tap/nepenthe
```

### From source

Requires [mise](https://mise.jdx.dev) for toolchain management.

```sh
git clone https://github.com/ryanwersal/nepenthe.git
cd nepenthe
mise install
mise run build
mise run install   # copies binary to /usr/local/bin
```

## Quick start

Launch the interactive TUI:

```sh
nepenthe
```

Or scan from the command line:

```sh
nepenthe scan
```

Run `nepenthe --help` for the full command reference.

## Development

```sh
mise install        # set up Go + tooling
mise run build      # build to dist/
mise run test       # run tests
mise run lint       # run golangci-lint
mise run clean      # remove dist/
```

## License

[GPL-3.0](LICENSE)
