# central-search

`central-search` searches packages and reads structured API documentation from [Ballerina Central](https://central.ballerina.io/).

## Build

```sh
go build ./...
```

## Search packages

```sh
central-search search <query>
central-search search http --limit 25
central-search search http --json
```

Results retain Central's relevance order. The JSON search output is a stable array containing `organization`, `package`, `version`, and `summary`.

## Read package documentation

Documentation is currently available as JSON:

```sh
central-search man <query> --json
```

The query is passed to Central search, so queries containing spaces must be quoted. The command follows every search page. A qualified query that exactly matches an `organization/package` is selected; otherwise, resolution requires exactly one distinct match. The command reports all candidates when a fuzzy query remains ambiguous.

After resolution, the command selects the first version returned by Central (currently the newest), scrapes the default module and every related module, and emits one schema-versioned JSON object followed by a newline. Package/module overviews and API documentation remain Markdown strings. Collection values are arrays, including when empty.

The top-level `complete` field is `true` when all related modules were retrieved. If a related module fails, the command still succeeds with the available modules, sets `complete` to `false`, adds an entry to `warnings`, and prints that warning to stderr. Failure to retrieve the default module is fatal and produces no JSON. Non-JSON terminal rendering is deferred; invoking `man` without `--json` reports how to request JSON.

### jq recipes

List default-module functions and signatures:

```sh
central-search man ballerina/http --json |
  jq '.modules[] | select(.isDefault) | .functions[] | {name, signature}'
```

List all declaration signatures:

```sh
central-search man ballerina/http --json |
  jq -r '.. | objects | select(.kind? and .signature?) | .signature'
```

Inspect Kafka clients and remote methods:

```sh
central-search man ballerinax/kafka --json |
  jq '.modules[].clients[] | {name, remoteMethods: [.remoteMethods[] | {name, signature}]}'
```

Inspect records and fields:

```sh
central-search man ballerina/http --json |
  jq '.modules[].records[] | {name, fields: [.fields[] | {name, type, signature}]}'
```

Find any symbol named `Client` recursively:

```sh
central-search man ballerina/http --json |
  jq '.. | objects | select(.kind? and .name? == "Client") | {kind, signature}'
```

## Validation

```sh
gofmt -w $(find . -name '*.go' -not -path './vendor/*')
go test ./...
go test -race ./...
go vet ./...
go build ./...
```
