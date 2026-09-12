package businessrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	sdk "github.com/domainry/domainry-agent-sdk"
)

type ClientOptions struct {
	BaseURL                string
	Token                  string
	Scope                  Scope
	ExpectedSourceIdentity string
	ExpectedContractSHA256 string
	HTTPClient             *http.Client
}

// Client's immutable descriptor is checked at startup and on every service call.
// The caller must use the same trusted Identity domain as the business service.
type Client struct {
	base, token string
	descriptor  Descriptor
	http        *http.Client
}

func Open(ctx context.Context, o ClientOptions) (*Client, error) {
	u, err := url.Parse(o.BaseURL)
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.RawPath != "" || u.Opaque != "" || u.Path != "" && u.Path != "/" || !canonical(o.BaseURL, 4096) || strings.Contains(o.BaseURL, "\\") {
		return nil, failure("bad_request")
	}
	if strings.HasSuffix(u.Host, ":") {
		return nil, failure("bad_request")
	}
	if port := u.Port(); port != "" {
		n, e := strconv.Atoi(port)
		if e != nil || n < 1 || n > 65535 {
			return nil, failure("bad_request")
		}
	}
	local := strings.EqualFold(u.Hostname(), "localhost")
	if ip := net.ParseIP(u.Hostname()); ip != nil {
		local = ip.IsLoopback()
	}
	if u.Scheme != "https" && !(u.Scheme == "http" && local) {
		return nil, failure("bad_request")
	}
	if !o.Scope.valid() || !canonical(o.Token, 16384) || len(o.Token) < 16 || !canonical(o.ExpectedSourceIdentity, 1024) || o.ExpectedContractSHA256 != ContractSHA256() {
		return nil, failure("bad_request")
	}
	client := http.Client{Timeout: 65 * time.Second}
	if o.HTTPClient != nil {
		client = *o.HTTPClient
		if client.Timeout <= 0 || client.Timeout > 65*time.Second {
			client.Timeout = 65 * time.Second
		}
	}
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	client.Jar = nil
	u.Path = ""
	c := &Client{base: u.String(), token: o.Token, descriptor: newDescriptor(o.Scope, o.ExpectedSourceIdentity), http: &client}
	raw, status, err := c.do(ctx, "GET", CapabilitiesPath, nil)
	var got Descriptor
	if err != nil || status != 200 || decode(raw, &got) != nil || got != c.descriptor {
		return nil, failure("unavailable")
	}
	return c, nil
}
func (c *Client) BusinessSourceIdentity() string { return c.descriptor.SourceIdentity }
func (c *Client) Descriptor() Descriptor         { return c.descriptor }
func (c *Client) do(ctx context.Context, method, path string, body []byte) ([]byte, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, failure("unavailable")
	}
	r, err := http.NewRequestWithContext(ctx, method, c.base+path, bytes.NewReader(body))
	if err != nil {
		return nil, 0, failure("unavailable")
	}
	// Even a supplied transport must not receive a replayable body from us.
	r.GetBody = nil
	r.Header.Set("Authorization", "Bearer "+c.token)
	r.Header.Set("Accept", "application/json")
	if method == "POST" {
		r.Header.Set("Content-Type", "application/json")
	}
	out, err := c.http.Do(r)
	if err != nil {
		return nil, 0, failure("unavailable")
	}
	defer out.Body.Close()
	media, _, err := mime.ParseMediaType(out.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		return nil, out.StatusCode, failure("unavailable")
	}
	raw, err := read(out.Body, MaxResponseBytes)
	return raw, out.StatusCode, err
}
func (c *Client) call(ctx context.Context, op string, a sdk.ConversationAuthority, payload, out any) error {
	if !validAuthority(a, c.descriptor.Scope) {
		return failure("forbidden")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return failure("bad_request")
	}
	encoded, err := json.Marshal(request{Descriptor: c.descriptor, Operation: op, Authority: a, Payload: raw})
	if err != nil || len(encoded) > MaxRequestBytes {
		return failure("bad_request")
	}
	raw, status, err := c.do(ctx, "POST", CallPath, encoded)
	if err != nil {
		return err
	}
	var result response
	if decode(raw, &result) != nil {
		return failure("unavailable")
	}
	if status != 200 {
		if result.Error != nil {
			class := result.Error.Class
			want := map[string]int{"bad_request": 400, "forbidden": 403, "not_found": 404, "conflict": 409, "unavailable": 503}[class]
			if want == status {
				return &sdk.Error{Class: class, Code: safeCode(result.Error.Code, class)}
			}
		}
		return failure("unavailable")
	}
	if result.Error != nil || decode(result.Data, out) != nil {
		return failure("unavailable")
	}
	return nil
}
func safeCode(code, class string) string {
	switch code {
	case "business_action_changed", "business_action_invalid", "business_record_version_conflict", "business_action_assurance_required", "business_action_receipt_conflict":
		return code
	case "backend.report.analysis.result_limit_exceeded","backend.report.analysis.spec_invalid","backend.report.analysis.arithmetic_limit","backend.report.analysis.source_changed","backend.report.analysis.result_stale_or_invalid":
		return code
	}
	return "agent.business_host." + class
}

var _ sdk.ConversationBusinessSource = (*Client)(nil)
var _ sdk.ConversationBusinessRelationSource = (*Client)(nil)
var _ sdk.ConversationBusinessActionSource = (*Client)(nil)
var _ sdk.ConversationBusinessWorkflowSource = (*Client)(nil)
var _ sdk.ConversationBusinessEvidenceSealer = (*Client)(nil)
var _ sdk.ConversationToolAuthorizer = (*Client)(nil)
