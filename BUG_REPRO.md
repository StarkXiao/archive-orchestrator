# Bug Reproduction

## Baseline

This workspace is compared with remote branch `base_bug_003_green`.

## Reproduction

Run:

```text
go test ./internal/application -count=1 -run '^TestBug003_ListRulesDoesNotMutateStoredNames$'
```

## Expected Result

Formatting rules for an API response must not mutate names stored in the repository.

## Observed Result

The returned rule slice aliases repository state, and response processing changes the persisted rule name, so the test fails consistently.
