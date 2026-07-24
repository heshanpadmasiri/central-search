# central-search

`central-search` is a CLI for finding packages and reading package documentation from [Ballerina Central](https://central.ballerina.io/).

> The CLI interface is implemented, but the Ballerina Central backend is not configured yet. Valid `search` and `man` requests currently return a backend-not-configured error.

## Build

```sh
go build ./...
```

## Commands

### Search packages

```sh
central-search search <query>
```

Search is case-insensitive and matches the query as a substring of either the organization or package name. It does not search package summaries. Results are sorted by organization and package.

The default output is one package per line followed by its summary:

```text
ballerina/http  HTTP client and server APIs
```

Use `--json` for structured output:

```sh
central-search search http --json
```

```json
[
  {
    "organization": "ballerina",
    "package": "http",
    "summary": "HTTP client and server APIs"
  }
]
```

A search without matches prints an error and exits with a non-zero status. With `--json`, it also writes `[]` to stdout.

### Read package documentation

```sh
central-search man <package>
central-search man <organization/package>
```

An unqualified package name must uniquely identify a package. Documentation always targets the latest package version. Documentation parsing and terminal rendering will be added when the Central API is integrated; the command currently treats service content as opaque bytes.

## Test

```sh
go test ./...
```
