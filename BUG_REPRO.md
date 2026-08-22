# Bug Reproduction

## Baseline

This workspace is compared with remote branch `base_bug_007_green`.

## Reproduction

Run:

```text
go test ./internal/worker -count=1 -run '^TestBug007_UnexpiredOrMissingLeaseIsNotRequeued$'
```

## Expected Result

Only jobs with an expired lease may be requeued; jobs with an active or missing lease must remain unchanged.

## Observed Result

The green baseline filters lease state correctly and the test passes consistently.
