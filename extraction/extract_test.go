package extraction

import (
	"context"
	"strings"
	"testing"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

func TestExtractFieldsKeepExactValuesAndExplainMissingInvalidAmbiguous(t *testing.T) {
	content := "客户：青禾公司\n金额：9,007,199,254,740,993.25 元\n签署：2026年9月11日\n生效：2026-02-30\n确认：是\n币种：USD\n负责人：甲\n负责人：乙\n"
	field := func(key, kind, pattern string) agentsdk.KnowledgeExtractionField {
		return agentsdk.KnowledgeExtractionField{KnowledgeExtractionColumn: agentsdk.KnowledgeExtractionColumn{Key: key, Type: kind, Required: true}, Pattern: pattern}
	}
	plan := agentsdk.KnowledgeExtractionPlan{Fields: []agentsdk.KnowledgeExtractionField{
		field("customer", "text", `客户：([^\n]+)`), field("amount", "decimal", `金额：([0-9,.]+)`), field("signed", "date", `签署：([^\n]+)`), field("effective", "date", `生效：([^\n]+)`), field("approved", "boolean", `确认：([^\n]+)`), field("email", "text", `邮箱：([^\n]+)`), field("owner", "text", `负责人：([^\n]+)`), field("currency", "text", `币种：([^\n]+)`),
	}}
	plan.Fields[7].AllowedValues = []string{"CNY"}
	passages := []agentsdk.KnowledgeDocumentPassage{{DocumentID: "doc", Content: content}}
	got, err := Extract(t.Context(), plan, passages)
	if err != nil {
		t.Fatal(err)
	}
	want := []struct{ status, value, code string }{{"valid", "青禾公司", ""}, {"valid", "9007199254740993.25", ""}, {"valid", "2026-09-11", ""}, {"invalid", "", "type_invalid"}, {"valid", "true", ""}, {"missing", "", "required_not_matched_in_available_text"}, {"ambiguous", "", "multiple_values"}, {"invalid", "", "value_not_allowed"}}
	for i, expected := range want {
		cell := got.Fields[i]
		if cell.Status != expected.status || cell.Code != expected.code {
			t.Errorf("field %d status=%s code=%s", i, cell.Status, cell.Code)
		}
		if expected.status == "valid" {
			if cell.Value == nil || *cell.Value != expected.value {
				t.Errorf("field %d lost exact value", i)
			}
		} else if cell.Value != nil {
			t.Errorf("field %d invented a value", i)
		}
		for _, span := range cell.Evidence {
			if content[span.Start:span.End] != span.Quote {
				t.Error("citation not taken from actual text")
			}
		}
	}
	if got.Coverage.OriginalComplete || got.Coverage.Basis != "available_parsed_passages" {
		t.Fatal("parsed text claimed complete original")
	}
	// The same value appearing twice remains supported, not a disagreement.
	plan.Fields = plan.Fields[:1]
	passages = append(passages, passages[0])
	got, err = Extract(t.Context(), plan, passages)
	if err != nil || got.Fields[0].Status != "valid" {
		t.Fatal("same repeated value became ambiguous", err)
	}
}

func TestExtractTablesPageAcrossPassagesAndValidateEachCell(t *testing.T) {
	plan := agentsdk.KnowledgeExtractionPlan{Tables: []agentsdk.KnowledgeExtractionTable{{Key: "items", Pattern: `(?m)^\| (?P<item>[^|\n]+) \| (?P<quantity>[^|\n]+) \| (?P<price>[^|\n]+) \|$`, Columns: []agentsdk.KnowledgeExtractionColumn{{Key: "item", Type: "text", Required: true}, {Key: "quantity", Type: "integer", Required: true}, {Key: "price", Type: "decimal", Required: true}}, Limit: 2}}}
	passages := []agentsdk.KnowledgeDocumentPassage{{DocumentID: "sheet", Content: "| 铅笔 | 2 | 1.20 |\n| 纸张 | 9007199254740993 | 12.50 |"}, {DocumentID: "sheet", Content: "| 订书机 | unknown | 24.00 |\n| 擦子 | 1 | 2.30 |"}}
	first, err := Extract(t.Context(), plan, passages)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Tables[0].Rows) != 2 || first.Tables[0].Complete || first.Tables[0].NextAfter == nil || *first.Tables[0].NextAfter != 2 || !first.Coverage.Limited {
		t.Fatal("table pagination or partial marker")
	}
	if value := first.Tables[0].Rows[1][1].Value; value == nil || *value != "9007199254740993" {
		t.Fatal("integer precision lost")
	}
	plan.Tables[0].After = *first.Tables[0].NextAfter
	second, err := Extract(t.Context(), plan, passages)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Tables[0].Rows) != 2 || !second.Tables[0].Complete || second.Tables[0].NextAfter != nil || second.Tables[0].Rows[0][1].Status != "invalid" {
		t.Fatal("second page or invalid cell lost")
	}
	if *second.Tables[0].Rows[0][0].Value != "订书机" || second.Tables[0].Rows[0][0].Evidence[0].Passage != 1 {
		t.Fatal("rows repeated or source fabricated")
	}
}

func TestExtractRejectsUnsafePlansBoundsAndCancellation(t *testing.T) {
	column := agentsdk.KnowledgeExtractionColumn{Key: "value", Type: "text"}
	for _, pattern := range []string{`(.+)\1`, `(?<=label)(.+)`, `(a)(b)`, `abc`, strings.Repeat("x", 2049)} {
		if err := Validate(agentsdk.KnowledgeExtractionPlan{Fields: []agentsdk.KnowledgeExtractionField{{KnowledgeExtractionColumn: column, Pattern: pattern}}}); err == nil {
			t.Fatal("unsupported/unbounded pattern accepted")
		}
	}
	for _, pattern := range []string{`(?P<other>.+)`, `(?P<value>a)(?P<value>b)`} {
		if err := Validate(agentsdk.KnowledgeExtractionPlan{Tables: []agentsdk.KnowledgeExtractionTable{{Key: "rows", Pattern: pattern, Columns: []agentsdk.KnowledgeExtractionColumn{column}}}}); err == nil {
			t.Fatal("table schema/capture mismatch accepted")
		}
	}
	plan := agentsdk.KnowledgeExtractionPlan{Fields: []agentsdk.KnowledgeExtractionField{{KnowledgeExtractionColumn: column, Pattern: `(.)`}}}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := Extract(ctx, plan, []agentsdk.KnowledgeDocumentPassage{{DocumentID: "doc", Content: "x"}}); err != context.Canceled {
		t.Fatal("cancelled extraction proceeded", err)
	}
	if _, err := Extract(t.Context(), plan, []agentsdk.KnowledgeDocumentPassage{{DocumentID: "doc", Content: strings.Repeat("x", 641*1024)}}); err != InvalidContent {
		t.Fatal("content bound bypassed", err)
	}
	plan.Fields[0].Pattern = `()`
	if _, err := Extract(t.Context(), plan, []agentsdk.KnowledgeDocumentPassage{{DocumentID: "doc", Content: "x"}}); err != InvalidPlan {
		t.Fatal("zero-width row accepted", err)
	}
}

func TestNormalizedNumbersAndDatesNeverComputeOrGuess(t *testing.T) {
	for _, tc := range []struct {
		input, kind, want string
		ok                bool
	}{{"001,234.500", "decimal", "1234.5", true}, {"-0.00", "decimal", "0", true}, {"1,23", "decimal", "", false}, {"1e3", "decimal", "", false}, {"1+2", "decimal", "", false}, {"12.5", "integer", "", false}, {"2024/2/29", "date", "2024-02-29", true}, {"02/03/2026", "date", "", false}, {"2026-2-29", "date", "", false}, {"1", "boolean", "", false}} {
		value, ok := normalize(tc.input, tc.kind)
		if ok != tc.ok || value != tc.want {
			t.Errorf("normalize(%q,%s)=%q,%v", tc.input, tc.kind, value, ok)
		}
	}
}

func TestPositionalTableCapturesAndActionablePatternFeedback(t *testing.T) {
	columns := []agentsdk.KnowledgeExtractionColumn{{Key: "item", Type: "text"}, {Key: "quantity", Type: "integer"}, {Key: "price", Type: "decimal"}}
	plan := agentsdk.KnowledgeExtractionPlan{Tables: []agentsdk.KnowledgeExtractionTable{{Key: "items", Pattern: `(Pencil|Paper)[[:space:]]+([0-9]+)[[:space:]]+([0-9.]+)`, Columns: columns}}}
	got, err := Extract(t.Context(), plan, []agentsdk.KnowledgeDocumentPassage{{DocumentID: "docx", Content: "Pencil\n2\n1.20\nPaper\n3\n12.50"}})
	if err != nil || len(got.Tables[0].Rows) != 2 || *got.Tables[0].Rows[1][2].Value != "12.5" {
		t.Fatal("positional columns failed", err)
	}
	plan.Tables[0].Pattern = `(?P<item>Pencil) ([0-9]+) ([0-9.]+)`
	if Validate(plan) == nil || Explain(plan)[0].Code != "column_captures" {
		t.Fatal("mixed naming must explain the exact constraint")
	}
	plan = agentsdk.KnowledgeExtractionPlan{Fields: []agentsdk.KnowledgeExtractionField{{KnowledgeExtractionColumn: agentsdk.KnowledgeExtractionColumn{Key: "customer", Type: "text"}, Pattern: `Customer:\\s*(.+)`}}}
	got, err = Extract(t.Context(), plan, []agentsdk.KnowledgeDocumentPassage{{DocumentID: "docx", Content: "Customer: Qinghe Fixture"}})
	if err != nil || got.Fields[0].Status != "missing" || len(got.Warnings) != 1 || got.Warnings[0].Code != "literal_backslash_no_match" {
		t.Fatal("double escaping was silently rewritten or not explained", err)
	}
	plan.Fields[0].Pattern = `Customer: .+`
	if Validate(plan) == nil || Explain(plan)[0].Code != "capture_count" {
		t.Fatal("missing capture was not explained")
	}
}
