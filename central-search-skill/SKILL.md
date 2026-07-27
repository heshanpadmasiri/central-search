# Central Search

Use `central-search` to find Ballerina packages and inspect their documentation.

## Find a package

```sh
central-search search '<query>' --json
```

## Read structured documentation

```sh
central-search man '<query>' --json
```

The query must resolve to exactly one package. The result contains package metadata and a `modules` array. Each module separates declarations into `functions`, `classes`, `clients`, `listeners`, `objects`, `services`, `records`, `enums`, `types`, `errors`, `constants`, `variables`, `configurables`, and `annotations`.

Every declaration has these stable fields:

```json
{
  "kind": "function",
  "name": "parseHeader",
  "signature": "function parseHeader(string header) returns map<string|string[]>",
  "documentation": "...",
  "deprecated": false
}
```

Function arguments, return values, record fields, methods, and enum members are also structured. Prefer `signature` when reporting how an API is declared, and use the structured fields when explaining individual parameters or fields.

## Useful queries

List modules:

```sh
central-search man '<query>' --json | jq -r '.modules[].name'
```

List functions and signatures in a module:

```sh
central-search man '<query>' --json |
  jq -r '.modules[] | select(.name == "<module>") | .functions[] | "\(.name): \(.signature)"'
```

List all functions with their arguments and documentation:

```sh
central-search man '<query>' --json |
  jq '.modules[] as $module | $module.functions[] | {module: $module.name, name, signature, documentation, arguments: [.arguments[] | {name, type, signature, documentation, optional, rest, defaultValue}]}'
```

Each result includes the module name, complete function signature and documentation. Its `arguments` array includes each argument's declaration, type, documentation, optional and rest markers, and default value when present.

List records and their signatures:

```sh
central-search man '<query>' --json |
  jq -r '.modules[].records[] | "\(.name): \(.signature)"'
```

Inspect a record's fields:

```sh
central-search man '<query>' --json |
  jq '.modules[].records[] | select(.name == "<record>") | .fields'
```

List all clients and their method definitions:

```sh
central-search man '<query>' --json |
  jq '.modules[] as $module | $module.clients[] | {module: $module.name, name, signature, methods: [(.methods + .remoteMethods + .resourceMethods)[] | {name, signature}]}'
```

Each result retains the module name and client signature, and groups the client's regular, remote, and resource methods under `methods`. Use the method `signature` for the complete declaration.

Find any symbol recursively:

```sh
central-search man '<query>' --json |
  jq --arg name '<symbol>' '.. | objects | select(.name? == $name and has("signature"))'
```

List every documented symbol:

```sh
central-search man '<query>' --json |
  jq '.. | objects | select(has("kind") and has("signature")) | {kind, name, signature}'
```
