# Bug Reproduction

## Baseline

This workspace is compared with remote branch `base_bug_008_green`.

## Reproduction

Run:

```text
go test ./internal/infrastructure -count=1 -run '^TestBug008_RestoreDoesNotCommitChecksumMismatch$'
```

## Expected Result

A restored file with a checksum mismatch must not replace the destination file.

## Observed Result

The green baseline validates the checksum before committing the restored file, leaves the destination unchanged, and the test passes consistently.
