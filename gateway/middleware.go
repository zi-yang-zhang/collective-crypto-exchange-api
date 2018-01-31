package gateway

import (
	"github.com/devopsfaith/krakend/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zi-yang-zhang/cryptopia-api/core"
	"github.com/zi-yang-zhang/go-oauth-authenticator"
	"net/http"
)

func configAuthenticator(authParams map[string]interface{}) *auth.AuthenticationProvider {
	authenticator := auth.New(authParams)
	return authenticator
}

func NewAuthenticationEnabledMiddleware(cfg config.ServiceConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.ExtraConfig["authParam"] != nil {
			authenticator := configAuthenticator(cfg.ExtraConfig["authParam"].(map[string]interface{}))
			authorization := c.GetHeader("Authorization")
			claims, ve := authenticator.Authenticate(authorization)
			if ve != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, core.CreateError(core.JWTError, ve.Error()))
				return
			}
			if authorization == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, core.CreateError(core.JWTMissing, "JWT missing in header"))
				return
			}
			c.Header(core.ClientID, claims.GetId())
			c.Header(core.ClientEmail, claims.GetEmail())
		}
		c.Next()
	}
}

func AddTracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Header.Set(core.TracingRoot, uuid.New().String())
		c.Request.Header.Set(core.TracingPath, uuid.New().String())
		c.Next()

	}
}
