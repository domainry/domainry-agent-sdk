package agentsdk

import (
	"fmt"
	"strings"
)

// TaskAttachmentMaxBytes is the per-file admission limit for durable Agent
// tasks. Task attachment bytes are stored outside task state; only the frozen
// identity below is persisted with a run.
const TaskAttachmentMaxBytes int64 = 16 << 20

const TaskAttachmentMaxCount = 4

// AgentTaskAttachmentSchema is the task-owned contract for binary inputs.
// Binary bodies remain outside the JSON input schema and durable task state;
// this contract declares which immutable image or file attachments a task may
// accept after the ingress has resolved and authenticated their bytes.
type AgentTaskAttachmentSchema struct {
	MinItems            int      `json:"min_items"`
	MaxItems            int      `json:"max_items"`
	MaxBytes            int64    `json:"max_bytes"`
	AllowedContentTypes []string `json:"allowed_content_types"`
	ImageDetail         string   `json:"image_detail,omitempty"`
}

// ValidateAgentTaskAttachmentSchema validates the portable part of an Agent
// task attachment contract. A nil schema deliberately means that the task is
// text-only and accepts no attachments.
func ValidateAgentTaskAttachmentSchema(schema *AgentTaskAttachmentSchema) error {
	if schema == nil {
		return nil
	}
	if schema.MinItems < 0 || schema.MaxItems < 1 || schema.MinItems > schema.MaxItems || schema.MaxItems > TaskAttachmentMaxCount {
		return fmt.Errorf("attachment item bounds must satisfy 0 <= min_items <= max_items <= %d", TaskAttachmentMaxCount)
	}
	if schema.MaxBytes < 1 || schema.MaxBytes > TaskAttachmentMaxBytes {
		return fmt.Errorf("attachment max_bytes must be between 1 and %d", TaskAttachmentMaxBytes)
	}
	if len(schema.AllowedContentTypes) == 0 {
		return fmt.Errorf("allowed_content_types is required")
	}
	seen := map[string]bool{}
	for _, value := range schema.AllowedContentTypes {
		contentType := strings.TrimSpace(value)
		if value != contentType || value != strings.ToLower(value) || !SupportedTaskAttachmentContentType(contentType) {
			return fmt.Errorf("unsupported or non-canonical attachment content type %q", value)
		}
		if seen[contentType] {
			return fmt.Errorf("duplicate attachment content type %q", contentType)
		}
		seen[contentType] = true
	}
	detail := strings.TrimSpace(schema.ImageDetail)
	if detail == "" {
		detail = "auto"
	}
	if detail != "auto" && detail != "low" && detail != "high" {
		return fmt.Errorf("image_detail must be auto, low or high")
	}
	return nil
}

// SupportedTaskAttachmentContentType reports the binary formats that the
// built-in model task runner can translate to provider image/file blocks.
func SupportedTaskAttachmentContentType(contentType string) bool {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/png", "image/jpeg", "image/gif", "image/webp", "application/pdf":
		return true
	default:
		return false
	}
}

// TaskAttachment freezes one owner-private input file for an Agent task.
// BodyRef is an opaque private-storage reference. Data is populated only at
// the authenticated HTTP boundary and immediately before provider execution;
// it is never serialized into task state or service requests.
type TaskAttachment struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Bytes       int64  `json:"bytes"`
	SHA256      string `json:"sha256"`
	BodyRef     string `json:"body_ref,omitempty"`
	Detail      string `json:"detail,omitempty"` // auto, low or high
	Data        []byte `json:"-"`
}
