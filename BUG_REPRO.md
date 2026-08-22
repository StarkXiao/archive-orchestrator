# Bug Reproduction

## Baseline

This workspace is compared with remote branch `base_bug_002_green`.

## Reproduction

Run:

```text
go test ./internal/worker -count=1 -run '^TestBug002_RetryWaitWithoutTimestampIsClaimable$'
```

## Expected Result

A retry-wait job without a retry timestamp must be handled without a panic and remain claimable according to its state.

## Observed Result

The worker dereferences the missing retry timestamp and panics, so the test fails consistently.
