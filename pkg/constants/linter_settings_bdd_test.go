package constants_test

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These specs pin the exact wire output of every curated default: the values
// users receive in their .golangci.yml. A change here without a deliberate
// spec update means every downstream project silently receives new behavior.
var _ = Describe("DefaultLinterSettings wire output", func() {
	Describe("complexity linters", func() {
		It("gocognit produces min-complexity 25", func() {
			m := constants.DefaultLinterSettings["gocognit"].ToMap()
			Expect(m).To(HaveKeyWithValue("min-complexity", 25))
			Expect(m).To(HaveLen(1))
		})

		It("gocyclo produces min-complexity 20", func() {
			m := constants.DefaultLinterSettings["gocyclo"].ToMap()
			Expect(m).To(HaveKeyWithValue("min-complexity", 20))
			Expect(m).To(HaveLen(1))
		})

		It("nestif produces min-complexity 6", func() {
			m := constants.DefaultLinterSettings["nestif"].ToMap()
			Expect(m).To(HaveKeyWithValue("min-complexity", 6))
			Expect(m).To(HaveLen(1))
		})
	})

	Describe("goconst", func() {
		It("produces min-len (NOT the schema-invalid min-length), min-occurrences, ignore-tests", func() {
			m := constants.DefaultLinterSettings["goconst"].ToMap()
			Expect(m).To(HaveKeyWithValue("min-len", 4))
			Expect(m).To(HaveKeyWithValue("min-occurrences", 5))
			Expect(m).To(HaveKeyWithValue("ignore-tests", true))
			Expect(m).To(HaveLen(3))
			Expect(m).NotTo(HaveKey("min-length"))
		})
	})

	Describe("tagalign", func() {
		It("produces align/sort/order with the curated ordering", func() {
			m := constants.DefaultLinterSettings["tagalign"].ToMap()
			Expect(m).To(HaveKeyWithValue("align", false))
			Expect(m).To(HaveKeyWithValue("sort", true))
			Expect(m).To(HaveKeyWithValue("order", []any{
				"binding", "json", "yaml", "xml", "toml", "validate", "mapstructure",
			}))
		})
	})

	Describe("mnd", func() {
		It("produces ignored-files for test files", func() {
			m := constants.DefaultLinterSettings["mnd"].ToMap()
			Expect(m).To(HaveKeyWithValue("ignored-files", []any{`_test\.go`}))
		})

		It("intentionally omits ignored-numbers (per-project decision, not universal)", func() {
			m := constants.DefaultLinterSettings["mnd"].ToMap()
			Expect(m).NotTo(HaveKey("ignored-numbers"))
		})
	})

	Describe("wrapcheck", func() {
		It("produces the stdlib signature regexps", func() {
			m := constants.DefaultLinterSettings["wrapcheck"].ToMap()
			regexps, ok := m["ignore-sig-regexps"].([]any)
			Expect(ok).To(BeTrue(), "ignore-sig-regexps must be a list")
			Expect(regexps).To(HaveLen(16))
			Expect(regexps).To(ContainElement("fmt\\."))
			Expect(regexps).To(ContainElement("errors\\."))
			Expect(regexps).To(ContainElement(`\(context\.Context\)\.Err`))
		})

		It("produces the error-constructor ignore sigs", func() {
			m := constants.DefaultLinterSettings["wrapcheck"].ToMap()
			sigs, ok := m["ignore-sigs"].([]any)
			Expect(ok).To(BeTrue(), "ignore-sigs must be a list")
			Expect(sigs).To(HaveLen(9))
			Expect(sigs).To(ContainElement("errors.New("))
			Expect(sigs).To(ContainElement(".Errorf("))
		})
	})

	Describe("errcheck", func() {
		It("enables check-type-assertions", func() {
			m := constants.DefaultLinterSettings["errcheck"].ToMap()
			Expect(m).To(HaveKeyWithValue("check-type-assertions", true))
		})

		It("excludes curated Close/Fprint functions", func() {
			m := constants.DefaultLinterSettings["errcheck"].ToMap()
			excluded, ok := m["exclude-functions"].([]any)
			Expect(ok).To(BeTrue(), "exclude-functions must be a list")
			Expect(excluded).To(HaveLen(16))
			Expect(excluded).To(ContainElement("(*os.File).Close"))
			Expect(excluded).To(ContainElement("(io.Closer).Close"))
			Expect(excluded).To(ContainElement("fmt.Fprintln"))
		})
	})

	Describe("varnamelen", func() {
		It("produces the typed threshold knobs", func() {
			m := constants.DefaultLinterSettings["varnamelen"].ToMap()
			Expect(m).To(HaveKeyWithValue("max-distance", 15))
			Expect(m).To(HaveKeyWithValue("min-name-length", 2))
			Expect(m).To(HaveKeyWithValue("ignore-map-index-ok", true))
			Expect(m).To(HaveKeyWithValue("ignore-type-assert-ok", true))
		})

		It("produces ignore-decls with stdlib-only typed declarations", func() {
			m := constants.DefaultLinterSettings["varnamelen"].ToMap()
			decls, ok := m["ignore-decls"].([]any)
			Expect(ok).To(BeTrue(), "ignore-decls must be a list")
			for _, required := range []string{"err error", "wg sync.WaitGroup", "mu sync.Mutex", "c context.Context", "db *sql.DB", "tx *sql.Tx", "t *testing.T"} {
				Expect(decls).To(ContainElement(required))
			}
			// Framework types are deliberately excluded: a shared default
			// referencing gin/httpx/koanf injects dead decls into every project
			// that does not use them and wrong expectations into those that do.
			for _, framework := range []string{"*gin.Context", "*httpx.Context", "*koanf.Koanf"} {
				Expect(decls).NotTo(ContainElement(framework),
					"shared varnamelen defaults must stay stdlib-only, found %q", framework)
			}
		})

		It("produces the curated ignore-names", func() {
			m := constants.DefaultLinterSettings["varnamelen"].ToMap()
			Expect(m).To(HaveKeyWithValue("ignore-names", []any{
				"err", "ok", "tt", "fn", "t", "i", "m", "g", "a", "b", "v",
			}))
		})
	})
})
