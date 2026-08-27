# Cross-cutting nitpick taxonomy — redis-zbench-go, real precedent only

Grounded in this repo's complete mined history at time of writing:
`gh pr list --repo redis-performance/redis-zbench-go --state all --limit 300` (12 PRs, #1–#12, 10 merged/1
open/1 the PR adding this skill's own workflow), `gh api .../pulls/<n>/reviews`, `/pulls/<n>/comments`, and
`/issues/<n>/comments` on each, `gh issue list --state all` (1 issue, #8, still open), and this repo's own
`AGENTS.md`/`CONTRIBUTING.md`. This is a genuinely small record — every category below is evidenced by **one**
real occurrence, not a repeated pattern. Say so plainly when citing them; do not imply repetition that isn't
there.

1. **Shared/reused randomness across per-client goroutines is this repo's single sharpest real bug, caught in
   review.** PR#6 (`andydunstall`, external contributor): the tool spawns one goroutine per simulated client,
   and each drew from what was effectively the same `rand` source/seed. filipecosta90's real review comment,
   traced exactly: *"If I read the code correctly this will force all clients to use the same seed, meaning
   client 1 first call to `r.Int63n(int64(keyspace_len))` will have exactly the same value as client 2 first
   call... I suggest we do `r := rand.New(rand.NewSource(seed+clientId))`. Agree?"* `andydunstall` agreed and
   fixed it in the same PR (merged as "remove global rand. Updating to a rand source per goroutine to avoid
   locking overhead"). Real precedent: any new code path that spawns per-client/per-goroutine concurrency and
   needs independent randomness, key sampling, or other per-client state must be checked for exactly this —
   is the source of that state created once and shared (all goroutines draw the same sequence), or created
   fresh per goroutine with something that actually distinguishes clients (a seed offset, a client index)?
   The merged fix pattern in this codebase is `*seed + int64(client_id)` passed into a fresh
   `rand.New(rand.NewSource(...))` inside each goroutine — a new PR reintroducing a shared/package-level
   source for anything client-facing should be flagged the same way.

2. **A CLI flag's documented default value must literally match what the consuming code expects it to equal —
   real, still-unfixed bug.** Issue #8 (`romange`, non-collaborator, filed 2023, still open at time of
   mining): the `--query` flag defaults to the string `"zrangebyscore"`, but the `switch *query` block that
   consumes it only matches `"zrange-byscore-rev"`, `"zrange-byscore"`, and `"zrevrangebylex"` — none of which
   equal the default. Confirmed still present in `redis-zbench-go.go` as of this mining (flag default and
   switch cases were re-checked directly against the file on `main`). filipecosta90's real reply: *"Thank you
   for noticing this @romange . will address it shortly"* — and it was not, in fact, addressed within the
   years since. Be honest about that if it comes up: a maintainer saying "will address it shortly" here is not
   evidence the fix actually landed promptly, or at all — check the current source, don't take the comment as
   confirmation. Real precedent for review: whenever a PR changes a flag's default value, or renames/adds a
   case string in a switch/if-else the flag feeds into, verify the two are exactly equal strings, not just
   "plausibly the same concept" (`zrangebyscore` vs. `zrange-byscore` is exactly this kind of near-miss).

3. **String-dispatch switch/if-else chains with no default/error branch fail silently — the second half of the
   same real, evidenced report.** Issue #8, verbatim: *"also, there is no error if the switch misses all the
   options."* Confirmed still true of the current `switch *query { ... }` block on `main` — no `default:` case,
   so an unrecognized value causes the tool to silently submit zero work for that client rather than erroring.
   Real precedent: any new or modified switch/if-else dispatching on a string CLI flag should have an explicit
   default/else that surfaces a clear error for an unrecognized value, not a silent no-op — this project has a
   real, on-record example of exactly that gap going unfixed.

4. **Bare, comment-free `APPROVED` is the norm for routine/CI/chore PRs.** Real, repeated evidence: PR#9 (adding
   `CONTRIBUTING.md`/`AGENTS.md` themselves), PR#11 (bumping GitHub Actions to node24-compatible versions), and
   PR#12 (pinning the CI Redis service image to `8.6`) were all approved by `paulorsousa` with an entirely empty
   review body. Don't manufacture a substantive comment on this class of PR to seem thorough; matching silence
   (or a one-line approval) is the authentic, evidenced behavior for CI/chore/docs-only changes here.

5. **Test coverage is written doctrine, but currently has zero enforcement — more starkly than a "low but
   nonzero" gap.** `CONTRIBUTING.md`: *"All new behaviour must be covered by tests... Coverage should not
   decrease."* `AGENTS.md` repeats "Always run tests before declaring a task complete." But at time of mining,
   **this repository contains no `_test.go` files at all** (confirmed via a full recursive tree listing) —
   `.github/workflows/test.yml`'s `go test -race -covermode=atomic ./...`, run on every push/PR across a Go
   1.18–1.21 matrix, currently exercises zero test cases. Cite the written rule as real if relevant, but do not
   claim or imply it functions as any kind of gate today, and do not flag "no new tests" on an ordinary PR as
   unusual — it would match every PR in this repo's history so far. If a PR under review does add genuine test
   files, that is a real first against this baseline and deserves specific, positive credit.

6. **No CodeQL, no Copilot review bot, no lint step actually runs in CI.** `.github/workflows/` here contains
   only `test.yml` (the Go 1.18–1.21 `make test` matrix) and `publish.yml` (release binaries via
   `wangyoucao577/go-release-action`, updated in PR#7). The `Makefile` defines a `lint` target
   (`golangci-lint run`), but no workflow ever invokes it. Unlike a project with GitHub Advanced Security wired
   up, there is no automated tool here catching import issues, unclosed resources, or unreachable branches
   before a human reads the diff — say so plainly rather than assuming or implying such tooling exists.

7. **"Don't add a new dependency without checking with the maintainer" is written doctrine with no evidenced
   enforcement example.** `AGENTS.md`: *"Do not introduce new dependencies without checking with the
   maintainer."* No mined PR in this repo's history was rejected, commented on, or delayed over adding a
   dependency — `go.mod` itself is short (HdrHistogram, `mediocregopher/radix/v3`, `golang.org/x/time`, plus a
   few indirect release-tooling deps). Cite the rule as real written doctrine if a PR adds a new import, but
   don't claim a citable precedent for how strictly it's actually enforced — there isn't one in the record.

8. **`filipecosta90` and `fcostaoliveira` are the same maintainer, not two reviewers.** GitHub API confirms
   distinct account IDs but the same real name, "Filipe Oliveira" — the personal account authored/reviewed
   PRs #1–#7, the Redis-org account PR#9 onward. He is the only person in this repo's mined history who has
   ever left a substantive (non-empty) review comment. Do not treat these as separate voices with separate
   standards.

## What this taxonomy is honestly thin or silent on

- **No long-form, multi-point, dialectic review exists in this repo's history at all.** PR#6 is a single
  comment → single fix → approval, not a numbered multi-point essay. If you're tempted to write a long,
  structured review because "that's what a thorough maintainer review looks like," recognize that no real
  example of that exists here to imitate — keep it as terse as the real record actually is.
- **No performance/benchmark-methodology critique exists in the mined review history**, despite this being a
  benchmarking tool — nobody has left a review comment questioning HDR histogram usage, pipelining semantics,
  RPS-limiter behavior, or latency-measurement correctness. If a PR touches that surface, reason about it from
  Go/benchmarking first principles; there is no citable precedent from this repo's own reviewers to lean on.
- **No security-sensitive code path has been the subject of any real review comment** (no auth, no untrusted
  network input parsing beyond CLI flags and Redis responses) — don't invent a "security review culture" for
  this repo.
- **No style/whitespace nitpicking exists in the record at all** — `gofmt` runs automatically in `make test`,
  so there's nothing left for a human to catch manually, and no mined comment ever raised one.
- **Cluster-mode/slot-routing logic** (`cluster_conn.go`, `crc16_slottable.go`) has never drawn a real review
  comment in the mined sample (PR#3's "Fixed cross-slot issue" and PR#4's "Added OSS cluster API support" were
  both self-merged with no review body found) — treat this as unreviewed-in-practice surface, not an area with
  real institutional scrutiny behind it, if a PR touches it.
