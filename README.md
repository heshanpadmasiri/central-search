# central-search

`central-search` is a CLI for finding packages and reading package documentation from [Ballerina Central](https://central.ballerina.io/).

> Package search uses the public Ballerina Central API. The package-documentation backend used by `man` is not configured yet.

## Build

```sh
go build ./...
```

## Commands

### Search packages

```sh
central-search search <query>
```

Results retain the relevance order returned by Ballerina Central. Central returns 15 results by default; use `--limit` to request a different maximum:

```sh
central-search search http --limit 25
```

The default output contains one package per line with its version and summary:

```text
ballerina/http  2.16.4  This module provides APIs for HTTP and HTTP2 endpoints.
```

Use `--json` for a stable, compact JSON representation owned by this CLI:

```sh
central-search search http --json
```

```json
[
  {
    "organization": "ballerina",
    "package": "http",
    "version": "2.16.4",
    "summary": "This module provides APIs for HTTP and HTTP2 endpoints."
  }
]
```

The version is the single version returned by Central's search endpoint, normally the latest package version. A search without matches prints an error and exits with a non-zero status. With `--json`, it also writes `[]` to stdout.

### Read package documentation

```sh
central-search man <package>
central-search man <organization/package>
```

An unqualified package name must uniquely identify a package. Documentation always targets the latest package version. Documentation parsing, rendering, and the Central documentation backend have not yet been implemented.

## Test

```sh
go test ./...
```
