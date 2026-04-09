# ADR-004: BDD Testing Approach with Ginkgo/Gomega

**Status:** Accepted  
**Date:** 2026-04-09  
**Author:** Lars Artmann (@larsartmann)

---

## Context

We need a testing approach that provides clear, readable test specifications that serve as documentation. The standard Go testing framework (`testing` package) is functional but produces verbose tests that are hard to read and maintain.

Prior to this decision, we considered:

1. **Standard Go testing**: Built-in, but verbose assertions, repetitive boilerplate
2. **Testify**: Popular assertion library, but doesn't solve the "story" problem
3. **Ginkgo/Gomega**: BDD framework with natural language descriptions

---

## Decision

We will use **Ginkgo v2** (BDD testing framework) with **Gomega** (matcher library) for all tests.

### Implementation

```go
// pkg/types/set_test.go
import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

func TestSet(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Set Suite")
}

var _ = Describe("Set", func() {
    var set Set[string]

    BeforeEach(func() {
        set = NewSet("a", "b", "c")
    })

    Context("when adding elements", func() {
        It("should contain the added element", func() {
            set = set.Add("d")
            Expect(set.Contains("d")).To(BeTrue())
        })

        It("should be idempotent", func() {
            set = set.Add("a")  // Already exists
            Expect(set.Len()).To(Equal(3))
        })
    })

    Context("when checking membership", func() {
        It("should return true for existing elements", func() {
            Expect(set.Contains("a")).To(BeTrue())
        })

        It("should return false for non-existing elements", func() {
            Expect(set.Contains("z")).To(BeFalse())
        })
    })
})
```

---

## Consequences

### Positive

- **Readable specifications**: Tests read like natural language
- **Documentation**: Test output serves as feature documentation
- **Structured organization**: `Describe` > `Context` > `It` hierarchy
- **Rich matchers**: Gomega provides extensive assertion helpers
- **Parallel execution**: Ginkgo supports parallel test execution

### Negative

- **Learning curve**: Team must learn Ginkgo/Gomega idioms
- **Dot imports**: `import . "github.com/onsi/ginkgo/v2"` is required for DSL
- **Dependency**: External test dependencies

### Trade-offs Accepted

- DSL complexity for readability and expressiveness

---

## Test Structure Convention

```
Describe("Component")
├── Context("when in state X")
│   ├── BeforeEach()        // Setup
│   ├── It("should do Y")
│   └── It("should not do Z")
├── Context("when in state W")
│   ├── BeforeEach()
│   └── It("should do V")
└── Describe("sub-component")
    └── It("should have property P")
```

---

## Gomega Matchers Cheatsheet

| Matcher | Purpose | Example |
|---------|---------|---------|
| `Equal()` | Exact equality | `Expect(x).To(Equal(42))` |
| `BeNil()` | Nil check | `Expect(err).To(BeNil())` |
| `BeTrue()/BeFalse()` | Boolean | `Expect(ok).To(BeTrue())` |
| `HaveLen()` | Length check | `Expect(slice).To(HaveLen(3))` |
| `ContainElement()` | Contains | `Expect(list).To(ContainElement("a"))` |
| `MatchError()` | Error message | `Expect(err).To(MatchError("not found"))` |
| `Succeed()` | No error | `Expect(fn()).To(Succeed())` |
| `And()/Or()` | Composition | `Expect(x).To(And(BeTrue(), HaveLen(2)))` |

---

## Table-Driven Tests with Ginkgo

```go
var _ = Describe("Priority", func() {
    DescribeTable("string conversion",
        func(priority LinterPriority, expected string) {
            Expect(priority.String()).To(Equal(expected))
        },
        Entry("critical", LinterPriorityCritical, "CRITICAL"),
        Entry("high", LinterPriorityHigh, "HIGH"),
        Entry("medium", LinterPriorityMedium, "MEDIUM"),
        Entry("optional", LinterPriorityOptional, "OPTIONAL"),
    )
})
```

---

## Running Tests

```bash
# Run all tests
ginkgo -r ./...

# Run specific package
ginkgo ./pkg/types/...

# Run with verbose output
ginkgo -v ./pkg/config/...

# Run specific test by name
ginkgo -r --focus="Set"

# Skip specific tests
ginkgo -r --skip="Integration"

# Run in parallel
ginkgo -p ./...

# Generate coverage
ginkgo -r --cover && go tool cover -html=coverage.out
```

---

## When to Use Standard Testing

Use standard Go testing when:

- Testing external interfaces (APIs, CLIs)
- Need minimal dependencies
- Simple unit tests without complex setup
- Benchmarking (Ginkgo doesn't support benchmarks well)

---

## Migration Strategy

1. **New code**: Always use Ginkgo/Gomega
2. **Existing tests**: Refactor when modifying
3. **Hybrid**: Can mix Ginkgo and standard tests in same package

---

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| Standard Go testing | No deps, familiar | Verbose, poor documentation | Rejected for most tests |
| Testify | Good assertions | Still imperative style | Rejected |
| **Ginkgo/Gomega** | BDD style, readable | DSL to learn, dot imports | **Accepted** |

---

## References

- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Documentation](https://onsi.github.io/gomega/)
- Example: [pkg/types/set_test.go](../../pkg/types/set_test.go)
- Example: [pkg/config/merger_test.go](../../pkg/config/merger_test.go)

---

_Accepted by: Lars Artmann_  
_Date: 2026-04-09_
