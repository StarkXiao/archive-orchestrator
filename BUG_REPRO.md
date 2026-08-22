# Bug Reproduction

## Baseline

This workspace is compared with remote branch `base_bug_006_green`.

## Reproduction

Run:

```text
go test ./internal/application -count=1 -run '^TestBug006_RestoreRejectsSiblingOfConfiguredRoot$'
```

## Expected Result

A restore target outside the configured root, including a sibling path with a shared prefix, must be rejected.

## Observed Result

The green baseline performs a path-boundary check, rejects the sibling path, and the test passes consistently.
