# Articles Coverage Report

This repo implements `go-task-3` (Articles module). To generate a coverage report:

```bash
go test ./internal/features/articles/... ./tests/articles/... -coverprofile=coverage_articles.out -covermode=atomic -v
```

Then print the summary:

```bash
go tool cover -func=coverage_articles.out
```

To extract the total line only:

```bash
go tool cover -func=coverage_articles.out | grep total
```

On Windows PowerShell:

```powershell
go tool cover -func=coverage_articles.out | Select-String -Pattern "total"
```

Target: `>= 90%` total coverage for the Articles module.
