package browsergateway

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"sort"
	"strings"
	"unicode"

	action "github.com/domainry/domainry-foundation/action"
	"github.com/domainry/domainry-foundation/modulehttp"
	identity "github.com/domainry/domainry-identity-sdk"
)

func (b *boundary) mountModules(mux *http.ServeMux) error {
	if len(b.options.ModuleAdapters) > 0 && b.options.Identity.Principals() == nil {
		return fmt.Errorf("browser modules require live Identity principal resolution")
	}
	seen := map[string]bool{}
	for _, adapter := range b.options.ModuleAdapters {
		if err := modulehttp.ValidateAdapter(adapter); err != nil {
			return fmt.Errorf("browser module adapter: %w", err)
		}
		owner := adapter.Owner()
		if owner == "" || strings.IndexFunc(owner, func(r rune) bool { return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-') }) >= 0 || owner == "agent" || owner == "app" || owner == "auth" {
			return fmt.Errorf("browser module must have a separate owner namespace")
		}
		for _, route := range adapter.Routes() {
			if route.Action.HTTP == nil || !strings.HasPrefix(route.Action.HTTP.RouteTemplate, "/"+owner+"/") || route.Action.Authorization.Strategy != action.AuthorizationAuthenticated || route.Action.Permission == nil || seen[route.Pattern()] {
				return fmt.Errorf("browser module route must declare its owner path and authenticated action permission")
			}
			seen[route.Pattern()] = true
			if err := mountOwnerRoute(mux, route.Pattern(), b.authenticated(true, b.modulePrincipal(adapter.Handler()))); err != nil {
				return err
			}
		}
	}
	return nil
}

// net/http detects ambiguous parameter patterns during registration. Invalid
// owner declarations are startup errors, without partially serving the module.
func mountOwnerRoute(mux *http.ServeMux, pattern string, handler http.Handler) (err error) {
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("browser module route is invalid or conflicts with another declaration")
		}
	}()
	mux.Handle(pattern, handler)
	return nil
}

func (b *boundary) modulePrincipal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current, ok := identity.RequestIdentityFromContext(r.Context())
		if !ok {
			writeCode(w, 401, "agent.web.login_required")
			return
		}
		principal := current.Principal
		resolved, err := b.options.Identity.Principals().Resolve(r.Context(), identity.PrincipalResolutionRequest{
			Application: identity.ApplicationScope{WorkspaceID: identity.WorkspaceID(principal.WorkspaceID), ApplicationKey: identity.ApplicationKey(b.options.ApplicationKey)},
			SubjectID:   identity.SubjectID(principal.UserID), RoleKey: principal.RoleKey,
		})
		if err != nil || !resolved.Principal.Known || resolved.Principal.MustChangePassword || resolved.Principal.WorkspaceID != principal.WorkspaceID || resolved.Principal.UserID != principal.UserID {
			writeCode(w, 403, "agent.web.module_access_denied")
			return
		}
		// The cookie projection alone contains no action/data policy. Resolve the
		// live bundle per request; the owner adapter decides its exact action scope.
		resolved.Principal.AccessBundle = &resolved.AccessBundle
		current.Principal = resolved.Principal
		next.ServeHTTP(w, r.WithContext(identity.WithRequestIdentity(r.Context(), current)))
	})
}

func (b *boundary) mountNavigationFiles(mux *http.ServeMux) (map[string]bool, error) {
	allowed := map[string]bool{}
	for route, filename := range b.options.NavigationFiles {
		if route == "/" || !strings.HasPrefix(route, "/") || path.Clean(route) != route || strings.ContainsAny(route, "{}?#%\\") || strings.IndexFunc(route, unicode.IsSpace) >= 0 || strings.IndexFunc(route, unicode.IsControl) >= 0 || strings.HasSuffix(route, "/") || !fs.ValidPath(filename) || path.Ext(filename) != ".html" {
			return nil, fmt.Errorf("navigation landing page requires an exact path and static HTML file")
		}
		// Reserve service namespaces for their authenticated handlers.
		for _, prefix := range []string{"/app", "/auth", "/agent"} {
			if route == prefix || strings.HasPrefix(route, prefix+"/") {
				return nil, fmt.Errorf("navigation path conflicts with a service namespace")
			}
		}
		for _, adapter := range b.options.ModuleAdapters {
			if route == "/"+adapter.Owner() || strings.HasPrefix(route, "/"+adapter.Owner()+"/") {
				return nil, fmt.Errorf("navigation path conflicts with a module namespace")
			}
		}
		content, err := fs.ReadFile(b.options.Files, filename)
		if err != nil || len(content) > 65536 {
			return nil, fmt.Errorf("navigation landing page is unavailable or too large")
		}
		allowed[route] = true
		mux.HandleFunc("GET "+route, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(content)
		})
	}
	return allowed, nil
}

func (b *boundary) moduleOwners() []string {
	owners := map[string]bool{}
	for _, adapter := range b.options.ModuleAdapters {
		if adapter != nil {
			owners[adapter.Owner()] = true
		}
	}
	result := []string{}
	for owner := range owners {
		result = append(result, owner)
	}
	sort.Strings(result)
	return result
}
