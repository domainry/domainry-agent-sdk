package agentsdk

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"unicode/utf8"

	toolsdk "github.com/domainry/domainry-tools-sdk"
)

const conversationBusinessCitationExcerptBytes = 3072

// ConversationBusinessRecordCitations creates immutable citations for a
// source-owned business projection. Products keep ownership of object schemas
// and tool definitions; this helper standardizes only evidence identity and
// bounded presentation. A non-positive limit includes every supplied record.
func ConversationBusinessRecordCitations(source, operation, objectKey string, records []ConversationBusinessRecord, limit int) []toolsdk.Citation {
	if len(records) == 0 || strings.TrimSpace(objectKey) == "" {
		return nil
	}
	source, operation, objectKey = strings.TrimSpace(source), strings.TrimSpace(operation), strings.TrimSpace(objectKey)
	if len(source) > 255 {
		digest := sha256.Sum256([]byte(source))
		source = "business:" + hex.EncodeToString(digest[:16])
	}
	if limit <= 0 || limit > len(records) {
		limit = len(records)
	}
	values := make([]toolsdk.Citation, 0, limit)
	for _, record := range records[:limit] {
		if strings.TrimSpace(record.ID) == "" {
			continue
		}
		raw, _ := json.Marshal(record.Data)
		digest := sha256.Sum256([]byte(strings.Join([]string{source, operation, objectKey, record.ID, record.Version, string(raw)}, "\x00")))
		values = append(values, toolsdk.Citation{
			ID: "business_" + hex.EncodeToString(digest[:16]), Source: source, Operation: operation,
			ObjectKey: objectKey, RecordID: record.ID, Title: conversationBusinessCitationTitle(record), Excerpt: strings.TrimSpace(conversationBusinessCitationText(string(raw), conversationBusinessCitationExcerptBytes)),
		})
	}
	return values
}

func conversationBusinessCitationTitle(record ConversationBusinessRecord) string {
	for _, key := range []string{"name", "subject", "title"} {
		var value string
		if json.Unmarshal(record.Data[key], &value) == nil && strings.TrimSpace(value) != "" {
			return conversationBusinessCitationText(strings.TrimSpace(value), 512)
		}
	}
	return ""
}

func conversationBusinessCitationText(value string, maxBytes int) string {
	if len(value) <= maxBytes {
		return value
	}
	value = value[:maxBytes]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value
}

// ConversationCitationsEqual compares the exact ordered evidence envelope.
// Nil and an empty slice are equivalent after durable JSON normalization.
func ConversationCitationsEqual(left, right []toolsdk.Citation) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		leftRaw, leftErr := json.Marshal(left[index])
		rightRaw, rightErr := json.Marshal(right[index])
		if leftErr != nil || rightErr != nil || string(leftRaw) != string(rightRaw) {
			return false
		}
	}
	return true
}
