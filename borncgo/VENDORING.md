# borncgo: a copy of `born`

`borncgo/` is a copy of the upstream `born` ML framework, not a Go module
dependency. The module path was rewritten from the upstream path to `blue/borncgo`,
and it has no separate `go.mod`, so it is built as part of the `blue` module.

## Consequences

- It cannot be updated with `go get` or a `replace` directive. Updating it means
  copying new upstream code in.
- It does not appear in `go.sum`; `go mod vendor` does not manage it.
- Blue-specific glue (`ml/`, `object/std_ml.go`) sits on top and is free to
  change without touching borncgo.

## Why copy instead of depend

The integration needed internal packages and a fast edit loop while the engine
was changing daily. Depending on a fork would have meant a `replace` to a fork
URL or a published module path, and would have made the vendored WebGPU native
libraries harder to manage. Copying keeps one build and one import path.

## Syncing with upstream

Upstream: `https://github.com/born-ml/born` (also checked out at
`/home/brice/wss/gitrepos/born`).

1. Fetch and diff the upstream tree against `borncgo/`.
2. Copy changed files in, then rewrite imports from the upstream module path to
   `blue/borncgo`.
3. Keep the local additions that blue relies on (for example the fused
   `MatMulBias` path and the tracer hooks) and re-apply them if upstream changed
   the same files.
4. Run `go build ./...`, `go test ./ml/... ./object/...`, and
   `/tmp/blue_static b_test_programs/test_ml.b`.

This is a manual process on purpose: the engine is still moving, and a silent
overwrite would break the blue-facing glue.
