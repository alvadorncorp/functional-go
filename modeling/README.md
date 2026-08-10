# modeling

`modeling` is an explicit, compile-time-safe Go metamodel for application-owned
domain types and behavior. It sits above the public capabilities in
[`functional`](../functional/) and below concrete application domains.

The module intentionally contains no reflection, runtime registry, dispatcher,
infrastructure, persistence, or concrete business model.

## Documentation

- [Documentation index](docs/README.md)
- [Architecture overview](docs/overview.md)

## Development

```sh
go test ./...
go test -race ./...
go vet ./...
go mod tidy -diff
```
