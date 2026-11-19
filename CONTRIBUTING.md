# Contributing

Thanks for helping. Bug reports, ideas and pull requests are welcome.

## Before you start

For larger changes, open an issue first so we can agree on the approach.

## Setup

See [docs/development.md](docs/development.md) for build, run and checks, and [docs/architecture.md](docs/architecture.md) for how the code is organized.

## Pull requests

- Keep a pull request focused on one change.
- Add or update tests for new behavior. Go edits get a test in `wasm/editor`, UI logic gets a Vitest test in `web/src/lib`.
- Run all checks from the development guide. CI runs the same ones.
- Use [conventional commits](https://www.conventionalcommits.org) for commit messages and the pull request title, for example `feat(web): snap to grid` or `fix(wasm): keep comments on rename`.
- Update the docs when you change what the user sees.

## License

By contributing you agree that your contribution is licensed under the [BSD-3-Clause license](LICENSE) of this project.
