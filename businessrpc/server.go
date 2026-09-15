package businessrpc

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"time"

	sdk "github.com/domainry/domainry-agent-sdk"
	toolsdk "github.com/domainry/domainry-tools-sdk"
)

type ServerOptions struct {
	Scope   Scope
	Token   string
	Backend Backend
}

// NewHandler is opt-in server-to-server transport, never a browser API. The
// embedding service owns its listener, TLS, admission and shutdown lifecycle.
func NewHandler(o ServerOptions) (http.Handler, error) {
	if !o.Scope.valid() || !canonical(o.Token, 16384) || len(o.Token) < 16 || nilBackend(o.Backend) || !canonical(o.Backend.BusinessSourceIdentity(), 1024) {
		return nil, failure("bad_request")
	}
	descriptor := newDescriptor(o.Scope, o.Backend.BusinessSourceIdentity())
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.URL.RawQuery != "" || r.URL.RawPath != "" || r.URL.Path != CapabilitiesPath && r.URL.Path != CallPath {
			writeFailure(w, failure("not_found"))
			return
		}
		if len(r.Header.Values("Authorization")) != 1 || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+o.Token)) != 1 {
			writeJSON(w, 401, response{Error: &wireError{Class: "forbidden", Code: "agent.business_host.unauthenticated"}})
			return
		}
		if r.URL.Path == CapabilitiesPath {
			if r.Method != "GET" {
				w.WriteHeader(405)
				return
			}
			writeJSON(w, 200, descriptor)
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" || r.Header.Get("Content-Encoding") != "" {
			writeFailure(w, failure("bad_request"))
			return
		}
		raw, err := read(r.Body, MaxRequestBytes)
		if err != nil {
			writeJSON(w, 413, response{Error: &wireError{Class: "bad_request", Code: "agent.business_host.request_too_large"}})
			return
		}
		var in request
		if decode(raw, &in) != nil {
			writeFailure(w, failure("bad_request"))
			return
		}
		if in.Descriptor != descriptor || o.Backend.BusinessSourceIdentity() != descriptor.SourceIdentity {
			writeFailure(w, failure("conflict"))
			return
		}
		if !validAuthority(in.Authority, o.Scope) {
			writeFailure(w, failure("forbidden"))
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 65*time.Second)
		defer cancel()
		allowed, err := o.Backend.AuthorizeConversationExecution(ctx, sdk.ConversationExecutionAuthorizationRequest{Authority: in.Authority, Stage: "tool"})
		if err != nil {
			writeFailure(w, err)
			return
		}
		if !allowed {
			writeFailure(w, failure("forbidden"))
			return
		}
		// Report reads, analyses and business receipt reads use the owner's
		// application permission boundary.
		// They are host ports, not published conversation tools. They still passed
		// scope and current execution authorization above; ReportReader must resolve
		// current Identity again and invoke its actual Report owner.
		if !reportOperation(in.Operation) && !analysisOperation(in.Operation) && !sharedResultReadOperation(in.Operation) && in.Operation != "business_result_read" {
			tool, err := operationTool(in)
			if err != nil {
				writeFailure(w, err)
				return
			}
			if in.Operation != "tool_authorize" {
				granted, e := o.Backend.AuthorizeConversationTool(ctx, sdk.ConversationToolRequest{Authority: in.Authority, Definition: tool})
				if e != nil {
					writeFailure(w, e)
					return
				}
				if !granted.Granted {
					writeFailure(w, failure("forbidden"))
					return
				}
			}
		}
		result, err := dispatch(ctx, o.Backend, in)
		if err != nil {
			writeFailure(w, err)
			return
		}
		encoded, err := json.Marshal(result)
		if err != nil {
			writeFailure(w, failure("unavailable"))
			return
		}
		writeJSON(w, 200, response{Data: encoded})
	}), nil
}
func writeFailure(w http.ResponseWriter, err error) {
	class := "unavailable"
	var coded *sdk.Error
	if errors.As(err, &coded) {
		switch coded.Class {
		case "bad_request", "forbidden", "not_found", "conflict":
			class = coded.Class
		}
	}
	code := "agent.business_host." + class
	if coded != nil {
		code = safeCode(coded.Code, class)
	}
	status := map[string]int{"bad_request": 400, "forbidden": 403, "not_found": 404, "conflict": 409, "unavailable": 503}[class]
	writeJSON(w, status, response{Error: &wireError{Class: class, Code: code}})
}
func definition(key string) (sdk.ConversationToolDefinition, bool) {
	defs := append(sdk.BusinessConversationTools(), sdk.BusinessRelationConversationTools()...)
	defs = append(defs, sdk.BusinessActionConversationTools()...)
	defs = append(defs, sdk.BusinessWorkflowConversationTools()...)
	defs = append(defs, toolsdk.AnalysisDefinitions()...)
	defs = append(defs, toolsdk.ReportQueryDefinitions()...)
	defs = append(defs, toolsdk.MCPDefinitions()...)
	for _, d := range defs {
		if d.Key == key {
			return d, true
		}
	}
	return sdk.ConversationToolDefinition{}, false
}
func operationTool(in request) (sdk.ConversationToolDefinition, error) {
	keys := map[string]string{"catalog": "business_catalog", "query": "query_records", "get": "get_record", "related": "query_related_records", "action_authorize": "invoke_action", "action_invoke": "invoke_action", "action_reconcile": "invoke_action", "action_revalidate": "invoke_action", "workflow_authorize": "workflow_start", "workflow_start": "workflow_start", "workflow_reconcile": "workflow_start", "workflow_get": "workflow_get"}
	key := keys[in.Operation]
	if in.Operation == "revalidate" || in.Operation == "seal" || in.Operation == "workflow_revalidate" {
		var e sdk.ConversationBusinessEvidence
		if decode(in.Payload, &e) != nil {
			return sdk.ConversationToolDefinition{}, failure("bad_request")
		}
		key = e.Operation
		if in.Operation == "workflow_revalidate" && key != "workflow_start" && key != "workflow_get" {
			return sdk.ConversationToolDefinition{}, failure("bad_request")
		}
		if in.Operation != "workflow_revalidate" && (key == "workflow_start" || key == "invoke_action") {
			return sdk.ConversationToolDefinition{}, failure("bad_request")
		}
	}
	if in.Operation == "tool_authorize" {
		var r sdk.ConversationToolRequest
		if decode(in.Payload, &r) != nil || r.Authority != in.Authority {
			return sdk.ConversationToolDefinition{}, failure("bad_request")
		}
		d, ok := definition(r.Definition.Key)
		if !ok || r.Definition.Version != d.Version || r.Definition.ActionKey != d.ActionKey {
			return sdk.ConversationToolDefinition{}, failure("bad_request")
		}
		return d, nil
	}
	d, ok := definition(key)
	if !ok {
		return d, failure("bad_request")
	}
	return d, nil
}

func nilBackend(b Backend) bool {
	if b == nil {
		return true
	}
	v := reflect.ValueOf(b)
	switch v.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Func, reflect.Slice, reflect.Interface, reflect.Chan:
		return v.IsNil()
	}
	return false
}
