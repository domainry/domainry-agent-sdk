package agentsdk

import (
	"context"
	"errors"
	"testing"
)

type sharedKnowledgeReadProbe struct {
	reader, producer ConversationAuthority
	calls            int
	denied           bool
}

func (p *sharedKnowledgeReadProbe) AuthorizeSharedKnowledgeResultRead(_ context.Context, _ ConversationKnowledgeResult, reader, producer ConversationAuthority) error {
	p.calls++
	if reader != p.reader || producer != p.producer || p.denied {
		return &Error{Class: "forbidden", Code: "agent.conversation.knowledge_access_denied"}
	}
	return nil
}

type oldKnowledgeReadProbe struct {
	reader ConversationAuthority
	calls  int
}

func (p *oldKnowledgeReadProbe) AuthorizeKnowledgeResultRead(_ context.Context, _ ConversationKnowledgeResult, a ConversationAuthority) error {
	p.calls++
	if a != p.reader {
		return &Error{Class: "forbidden", Code: "agent.conversation.knowledge_access_denied"}
	}
	return nil
}

func TestSharedKnowledgeReadPreservesAuthorityAndPrivateSourceBoundary(t *testing.T) {
	reader := ConversationAuthority{Known: true, RuntimeID: "runtime", WorkspaceID: "workspace", UserID: "reader", RoleKey: "read-only"}
	producer := reader
	producer.UserID, producer.RoleKey = "producer", "professional"
	p := &sharedKnowledgeReadProbe{reader: reader, producer: producer}
	saved := ConversationKnowledgeResult{}
	if err := AuthorizeSharedKnowledgeResultRead(t.Context(), p, saved, reader, producer); err != nil || p.calls != 1 {
		t.Fatal(err, p.calls)
	}
	for _, invalid := range []string{"unknown", "runtime", "workspace", "user", "private-result", "private-citation"} {
		bad, a := saved, reader
		switch invalid {
		case "unknown":
			a.Known = false
		case "runtime":
			a.RuntimeID = "foreign"
		case "workspace":
			a.WorkspaceID = "foreign"
		case "user":
			a.UserID = ""
		case "private-result":
			bad.ConversationID = "private"
		case "private-citation":
			bad.Citations = []ConversationCitation{{ConversationID: "private"}}
		}
		calls := p.calls
		if err := AuthorizeSharedKnowledgeResultRead(t.Context(), p, bad, a, producer); err == nil || p.calls != calls {
			t.Fatal("invalid source reached owner", invalid, err)
		}
		if _, err := SharedKnowledgeExtractionPassages(t.Context(), p, bad, a, producer); err == nil {
			t.Fatal("invalid extraction source reached owner", invalid)
		}
	}
	p.denied = true
	if err := AuthorizeSharedKnowledgeResultRead(t.Context(), p, saved, reader, producer); err == nil {
		t.Fatal("owner denial ignored")
	}
	old := &oldKnowledgeReadProbe{reader: reader}
	var coded *Error
	if err := AuthorizeSharedKnowledgeResultRead(t.Context(), old, saved, reader, producer); !errors.As(err, &coded) || coded.Code != KnowledgeResultReadUnsupportedCode || old.calls != 0 {
		t.Fatal("cross-user old owner invoked", err)
	}
	producer.UserID = reader.UserID
	if err := AuthorizeSharedKnowledgeResultRead(t.Context(), old, saved, reader, producer); err != nil || old.calls != 1 {
		t.Fatal("old same-user owner lost actual role", err)
	}
	p.producer = producer
	p.denied = false
	if err := AuthorizeSharedKnowledgeResultRead(t.Context(), p, saved, reader, producer); err != nil {
		t.Fatal("same-user different role lost shared port", err)
	}
}
