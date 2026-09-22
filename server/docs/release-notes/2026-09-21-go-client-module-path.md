# Go client: a resolvable module path

Date: 2026-09-21

The Go client's module path was `toolplane-go-client` — not a path any
Go toolchain could resolve, so `go get` could never work. The module is
now `github.com/Abhishek-chohan/tool-control-plane/clients/go-client`,
matching its repository location; from the first tagged release, the
canonical install is:

```bash
go get github.com/Abhishek-chohan/tool-control-plane/clients/go-client@v0.1.0
```

All in-module imports (client, proto, examples) were rewritten to the
new path; `go build` and `go vet` are green across the module. The
server module (`toolplane`) is unchanged — it is a self-contained
binary build that never imports the client, so nothing outside
`clients/go-client` is affected.

Notes: the import path is a public surface from the first release,
which is why this lands before the tag rather than after. The server's
proto stubs remain the canonical generated contract; the client's
`proto` package is its vendored copy, as before.
