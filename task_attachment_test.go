package agentsdk

import "testing"

func TestValidateAgentTaskAttachmentSchema(t *testing.T) {
	valid := &AgentTaskAttachmentSchema{
		MinItems: 1, MaxItems: 1, MaxBytes: 5 << 20,
		AllowedContentTypes: []string{"image/png", "image/jpeg", "application/pdf"},
		ImageDetail:         "high",
	}
	if err := ValidateAgentTaskAttachmentSchema(valid); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAgentTaskAttachmentSchema(nil); err != nil {
		t.Fatal(err)
	}
	cases := map[string]*AgentTaskAttachmentSchema{
		"invalid count bounds": {
			MinItems: 2, MaxItems: 1, MaxBytes: 1, AllowedContentTypes: []string{"image/png"},
		},
		"too many files": {
			MinItems: 1, MaxItems: TaskAttachmentMaxCount + 1, MaxBytes: 1, AllowedContentTypes: []string{"image/png"},
		},
		"file too large": {
			MinItems: 1, MaxItems: 1, MaxBytes: TaskAttachmentMaxBytes + 1, AllowedContentTypes: []string{"image/png"},
		},
		"missing content types": {
			MinItems: 1, MaxItems: 1, MaxBytes: 1,
		},
		"unsupported content type": {
			MinItems: 1, MaxItems: 1, MaxBytes: 1, AllowedContentTypes: []string{"text/plain"},
		},
		"non canonical content type": {
			MinItems: 1, MaxItems: 1, MaxBytes: 1, AllowedContentTypes: []string{"Image/PNG"},
		},
		"duplicate content type": {
			MinItems: 1, MaxItems: 1, MaxBytes: 1, AllowedContentTypes: []string{"image/png", "image/png"},
		},
		"invalid image detail": {
			MinItems: 1, MaxItems: 1, MaxBytes: 1, AllowedContentTypes: []string{"image/png"}, ImageDetail: "original",
		},
	}
	for name, schema := range cases {
		t.Run(name, func(t *testing.T) {
			if err := ValidateAgentTaskAttachmentSchema(schema); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
