# Development

## Requirements

- Go, version as in `wasm/go.mod`
- Node 22
- golangci-lint v2.15.0 (optional, CI runs it)

## Build and run

```sh
bash wasm/scripts/build.sh
cd web && npm ci && npm run dev
```

Rebuild the WASM after changing Go code. The dev server picks up the new file on reload.

`npm run build` writes the static site to `web/dist/`. Set `VERSION` to change the version shown in the top bar (default `dev`).

## Checks

Run what CI runs before you push:

```sh
cd wasm
golangci-lint run
GOOS=js GOARCH=wasm golangci-lint run
go test ./...

cd ../web
npm run check
npx vitest run
```

`web/src/lib/wasm.smoke.test.ts` loads the real `editor.wasm`, so build it first.

TALA timings can be measured with the benchmark in `wasm/bench`, see [perf.md](perf.md).

## Workflows

| Workflow | Runs on | Does |
|----------|---------|------|
| `ci.yml` | pull requests, pushes to `main`, called by release | lint, tests, WASM and web build, uploads `dist` |
| `release.yml` | `v*` tags | runs CI with the tag as version, deploys to GitHub Pages, creates a GitHub Release with the zipped site |
| `codeql.yml` | pull requests, pushes to `main`, weekly | CodeQL for Go and TypeScript |

Dependabot opens weekly grouped updates for Go modules, npm packages and GitHub Actions.

## Releasing

The live site only changes on a release.

```sh
git checkout main && git pull
git tag v0.2.0
git push origin v0.2.0
```

Use [semantic versioning](https://semver.org). The release notes are generated from the merged pull requests.

## Updating D2 and TALA

D2 and TALA come from the Go module `github.com/d2lang/d2`.

1. `cd wasm && go get github.com/d2lang/d2@<version> && go mod tidy`
2. Run all checks and try the examples in the app.
3. Update the version in the README license section. The license files in the build are copied from the module by `build.sh`.

## Repository settings

These are set once on GitHub:

- Pages: source "GitHub Actions", custom domain `d2-tala-editor.linogics.io`, HTTPS enforced.
- Environment `github-pages`: deployments only from `v*` tags.
- Ruleset on `main`: changes through pull requests, the `check` job must pass, no force pushes.
