# Error Handling Patterns

## Custom Error Types (in `pkg/errors/errors.go`)

```go
type ConfigError struct {
    Message string
    Path    string
    Cause   error
}

type AnalysisError struct {
    Message string
    File    string
    Cause   error
}

type ReportError struct {
    Message string
    Path    string
    Cause   error
}
```

## Error Handling Best Practices

1. **Wrap errors with context**: Use `%w` for wrapping, always preserve chain
2. **Use custom error types**: For domain-specific errors (ConfigError, AnalysisError)
3. **Provide actionable messages**: Include file paths, config names
4. **Log errors before returning**: Use structured logging with logger
5. **Recover from errors**: Try multiple strategies (e.g., JSON version parsing → text fallback)

Example from `pkg/linter/analyzer.go`:

```go
if err := json.Unmarshal(output, &versionInfo); err != nil {
    a.logger.Debugf("Failed to parse JSON version output, falling back to text: %v", err)
    return a.checkVersionText()  // Fallback strategy
}
```

## Error Patterns

```go
// Return wrapped errors
return fmt.Errorf("failed to load config: %w", err)

// Use custom error types
return errors.NewConfigError("failed to parse config", path, err)
```
