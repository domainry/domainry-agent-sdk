// Package browsergateway is the same-origin browser boundary over the Identity and Agent
// modules. Identity owns credentials, token verification and session revocation.
package browsergateway

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"strings"
	"time"

	agentsdk "github.com/domainry/domainry-agent-sdk"
	actioncontract "github.com/domainry/domainry-foundation/action"
	"github.com/domainry/domainry-foundation/modulehttp"
	identitysdk "github.com/domainry/domainry-identity-sdk"
	"github.com/domainry/domainry-identity-sdk/browsergateway"
)

const accessCookie = "domainry_agent_access"

type Options struct {
	// CookiePrefix isolates independently deployed products on the same hostname.
	CookiePrefix                                          string
	Identity                                              identitysdk.Binding
	Agent                                                 agentsdk.Binding
	RuntimeID, WorkspaceID, ApplicationKey, Origin, Model string
	Files                                                 fs.FS
	// Optional in-process host router. The mounted conversation routes still
	// pass the browser boundary first, then retain the host's admission/audit
	// middleware. Only the verified cookie's token is forwarded server-side.
	ConversationHandler http.Handler
	// Product routes use the same current-session, password and scope checks.
	ApplicationRoutes map[string]http.Handler
	// Module adapters keep their owner routes and action authorization. The
	// browser host supplies only a live Identity principal and same-origin guard.
	ModuleAdapters []modulehttp.Adapter
	// NavigationFiles maps explicit cross-site GET landing paths to static HTML
	// files in Files. These pages receive no identity and cannot execute commands.
	NavigationFiles map[string]string
}

func Scope(runtimeID, workspaceID, userID string) string {
	raw, _ := json.Marshal([]string{runtimeID, workspaceID, userID})
	hash := sha256.Sum256(raw)
	return hex.EncodeToString(hash[:])
}

func NewHandler(options Options) (http.Handler, error) {
	origin, err := url.Parse(options.Origin)
	if err != nil || origin.Host == "" || origin.User != nil || origin.RawQuery != "" || origin.Fragment != "" || origin.Path != "" || (origin.Scheme != "https" && origin.Scheme != "http") {
		return nil, fmt.Errorf("web origin must be an exact HTTP(S) origin without a path")
	}
	if options.RuntimeID == "" || options.Identity == nil || options.Agent == nil || options.Files == nil {
		return nil, fmt.Errorf("web host modules, runtime identity and UI are required")
	}
	if options.CookiePrefix != "" {
		if err := (&http.Cookie{Name: options.CookiePrefix + "_access"}).Valid(); err != nil {
			return nil, fmt.Errorf("invalid product cookie prefix")
		}
	}
	b := &boundary{options: options, secure: origin.Scheme == "https", external: options.Identity.Descriptor().Mode == identitysdk.DeploymentModeExternal}
	mux := http.NewServeMux()
	if b.external {
		if err := b.mountExternal(mux); err != nil {
			return nil, err
		}
	} else {
		gateway, err := browsergateway.New(options.Identity, browsergateway.Config{
			ApplicationKey: identitysdk.ApplicationKey(options.ApplicationKey), DefaultWorkspaceID: identitysdk.WorkspaceID(options.WorkspaceID),
			Cookie: browsergateway.CookieConfig{Name: b.refreshCookie(), Path: "/auth", Secure: origin.Scheme == "https", SameSite: http.SameSiteStrictMode},
		})
		if err != nil {
			return nil, err
		}
		// Mount the SDK-owned session handlers only. Identity management and service
		// credential interfaces are deliberately not part of this chat application.
		for pattern, handler := range map[string]http.HandlerFunc{
			"POST /auth/login": gateway.Login, "POST /auth/refresh": gateway.Refresh,
			"POST /auth/logout": gateway.Logout, "GET /auth/session": gateway.Session,
			"POST /auth/password/change": gateway.ChangePassword,
		} {
			mux.Handle(pattern, b.cookieTransport(handler))
		}
	}
	mux.HandleFunc("GET /app/config", func(w http.ResponseWriter, r *http.Request) {
		workspace, authentication := options.WorkspaceID, "managed"
		if b.external {
			workspace, authentication = "", "external"
		}
		writeJSON(w, 200, map[string]any{"mode": "identity", "authentication": authentication, "workspace_id": workspace, "runtime_id": options.RuntimeID, "modules": b.moduleOwners()})
	})
	mux.Handle("GET /app/session", b.authenticated(false, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, _ := identitysdk.RequestIdentityFromContext(r.Context())
		p := identity.Principal
		ready := false
		if binding, ok := options.Agent.(agentsdk.ConversationBinding); ok {
			if status, ok := binding.Conversations().(agentsdk.ConversationStatusProvider); ok {
				ready = status.ConversationReady(r.Context()) == nil
			}
		}
		writeJSON(w, 200, map[string]any{"mode": "identity", "runtime_id": options.RuntimeID, "workspace_id": p.WorkspaceID, "user_id": p.UserID, "name": p.User.Name, "scope": Scope(options.RuntimeID, p.WorkspaceID, p.UserID), "must_change_password": p.MustChangePassword, "ready": ready, "model": options.Model, "modules": b.moduleOwners()})
	})))
	provider, ok := options.Agent.(modulehttp.Provider)
	if !ok {
		return nil, fmt.Errorf("Agent module HTTP adapters are required")
	}
	mounted := 0
	for _, adapter := range provider.HTTPAdapters() {
		for _, route := range adapter.Routes() {
			if !strings.HasPrefix(route.Action.Key, agentsdk.ConversationActionPrefix) {
				continue
			}
			if route.Action.Authorization.Strategy != actioncontract.AuthorizationAuthenticated {
				return nil, fmt.Errorf("unsupported conversation authorization strategy")
			}
			handler := adapter.Handler()
			if options.ConversationHandler != nil {
				handler = http.HandlerFunc(b.forwardConversation)
			}
			mux.Handle(route.Pattern(), b.authenticated(true, handler))
			mounted++
		}
	}
	if mounted == 0 {
		return nil, fmt.Errorf("Agent conversation routes are required")
	}
	for pattern, handler := range options.ApplicationRoutes {
		parts := strings.SplitN(pattern, " ", 2)
		path := parts[len(parts)-1]
		if !strings.HasPrefix(path, "/app/product/") || handler == nil {
			return nil, fmt.Errorf("product routes must live under /app/product/")
		}
		mux.Handle(pattern, b.authenticated(true, handler))
	}
	if err := b.mountModules(mux); err != nil {
		return nil, err
	}
	navigation, err := b.mountNavigationFiles(mux)
	if err != nil {
		return nil, err
	}
	mux.Handle("/", http.FileServer(http.FS(options.Files)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
		unsafe := r.Method != http.MethodGet && r.Method != http.MethodHead
		landing := r.Method == http.MethodGet && navigation[r.URL.Path]
		if r.Host != origin.Host || (!landing && r.Header.Get("Sec-Fetch-Site") == "cross-site") || (!landing && r.Header.Get("Origin") != "" && r.Header.Get("Origin") != options.Origin) || (unsafe && r.Header.Get("Origin") != options.Origin) {
			writeCode(w, 403, "agent.web.origin_required")
			return
		}
		mux.ServeHTTP(w, r)
	}), nil
}

func (b *boundary) forwardConversation(w http.ResponseWriter, r *http.Request) {
	identity, ok := identitysdk.RequestIdentityFromContext(r.Context())
	if !ok || identity.AccessToken == "" {
		writeCode(w, 401, "agent.web.login_required")
		return
	}
	// A browser session establishes the owner, but does not contain the host's
	// access bundle. Do not let that partial identity short-circuit the host's
	// SDK authentication middleware: it must resolve the verified credential.
	forward := r.Clone(identitysdk.WithRequestIdentity(r.Context(), identitysdk.RequestIdentity{}))
	forward.Header.Set("Authorization", "Bearer "+identity.AccessToken)
	forward.Header.Set("X-Workspace-ID", identity.Principal.WorkspaceID)
	if !b.external {
		forward.Header.Del("Cookie")
	}
	b.options.ConversationHandler.ServeHTTP(w, forward)
}

type boundary struct {
	options  Options
	secure   bool
	external bool
}

func (b *boundary) identity(r *http.Request) (identitysdk.RequestIdentity, error) {
	if b.external {
		return b.externalIdentity(r)
	}
	var result identitysdk.RequestIdentity
	cookie, err := r.Cookie(b.accessCookie())
	if err != nil || cookie.Value == "" {
		return result, errors.New("missing access cookie")
	}
	// CurrentSession checks the live Identity session on every request, including
	// revocation and password state. No host-side principal cache can outlive logout.
	verified, err := b.options.Identity.Tokens().Verify(r.Context(), identitysdk.VerifyTokenRequest{AccessToken: cookie.Value, Audience: identitysdk.ApplicationKey(b.options.ApplicationKey)})
	if err != nil {
		return result, err
	}
	session, err := b.options.Identity.Authentication().CurrentSession(r.Context(), identitysdk.CurrentSessionRequest{AccessToken: cookie.Value})
	if err != nil {
		return result, err
	}
	if session.WorkspaceID != identitysdk.WorkspaceID(b.options.WorkspaceID) || session.SubjectID == "" || session.User.ID != string(session.SubjectID) || session.WorkspaceID != verified.WorkspaceID || session.SubjectID != verified.SubjectID {
		return result, errors.New("Identity session scope mismatch")
	}
	// This projection is sufficient for browser owner/scope checks. Tool and
	// interaction permissions are checked by their services; a mounted host
	// router authenticates the credential to obtain its full access bundle.
	result = identitysdk.RequestIdentity{AccessToken: cookie.Value, Principal: identitysdk.Principal{Known: true, WorkspaceID: string(session.WorkspaceID), UserID: string(session.SubjectID), RoleKey: session.DefaultRole, User: session.User, MustChangePassword: session.MustChangePassword}}
	return result, nil
}

func (b *boundary) authenticated(requireScope bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, err := b.identity(r)
		if err != nil {
			writeCode(w, 401, "agent.web.login_required")
			return
		}
		p := identity.Principal
		if requireScope {
			if p.MustChangePassword {
				writeCode(w, 403, "agent.web.password_change_required")
				return
			}
			scope := r.Header.Get("X-Agent-Scope")
			stream := strings.HasSuffix(r.URL.Path, "/events/stream")
			if stream {
				scope = r.URL.Query().Get("scope")
			}
			if scope != Scope(b.options.RuntimeID, p.WorkspaceID, p.UserID) {
				writeCode(w, 409, "agent.web.identity_changed")
				return
			}
			if stream {
				w = &sessionStreamWriter{ResponseWriter: w, check: func() error {
					current, err := b.identity(r)
					if err != nil {
						return err
					}
					if current.Principal.MustChangePassword {
						return errors.New("password change required")
					}
					if current.Principal.WorkspaceID != p.WorkspaceID || current.Principal.UserID != p.UserID {
						return errors.New("session owner changed")
					}
					return nil
				}}
			}
		}
		next.ServeHTTP(w, r.WithContext(identitysdk.WithRequestIdentity(r.Context(), identity)))
	})
}

type sessionStreamWriter struct {
	http.ResponseWriter
	check func() error
}

func (w *sessionStreamWriter) Write(data []byte) (int, error) {
	if err := w.check(); err != nil {
		return 0, err
	}
	return w.ResponseWriter.Write(data)
}
func (w *sessionStreamWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Cookie transport preserves Identity's rotating refresh-cookie protocol and
// carries its access token in an additional HttpOnly cookie so native SSE can
// authenticate. It stores no sessions and returns no credentials to JavaScript.
func (b *boundary) cookieTransport(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.Clone(r.Context())
		r.Header.Del("Authorization")
		if cookie, err := r.Cookie(b.accessCookie()); err == nil {
			r.Header.Set("Authorization", "Bearer "+cookie.Value)
		}
		out := &bufferedResponse{header: make(http.Header), status: 200}
		next(out, r)
		var body map[string]json.RawMessage
		if out.body.Len() > 0 {
			_ = json.Unmarshal(out.body.Bytes(), &body)
		}
		if raw, ok := body["access_token"]; ok && out.status < 300 {
			var token, expires string
			_ = json.Unmarshal(raw, &token)
			_ = json.Unmarshal(body["expires_at"], &expires)
			deadline, err := time.Parse(time.RFC3339, expires)
			if err != nil || token == "" || !deadline.After(time.Now()) {
				writeCode(w, 502, "agent.web.session_invalid")
				return
			}
			http.SetCookie(w, &http.Cookie{Name: b.accessCookie(), Value: token, Path: "/", HttpOnly: true, Secure: b.secure, SameSite: http.SameSiteStrictMode, Expires: deadline})
			delete(body, "access_token")
			out.body.Reset()
			_ = json.NewEncoder(&out.body).Encode(body)
		}
		if r.URL.Path == "/auth/logout" {
			http.SetCookie(w, &http.Cookie{Name: b.accessCookie(), Path: "/", HttpOnly: true, Secure: b.secure, SameSite: http.SameSiteStrictMode, MaxAge: -1})
		}
		for key, values := range out.header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(out.status)
		_, _ = w.Write(out.body.Bytes())
	})
}

type bufferedResponse struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func (w *bufferedResponse) Header() http.Header            { return w.header }
func (w *bufferedResponse) WriteHeader(status int)         { w.status = status }
func (w *bufferedResponse) Write(data []byte) (int, error) { return w.body.Write(data) }
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
func writeCode(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"code": code})
}

func (b *boundary) accessCookie() string {
	if b.options.CookiePrefix != "" {
		return b.options.CookiePrefix + "_access"
	}
	return accessCookie
}
func (b *boundary) refreshCookie() string {
	if b.options.CookiePrefix != "" {
		return b.options.CookiePrefix + "_refresh"
	}
	return "domainry_agent_refresh"
}
