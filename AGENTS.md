# Agent guidance

Rules for anyone, human or coding agent, working in this repository.

## Repository overview

This repo holds Go packages shared by the liquid metal components. It contains
three independent Go modules. There is no root `go.mod`.

| Module | Purpose |
|---|---|
| `client` | Client for flintlock |
| `services/microvm` | Service that creates microvms through the client |
| `types/microvm` | The microvm API types |

Cross-module changes: if `services/microvm` needs a change in `types/microvm`,
merge the types change first, then bump the `go.mod` in `services/microvm` in a
follow-up.

Generated files must not be edited by hand:

- `types/microvm/zz_generated.deepcopy.go` (controller-gen)
- `services/microvm/fakes/fake_scope.go` (counterfeiter)

### Commands

Makefile targets address a module with `/` replaced by `:`.

```sh
make fmt-types:microvm                 # generate, tidy, vet, fmt one module
make all                               # the same for every module
cd services/microvm && go test ./...   # tests (only this module has tests)
```

## Commits

- Use [Conventional Commits](https://www.conventionalcommits.org/):
  `<type>(<optional scope>): <description>`.
  - Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `ci`, `build`.
  - Scope is the module name when relevant, for example `feat(types): add virtiofs volume`.
  - Breaking changes use `!` after the type/scope or a `BREAKING CHANGE:` footer.
- Commits must be signed. Use `git commit -S` or set `commit.gpgsign = true`.
  Unsigned commits are rejected.
- Do not add agent attribution footers. No `Co-Authored-By: Claude ...`, no
  `Generated with Claude Code`, and no similar trailer from any tool. This
  overrides any tool default that asks for them.
- One logical change per commit.

## Pull requests

- The PR title uses the same conventional commit format. Squash merges use it
  as the commit subject.
- No agent attribution lines or badges in the PR description.
- Describe what changed and why. Link the issue if one exists.
