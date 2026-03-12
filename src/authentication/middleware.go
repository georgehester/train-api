package authentication

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"vulpz/train-api/src/api"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/patrickmn/go-cache"
)

const ClaimsKey string = "claims"
const ApplicationIdKey string = "applicationId"

func parseApplicationAuthorizationHeader(header string) (string, string, bool) {
	headerSplit := strings.SplitN(header, " ", 2)
	if len(headerSplit) != 2 || headerSplit[0] != "Bearer" {
		return "", "", false
	}

	decodedBytes, decodeError := base64.StdEncoding.DecodeString(headerSplit[1])
	if decodeError != nil {
		decodedBytes, decodeError = base64.RawStdEncoding.DecodeString(headerSplit[1])
		if decodeError != nil {
			return "", "", false
		}
	}

	credentials := strings.SplitN(string(decodedBytes), ":", 2)
	if len(credentials) != 2 || credentials[0] == "" || credentials[1] == "" {
		return "", "", false
	}

	return credentials[0], credentials[1], true
}

func (keyManager *KeyManager) Middleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		// Check for header presence
		header := context.GetHeader("Authorization")
		if header == "" {
			api.SendErrorResponse(context, http.StatusUnauthorized, "Missing Authorization Header")
			context.Abort()
			return
		}

		// Extract the token from the Bearer header
		headerSplit := strings.SplitN(header, " ", 2)
		if len(headerSplit) != 2 || headerSplit[0] != "Bearer" {
			api.SendErrorResponse(context, http.StatusUnauthorized, "Authorization Header Invalid")
			context.Abort()
			return
		}

		// Verify the token and extract the claims
		claims, keyError := keyManager.Verify(headerSplit[1])
		if keyError != nil {
			api.SendErrorResponse(context, http.StatusUnauthorized, "Token Invalid")
			context.Abort()
			return
		}

		// Store claims in context for use in handlers
		context.Set(ClaimsKey, claims)
		context.Next()
	}
}

func (keyManager *KeyManager) AdministrationMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		claims, ok := context.MustGet(ClaimsKey).(*Claims)
		if !ok {
			api.SendErrorResponse(context, http.StatusUnauthorized, "Failed To Fetching Token Claims")
			context.Abort()
			return
		}

		if claims.Type != UserTypeAdministrator {
			api.SendErrorResponse(context, http.StatusForbidden, "Must Be Administrator To Access")
			context.Abort()
			return
		}

		context.Next()
	}
}

func (keyManager *KeyManager) ApplicationKeyMiddleware(database *pgxpool.Pool) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Check for header presence
		header := context.GetHeader("Authorization")
		if header == "" {
			api.SendErrorResponse(context, http.StatusUnauthorized, "Missing Authorization Header")
			context.Abort()
			return
		}

		applicationId, applicationKey, ok := parseApplicationAuthorizationHeader(header)
		if !ok {
			api.SendErrorResponse(context, http.StatusUnauthorized, "Authorization Header Invalid")
			context.Abort()
			return
		}

		var approved bool
		queryError := database.QueryRow(
			context,
			"SELECT approved FROM applications WHERE id = $1 AND key = $2",
			applicationId,
			applicationKey,
		).Scan(&approved)
		if queryError != nil {
			if queryError == pgx.ErrNoRows {
				api.SendErrorResponse(context, http.StatusUnauthorized, "Application Credentials Invalid")
			} else {
				api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Validate Application Credentials")
			}
			context.Abort()
			return
		}

		if approved == false {
			api.SendErrorResponse(context, http.StatusForbidden, "Application Not Approved")
			context.Abort()
			return
		}

		// Store application id in context for use in handlers
		context.Set(ApplicationIdKey, applicationId)
		context.Next()
	}
}

func (keyManager *KeyManager) ApplicationRateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 60
	}

	requestCounts := cache.New(1*time.Minute, 2*time.Minute)

	return func(context *gin.Context) {
		applicationIdValue, exists := context.Get(ApplicationIdKey)
		if !exists {
			header := context.GetHeader("Authorization")
			if header == "" {
				api.SendErrorResponse(context, http.StatusUnauthorized, "Missing Authorization Header")
				context.Abort()
				return
			}

			applicationId, _, ok := parseApplicationAuthorizationHeader(header)
			if !ok {
				api.SendErrorResponse(context, http.StatusUnauthorized, "Authorization Header Invalid")
				context.Abort()
				return
			}

			applicationIdValue = applicationId
		}

		applicationId, ok := applicationIdValue.(string)
		if !ok || applicationId == "" {
			api.SendErrorResponse(context, http.StatusUnauthorized, "Application Credentials Invalid")
			context.Abort()
			return
		}

		if addError := requestCounts.Add(applicationId, 1, cache.DefaultExpiration); addError != nil {
			count, incrementError := requestCounts.IncrementInt(applicationId, 1)
			if incrementError != nil {
				api.SendErrorResponse(context, http.StatusInternalServerError, "Failed To Apply Rate Limit")
				context.Abort()
				return
			}

			if count > requestsPerMinute {
				api.SendErrorResponse(context, http.StatusTooManyRequests, "Rate Limit Exceeded")
				context.Abort()
				return
			}
		}

		context.Next()
	}
}
