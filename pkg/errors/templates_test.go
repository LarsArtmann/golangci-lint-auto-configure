package apperrors

import (
	errorfamily "github.com/larsartmann/go-error-family"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Domain message templates", func() {
	DescribeTable("every registered template has a non-empty What and Fix",
		func(code string) {
			tmpl, ok := errorfamily.TemplateForCode(code)
			Expect(ok).To(BeTrue(), "template %q should be registered", code)
			Expect(tmpl.What).NotTo(BeEmpty(), "template %q should have a What message", code)
			Expect(tmpl.Fix).NotTo(BeEmpty(), "template %q should have a Fix suggestion", code)
		},
		Entry("git.not_repository", "git.not_repository"),
		Entry("git.not_work_tree", "git.not_work_tree"),
		Entry("version.too_old", "version.too_old"),
		Entry("version.invalid_format", "version.invalid_format"),
		Entry("version.parse_json", "version.parse_json"),
		Entry("version.parse_text", "version.parse_text"),
		Entry("config.unsupported_format", "config.unsupported_format"),
		Entry("config.linters_parse", "config.linters_parse"),
		Entry("config.validation.version", "config.validation.version"),
		Entry("config.validation.issues_exit_code", "config.validation.issues_exit_code"),
		Entry("config.validation.concurrency", "config.validation.concurrency"),
		Entry("config.validation.max_issues_per_linter", "config.validation.max_issues_per_linter"),
		Entry("config.validation.max_same_issues", "config.validation.max_same_issues"),
		Entry("migration.parse_yaml", "migration.parse_yaml"),
		Entry("migration.encode_yaml", "migration.encode_yaml"),
		Entry("migration.validate_config", "migration.validate_config"),
		Entry("report.create_output", "report.create_output"),
		Entry("report.render", "report.render"),
		Entry("report.json_marshal", "report.json_marshal"),
		Entry("report.json_write", "report.json_write"),
		Entry("detector.open_file", "detector.open_file"),
		Entry("detector.walk_dir", "detector.walk_dir"),
		Entry("detector.open_gomod", "detector.open_gomod"),
		Entry("detector.scan_gomod", "detector.scan_gomod"),
		Entry("detector.walk_failed", "detector.walk_failed"),
		Entry("retry.interrupted", "retry.interrupted"),
		Entry("retry.exhausted", "retry.exhausted"),
	)

	It("every template in domainMessageTemplates resolves via TemplateForCode", func() {
		for code := range domainMessageTemplates {
			_, ok := errorfamily.TemplateForCode(code)
			Expect(ok).To(BeTrue(), "template %q should resolve", code)
		}
	})
})
