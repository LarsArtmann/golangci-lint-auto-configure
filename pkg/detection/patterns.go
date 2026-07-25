package detection

// HTTPFrameworks contains common HTTP framework import paths.
var HTTPFrameworks = []string{
	"github.com/gin-gonic/gin",
	"github.com/labstack/echo",
	"github.com/gofiber/fiber",
	"github.com/gorilla/mux",
	"github.com/go-chi/chi",
	"net/http",
	"github.com/valyala/fasthttp",
	"goa.design/goa",
}

// CLIFrameworks contains common CLI framework import paths.
var CLIFrameworks = []string{
	"github.com/spf13/cobra",
	"github.com/urfave/cli",
	"github.com/alecthomas/kingpin",
	"github.com/charmbracelet/bubbletea",
	"github.com/charmbracelet/lipgloss",
	"github.com/manifoldco/promptui",
}

// APIPatterns contains common API code patterns to search for.
var APIPatterns = []string{
	"json.Marshal",
	"json.Unmarshal",
	"http.Handler",
	"grpc.",
	"proto.",
	"REST",
	"API",
}

// SwaggoImports contains swaggo-related import paths.
var SwaggoImports = []string{
	"github.com/swaggo/swag",
	"github.com/swaggo/gin-swagger",
	"github.com/swaggo/echo-swagger",
	"github.com/swaggo/fiber-swagger",
	"github.com/swaggo/http-swagger",
	"github.com/swaggo/swag/cmd/swag",
	"github.com/swaggo/files",
}

// ClickHouseImports contains ClickHouse driver import paths.
var ClickHouseImports = []string{
	"github.com/ClickHouse/clickhouse-go",
	"github.com/mailru/go-clickhouse",
}

// ArangoDBImports contains ArangoDB driver import paths.
var ArangoDBImports = []string{
	"github.com/arangodb/go-driver",
}

// SwaggoPatterns contains swaggo annotation patterns to search for in code.
var SwaggoPatterns = []string{
	"@Summary",
	"@Description",
	"@Tags",
	"@Accept",
	"@Produce",
	"@Param",
	"@Success",
	"@Failure",
	"@Router",
	"@Security",
	"@Header",
	"@ID",
	"swagger",
}

// coreLinters are the essential linters recommended for every project type.
// This matches the minimalLinters set from pkg/constants/presets.go.
var coreLinters = []string{
	"gosec", "errcheck", "staticcheck", "govet", "ineffassign",
}

// withCore returns a new slice containing coreLinters followed by extra.
// Each call produces an independent slice to avoid shared backing arrays.
func withCore(extra ...string) []string {
	result := make([]string, 0, len(coreLinters)+len(extra))
	result = append(result, coreLinters...)

	return append(result, extra...)
}

// RecommendedLinters maps project types to their recommended linters.
var RecommendedLinters = map[ProjectType][]string{
	ProjectTypeCLI:      withCore("wrapcheck", "errorlint", "gocritic", "nolintlint"),
	ProjectTypeLibrary:  withCore("wrapcheck", "errorlint", "gocritic", "musttag"),
	ProjectTypeWeb:      withCore("noctx", "bodyclose", "wrapcheck", "errorlint"),
	ProjectTypeAPI:      withCore("noctx", "bodyclose", "wrapcheck", "errorlint", "musttag"),
	ProjectTypeMonorepo: withCore("gocritic", "errorlint"),
	ProjectTypeUnknown:  withCore(),
}

// GetRecommendedLinters returns recommended linters for a project type.
func GetRecommendedLinters(projectType ProjectType) []string {
	if linters, ok := RecommendedLinters[projectType]; ok {
		return linters
	}

	return RecommendedLinters[ProjectTypeUnknown]
}
