# Working in this repository

This repo extends the Aspect CLI with custom commands written in AXL (Starlark).
They live in `.aspect/*.axl` and are real subcommands, not scripts.

## Find out what exists

```shell
aspect describe
```

Lists every command available here as JSON, including the ones this repository
defines, each with its flags, their types, their defaults and their allowed
values. Use it instead of guessing or reading source. For one command:

```shell
aspect describe impact
```

`defined_in` names the file to read if you need the implementation.

## Useful commands here

| Command | Answers |
|---|---|
| `aspect changed` | which files the working tree has touched, including uncommitted and untracked |
| `aspect impact` | which tests the working tree could break, before anything is committed |
| `aspect policy` | whether the change violates a declared architecture rule; exits non-zero if so |
| `aspect observe` | runs the tests and reports what Bazel's event stream says about each one |

`changed` and `impact` take `--format=json` and write one document to
**stdout**; progress and the task banner go to **stderr**, so stdout pipes
without filtering.

```shell
aspect impact --format=json | jq -r '.tests[]'
```

## Before you hand work back

```shell
aspect impact        # the tests your change affects
aspect policy        # exits 1 on a layering violation
```

`aspect impact` reads your dirty tree, so it works mid-change. It is advisory:
it maps changed files to their Bazel packages, so a change that affects the
build without changing a package — a `.bzl` file, `MODULE.bazel` — is invisible
to it. Run the full suite before you claim a change is safe.

`internal/domain/forecast` contains a deliberately flaky test. A single red run
there is expected and is not your change.
