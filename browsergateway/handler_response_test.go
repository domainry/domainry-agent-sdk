package browsergateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/domainry/domainry-foundation/httpresponse"
	"github.com/domainry/domainry-foundation/requestcontext"
)

func TestWriteCodeUsesCanonicalHTTPErrorResponse(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/app/session", nil)
	request = request.WithContext(requestcontext.WithRequestID(request.Context(), "request-1"))
	response := httptest.NewRecorder()
	writeCode(response, request, http.StatusUnauthorized, "agent.web.login_required")
	var failure httpresponse.ErrorResponse
	if err := json.NewDecoder(response.Body).Decode(&failure); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusUnauthorized || failure.Code != "agent.web.login_required" || failure.Message == "" || failure.RequestID != "request-1" || failure.Params != nil {
		t.Fatalf("unexpected canonical failure: status=%d response=%+v", response.Code, failure)
	}
}
