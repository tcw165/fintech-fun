# Dev rules

High-level only. Details live in the code and `README.md`.

## Naming

- Variables, parameters, and unexported fields/helpers are **snake_case**, lowercase.
- Exported interfacing functions stay readable PascalCase. Client accessors end in `Client`.

```go
type AppContext struct {
	graph_db  neo4j.Client
	vector_db qdrant.Client
	embedder  embed.Embedder
	closer    func(context.Context) error
}

func (c *AppContext) GraphClient() neo4j.Client { return c.graph_db }
func (c *AppContext) VectorsClient() qdrant.Client { return c.vector_db }
```

- Exported types and public data fields stay PascalCase (Go export). JSON tags stay snake_case.

## Modules

- Every folder is a Bazel module.
- Models and contracts (pure data / protocol) are their own module.
- Interface and implementation live in different modules.
- Dependencies point child → parent. Inject interfaces; do not import impl from parents.
- `di/` is for runnable composition roots only.

## Python

- New Python uses Bazel. Do not add `__init__.py`.
- Do not use a `test_tuit` target.

## Multi-line args

Call sites with three or more arguments, or one long string argument, put each argument on its own line. The same rule applies to Cobra `SetArgs`, flag helpers, and composite literals in tests so wrapped headlines and multi-word queries stay readable.

```go
got := execute_cli(
	t,
	"search",
	"--query",
	"LivePerson\nstock\nmerger",
	"--limit",
	"3",
)
```

Headline and query strings may contain newlines. Collapse that whitespace the same way the tracker parser does (`\s+` → one space) before classify, search, or fold.

## PRs

Always a **linear Graphite stack**. Not optional. Not a single fat PR. Not four PRs all based on `main`.

```
main ← PR1 ← PR2 ← PR3 ← …
```

1. `gt sync` on `main` before the first PR.
2. `gt create -m "[tag] …"` for PR 1. Its parent is `main`.
3. `gt create` again for each next change. Parent is the previous branch. One commit per branch.
4. Open and update the stack with `gt submit --stack --publish` only. Do not create PRs with `gh` or any other GitHub tool first — Graphite will not see the parent chain and the dashboard will show a lone PR on `main`.
5. After submit, share the Graphite URLs (`https://app.graphite.com/github/pr/OWNER/REPO/N`), not only GitHub.
6. Never land work on `dev/wt*` placeholder branches.
