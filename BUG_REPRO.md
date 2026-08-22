# Bug Reproduction

## Baseline

This workspace is compared with remote branch `base_bug_009_green`.

## Reproduction

Run:

```text
go test ./internal/application -count=1 -run '^TestBug009_InvalidPatternErrorSurvivesApplicationBoundary$'
```

## Expected Result

An invalid pattern error must retain its error chain when it crosses the application boundary.

## Observed Result

The green baseline wraps the underlying error without discarding its identity, and the test passes consistently.
