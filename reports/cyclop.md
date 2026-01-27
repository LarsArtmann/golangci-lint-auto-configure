# cyclop Linter - Comprehensive Analysis

## What the Linter Does

**cyclop** calculates **cyclomatic complexity** of functions and packages to identify overly complex code that is difficult to understand, test, and maintain. Cyclomatic complexity measures the number of linearly independent paths through code - essentially, how complex your code's control flow is.

### The Concept

Cyclomatic complexity (also known as McCabe complexity) counts:
- The number of decision points (if, switch, loops)
- Each adds branching paths through the code
- Higher complexity = more paths = harder to test and maintain

**Formula**: `Complexity = Number of decision points + 1`

### What It Measures

- **Function complexity**: Measures paths within individual functions
- **Package average**: Calculates average complexity across a package  
- **Default threshold**: 10 (functions exceeding this are flagged)
- **Repository**: https://github.com/bkielbasa/cyclop

### Example

```go
// Complexity = 1 (no decision points)
func simple() {
    fmt.Println("hello")
}

// Complexity = 2 (1 if statement + 1)
func withIf(x int) {
    if x > 0 {
        fmt.Println("positive")
    }
}

// Complexity = 3 (2 if statements + 1)
func withTwoIfs(x int) {
    if x > 0 {
        fmt.Println("positive")
    }
    if x < 0 {
        fmt.Println("negative")
    }
}

// Complexity = 4 (1 if + 1 switch with 2 cases + 1)
func complex(x int) {
    if x > 0 {
        switch x {
        case 1:
            fmt.Println("one")
        case 2:
            fmt.Println("two")
        }
    }
}
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**
- **Medium to large codebases** - Where complexity debt accumulates quickly
- **Web APIs and microservices** - Maintainability is critical
- **Long-term projects** - Code will be read more than written
- **Team environments** - Multiple developers need to understand code
- **Libraries and SDKs** - Public APIs need clarity
- **Codebases with technical debt** - Measure and reduce complexity gradually

**Team Scenarios:**
- **Junior developers** - Provides guardrails against overly complex code
- **Code reviews** - Objective metric for complexity discussions
- **Refactoring efforts** - Identifies priority targets
- **Quality gates** - Prevents complexity from increasing
- **Onboarding** - Simpler code is easier for new team members

**Code Quality Goals:**
- **Improved testability** - Simple functions are easier to unit test
- **Better maintainability** - Less cognitive load to understand code
- **Fewer bugs** - Complex code has more hiding spots for bugs
- **Faster reviews** - Simple code is quicker to review
- **Documentation** - Forces developers to document complex logic

### Priority Assessment

- **Default Priority**: Medium-High
- **Essential for**: Large codebases, team projects, long-term maintenance
- **Recommended for**: Most production code
- **Less critical for**: Prototypes, small utilities, one-off scripts

## When It Should Be Disabled

### ❌ Consider Disabling For:

**Project Types:**
- **Small utility scripts** - Simplicity is inherent
- **Prototypes and POCs** - Speed over maintainability
- **One-off tools** - Won't be maintained long-term
- **Inherently complex algorithms** - Parsers, compilers, mathematical solvers
- **Generated code** - Complexity is in the generator, not the output

**Specific Scenarios:**
- **Existing complexity linters** - If already using `gocyclo` or `gocognit` (redundant)
- **Legacy codebases** - If refactoring is not planned
- **Performance-critical code** - Sometimes complexity is necessary for performance
- **Deadline crunch** - Temporary disable to meet deadlines (not recommended)

### Better Alternative to Disabling

Use **exclusions** for legitimate complexity:

```yaml
# .golangci.yml
linters:
  enable:
    - cyclop

issues:
  exclusions:
    rules:
      # Exclude test files (often have higher complexity)
      - path: '(.+)_test\.go'
        linters:
          - cyclop
      
      # Exclude generated code
      - path: '(.+)_generated\.go'
        linters:
          - cyclop
      
      # Exclude specific complex but necessary functions
      - path: pkg/parser/compiler.go
        text: 'function .* has cyclomatic complexity \d+'
        linters:
          - cyclop
      
      # Exclude specific known functions by name
      - linters: [cyclop]
        text: "function 'ComplexButNecessaryAlgorithm'"
        path: pkg/algorithms/
```

Or use inline `nolint` directives:

```go
// ComplexButNecessaryAlgorithm has high complexity due to business rules
func ComplexButNecessaryAlgorithm(input string) Result { //nolint:cyclop // Business logic complexity
    // Complex implementation...
}
```

## How It Should Be Configured

### Configuration Options

**cyclop** supports these settings in golangci-lint:

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    cyclop:
      # Maximum complexity allowed for a function (default: 10)
      max-complexity: 10
      
      # Maximum average complexity for a package (default: 0, disabled)
      # Set to non-zero to enable package-level checking
      package-average: 0
```

### Configuration Options Explained

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `max-complexity` | `int` | `10` | Maximum complexity allowed for a function |
| `package-average` | `int` | `0` | Maximum average complexity across a package (0 = disabled) |

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)
```yaml
# Most Go projects
version: "2"
linters:
  settings:
    cyclop:
      max-complexity: 15  # Slightly more permissive than default
      package-average: 0  # Focus on function-level initially
```

#### ✅ Strict Configuration
```yaml
# Library code requiring high maintainability
version: "2"
linters:
  settings:
    cyclop:
      max-complexity: 10        # Conservative limit per function
      package-average: 8        # Enforce package-wide simplicity
```

#### ✅ Relaxed Configuration
```yaml
# Application code with complex business logic
version: "2"
linters:
  settings:
    cyclop:
      max-complexity: 25        # Allow more complexity per function
      package-average: 0        # Disable package checks
```

#### ✅ Progressive Configuration
```yaml
# Start high and gradually lower as you refactor
version: "2"
linters:
  settings:
    cyclop:
      max-complexity: 30        # Start high for existing code
```

Then over time:
```yaml
# Later, after refactoring
version: "2"
linters:
  settings:
    cyclop:
      max-complexity: 20        # Lower as code improves
```

### Best Practices

1. **Start conservative**: Begin with default (10) or slightly higher (15)
2. **Use exclusions for exceptions**: Complex but necessary functions can be excluded
3. **Gradually tighten**: Lower thresholds as you refactor
4. **Package-level checks**: Consider `package-average` for new projects
5. **Exclude tests**: Tests often have higher complexity - exclude them by default
6. **Pair with funlen**: Complexity + length gives complete picture
7. **Document exclusions**: Use `//nolint:cyclop` with explanatory comments

## How It Interferes or Works Together With Other Linters

### Complexity Preset

cyclop is part of the `complexity` preset along with:
- **`funlen`** - Function length limits
- **`gocognit`** - Cognitive complexity
- **`gocyclo`** - Cyclomatic complexity (alternative to cyclop)
- **`maintidx`** - Maintainability index
- **`nestif`** - Nested if statements

### ✅ Synergistic Combinations

**Harmonious Configuration:**
```yaml
# cyclop + funlen + maintidx = comprehensive complexity analysis
linters:
  enable:
    - cyclop
    - funlen
    - maintidx
  
  settings:
    cyclop:
      max-complexity: 15
    
    funlen:
      lines: 80
      statements: 50
    
    maintidx:
      under: 20  # High maintainability threshold
```

**Each linter measures different aspects:**
- **cyclop**: Branching paths (if/switch/loops)
- **funlen**: Physical size (lines/statements)
- **maintidx**: Mix of complexity and size metrics

**Example Synergy:**
```go
// cyclop measures branching complexity
// funlen measures function length
// maintidx calculates overall maintainability
func processOrder(order Order) error { //nolint:funlen // Long but necessary
    if err := validateOrder(order); err != nil {
        return err
    }
    
    switch order.Status {  // cyclop counts each case
    case StatusNew:
        if err := processNewOrder(order); err != nil {
            return err
        }
    case StatusPending:
        if err := processPendingOrder(order); err != nil {
            return err
        }
    case StatusProcessing:
        if complexCondition {  // Nested conditions increase complexity
            // Many more branches...
        }
    default:
        return fmt.Errorf("unknown status: %s", order.Status)
    }
    
    return nil
}
```

### ⚠️ Redundant Combinations

 **Avoid using cyclop + gocyclo**  : Both measure cyclomatic complexity in similar ways
```yaml
# Don't do this - redundant
linters:
  enable:
    - cyclop
    - gocyclo  # ⛔ Measures the same thing
```

**Pick one complexity linter:**
- **cyclop**: Newer, actively maintained, with package-average feature
- **gocyclo**: Older, widely used, simpler
- **gocognit**: Measures cognitive complexity (different concept)

**Better approach:**
```yaml
# Choose cyclop OR gocyclo, not both
linters:
  enable:
    - cyclop       # Pick this one
    # - gocyclo    # Or this one, not both
    - gocognit     # This is different - can include
```

### 📊 Performance Impact

- **Low overhead**: Single-pass AST analysis
- **Fast execution**: Complexity calculation is O(n) where n = decision points
- **Recommended in CI/CD**: No performance concerns
- **Scales linearly**: With number of functions in codebase

## Practical Examples

### ✅ Refactoring to Reduce Complexity

```go
// ❌ Before: Complex function (complexity = 8)
func processOrder(order Order) error {
    if order == nil {
        return errors.New("nil order")
    }
    
    if order.Total <= 0 {
        return errors.New("invalid total")
    }
    
    if order.CustomerID == "" {
        return errors.New("missing customer")
    }
    
    switch order.Status {
    case StatusNew:
        if err := chargePayment(order); err != nil {
            return err
        }
        order.Status = StatusProcessing
        if err := saveOrder(order); err != nil {
            return err
        }
    case StatusPending:
        if err := retryPayment(order); err != nil {
            return err
        }
    default:
        return fmt.Errorf("invalid status: %s", order.Status)
    }
    
    return nil
}
```

```go
// ✅ After: Refactored (complexity = 3 per function)
func processOrder(order Order) error {
    if err := validateOrder(order); err != nil {
        return err
    }
    
    return processOrderByStatus(order)
}

func validateOrder(order Order) error {  // Complexity = 3
    if order == nil {
        return errors.New("nil order")
    }
    if order.Total <= 0 {
        return errors.New("invalid total")
    }
    if order.CustomerID == "" {
        return errors.New("missing customer")
    }
    return nil
}

func processOrderByStatus(order Order) error {  // Complexity = 3
    switch order.Status {
    case StatusNew:
        return processNewOrder(order)
    case StatusPending:
        return processPendingOrder(order)
    default:
        return fmt.Errorf("invalid status: %s", order.Status)
    }
}
```

### ✅ Using Exclusions for Necessary Complexity

```go
// Parser implementation - complexity is inherent to the problem
func parseExpression(tokens []Token) (ASTNode, error) { //nolint:cyclop // Parser complexity unavoidable
    if len(tokens) == 0 {
        return nil, errors.New("empty tokens")
    }
    
    switch tokens[0].Type {
    case TokenNumber:
        return parseNumber(tokens)
    case TokenString:
        return parseString(tokens)
    case TokenIdentifier:
        return parseIdentifier(tokens)
    case TokenOperator:
        return parseOperator(tokens)
    case TokenLParen:
        return parseGroup(tokens)
    case TokenLBracket:
        return parseArray(tokens)
    case TokenLBrace:
        return parseObject(tokens)
    case TokenKeyword:
        switch tokens[0].Value {
        case "if":
            return parseIf(tokens)
        case "while":
            return parseWhile(tokens)
        case "for":
            return parseFor(tokens)
        case "function":
            return parseFunction(tokens)
        case "return":
            return parseReturn(tokens)
        default:
            return nil, fmt.Errorf("unknown keyword: %s", tokens[0].Value)
        }
    default:
        return nil, fmt.Errorf("unexpected token: %s", tokens[0].Type)
    }
}
```

### ✅ Package-Level Complexity Control

```yaml
# .golangci.yml
linters:
  settings:
    cyclop:
      max-complexity: 15
      package-average: 12  # Enable package-level checks
```

The linter will now additionally flag packages where the average complexity exceeds 12, encouraging distribution of complexity across multiple files/functions.

## Common Scenarios

### Scenario 1: HTTP Handler Complexity

```go
// ❌ High complexity due to multiple validation branches
func handleCreateUser(w http.ResponseWriter, r *http.Request) { // Complex: 12
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    if r.Header.Get("Content-Type") != "application/json" {
        http.Error(w, "Invalid content type", http.StatusBadRequest)
        return
    }
    
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Failed to read body", http.StatusBadRequest)
        return
    }
    
    var user User
    if err := json.Unmarshal(body, &user); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    if user.Email == "" {
        http.Error(w, "Email required", http.StatusBadRequest)
        return
    }
    
    if user.Name == "" {
        http.Error(w, "Name required", http.StatusBadRequest)
        return
    }
    
    if err := validateEmail(user.Email); err != nil {
        http.Error(w, "Invalid email", http.StatusBadRequest)
        return
    }
    
    if err := saveUser(r.Context(), &user); err != nil {
        http.Error(w, "Failed to create user", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}
```

**Solution**: Extract validation logic
```go
// ✅ Refactored - each function has complexity < 8
func handleCreateUser(w http.ResponseWriter, r *http.Request) { // Complex: 6
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    user, err := parseAndValidateUser(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    if err := saveUser(r.Context(), user); err != nil {
        http.Error(w, "Failed to create user", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

func parseAndValidateUser(r *http.Request) (*User, error) { // Complex: 7
    if r.Header.Get("Content-Type") != "application/json" {
        return nil, errors.New("invalid content type")
    }
    
    body, err := io.ReadAll(r.Body)
    if err != nil {
        return nil, errors.New("failed to read body")
    }
    
    var user User
    if err := json.Unmarshal(body, &user); err != nil {
        return nil, errors.New("invalid JSON")
    }
    
    if err := validateUser(&user); err != nil {
        return nil, err
    }
    
    return &user, nil
}

func validateUser(user *User) error { // Complex: 4
    if user.Email == "" {
        return errors.New("email required")
    }
    if user.Name == "" {
        return errors.New("name required")
    }
    if err := validateEmail(user.Email); err != nil {
        return errors.New("invalid email")
    }
    return nil
}
```

### Scenario 2: Switch Statement Complexity

```go
// ❌ Complex due to many switch cases
func processEvent(event Event) error { // Complex: 15
    switch event.Type {
    case EventUserCreated:
        return handleUserCreated(event)
    case EventUserUpdated:
        return handleUserUpdated(event)
    case EventUserDeleted:
        return handleUserDeleted(event)
    case EventOrderCreated:
        return handleOrderCreated(event)
    case EventOrderUpdated:
        return handleOrderUpdated(event)
    case EventOrderCancelled:
        return handleOrderCancelled(event)
    case EventPaymentReceived:
        return handlePaymentReceived(event)
    case EventPaymentFailed:
        return handlePaymentFailed(event)
    case EventShipmentCreated:
        return handleShipmentCreated(event)
    case EventShipmentDelivered:
        return handleShipmentDelivered(event)
    case EventInventoryLow:
        return handleInventoryLow(event)
    case EventInventoryRestocked:
        return handleInventoryRestocked(event)
    default:
        return fmt.Errorf("unknown event type: %s", event.Type)
    }
}
```

**Solution**: Use a map-based dispatcher
```go
// ✅ Refactored - dispatcher has complexity = 3
func processEvent(event Event) error { // Complex: 3
    handler, ok := eventHandlers[event.Type]
    if !ok {
        return fmt.Errorf("unknown event type: %s", event.Type)
    }
    return handler(event)
}

// Map initialization (in init() or package var)
var eventHandlers = map[EventType]func(Event) error{
    EventUserCreated:      handleUserCreated,
    EventUserUpdated:      handleUserUpdated,
    EventUserDeleted:      handleUserDeleted,
    EventOrderCreated:     handleOrderCreated,
    EventOrderUpdated:     handleOrderUpdated,
    EventOrderCancelled:   handleOrderCancelled,
    EventPaymentReceived:  handlePaymentReceived,
    EventPaymentFailed:    handlePaymentFailed,
    EventShipmentCreated:  handleShipmentCreated,
    EventShipmentDelivered: handleShipmentDelivered,
    EventInventoryLow:     handleInventoryLow,
    EventInventoryRestocked: handleInventoryRestocked,
}

// Each handler now has individual complexity limits
func handleUserCreated(event Event) error { // Complex: 2
    // Implementation
    return nil
}
```

## Best Practices Summary

1. **Start with default or slightly higher threshold (10-15)**
2. **Refactor incrementally** - Don't try to fix everything at once
3. **Use exclusions judiciously** - Document why complexity is necessary
4. **Pair with funlen** - Function length + complexity = complete picture
5. **Focus on new code** - Use `new-from-rev` to only check new changes
6. **Extract helper functions** - The primary refactoring technique
7. **Use strategy pattern** - Replace complex conditionals with polymorphism
8. **Test complex functions thoroughly** - Higher complexity = more tests needed

## Summary

**cyclop** is a **valuable complexity metric linter** that helps maintain code quality by:
- ✅ Identifying hard-to-maintain functions
- ✅ Encouraging better code structure
- ✅ Providing objective complexity measurement
- ✅ Guiding refactoring efforts
- ✅ Preventing complexity debt accumulation

**Recommendation**: **ENABLE** for most production codebases, especially team projects and long-term applications. Set thresholds based on your team's needs and refactor incrementally. Use with `funlen` and `maintidx` for comprehensive code quality analysis.

---

**Reference**: https://github.com/bkielbasa/cyclop
