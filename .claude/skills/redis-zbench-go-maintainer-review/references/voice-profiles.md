# Voice profiles — real redis-zbench-go participants

Mined from the complete real GitHub history on `redis-performance/redis-zbench-go` at time of writing:
`gh pr list --state all --limit 300` (12 PRs, #1–#12), `gh api .../pulls/<n>/reviews`, `/pulls/<n>/comments`,
`/issues/<n>/comments` on each, and `gh issue view 8` (the repo's one issue). Read this alongside
`nitpick-taxonomy.md` before writing anything.

**Be honest about what this repo's history actually is, up front:** this is an extremely small record — 12
PRs and 1 issue, total. There is exactly **one** PR with a substantive, non-empty review comment (PR#6) and
exactly **one** issue with a real bug report from an outside user (issue #8). Every other PR in the sample is
either a same-day, zero-comment `APPROVED`, or has no review activity recorded at all. Do not manufacture a
richer review culture than this. When a PR doesn't resemble PR#6 or issue #8, the honest response is a short,
light-touch comment (or `skip_comment`), not an invented "maintainer voice."

## filipecosta90 / fcostaoliveira — Filipe Oliveira (the sole maintainer with any real review voice)

**Important:** `filipecosta90` (personal GitHub account, active on PRs #1–#7) and `fcostaoliveira` (Redis-org
account, active from PR#9 onward) are confirmed via the GitHub API to be the same real person, "Filipe
Oliveira." Treat them as one continuous voice, not two reviewers.

**Voice**: terse, warm, and precise when something concrete is worth naming; otherwise near-silent. The full
set of real substantive quotes found in this repo's mined history:

- PR#6 review comment (the one clearly evidenced technical catch in this repo's history), tracing a real
  concurrency bug through the exact call sequence: *"@andydunstall If I read the code correctly this will
  force all clients to use the same seed, meaning client 1 first call to `r.Int63n(int64(keyspace_len))` will
  have exactly the same value as client 2 first call to `r.Int63n(int64(keyspace_len))`. I suggest we do `r :=
  rand.New(rand.NewSource(seed+clientId))`. Agree?"* — concrete, names the exact mechanism, proposes a specific
  fix, and asks for agreement rather than mandating it.
- Same thread, after the fix: *"thank you! btw, this is exactly what I'm doing at
  https://github.com/redis-performance/openstreaming-benchmark/blob/main/cmd/producer.go#L109. Feel free to
  poke on the other tools. Some of them are outdated/need some love :)"* — a real, evidenced habit of pointing
  a contributor at a sibling redis-performance-org project as a concrete reference, and inviting further
  contribution warmly.
- Merge comment on PR#6: *"Thank you @andydunstall . Merging... :) EDIT: released as part of
  https://github.com/redis-performance/redis-zbench-go/releases/tag/v0.0.3"* — thanks by name, notes the
  release tag it shipped in.
- Issue #8 reply to a non-collaborator bug report: *"Thank you for noticing this @romange . will address it
  shortly."* **Be honest about the outcome, not just the words**: as of this mining, years later, the bug
  described in issue #8 is still present in `redis-zbench-go.go` on `main` — a maintainer saying "will address
  it shortly" here is not evidence a fix actually landed. Don't cite this comment as if the issue were
  resolved.
- Thanks contributors by name in these real quotes (`@andydunstall`, `@romange`) — a human habit; **do not**
  have the bot literally `@`-mention anyone (see SKILL.md).

**What this means for the bot's voice**: when something is worth commenting on at all, name the specific
mechanism (which shared state, which flag/string mismatch) the way the PR#6 comment does — not a generic
"looks good" or a generic "consider adding tests." When nothing stands out, silence or a short line is
authentic; there is no real example in this repo's history of a long, itemized review.

## paulorsousa — the only other reviewer with any recorded approvals

**Voice**: in the mined sample, `paulorsousa` approved three PRs — #9 (adding `CONTRIBUTING.md`/`AGENTS.md`),
#11 (bumping GitHub Actions to node24-compatible versions), and #12 (pinning the CI Redis image to `8.6`) —
and left an entirely empty review body on all three. There is **no** written comment from `paulorsousa`
anywhere in this repo's mined history (unlike, e.g., a sibling project where a short warm line exists on
record). Be honest that this skill has no `paulorsousa` voice profile beyond "approves routine/CI PRs with no
comment" — do not extrapolate a richer style than that; there's nothing in the record to extrapolate from.

## andydunstall — one-time external contributor (PR#6)

The author on the one PR with a substantive review exchange. Responded to the shared-RNG catch collegially and
without defensiveness: *"ah yep - sure will update"*, then after merge: *"Great thanks for reviewing :)"*. Not
a maintainer, but worth noting as the one real example in this repo's history of an external contributor
accepting a concrete correctness catch and shipping the fix in the same PR, same day.

## romange — external issue reporter (issue #8), not a collaborator

Filed the repo's one and only issue: a precise, technical, two-sentence bug report citing an exact source line
(`redis-zbench-go.go#L137` at the time) and naming two distinct problems tersely — a flag-default/switch-case
mismatch, and a missing error path on an unmatched switch. No back-and-forth beyond the maintainer's one-line
acknowledgment; the issue remains open. Useful as the one real example of what a good, actionable external bug
report on this repo looks like, if a PR under review claims to fix a similarly-shaped issue.

## No automated tooling does load-bearing review work here

Unlike a project with GitHub Advanced Security / CodeQL and a Copilot review bot wired up, `redis-zbench-go`'s
CI (`.github/workflows/test.yml`) runs only `make test` (`gofmt` + `go test -race -covermode=atomic ./...`)
across a Go 1.18–1.21 matrix, and `.github/workflows/publish.yml` handles release binaries. There is no lint
step in CI (the Makefile's `lint` target is never invoked by a workflow), no static-analysis bot, and no
coverage-percentage bot commenting on PRs. Whatever a human (or this skill) catches by reading the diff is the
only review this repo's code actually gets before merge — don't imply otherwise, and don't assume an automated
tool already caught something a manual read would otherwise flag.
