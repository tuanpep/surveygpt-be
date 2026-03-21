package handlers_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/surveyflow/be/internal/testutil"
)

// stubTest defines a single stub endpoint test case.
type stubTest struct {
	name       string
	method     string
	path       string
	body       any
	auth       bool // whether to send an auth token
	wantStatus int
}

func allStubTests() []stubTest {
	return []stubTest{
		// OAuth stubs (public, no auth)
		{name: "AuthGoogle", method: http.MethodGet, path: "/api/v1/auth/google", auth: false},
		{name: "AuthGoogleCallback", method: http.MethodGet, path: "/api/v1/auth/google/callback", auth: false},
		{name: "AuthMicrosoft", method: http.MethodGet, path: "/api/v1/auth/microsoft", auth: false},
		{name: "AuthMicrosoftCallback", method: http.MethodGet, path: "/api/v1/auth/microsoft/callback", auth: false},

		// User profile stubs (auth required)
		{name: "UploadAvatar", method: http.MethodPost, path: "/api/v1/me/avatar", auth: true},
		{name: "DeleteAccount", method: http.MethodDelete, path: "/api/v1/me/account", auth: true},
		{name: "Enable2FA", method: http.MethodPost, path: "/api/v1/me/2fa/enable", auth: true},
		{name: "Disable2FA", method: http.MethodPost, path: "/api/v1/me/2fa/disable", auth: true},
		{name: "Verify2FA", method: http.MethodPost, path: "/api/v1/me/2fa/verify", auth: true},

		// Organization stubs
		{name: "DeleteOrg", method: http.MethodDelete, path: "/api/v1/organizations/00000000-0000-0000-0000-000000000001", auth: true},

		// Response stubs (public)
		{name: "UpdateResponse", method: http.MethodPatch, path: "/api/v1/surveys/00000000-0000-0000-0000-000000000001/responses/00000000-0000-0000-0000-000000000002", auth: false},

		// File stubs
		{name: "UploadFile", method: http.MethodPost, path: "/api/v1/files/upload", auth: true},
		{name: "ListFiles", method: http.MethodGet, path: "/api/v1/files", auth: true},
		{name: "DeleteFile", method: http.MethodDelete, path: "/api/v1/files/00000000-0000-0000-0000-000000000001", auth: true},

		// Integration stubs
		{name: "ListIntegrations", method: http.MethodGet, path: "/api/v1/integrations", auth: true},
		{name: "CreateIntegration", method: http.MethodPost, path: "/api/v1/integrations", auth: true},
		{name: "UpdateIntegration", method: http.MethodPut, path: "/api/v1/integrations/00000000-0000-0000-0000-000000000001", auth: true},
		{name: "DeleteIntegration", method: http.MethodDelete, path: "/api/v1/integrations/00000000-0000-0000-0000-000000000001", auth: true},

		// API key stubs
		{name: "ListAPIKeys", method: http.MethodGet, path: "/api/v1/api-keys", auth: true},
		{name: "CreateAPIKey", method: http.MethodPost, path: "/api/v1/api-keys", auth: true},
		{name: "RevokeAPIKey", method: http.MethodDelete, path: "/api/v1/api-keys/00000000-0000-0000-0000-000000000001", auth: true},

		// Webhook stubs
		{name: "ListWebhooks", method: http.MethodGet, path: "/api/v1/webhooks", auth: true},
		{name: "CreateWebhook", method: http.MethodPost, path: "/api/v1/webhooks", auth: true},
		{name: "UpdateWebhook", method: http.MethodPut, path: "/api/v1/webhooks/00000000-0000-0000-0000-000000000001", auth: true},
		{name: "DeleteWebhook", method: http.MethodDelete, path: "/api/v1/webhooks/00000000-0000-0000-0000-000000000001", auth: true},

		// Audit log stubs
		{name: "ListAuditLogs", method: http.MethodGet, path: "/api/v1/audit-logs", auth: true},
	}
}

func TestStubs_Return501(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	for _, tc := range allStubTests() {
		t.Run(tc.name, func(t *testing.T) {
			tok := ""
			if tc.auth {
				tok = token
			}

			rec := testutil.DoRequest(t, e, tc.method, tc.path, tc.body, tok)
			require.Equal(t, http.StatusNotImplemented, rec.Code, "stub %s should return 501", tc.name)

			body := testutil.ParseJSON(t, rec)
			assert.Equal(t, "NOT_IMPLEMENTED", body["code"])
		})
	}
}
