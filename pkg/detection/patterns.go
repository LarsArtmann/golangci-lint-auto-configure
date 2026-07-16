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

// RecommendedLinters maps project types to their recommended linters.
var RecommendedLinters = map[ProjectType][]string{
	ProjectTypeCLI: {
		"gosec", "errcheck", "staticcheck", "govet", "ineffassign",
		"wrapcheck", "errorlint", "gocritic", "nolintlint",
	},
	ProjectTypeLibrary: {
		"gosec", "errcheck", "staticcheck", "govet", "ineffassign",
		"wrapcheck", "errorlint", "gocritic", "musttag",
	},
	ProjectTypeWeb: {
		"gosec", "errcheck", "staticcheck", "govet", "ineffassign",
		"noctx", "bodyclose", "wrapcheck", "errorlint",
	},
	ProjectTypeAPI: {
		"gosec", "errcheck", "staticcheck", "govet", "ineffassign",
		"noctx", "bodyclose", "wrapcheck", "errorlint", "musttag",
	},
	ProjectTypeMonorepo: {
		"gosec", "errcheck", "staticcheck", "govet", "ineffassign",
		"gocritic", "errorlint",
	},
	ProjectTypeUnknown: {
		"gosec", "errcheck", "staticcheck", "govet", "ineffassign",
	},
}

// GetRecommendedLinters returns recommended linters for a project type.
func GetRecommendedLinters(projectType ProjectType) []string {
	if linters, ok := RecommendedLinters[projectType]; ok {
		return linters
	}

	return RecommendedLinters[ProjectTypeUnknown]
}
