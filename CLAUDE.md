# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & run

```bash
go build -o degunk .   # produces the ./degunk binary at repo root (gitignored)
./degunk               # interactive CLI
```

There are no tests, no lint config, and no CI. The binary is NOT checked in — it's in `.gitignore`; users build their own (or `go install github.com/felipehidra/degunk@latest`).

Requires Go (module declares `go 1.24.5`). macOS only — see architecture notes.

## Publishing model

- `main` is the public branch, pushed to github.com/felipehidra/degunk.
- `private-history` is local-only pre-release history — **never push it**.
- Repo-local git identity is set to the GitHub noreply address (`felipehidra@users.noreply.github.com`); keep it that way so no personal email lands in public commits.
- Commit messages are hook-enforced: `type: Capitalized description` with types `feat|fix|docs|chore|test`.

## Architecture

Single-file program (`main.go`). The design is intentionally flat; don't add packages/abstractions unless a change genuinely needs them.

**macOS-only by construction.** Two hard dependencies on macOS:
- `moveToTrash` runs `NSFileManager.trashItem` via `swift -e`; for root-owned system paths it falls back to `sudo mv` into `~/.Trash`, and if that also fails it prints the manual command and touches nothing. There is no `rm` fallback; this is the safety contract of the tool (items go to Trash, user empties it manually).
- The `targets` table is hardcoded with `~/Library/...` paths (Xcode, Homebrew, Safari, Slack, etc.). Porting to Linux/Windows means replacing both of these, not tweaking them.

**Scan model.** `scanTargets` does a two-level scan per target: it lists direct children of each target path, then recursively sums each child subtree via `dirSize` (a `filepath.Walk` that silently swallows permission errors — critical for running without `sudo` against `/Library/Caches`). A child that is itself another target's path is not listed twice — its dedicated target scans it. Entries below `minSize` (hardcoded 10 MB in `main`) are dropped from the candidate list but still contribute to the per-target total line printed during the scan. Results are sorted descending and capped at 30 for display. Selection happens in a bubbletea checkbox picker (requires a TTY), followed by a typed `y` confirmation.

**`CleanTarget.SafeToWipe`** flags whether a location is fully regenerable (caches, DerivedData, Trash) vs. user data that must be reviewed (Downloads, iOS backups, Xcode Archives, Docker VM data). The current UI doesn't act on this flag — it's metadata/documentation. If adding destructive flows (e.g. `--yes` / non-interactive), gate them on `SafeToWipe=true` rather than removing the distinction.

**Adding a new cleanup target** = one entry in the `targets` slice in `main.go`. Tilde paths are expanded by `expandHome`; don't pre-expand.

## README highlights worth knowing

- Interaction is always gated behind a typed `y` confirmation; the tool never deletes on its own.
- README lists extension ideas (`docker system prune`/`brew cleanup` shortcuts, `--dry-run`, JSON output) — useful context if the user asks for "a flag to …" or "an option for …".
