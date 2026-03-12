package authentication

import (
	"crypto/ed25519"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseApplicationAuthorizationHeader(t *testing.T) {
	tests := []struct {
		name              string
		header            string
		expectedID        string
		expectedKey       string
		expectedValidBool bool
	}{
		{
			name:              "valid standard base64",
			header:            "Bearer " + base64.StdEncoding.EncodeToString([]byte("app-id:app-key")),
			expectedID:        "app-id",
			expectedKey:       "app-key",
			expectedValidBool: true,
		},
		{
			name:              "valid raw base64",
			header:            "Bearer " + base64.RawStdEncoding.EncodeToString([]byte("app-id:app-key")),
			expectedID:        "app-id",
			expectedKey:       "app-key",
			expectedValidBool: true,
		},
		{
			name:              "invalid scheme",
			header:            "Basic " + base64.StdEncoding.EncodeToString([]byte("app-id:app-key")),
			expectedValidBool: false,
		},
		{
			name:              "invalid base64",
			header:            "Bearer not-base64",
			expectedValidBool: false,
		},
		{
			name:              "missing separator",
			header:            "Bearer " + base64.StdEncoding.EncodeToString([]byte("appidonly")),
			expectedValidBool: false,
		},
		{
			name:              "empty id",
			header:            "Bearer " + base64.StdEncoding.EncodeToString([]byte(":app-key")),
			expectedValidBool: false,
		},
		{
			name:              "empty key",
			header:            "Bearer " + base64.StdEncoding.EncodeToString([]byte("app-id:")),
			expectedValidBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, key, ok := parseApplicationAuthorizationHeader(tt.header)
			if ok != tt.expectedValidBool {
				t.Fatalf("expected valid=%v, got %v", tt.expectedValidBool, ok)
			}

			if id != tt.expectedID {
				t.Fatalf("expected id=%q, got %q", tt.expectedID, id)
			}

			if key != tt.expectedKey {
				t.Fatalf("expected key=%q, got %q", tt.expectedKey, key)
			}
		})
	}
}

func TestSignAndVerifyToken(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	manager := &KeyManager{
		privateKey: privateKey,
		publicKey:  publicKey,
	}

	token, signErr := manager.Sign(UserTypeCustomer, "123", "email@example.com", "Jane", "Doe")
	if signErr != nil {
		t.Fatalf("expected no sign error, got %v", signErr)
	}

	claims, verifyErr := manager.Verify(token)
	if verifyErr != nil {
		t.Fatalf("expected no verify error, got %v", verifyErr)
	}

	if claims.Type != UserTypeCustomer || claims.Id != "123" || claims.Email != "email@example.com" {
		t.Fatalf("unexpected claims returned: %+v", claims)
	}
}

func TestSignRejectsInvalidUserType(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	manager := &KeyManager{privateKey: privateKey}

	if _, signErr := manager.Sign(UserType("invalid"), "1", "e", "f", "s"); signErr == nil {
		t.Fatalf("expected sign error for invalid user type")
	}
}

func TestVerifyRejectsInvalidToken(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate public key: %v", err)
	}

	manager := &KeyManager{publicKey: publicKey}

	if _, verifyErr := manager.Verify("not-a-token"); verifyErr == nil {
		t.Fatalf("expected verify error for invalid token")
	}
}

func TestApplicationRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := &KeyManager{}
	router := gin.New()
	router.Use(manager.ApplicationRateLimitMiddleware(1))
	router.GET("/", func(context *gin.Context) {
		context.Status(http.StatusOK)
	})

	credentials := base64.StdEncoding.EncodeToString([]byte("app-id:app-key"))
	headerValue := "Bearer " + credentials

	firstRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	firstRequest.Header.Set("Authorization", headerValue)
	firstRecorder := httptest.NewRecorder()
	router.ServeHTTP(firstRecorder, firstRequest)

	if firstRecorder.Code != http.StatusOK {
		t.Fatalf("expected first request status %d, got %d", http.StatusOK, firstRecorder.Code)
	}

	secondRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	secondRequest.Header.Set("Authorization", headerValue)
	secondRecorder := httptest.NewRecorder()
	router.ServeHTTP(secondRecorder, secondRequest)

	if secondRecorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request status %d, got %d", http.StatusTooManyRequests, secondRecorder.Code)
	}

	if !strings.Contains(secondRecorder.Body.String(), "Rate Limit Exceeded") {
		t.Fatalf("expected rate limit error message, got %s", secondRecorder.Body.String())
	}
}
