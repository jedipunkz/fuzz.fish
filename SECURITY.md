# Security Policy

## Supported Versions

Only the latest release receives security fixes. Binaries are published for
macOS and Linux (amd64 / arm64) as tarballs plus SHA-256 checksums on the
GitHub Releases page.

| Version           | Supported |
| ----------------- | --------- |
| latest release    | yes       |
| older releases    | no        |

We recommend installing with [fisher](https://github.com/jorgebucaran/fisher),
or rebuilding the binary from source with `_fuzz_fish_rebuild_binary`, so
updates arrive with the plugin itself.

## Reporting a Vulnerability

Use [private vulnerability reporting](https://github.com/jedipunkz/fuzz.fish/security/advisories/new)
so details stay confidential until a fix is released.

- What to include: affected version, reproducible steps or a minimal PoC, and
  the impact you observed.
- You can expect an initial response within 7 days, and status updates at
  least every 14 days while a report is being triaged or fixed.
- Accepted reports are fixed and disclosed through a GitHub Security Advisory
  with coordinated disclosure. Reporters are credited in the advisory unless
  they prefer to stay anonymous.
- Reports judged not to be vulnerabilities are declined with an explanation.

## Scope

The project consists of a Go binary (`cmd/`, `internal/`) and the Fish shell
integration script (`conf.d/fuzz.fish`). Both are in scope, in particular:

- The Fish integration script, including the `FUZZ_FISH_BIN_PATH` handling
  (binary install and `rm -f` paths) and the `git switch` invocation for
  branch selection.
- The install/update path that clones the canonical GitHub repository
  (`_fuzz_fish_rebuild_binary`) and the `git pull` follow-up on the selected
  branch.
- Anything that writes outside the user's own `~/.config/fish`,
  `~/.local/share/fish/fish_history` (read-only in this project), or the
  configured `worktree_dir`, or that lets a remote branch name reach a shell
  unsafely.
- The integrity of the reachable code paths when release artifacts are
  tampered with downstream; users should install against the published
  SHA-256 checksums.

Out of scope:

- Vulnerabilities in Fish shell, `gh`, `git`, or Go's standard library.
  Report these to the respective upstream projects.
- Vulnerabilities in third-party Go dependencies; report them upstream, but
  also let us know if the dependency is mandatory for our use case so we can
  switch to a safer alternative or version.
- Attacks that require phishing or social engineering.
- Attacks that need the user to run the binary as root or with a mutually
  shared, already-writable `FUZZ_FISH_BIN_PATH` (partly documented in
  CLAUDE.md under Security as a known multi-user limitation).
