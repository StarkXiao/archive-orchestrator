# Bug Reproduction

## Baseline

This workspace is compared with remote branch `base_bug_004_green`.

## Reproduction

Run:

```text
go test ./internal/application -count=1 -run '^TestBug004_CreatePreservesInvalidRuleError$'
```

## Expected Result

The application boundary must preserve the invalid-rule error chain so callers can identify it with `errors.Is` or `errors.As`.

## Observed Result

The error is converted to plain text and loses its identity across the application boundary, so the test fails consistently.
