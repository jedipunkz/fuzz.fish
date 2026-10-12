## Summary

<!-- Describe what this PR changes and why. -->

## Changes

- <!-- bullet per change -->

## Design Checklist

- [ ] No new per-keystroke allocations in `internal/app/filter.go`.
- [ ] `exec.Command` uses argument lists, never `sh -c` with external input.
- [ ] Package dependency direction is preserved.
- [ ] Tests added or updated.
