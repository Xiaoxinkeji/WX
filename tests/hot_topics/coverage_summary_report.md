
# Hot Topics Coverage Report

This repo implements `go-task-2` (Hot Topics module). To generate a coverage report:

```bash
go test ./internal/features/hot_topics/... ./tests/hot_topics/... -coverprofile=coverage_hot_topics.out -covermode=atomic -v
```

Then print the summary:

```bash
go tool cover -func=coverage_hot_topics.out
```

To extract the total line only:

```bash
go tool cover -func=coverage_hot_topics.out | grep total
```

On Windows PowerShell:

```powershell
go tool cover -func=coverage_hot_topics.out | Select-String -Pattern "total"
```

Target: `>= 90%` total coverage for the Hot Topics module.
