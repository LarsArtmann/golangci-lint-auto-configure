package constants

// FormatterInfo provides metadata about formatters.
// Derived from the formatter policy table (formatter_policy_table.go).
var FormatterInfo = formatterInfoView()

// FormatterPriorities defines priority levels for formatters.
// Derived from the formatter policy table.
var FormatterPriorities = formatterPrioritiesView()

// FormatterReasons provides human-readable reasons for each formatter recommendation.
// Derived from the formatter policy table.
var FormatterReasons = formatterReasonsView()
