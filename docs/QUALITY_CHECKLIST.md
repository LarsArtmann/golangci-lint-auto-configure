# Linter Documentation Quality Checklist

Before marking a linter as **completed**, verify all criteria below are met:

## A. Content Completeness ✅

### Required Sections
- [ ] All required sections present:
  - [ ] "What the Linter Does"
  - [ ] "When It Should Be Enabled"
  - [ ] "How It Should Be Configured"
  - [ ] "How It Interferes or Works Together With Other Linters"

### Content Depth
- [ ] Linter purpose clearly explained (1-2 paragraphs)
- [ ] Problem detection details provided
- [ ] Why it matters explained (bullet points)
- [ ] Technical implementation explained (how it works)
- [ ] At least 5 practical examples with BAD/GOOD code
- [ ] Examples have clear explanations

## B. Technical Accuracy ✅

### Configuration Examples
- [ ] All configuration options documented with:
  - [ ] Type (bool, string, []string, etc.)
  - [ ] Default value
  - [ ] Detailed description
  - [ ] Example YAML configuration

### Configuration Correctness
- [ ] All YAML examples are valid syntax
- [ ] No invalid YAML structure
- [ ] No unknown configuration options

### Inter-Linter Relationships
- [ ] Complementary linters table included
- [ ] Conflicts/overlap table included
- [ ] Example of linter synergy provided
- [ ] Explanation of why no conflicts

### Priority Information
- [ ] Priority matches linter_data.go (CRITICAL/HIGH/MEDIUM)
- [ ] Priority justification provided
- [ ] Enable/disable scenarios specific and actionable
- [ ] Recommendation is clear (ALWAYS/RECOMMEND/CONSIDER)

## C. Consistency ✅

### Format Consistency
- [ ] Follows template structure
- [ ] Section order matches template
- [ ] Consistent use of ✅ GOOD vs ❌ BAD examples
- [ ] Consistent code block formatting (```go)
- [ ] Consistent YAML block formatting (```yaml)

### Terminology Consistency
- [ ] Linter name used consistently
- [ ] Terminology matches other documentation
- [ ] Acronyms explained (e.g., CRITICAL, HIGH, MEDIUM)
- [ ] Section headings use consistent style (##, ###)

### Style Consistency
- [ ] Table formatting consistent
- [ ] Bullet point formatting consistent
- [ ] Code comments use consistent style
- [ ] Lists use consistent indentation

## D. Validation ✅

### File Validation
- [ ] File exists in /reports/ directory
- [ ] File is not empty
- [ ] File size > 10 KB (comprehensive documentation)
- [ ] No markdown syntax errors

### Validation Script Check
- [ ] `/scripts/validate_linter_doc.sh {{LINTER_NAME}}` passes
- [ ] All required sections detected
- [ ] Examples present
- [ ] Configuration examples present
- [ ] Inter-linter relationships documented

## E. Quality Standards ✅

### Examples Quality
- [ ] At least 5 practical examples
- [ ] Each example has BAD (❌) code showing issue
- [ ] Each example has GOOD (✅) code showing fix
- [ ] Each example has brief explanation
- [ ] Examples cover different scenarios

### Best Practices
- [ ] At least 10 best practices listed
- [ ] Practices are actionable and specific
- [ ] Practices are relevant to linter
- [ ] Practices are practical, not theoretical

### Common Scenarios
- [ ] At least 5 common scenarios documented
- [ ] Each scenario has Problem description
- [ ] Each scenario has Solution
- [ ] Each scenario has example YAML configuration
- [ ] Scenarios are realistic and helpful

### Summary Section
- [ ] Summary lists key benefits (✅)
- [ ] Summary lists limitations (⚠️)
- [ ] Recommendation is clear and specific
- [ ] Top 3 configuration tips provided
- [ ] Reference link included

## F. Documentation Structure ✅

### Section Requirements
- [ ] "What the Linter Does" section present and complete
- [ ] "When It Should Be Enabled" section present and complete
- [ ] "How It Should Be Configured" section present and complete
- [ ] "How It Interferes or Works Together With Other Linters" section present and complete

### Sub-section Requirements
- [ ] "The Problem It Detects" in "What" section
- [ ] "How It Works" in "What" section
- [ ] "Examples" in "What" section (at least 5)
- [ ] "Priority Assessment" in "When" section
- [ ] "Configuration Options" in "How" section
- [ ] "Recommended Configurations" in "How" section (at least 3)
- [ ] "Practical Examples" section (at least 5 examples)
- [ ] "Best Practices" section (at least 10 items)
- [ ] "Common Scenarios and Solutions" section (at least 5 scenarios)
- [ ] "Summary" section with benefits, limitations, recommendation

## G. Completion Verification ✅

### Pre-Completion Checks
- [ ] Research complete (all sources reviewed)
- [ ] All configuration options documented
- [ ] All examples tested/verified (where applicable)
- [ ] Inter-linter relationships cross-checked

### Post-Completion Validation
- [ ] `/scripts/validate_linter_doc.sh {{LINTER_NAME}}` executed successfully
- [ ] All validation checks passed
- [ ] File size adequate (> 10 KB)
- [ ] No markdown errors in file

### Git Status
- [ ] File added to git
- [ ] File committed with descriptive message
- [ ] Pushed to remote repository

---

## Approval Criteria

**Linter documentation is APPROVED when:**

1. ✅ All items in sections A-F are checked
2. ✅ All validation scripts pass
3. ✅ File is committed and pushed to git
4. ✅ File size > 10 KB
5. ✅ Total linter count verification passes

---

**Use this checklist for EVERY linter documented.**
