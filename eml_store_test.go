package eml

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const (
	baseUrl      = "https://eml.com"
	clientId     = "client"
	clientSecret = "secret"
	accessToken  = "aToken"
)

var ctx = context.Background()

func mockTokenResponse(t *testing.T) {
	httpmock.RegisterResponder("POST", "https://eml.com/3.0/token", func(req *http.Request) (*http.Response, error) {
		bodyBytes, _ := io.ReadAll(req.Body)
		assert.Equal(t, "grant_type=client_credentials", string(bodyBytes), "Malformed token body %s", string(bodyBytes))
		acceptHeader := req.Header.Get("Accept")
		assert.Equal(t, "application/vnd.eml+json", acceptHeader, "expected eml vendor accept content type, got %s", acceptHeader)
		authHeader := req.Header.Get("Authorization")
		assert.Equal(t, "Basic Y2xpZW50OnNlY3JldA==", authHeader, "expected basic auth header, got %s", authHeader)
		return httpmock.NewJsonResponse(200, TokenResponse{AccessToken: accessToken, TokenType: "bearer", ExpiresIn: 60 * 60})
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
		assert.Equal(t, "Bearer "+accessToken, authHeader, "expected bearer %s auth header, got %s", accessToken, authHeader)
		return httpmock.NewJsonResponse(200, accSummary)
	})

	httpmock.RegisterResponder("GET", "https://eml.com/3.0/accounts/eaid?with_freetext=1&with_personal=1", func(req *http.Request) (*http.Response, error) {
		acceptHeader := req.Header.Get("Accept")
		assert.Equal(t, "application/vnd.eml+json", acceptHeader, "expected eml vendor accept content type, got %s", acceptHeader)
		authHeader := req.Header.Get("Authorization")
		assert.Equal(t, "Bearer "+accessToken, authHeader, "expected bearer %s auth header, got %s", accessToken, authHeader)
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

// Test the complete renewal flow: authenticate -> initiate -> activate
func Test_emlStore_RenewalFlow(t *testing.T) {
	e := &Settings{EmlRestId: clientId, EmlHostUrl: baseUrl}
	store := &emlStore{_restSecret: clientSecret, _env: e}
	store.lazyInit(ctx)

	httpmock.ActivateNonDefault(store._lazyClient.GetClient())
	defer httpmock.DeactivateAndReset()
	mockTokenResponse(t)

	// Mock authenticate endpoint
	authResponse := &AuthenticateResponse{
		TokenID:      "token123",
		EmailAddress: "test@example.com",
		Mobile:       "+61412345678",
	}
	httpmock.RegisterResponder("POST", "https://eml.com/3.0/accounts/eaid123/authenticate", func(req *http.Request) (*http.Response, error) {
		// Verify headers
		contentTypeHeader := req.Header.Get("Content-Type")
		assert.Equal(t, "application/vnd.eml+json", contentTypeHeader, "expected eml vendor content type, got %s", contentTypeHeader)
		authHeader := req.Header.Get("Authorization")
		assert.Equal(t, "Bearer "+accessToken, authHeader, "expected bearer %s auth header, got %s", accessToken, authHeader)

		// Verify request body
		bodyBytes, _ := io.ReadAll(req.Body)
		var authReq AuthenticateRequest
		err := json.Unmarshal(bodyBytes, &authReq)
		assert.NoError(t, err, "failed to unmarshal authenticate request")
		assert.Equal(t, "app123", authReq.ApplicationID, "expected application_id = app123, got %s", authReq.ApplicationID)
		assert.Equal(t, "192.168.1.1", authReq.IPAddress, "expected ip_address = 192.168.1.1, got %s", authReq.IPAddress)

		return httpmock.NewJsonResponse(200, authResponse)
	})

	// Mock initiate endpoint
	initiateResponse := &InitiateResponse{
		OperationID: "op456",
	}
	httpmock.RegisterResponder("POST", "https://eml.com/3.0/accounts/eaid123/initiate", func(req *http.Request) (*http.Response, error) {
		// Verify headers
		contentTypeHeader := req.Header.Get("Content-Type")
		assert.Equal(t, "application/vnd.eml+json", contentTypeHeader, "expected eml vendor content type, got %s", contentTypeHeader)

		// Verify request body
		bodyBytes, _ := io.ReadAll(req.Body)
		var initiateReq InitiateRequest
		err := json.Unmarshal(bodyBytes, &initiateReq)
		assert.NoError(t, err, "failed to unmarshal initiate request")
		assert.Equal(t, "192.168.1.1", initiateReq.IPAddress, "expected ip_address = 192.168.1.1, got %s", initiateReq.IPAddress)
		assert.Equal(t, "token123", initiateReq.TokenID, "expected token_id = token123, got %s", initiateReq.TokenID)
		assert.Equal(t, CommunicateMethodSMS, initiateReq.CommunicateMethod, "expected communicate_method = sms, got %s", initiateReq.CommunicateMethod)
		assert.Equal(t, OperationTypeOneTimePassCode, initiateReq.OperationType, "expected operation_type = renewal, got %s", initiateReq.OperationType)

		return httpmock.NewJsonResponse(200, initiateResponse)
	})

	// Mock activate endpoint
	httpmock.RegisterResponder("POST", "https://eml.com/3.0/accounts/eaid123/activate", func(req *http.Request) (*http.Response, error) {
		// Verify headers
		contentTypeHeader := req.Header.Get("Content-Type")
		assert.Equal(t, "application/vnd.eml+json", contentTypeHeader, "expected eml vendor content type, got %s", contentTypeHeader)

		// Verify request body
		bodyBytes, _ := io.ReadAll(req.Body)
		var activateReq ActivateRequest
		err := json.Unmarshal(bodyBytes, &activateReq)
		assert.NoError(t, err, "failed to unmarshal activate request")
		assert.Equal(t, "op456", activateReq.ValidationData.OperationID, "expected operation_id = op456, got %s", activateReq.ValidationData.OperationID)
		assert.Equal(t, "123456", activateReq.ValidationData.SecurityCode, "expected security_code = 123456, got %s", activateReq.ValidationData.SecurityCode)
		assert.Equal(t, "192.168.1.1", activateReq.ValidationData.IPAddress, "expected ip_address = 192.168.1.1, got %s", activateReq.ValidationData.IPAddress)
		assert.NotNil(t, activateReq.EnablePlastic, "expected enable_plastic to be set")
		assert.Equal(t, "true", *activateReq.EnablePlastic, "expected enable_plastic = true, got %s", *activateReq.EnablePlastic)

		return httpmock.NewStringResponse(200, ""), nil
	})

	// Execute the full flow
	authReq := AuthenticateRequest{
		ApplicationID: "app123",
		IPAddress:     "192.168.1.1",
	}
	authResp, err := store.Authenticate(ctx, "eaid123", authReq)
	assert.NoError(t, err, "error authenticating %v", err)
	assert.Equal(t, "token123", authResp.TokenID, "expected token_id = token123, got %s", authResp.TokenID)
	assert.Equal(t, "test@example.com", authResp.EmailAddress, "expected email = test@example.com, got %s", authResp.EmailAddress)

	initiateReq := InitiateRequest{
		IPAddress:           "192.168.1.1",
		TokenID:             authResp.TokenID,
		CommunicationMethod: CommunicationMethodSMS,
		OperationType:       OperationTypeOneTimePassCode,
	}
	initiateResp, err := store.Initiate(ctx, "eaid123", initiateReq)
	assert.NoError(t, err, "error initiating %v", err)
	assert.Equal(t, "op456", initiateResp.OperationID, "expected operation_id = op456, got %s", initiateResp.OperationID)

	activateReq := ActivateRequest{
		ValidationData: ValidationData{
			OperationID:  initiateResp.OperationID,
			SecurityCode: "123456",
			IPAddress:    "192.168.1.1",
		},
		EnablePlastic: BoolToStringPtr(true),
	}
	err = store.Activate(ctx, "eaid123", activateReq)
	assert.NoError(t, err, "error activating %v", err)

	// Verify call counts
	assert.Equal(t, 4, httpmock.GetTotalCallCount(), "Expected 4 calls (1 token + 3 renewals), got %d", httpmock.GetTotalCallCount())
	info := httpmock.GetCallCountInfo()
	tokenCalls := info["POST https://eml.com/3.0/token"]
	assert.Equal(t, 1, tokenCalls, "Expected 1 token call, got %d", tokenCalls)
}

// Test initiate with email communication method
func Test_emlStore_InitiateWithEmail(t *testing.T) {
	e := &Settings{EmlRestId: clientId, EmlHostUrl: baseUrl}
	store := &emlStore{_restSecret: clientSecret, _env: e}
	store.lazyInit(ctx)

	httpmock.ActivateNonDefault(store._lazyClient.GetClient())
	defer httpmock.DeactivateAndReset()
	mockTokenResponse(t)

	initiateResponse := &InitiateResponse{OperationID: "op789"}
	httpmock.RegisterResponder("POST", "https://eml.com/3.0/accounts/eaid123/initiate", func(req *http.Request) (*http.Response, error) {
		bodyBytes, _ := io.ReadAll(req.Body)
		var initiateReq InitiateRequest
		err := json.Unmarshal(bodyBytes, &initiateReq)
		assert.NoError(t, err, "failed to unmarshal initiate request")
		assert.Equal(t, CommunicationMethodEmail, initiateReq.CommunicationMethod, "expected communicate_method = email, got %s", initiateReq.CommunicateMethod)

		return httpmock.NewJsonResponse(200, initiateResponse)
	})

	initiateReq := InitiateRequest{
		IPAddress:           "10.0.0.1",
		TokenID:             "token456",
		CommunicationMethod: CommunicationMethodEmail,
		OperationType:       "renewal",
	}
	resp, err := store.Initiate(ctx, "eaid123", initiateReq)
	assert.NoError(t, err, "error initiating with email %v", err)
	assert.Equal(t, "op789", resp.OperationID, "expected operation_id = op789, got %s", resp.OperationID)
}

// Test activate without enable_plastic (should be omitted from JSON)
func Test_emlStore_ActivateWithoutPlastic(t *testing.T) {
	e := &Settings{EmlRestId: clientId, EmlHostUrl: baseUrl}
	store := &emlStore{_restSecret: clientSecret, _env: e}
	store.lazyInit(ctx)

	httpmock.ActivateNonDefault(store._lazyClient.GetClient())
	defer httpmock.DeactivateAndReset()
	mockTokenResponse(t)

	httpmock.RegisterResponder("POST", "https://eml.com/3.0/accounts/eaid123/activate", func(req *http.Request) (*http.Response, error) {
		bodyBytes, _ := io.ReadAll(req.Body)
		var activateReq ActivateRequest
		err := json.Unmarshal(bodyBytes, &activateReq)
		assert.NoError(t, err, "failed to unmarshal activate request")
		assert.Nil(t, activateReq.EnablePlastic, "expected enable_plastic to be nil/omitted")

		// Verify the raw JSON doesn't contain enable_plastic
		bodyStr := string(bodyBytes)
		assert.NotContains(t, bodyStr, "enable_plastic", "enable_plastic should be omitted from JSON when nil")

		return httpmock.NewStringResponse(200, ""), nil
	})

	activateReq := ActivateRequest{
		ValidationData: ValidationData{
			OperationID:  "op123",
			SecurityCode: "654321",
			IPAddress:    "172.16.0.1",
		},
		// EnablePlastic is nil, should be omitted
	}
	err := store.Activate(ctx, "eaid123", activateReq)
	assert.NoError(t, err, "error activating without plastic %v", err)
}

// Test error handling for renewal endpoints
func Test_emlStore_RenewalErrorHandling(t *testing.T) {
	e := &Settings{EmlRestId: clientId, EmlHostUrl: baseUrl}
	store := &emlStore{_restSecret: clientSecret, _env: e}
	store.lazyInit(ctx)

	httpmock.ActivateNonDefault(store._lazyClient.GetClient())
	defer httpmock.DeactivateAndReset()
	mockTokenResponse(t)

	// Mock error response
	errorResponse := &ErrorModel{
		Code:        "invalid_request",
		Description: "Invalid application ID",
	}
	httpmock.RegisterResponder("POST", "https://eml.com/3.0/accounts/eaid123/authenticate", func(req *http.Request) (*http.Response, error) {
		return httpmock.NewJsonResponse(400, errorResponse)
	})

	authReq := AuthenticateRequest{
		ApplicationID: "invalid",
		IPAddress:     "192.168.1.1",
	}
	_, err := store.Authenticate(ctx, "eaid123", authReq)
	assert.Error(t, err, "expected error for invalid authentication")
	assert.Contains(t, err.Error(), "invalid_request", "expected error to contain error code")
}
