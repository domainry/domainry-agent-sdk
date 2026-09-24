package extraction

import (
	"fmt"
	"regexp"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

func backslashIssue(group string, index int) agentsdk.KnowledgeExtractionIssue {
	return agentsdk.KnowledgeExtractionIssue{Path: fmt.Sprintf("/%s/%d/pattern", group, index), Code: "literal_backslash_no_match", Message: "This pattern matches a literal backslash and found nothing. Check JSON escaping before concluding a field is missing. For whitespace prefer [[:space:]] or [ ] to avoid an extra escaping layer. The pattern is never silently rewritten."}
}

// Explain reports bounded, actionable rule errors without reading a document.
func Explain(plan agentsdk.KnowledgeExtractionPlan) []agentsdk.KnowledgeExtractionIssue {
	issues := []agentsdk.KnowledgeExtractionIssue{}
	if len(plan.Fields) > 32 || len(plan.Tables) > 4 {
		return []agentsdk.KnowledgeExtractionIssue{{Path: "/", Code: "rule_limit", Message: "Use at most 32 scalar fields and 4 tables."}}
	}
	inspect := func(path, pattern string, columns []agentsdk.KnowledgeExtractionColumn, table bool) {
		if len(issues) >= 8 {
			return
		}
		if _, err := compile(pattern, columns, table); err == nil {
			return
		}
		message := "Check unique field keys, supported types and valid allowed_values."
		code := "rule_invalid"
		expression, err := regexp.Compile(pattern)
		if len(pattern) > 2048 {
			code = "pattern_limit"
			message = "Use at most 2048 bytes for one pattern."
		} else if err != nil {
			code = "pattern_syntax"
			message = "Use Go RE2 syntax. Lookaround and backreferences are unsupported."
		} else if expression.NumSubexp() != len(columns) {
			code = "capture_count"
			message = fmt.Sprintf("This rule requires exactly %d capturing group(s), one per value. Wrap the value in parentheses; use (?:...) for grouping without an extra capture.", len(columns))
		} else if table {
			code = "column_captures"
			message = "Use either unnamed captures in the declared column order, or name EVERY capture exactly after its column key, for example (?P<item>...) and (?P<quantity>...). Do not mix named and unnamed captures."
		}
		issues = append(issues, agentsdk.KnowledgeExtractionIssue{Path: path, Code: code, Message: message})
	}
	for i, f := range plan.Fields {
		inspect(fmt.Sprintf("/fields/%d", i), f.Pattern, []agentsdk.KnowledgeExtractionColumn{f.KnowledgeExtractionColumn}, false)
	}
	for i, t := range plan.Tables {
		inspect(fmt.Sprintf("/tables/%d", i), t.Pattern, t.Columns, true)
	}
	if len(issues) == 0 {
		issues = append(issues, agentsdk.KnowledgeExtractionIssue{Path: "/", Code: "plan_shape", Message: "Supply at least one field or table, unique keys, nonnegative after and limit between 1 and 50."})
	}
	return issues
}
