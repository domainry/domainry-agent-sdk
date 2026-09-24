// Package extraction evaluates bounded read-only plans over already parsed
// text. It owns no file parsing, network access, credentials or model calls.
package extraction

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	InvalidPlan    Error = "extraction_plan_invalid"
	InvalidContent Error = "extraction_content_invalid"
)

var keyPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,63}$`)

type compiledRule struct {
	expression *regexp.Regexp
	columns    []agentsdk.KnowledgeExtractionColumn
	groups     []int
}

func columnValid(column agentsdk.KnowledgeExtractionColumn) bool {
	if !keyPattern.MatchString(column.Key) || len(column.AllowedValues) > 32 {
		return false
	}
	switch column.Type {
	case "text", "integer", "decimal", "date", "boolean":
	default:
		return false
	}
	seen := map[string]bool{}
	for _, v := range column.AllowedValues {
		normalized, ok := normalize(v, column.Type)
		if !ok || seen[normalized] {
			return false
		}
		seen[normalized] = true
	}
	return true
}
func compile(pattern string, columns []agentsdk.KnowledgeExtractionColumn, table bool) (compiledRule, error) {
	out := compiledRule{columns: columns}
	if len(pattern) == 0 || len(pattern) > 2048 || !utf8.ValidString(pattern) || len(columns) == 0 || len(columns) > 12 && table {
		return out, InvalidPlan
	}
	expression, err := regexp.Compile(pattern)
	if err != nil || expression.NumSubexp() != len(columns) {
		return out, InvalidPlan
	}
	seen := map[string]bool{}
	named := false
	for _, name := range expression.SubexpNames()[1:] {
		named = named || name != ""
	}
	for i, c := range columns {
		if !columnValid(c) || seen[c.Key] {
			return out, InvalidPlan
		}
		seen[c.Key] = true
		group := i + 1
		if table && named {
			group = expression.SubexpIndex(c.Key)
			if group < 1 {
				return out, InvalidPlan
			}
		}
		out.groups = append(out.groups, group)
	}
	if table && named {
		names := map[string]bool{}
		for _, name := range expression.SubexpNames()[1:] {
			if name == "" || names[name] || !seen[name] {
				return out, InvalidPlan
			}
			names[name] = true
		}
	}
	out.expression = expression
	return out, nil
}

func Validate(plan agentsdk.KnowledgeExtractionPlan) error {
	if len(plan.Fields) > 32 || len(plan.Tables) > 4 || len(plan.Fields)+len(plan.Tables) == 0 {
		return InvalidPlan
	}
	seen := map[string]bool{}
	for _, f := range plan.Fields {
		if seen[f.Key] {
			return InvalidPlan
		}
		seen[f.Key] = true
		if _, e := compile(f.Pattern, []agentsdk.KnowledgeExtractionColumn{f.KnowledgeExtractionColumn}, false); e != nil {
			return e
		}
	}
	for _, table := range plan.Tables {
		if !keyPattern.MatchString(table.Key) || seen[table.Key] || table.After < 0 || table.After > 1000000 || table.Limit < 0 || table.Limit > 50 {
			return InvalidPlan
		}
		seen[table.Key] = true
		if _, e := compile(table.Pattern, table.Columns, true); e != nil {
			return e
		}
	}
	return nil
}
func emptyCell(column agentsdk.KnowledgeExtractionColumn) agentsdk.KnowledgeExtractionCell {
	code := "not_matched_in_available_text"
	if column.Required {
		code = "required_not_matched_in_available_text"
	}
	return agentsdk.KnowledgeExtractionCell{Key: column.Key, Type: column.Type, Status: "missing", Code: code, Evidence: []agentsdk.KnowledgeExtractionSpan{}}
}
func matchedCell(column agentsdk.KnowledgeExtractionColumn, passage int, text string, start, end int) agentsdk.KnowledgeExtractionCell {
	out := emptyCell(column)
	if start < 0 || end <= start {
		return out
	}
	raw := text[start:end]
	if len(raw) > 1024 {
		out.Status = "invalid"
		out.Code = "value_too_large"
		return out
	}
	out.Evidence = append(out.Evidence, agentsdk.KnowledgeExtractionSpan{Passage: passage, Start: start, End: end, Quote: raw})
	value, valid := normalize(raw, column.Type)
	if !valid {
		out.Status = "invalid"
		out.Code = "type_invalid"
		return out
	}
	if len(column.AllowedValues) > 0 {
		valid = false
		for _, allowed := range column.AllowedValues {
			expected, _ := normalize(allowed, column.Type)
			valid = valid || value == expected
		}
		if !valid {
			out.Status = "invalid"
			out.Code = "value_not_allowed"
			return out
		}
	}
	out.Status = "valid"
	out.Code = ""
	out.Value = &value
	return out
}

func Extract(ctx context.Context, plan agentsdk.KnowledgeExtractionPlan, passages []agentsdk.KnowledgeDocumentPassage) (agentsdk.KnowledgeExtractionData, error) {
	out := agentsdk.KnowledgeExtractionData{Fields: []agentsdk.KnowledgeExtractionCell{}, Tables: []agentsdk.KnowledgeExtractionRows{}, Warnings: []agentsdk.KnowledgeExtractionIssue{}, Coverage: agentsdk.KnowledgeExtractionCoverage{Basis: "available_parsed_passages", Passages: len(passages)}}
	if e := Validate(plan); e != nil {
		return out, e
	}
	if len(passages) == 0 || len(passages) > 1000 {
		return out, InvalidContent
	}
	total := 0
	for _, p := range passages {
		total += len(p.Content)
		if p.DocumentID == "" || len(p.Content) == 0 || !utf8.ValidString(p.Content) || strings.ContainsRune(p.Content, 0) || total > 640*1024 {
			return out, InvalidContent
		}
	}
	for fieldIndex, field := range plan.Fields {
		rule, _ := compile(field.Pattern, []agentsdk.KnowledgeExtractionColumn{field.KnowledgeExtractionColumn}, false)
		cell := emptyCell(field.KnowledgeExtractionColumn)
		seen := map[string]bool{}
		distinct := 0
		stop := false
		for index, p := range passages {
			if e := ctx.Err(); e != nil {
				return out, e
			}
			matches := rule.expression.FindAllStringSubmatchIndex(p.Content, 1001)
			if len(matches) > 1000 {
				matches = matches[:1000]
				out.Coverage.Limited = true
				cell.Status = "ambiguous"
				cell.Code = "match_limit"
				stop = true
			}
			for _, match := range matches {
				if match[0] == match[1] {
					return out, InvalidPlan
				}
				current := matchedCell(field.KnowledgeExtractionColumn, index, p.Content, match[2], match[3])
				if current.Status == "missing" {
					continue
				}
				identity := current.Status + ":" + current.Code
				if current.Value != nil {
					identity += ":" + *current.Value
				} else if len(current.Evidence) > 0 {
					identity += ":" + current.Evidence[0].Quote
				}
				if seen[identity] {
					continue
				}
				seen[identity] = true
				distinct++
				if distinct == 1 && !stop {
					cell = current
				} else {
					cell.Status = "ambiguous"
					cell.Value = nil
					cell.Code = "multiple_values"
					if len(cell.Evidence) < 3 {
						cell.Evidence = append(cell.Evidence, current.Evidence...)
					}
				}
				if distinct >= 3 {
					out.Coverage.Limited = true
					stop = true
					break
				}
			}
			if stop {
				break
			}
		}
		out.Fields = append(out.Fields, cell)
		if cell.Status == "missing" && strings.Contains(field.Pattern, `\\`) {
			out.Warnings = append(out.Warnings, backslashIssue("fields", fieldIndex))
		}
	}
	for tableIndex, table := range plan.Tables {
		rule, _ := compile(table.Pattern, table.Columns, true)
		limit := table.Limit
		if limit == 0 {
			limit = 20
		}
		result := agentsdk.KnowledgeExtractionRows{Key: table.Key, Rows: [][]agentsdk.KnowledgeExtractionCell{}, After: table.After, Complete: true}
		matched := 0
		done := false
		for index, p := range passages {
			if e := ctx.Err(); e != nil {
				return out, e
			}
			// At most remaining offset + requested rows + one lookahead. No total
			// count or complete-file claim is derived from this bounded scan.
			need := table.After - matched + limit - len(result.Rows) + 1
			if need < 1 {
				need = 1
			}
			scan := need
			if scan > 10001 {
				scan = 10001
			}
			matches := rule.expression.FindAllStringSubmatchIndex(p.Content, scan)
			if need > scan && len(matches) == scan {
				return out, Error("extraction_scan_limit")
			}
			for _, match := range matches {
				if match[0] == match[1] {
					return out, InvalidPlan
				}
				if matched < table.After {
					matched++
					continue
				}
				if len(result.Rows) == limit {
					next := table.After + len(result.Rows)
					result.NextAfter = &next
					result.Complete = false
					out.Coverage.Limited = true
					done = true
					break
				}
				row := make([]agentsdk.KnowledgeExtractionCell, 0, len(rule.columns))
				for i, column := range rule.columns {
					g := rule.groups[i]
					row = append(row, matchedCell(column, index, p.Content, match[g*2], match[g*2+1]))
				}
				result.Rows = append(result.Rows, row)
				matched++
			}
			if done {
				break
			}
		}
		out.Tables = append(out.Tables, result)
		if len(result.Rows) == 0 && strings.Contains(table.Pattern, `\\`) {
			out.Warnings = append(out.Warnings, backslashIssue("tables", tableIndex))
		}
	}
	if e := ctx.Err(); e != nil {
		return out, e
	}
	return out, nil
}
