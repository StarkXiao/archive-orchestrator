# Bug Reproduction

## Baseline

This workspace is compared with remote branch `base_bug_001_green`.

## Reproduction

Run:

```text
go test ./internal/infrastructure -count=1 -run '^TestBug001_RenewLeaseRejectsForeignWorker$'
```

## Expected Result

A worker must not renew a lease owned by another worker, and the lease owner must remain unchanged.

## Observed Result

The renewal succeeds for a foreign worker and changes lease state, so the test fails consistently.
