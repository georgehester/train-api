package authentication

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"vulpz/train-api/src/api"
)

const ClaimsKey string = "claims"
const ApplicationIdKey string = "applicationId"

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

func (keyManager *KeyManager) ApplicationKeyMiddleware(database *pgx.Conn) gin.HandlerFunc {
	return func(context *gin.Context) {
		// Check for header presence
		header := context.GetHeader("Authorization")
		if header == "" {
			api.SendErrorResponse(context, http.StatusUnauthorized, "Missing Authorization Header")
			context.Abort()
			return
		}

		headerSplit := strings.SplitN(header, " ", 2)
		if len(headerSplit) != 2 || headerSplit[0] != "Bearer" {
			api.SendErrorResponse(context, http.StatusUnauthorized, "Authorization Header Invalid")
			context.Abort()
			return
		}

		decodedBytes, decodeError := base64.StdEncoding.DecodeString(headerSplit[1])
		if decodeError != nil {
			decodedBytes, decodeError = base64.RawStdEncoding.DecodeString(headerSplit[1])
			if decodeError != nil {
				api.SendErrorResponse(context, http.StatusUnauthorized, "Authorization Header Invalid")
				context.Abort()
				return
			}
		}

		credentials := strings.SplitN(string(decodedBytes), ":", 2)
		if len(credentials) != 2 || credentials[0] == "" || credentials[1] == "" {
			api.SendErrorResponse(context, http.StatusUnauthorized, "Authorization Header Invalid")
			context.Abort()
			return
		}

		applicationId := credentials[0]
		applicationKey := credentials[1]

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
		context.Set(applicationKey, applicationId)
		context.Next()
	}
}
