# AI Writing Coverage Report

This repo implements `go-task-4` (AI Writing module), including provider implementations for:

- `openai` (`data.OpenAIClient`)
- `claude` (`data.ClaudeClient`)
- `gemini` (`data.GeminiClient`)

## Run tests + generate coverage profile

```bash
go test ./internal/features/ai_writing/... ./tests/ai_writing/... -coverprofile=coverage_ai_writing.out -covermode=atomic -v
```

## Print coverage summary

```bash
go tool cover -func=coverage_ai_writing.out
```

Extract the total line:

```bash
go tool cover -func=coverage_ai_writing.out | grep total
```

On Windows PowerShell:

```powershell
go tool cover -func=coverage_ai_writing.out | Select-String -Pattern "total"
```

## Target

- `>= 90%` total coverage for the AI Writing module.

## Optional: write a one-line markdown summary

```powershell
$line = (go tool cover -func=coverage_ai_writing.out | Select-String -Pattern "total").ToString()
Set-Content -Encoding UTF8 -Path tests/ai_writing/coverage_total.md -Value ("Total coverage: `"$line`"")
```
