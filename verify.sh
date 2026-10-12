#!/usr/bin/env bash
# Checks that the aspect-160 course pages still describe this repository.
#
# Every assertion here corresponds to a number, a transcript or a claim printed
# on https://aspect.build/learning/aspect-160. The pages quote exact counts, so
# a CLI upgrade or an edit to a step branch can silently make them wrong; this
# is what catches that. Run it before a workshop, and after bumping the version
# pin in .bazelversion or .aspect/config.axl.
#
# Usage:
#   ./verify.sh            # everything (needs Bazel, ~5 minutes warm)
#   ./verify.sh --quick    # structure only (no Bazel, seconds)
#
# Run it from a checkout of this repository. It never touches your working
# tree: each step is checked out into its own throwaway worktree under
# .verify-wt/, which is removed on exit.

set -uo pipefail

QUICK=0
[[ "${1:-}" == "--quick" ]] && QUICK=1

REPO="$(git rev-parse --show-toplevel)"
WT="$REPO/.verify-wt"
PASS=0
FAIL=0
FAILED_NAMES=()

cleanup() {
  cd "$REPO" || return
  for d in "$WT"/*; do
    [[ -d "$d" ]] && git worktree remove --force "$d" >/dev/null 2>&1
  done
  rmdir "$WT" >/dev/null 2>&1
  return 0
}
trap cleanup EXIT

# ok <name> <expected> <actual>
ok() {
  if [[ "$2" == "$3" ]]; then
    printf '  \033[32m✓\033[0m %s\n' "$1"
    PASS=$((PASS + 1))
  else
    printf '  \033[31m✗\033[0m %s\n      expected: %s\n      actual:   %s\n' "$1" "$2" "$3"
    FAIL=$((FAIL + 1))
    FAILED_NAMES+=("$1")
  fi
}

# contains <name> <needle> <haystack>
contains() {
  if [[ "$3" == *"$2"* ]]; then
    printf '  \033[32m✓\033[0m %s\n' "$1"
    PASS=$((PASS + 1))
  else
    printf '  \033[31m✗\033[0m %s\n      looked for: %s\n      in:\n%s\n' "$1" "$2" "$(sed 's/^/        /' <<<"$3")"
    FAIL=$((FAIL + 1))
    FAILED_NAMES+=("$1")
  fi
}

# Prepares a worktree for a step branch and echoes its path.
step() {
  local branch="$1" dir="$WT/$1"
  mkdir -p "$WT"
  git worktree add --force --detach "$dir" "$branch" >/dev/null 2>&1 || {
    echo "could not create a worktree for $branch" >&2
    exit 2
  }
  echo "$dir"
}

# A missing step branch would otherwise read as a page quoting a stale number,
# which is the opposite of what it means.
missing=()
for b in main step-0 step-1 step-2 step-3 step-4 step-5 step-6 step-7 step-8; do
  git rev-parse --verify --quiet "$b^{commit}" >/dev/null || missing+=("$b")
done
if (( ${#missing[@]} )); then
  cat >&2 <<MSG
This needs a local branch per step, and a fresh clone has only the one you are
standing on. Missing: ${missing[*]}

  for b in \$(git branch -r --format='%(refname:lstrip=3)' | grep '^step-'); do
    git branch --force "\$b" "origin/\$b"
  done
  git branch --force main origin/main
MSG
  exit 2
fi

echo
echo "== Structure: what the index page claims"

# "changed.axl is 49 lines, impact.axl 85, policy.axl 62, observe.axl 56",
# config.axl 27, and "the Gazelle extension ... is the longest at 128".
while read -r branch path want; do
  got=$(git show "$branch:$path" | wc -l | tr -d ' ')
  ok "$path on $branch is $want lines" "$want" "$got"
done <<'EOF'
step-2 .aspect/changed.axl 49
step-3 .aspect/impact.axl 85
step-4 .aspect/policy.axl 62
step-5 .aspect/observe.axl 56
step-6 .aspect/config.axl 27
step-7 tools/gazelle/service_image.axl 128
EOF

# "Every step's diff is one file, with three exceptions: the step that drops a
# whole Go monorepo in, the Gazelle step, and the policy step."
for n in 1 2 3 4 5 6 7; do
  m=$((n + 1))
  count=$(git diff --name-only "step-$n" "step-$m" | wc -l | tr -d ' ')
  case "$n" in
    3) want=2 ;;  # policy.axl + the patch its exercise applies
    6) want=11 ;; # the Gazelle step generates BUILD files as its output
    *) want=1 ;;
  esac
  ok "step-$n → step-$m diff is $want file(s)" "$want" "$count"
done

ok "main is step-8" "$(git rev-parse step-8)" "$(git rev-parse main)"

if [[ $QUICK == 1 ]]; then
  echo
  echo "$PASS passed, $FAIL failed (structure only)"
  [[ $FAIL == 0 ]] || { printf 'failed: %s\n' "${FAILED_NAMES[*]}"; exit 1; }
  exit 0
fi

echo
echo "== step-0: a project with one file"
d=$(step step-0); cd "$d"
ok "aspect describe lists 33 commands" "33" "$(aspect describe 2>/dev/null | jq '.commands | length')"

echo
echo "== step-1: the repository that arrives"
d=$(step step-1); cd "$d"
ok "55 go_library targets" "55" "$(bazel query 'kind(go_library, //...)' 2>/dev/null | wc -l | tr -d ' ')"
ok "55 go_test targets"    "55" "$(bazel query 'kind(go_test, //...)' 2>/dev/null | wc -l | tr -d ' ')"
ok "3 binaries"             "3" "$(bazel query 'kind(go_binary, //cmd/...)' 2>/dev/null | wc -l | tr -d ' ')"
ok "13 lib/ packages"      "13" "$(bazel query 'kind(go_library, //lib/...)' 2>/dev/null | wc -l | tr -d ' ')"
ok "39 domain packages"    "39" "$(bazel query 'kind(go_library, //internal/domain/...)' 2>/dev/null | wc -l | tr -d ' ')"

echo
echo "== step-2: your first command"
d=$(step step-2); cd "$d"
ok "a clean tree reports nothing" "0" "$(aspect changed 2>/dev/null | wc -l | tr -d ' ')"
printf '\n// a change\n' >> lib/log/log.go
ok "one edited file, one line of stdout" "lib/log/log.go" "$(aspect changed 2>/dev/null)"
aspect changed --format=json 2>/dev/null | jq -e . >/dev/null
ok "--format=json is valid JSON on stdout alone" "0" "$?"
aspect changed --format=bogus >/dev/null 2>&1
ok "--format=bogus exits 2, before any Starlark runs" "2" "$?"
ok "--base=HEAD~2 reports 256 files" "256" "$(aspect changed --base=HEAD~2 2>/dev/null | wc -l | tr -d ' ')"
ok "describe names both args" "base format" "$(aspect describe changed 2>/dev/null | jq -r '[.args[].name] | join(" ")')"

echo
echo "== step-3: what does my tree affect"
d=$(step step-3); cd "$d"
printf '\n// a change\n' >> lib/log/log.go
contains "a change to lib/log affects 33 tests" "33 affected tests" "$(aspect impact 2>&1)"
git checkout -- lib/log/log.go
printf '\n// a change\n' >> internal/domain/geo/model.go
contains "a change to internal/domain/geo affects 4" "4 affected tests" "$(aspect impact 2>&1)"
git checkout -- internal/domain/geo/model.go
echo 'package geo' > internal/domain/geo/newfile.go
contains "an untracked file is seen, and over-reports its package" "4 affected tests" "$(aspect impact 2>&1)"
rm -f internal/domain/geo/newfile.go

echo
echo "== step-4: a rule you can enforce"
d=$(step step-4); cd "$d"
out=$(aspect policy 2>&1); rc=$?
contains "a clean repository passes 3 rules" "3 rules checked, no violations" "$out"
ok "and exits 0" "0" "$rc"
git apply .course/exercises/policy-violation.patch
out=$(aspect policy 2>&1); rc=$?
ok "a violation exits 1" "1" "$rc"
contains "and names the rule it broke" "foundation-libs-stay-leaf" "$out"
contains "and both offending targets" "//lib/validate:validate_test" "$out"
git checkout -- .

echo
echo "== step-5: watch Bazel work"
d=$(step step-5); cd "$d"
out=$(aspect observe //lib/queue:queue_test //lib/log:log_test //lib/retry:retry_test 2>&1)
ok "three targets give three result lines" "3" "$(grep -c 'PASSED\|FAILED\|FLAKY' <<<"$out")"
contains "and the output comes after the build, not during" "queue_test" "$out"

echo
echo "== step-6: you won't ship that"
d=$(step step-6); cd "$d"
out=$(aspect test //lib/log:log_test --announce-bazel-command=true 2>&1)
contains "config.axl puts --flaky_test_attempts=3 on the Bazel command" "--flaky_test_attempts=3" "$out"
out=$(aspect test //lib/log:log_test 2>&1)
contains "a passing test is not flagged" "Passed test task" "$out"

echo
echo "== step-7: generate it"
d=$(step step-7); cd "$d"
git restore --source=step-6 -- cmd/*/BUILD.bazel
contains "gazelle regenerates 3 BUILD files" "Updated 3 BUILD files" "$(aspect gazelle 2>&1)"
ok "byte-identical to what the step committed" "" "$(git diff --name-only -- cmd/)"
contains "and is idempotent" "up to date" "$(aspect gazelle 2>&1)"
sed -i'' -e 's/port: 8080/port: 9090/' cmd/server/service.yaml
aspect gazelle >/dev/null 2>&1
contains "MergeableAttrs moves exposed_ports when port: changes" '9090/tcp' "$(cat cmd/server/BUILD.bazel)"
git checkout -- .

echo
echo "== step-8: hand it to an agent"
d=$(step step-8); cd "$d"
ok "aspect describe lists 37 commands: 33 shipped, 4 yours" "37" "$(aspect describe 2>/dev/null | jq '.commands | length')"
ok "and impact's own flags are in there" "base format" "$(aspect describe impact 2>/dev/null | jq -r '[.args[].name] | join(" ")')"

echo
echo "$PASS passed, $FAIL failed"
if [[ $FAIL != 0 ]]; then
  printf 'failed:\n'
  printf '  - %s\n' "${FAILED_NAMES[@]}"
  exit 1
fi
