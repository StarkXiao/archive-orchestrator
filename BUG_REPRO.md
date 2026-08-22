# Bug Reproduction

## Baseline

This workspace is compared with remote branch `base_bug_005_green`.

## Reproduction

Run:

```text
go test ./internal/application -count=1 -run '^TestBug005_CancelledArchiveDoesNotCreateBatch$'
```

## Expected Result

A cancelled archive request must propagate cancellation and must not create a batch.

## Observed Result

The green baseline preserves the caller context, creates no batch after cancellation, and the test passes consistently.
