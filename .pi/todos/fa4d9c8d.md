{
  "id": "fa4d9c8d",
  "title": "Implement Central-backed package search",
  "tags": [
    "implementation"
  ],
  "status": "closed",
  "created_at": "2026-07-24T02:19:22.576Z"
}

Implemented Central-backed package search with an injectable HTTP client/base URL, typed Central DTOs, a separate search use-case package, optional positive --limit forwarding, relevance-order preservation, singular version output, compact CLI-owned JSON, split search/documentation dependencies, production wiring, updated README, and comprehensive tests. Verified with gofmt, go test ./..., go vet ./..., go test -race ./..., git diff --check, and a live `go run . search http --limit 1 --json` request.
