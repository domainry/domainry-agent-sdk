package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

const maxVerificationResponseBytes = 1 << 20

type ConversationSourceVerifierConfig struct {
	Endpoint    string
	AccessToken string
	UserAgent   string
	HTTPClient  *http.Client
	Timeout     time.Duration
}

func ConversationSourceVerifierConfigFromEnvironment() ConversationSourceVerifierConfig {
	return ConversationSourceVerifierConfig{
		Endpoint: strings.TrimSpace(os.Getenv("AGENT_ENDPOINT")), AccessToken: strings.TrimSpace(os.Getenv("AGENT_SERVICE_ACCESS_TOKEN")),
		UserAgent: strings.TrimSpace(os.Getenv("AGENT_USER_AGENT")),
	}
}

type conversationSourceVerifier struct {
	endpoint, accessToken, userAgent string
	client                           *http.Client
}

func NewConversationSourceVerifier(config ConversationSourceVerifierConfig) (agentsdk.ConversationSourceVerifier, error) {
	endpoint := strings.TrimRight(strings.TrimSpace(config.Endpoint), "/")
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, fmt.Errorf("Agent source verifier requires an absolute HTTP(S) endpoint")
	}
	if strings.TrimSpace(config.AccessToken) == "" {
		return nil, fmt.Errorf("Agent source verifier requires an access token")
	}
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: config.Timeout}
	}
	return &conversationSourceVerifier{endpoint: endpoint, accessToken: strings.TrimSpace(config.AccessToken), userAgent: strings.TrimSpace(config.UserAgent), client: client}, nil
}

func (verifier *conversationSourceVerifier) VerifyConversationSources(ctx context.Context, input agentsdk.ConversationSourceVerificationRequest) (agentsdk.ConversationSourceVerificationReceipt, error) {
	if err := input.Validate(); err != nil {
		return agentsdk.ConversationSourceVerificationReceipt{}, err
	}
	body, err := json.Marshal(agentsdk.ConversationRPCRequest{Authority: input.Reader, SourceVerification: input})
	if err != nil {
		return agentsdk.ConversationSourceVerificationReceipt{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, verifier.endpoint+"/agent/conversation-sources/verify", bytes.NewReader(body))
	if err != nil {
		return agentsdk.ConversationSourceVerificationReceipt{}, err
	}
	request.Header.Set("Authorization", "Bearer "+verifier.accessToken)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if verifier.userAgent != "" {
		request.Header.Set("User-Agent", verifier.userAgent)
	}
	response, err := verifier.client.Do(request)
	if err != nil {
		return agentsdk.ConversationSourceVerificationReceipt{}, err
	}
	defer response.Body.Close()
	limited := io.LimitReader(response.Body, maxVerificationResponseBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return agentsdk.ConversationSourceVerificationReceipt{}, err
	}
	if len(raw) > maxVerificationResponseBytes {
		return agentsdk.ConversationSourceVerificationReceipt{}, fmt.Errorf("Agent source verification response exceeds the limit")
	}
	if response.StatusCode != http.StatusOK {
		return agentsdk.ConversationSourceVerificationReceipt{}, fmt.Errorf("Agent source verification failed with HTTP %d", response.StatusCode)
	}
	var receipt agentsdk.ConversationSourceVerificationReceipt
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return agentsdk.ConversationSourceVerificationReceipt{}, fmt.Errorf("decode Agent source verification receipt: %w", err)
	}
	return receipt, nil
}

var _ agentsdk.ConversationSourceVerifier = (*conversationSourceVerifier)(nil)
