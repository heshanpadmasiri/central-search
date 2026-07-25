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

## LLM instructions

Print the bundled skill Markdown as plain text for use by language models:

```sh
central-search llm
```

## Read package documentation

Read human-friendly documentation or request structured JSON:

```sh
central-search man <query>
central-search man <query> --json
```

The query is passed to Central search, so queries containing spaces must be quoted. The command follows every search page. A qualified query that exactly matches an `organization/package` is selected; otherwise, resolution requires exactly one distinct match. The command reports all candidates when a fuzzy query remains ambiguous.

After resolution, the command selects the first version returned by Central (currently the newest) and scrapes the default module and every related module. Human output includes every non-empty package, module, declaration, and nested-member field. It is rendered as styled, width-aware Markdown in an interactive terminal and as raw Markdown when piped or redirected. Stable headings such as `## Module: http`, `### Functions`, and `### Enums` make sections easy to extract:

```sh
central-search man ballerina/http |
  awk '/^### Functions$/ {show=1; next} /^### / {show=0} show'
```

Interactive output uses `$MANPAGER`, then `$PAGER`, then `less -R` when available. Set `NO_COLOR` to disable color while retaining terminal layout. When `TERM=dumb`, the command writes raw Markdown without a pager.

With `--json`, the command emits one schema-versioned JSON object followed by a newline. Package/module overviews and API documentation remain Markdown strings, and collection values are arrays, including when empty.

The top-level `complete` field is `true` when all related modules were retrieved. If a related module fails, the command still succeeds with the available modules, sets `complete` to `false`, adds an entry to `warnings`, and prints that warning to stderr. Human output also includes the warning. Failure to retrieve the default module is fatal and produces no documentation.

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
