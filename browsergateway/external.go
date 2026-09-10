package browsergateway

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	action "github.com/domainry/domainry-foundation/action"
	"github.com/domainry/domainry-foundation/modulehttp"
	identity "github.com/domainry/domainry-identity-sdk"
)

// External credentials are interpreted by the selected binding. The Agent
// boundary never issues a substitute session or uses the installation as a
// user's Workspace.
func (b *boundary) externalIdentity(r *http.Request) (identity.RequestIdentity, error) {
	source := b.options.Identity.(identity.RequestCredentialBinding)
	credential, err := source.ReadAccessCredential(r)
	if err != nil {
		return identity.RequestIdentity{}, err
	}
	if credential == "" {
		return identity.RequestIdentity{}, errors.New("missing external credential")
	}
	authenticator := b.options.Identity.(identity.PrincipalAuthenticationBinding).PrincipalAuthenticator()
	p, err := authenticator.Authenticate(r.Context(), credential)
	if err != nil {
		return identity.RequestIdentity{}, err
	}
	if !p.Known || p.WorkspaceID == "" || p.UserID == "" || p.User.ID != p.UserID {
		return identity.RequestIdentity{}, errors.New("external principal is incomplete")
	}
	return identity.RequestIdentity{AccessToken: credential, Principal: p}, nil
}

func (b *boundary) mountExternal(mux *http.ServeMux) error {
	source, ok := b.options.Identity.(identity.PrincipalAuthenticationBinding)
	if !ok || source.PrincipalAuthenticator() == nil {
		return fmt.Errorf("external principal authenticator is required")
	}
	if _, ok := b.options.Identity.(identity.RequestCredentialBinding); !ok {
		return fmt.Errorf("external request credential binding is required")
	}
	provider, ok := b.options.Identity.(modulehttp.Provider)
	if !ok {
		return fmt.Errorf("external browser discovery adapter is required")
	}
	discovery := false
	for _, adapter := range provider.HTTPAdapters() {
		for _, route := range adapter.Routes() {
			if !strings.HasPrefix(route.Action.Key, "identity.external.") {
				continue
			}
			handler := adapter.Handler()
			switch route.Action.Authorization.Strategy {
			case action.AuthorizationAnonymous:
			case action.AuthorizationAuthenticated:
				handler = b.authenticated(false, handler)
			default:
				return fmt.Errorf("unsupported external browser authorization strategy")
			}
			if route.Pattern() == "GET /auth/external/config" {
				discovery = true
			}
			mux.Handle(route.Pattern(), handler)
		}
	}
	if !discovery {
		return fmt.Errorf("external browser configuration route is required")
	}
	return nil
}
