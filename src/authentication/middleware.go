package authentication

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"vulpz/train-api/src/api"
)

const ClaimsKey string = "claims"

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
