package eml

import (
	"context"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

const (
	baseUrl      = "https://eml.com"
	clientId     = "client"
	clientSecret = "secret"
)

var ctx = context.Background()

func mockTokenResponse(t *testing.T) {
	httpmock.RegisterResponder("POST", "https://eml.com/3.0/token", func(req *http.Request) (*http.Response, error) {
		bodyBytes, _ := ioutil.ReadAll(req.Body)
		assert.Equal(t, "grant_type=client_credentials", string(bodyBytes), "Malformed token body %s", string(bodyBytes))
		acceptHeader := req.Header.Get("Accept")
		assert.Equal(t, "application/vnd.eml+json", acceptHeader, "expected eml vendor accept content type, got %s", acceptHeader)
		authHeader := req.Header.Get("Authorization")
		assert.Equal(t, "Basic Y2xpZW50OnNlY3JldA==", authHeader, "expected basic auth header, got %s", authHeader)
		return httpmock.NewJsonResponse(200, TokenResponse{AccessToken: "aToken", TokenType: "bearer", ExpiresIn: 60 * 60})
	})
}

// Test 2 EML calls produces a single token request, validating headers
func Test_emlStore_GetAccount(t *testing.T) {
	e := &Settings{EmlRestId: clientId, EmlHostUrl: baseUrl}
	store := &emlStore{_restSecret: clientSecret, _env: e}
	store.lazyInit(ctx)

	httpmock.ActivateNonDefault(store._lazyClient.GetClient())
	defer httpmock.DeactivateAndReset()
	mockTokenResponse(t)

	accSummary := &AccountSummary{ExternalAccountId: "eaid"}
	httpmock.RegisterResponder("POST", "https://eml.com/3.0/accounts", func(req *http.Request) (*http.Response, error) {
		contentTypeHeader := req.Header.Get("Content-Type")
		assert.Equal(t, "application/vnd.eml+json", contentTypeHeader, "expected eml vendor content type, got %s", contentTypeHeader)
		authHeader := req.Header.Get("Authorization")
		assert.Equal(t, "Bearer aToken", authHeader, "expected bearer aToken auth header, got %s", authHeader)
		return httpmock.NewJsonResponse(200, accSummary)
	})

	httpmock.RegisterResponder("GET", "https://eml.com/3.0/accounts/eaid?with_freetext=1&with_personal=1", func(req *http.Request) (*http.Response, error) {
		acceptHeader := req.Header.Get("Accept")
		assert.Equal(t, "application/vnd.eml+json", acceptHeader, "expected eml vendor accept content type, got %s", acceptHeader)
		authHeader := req.Header.Get("Authorization")
		assert.Equal(t, "Bearer aToken", authHeader, "expected bearer aToken auth header, got %s", authHeader)
		return httpmock.NewJsonResponse(200, AccountInfo{AccountSummary: accSummary, AccountId: "eaid"})
	})

	summary, err := store.CreateAccount(ctx, &CreateAccountRequest{})
	assert.NoError(t, err, "error creating account %v", err)
	assert.Equal(t, "eaid", summary.ExternalAccountId, "Expected EAID = eaid, got %s", summary.ExternalAccountId)
	account, err := store.GetAccount(ctx, "eaid", WithFreeText, WithPersonal)
	assert.NoError(t, err, "error getting account %v", err)
	assert.Equal(t, "eaid", account.ExternalAccountId, "Expected EAID = eaid, got %s", account.ExternalAccountId)
	assert.Equal(t, "eaid", account.AccountId, "Expected EAID = eaid, got %s", account.AccountId)

	assert.Equal(t, 3, httpmock.GetTotalCallCount(), "Expected 3 calls, got %d", httpmock.GetTotalCallCount())
	info := httpmock.GetCallCountInfo()
	tokenCalls := info["POST https://eml.com/3.0/token"]
	assert.Equal(t, 1, tokenCalls, "Expected 1 token call, got %d", tokenCalls)
}
