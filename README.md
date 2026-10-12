# Course verification

This branch is not part of the course. It carries one script, `verify.sh`,
which checks that the [aspect-160](https://aspect.build/learning/aspect-160)
pages still describe this repository.

It lives on its own branch on purpose. `main` is `step-8`, so anything added
there would show up in the students' final checkout and in
`git diff step-7 step-8` — a diff the course advertises as one file.

## Why

The pages quote exact numbers: 33 commands at `step-0` and 37 at `step-8`, 33
affected tests for a change to `lib/log`, 256 files for `--base=HEAD~2`, five
`.axl` line counts. Those are the first thing to rot when the pinned CLI
version moves or a step branch is edited, and nothing about a stale number
looks broken until a student reads it aloud. Each assertion below corresponds
to something a page claims.

## Running it

```shell
git checkout course-verify
./verify.sh --quick    # structure only, no Bazel, seconds
./verify.sh            # everything, needs Bazel, a few minutes warm
```

It checks each step out into a throwaway worktree under `.verify-wt/` and
removes them on exit, so it never touches your working tree.

It needs a local branch per step, and a fresh clone has only the one you are
standing on. It says so and exits 2 if they are missing:

```shell
for b in $(git branch -r --format='%(refname:lstrip=3)' | grep '^step-'); do
  git branch --force "$b" "origin/$b"
done
git branch --force main origin/main
```

## What it does not check

- The flaky test in `internal/domain/forecast` is random, so `step-5`'s
  `failed` result and `step-6`'s `FLAKE CAUGHT` banner are not asserted. What
  is asserted is the mechanism: that `config.axl` puts
  `--flaky_test_attempts=3` on the Bazel command line.
- `aspect delivery` pushes to ttl.sh and needs the network.
- Prose. It checks numbers and transcripts, not whether a sentence is true.

## When it fails

A failure names what it expected and what it got. Decide which one is wrong:
if the repository changed on purpose, the page needs an edit; if it did not,
something upstream moved.
