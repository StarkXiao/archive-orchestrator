# Bug Reproduction

## Baseline

This workspace is compared with remote branch `base_bug_010_green`.

## Reproduction

Run:

```text
go test ./internal/infrastructure -count=1 -run '^TestBug010_CopyFileKeepsWrittenDestination$'
```

## Expected Result

A successful file copy must keep the completed destination file while still cleaning up temporary data on failures.

## Observed Result

The green baseline limits cleanup to failure paths, keeps the copied destination, and the test passes consistently.
