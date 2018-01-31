package core

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zi-yang-zhang/go-oauth-authenticator"
	"strconv"
	"strings"
)

const (
	ErrorResponseMessageKey = "errorMessage"
	ErrorResponseStatusKey  = "errorCode"
	ErrorResponseReturnKey  = "errorResponse"
	ResponseReturnKey       = "data"
	JwtKey                  = "jwt_key"
	ClientID                = "X-Request-Client-ID"
	ClientEmail             = "X-Request-Client-Email"
	TracingRoot             = "X-Request-Root-Id"
	TracingPath             = "X-Request-Path-Id"
	JWTError                = "JWTE1"
	JWTMissing              = "JWTE2"
	GeneralError            = "GE1"
)

type Config struct {
	AuthParams map[string]interface{} `json:"authParams"`
}

type heartbeatResponse struct {
	Status string `json:"status"`
}

//Default gin engine
func Default() *gin.Engine {
	engine := gin.Default()
	engine.Use(headerHandler())
	return engine
}

//AuthEnabled gin engine
func SignUpEnabled(authParams map[string]interface{}) *gin.Engine {
	engine := Default()
	engine.Use(signUpMiddleWare(authParams))
	return engine
}

func configAuthenticator(authParams map[string]interface{}) *auth.AuthenticationProvider {
	authenticator := auth.New(authParams)
	return authenticator
}

func CreateError(code string, message string) gin.H {
	return gin.H{
		ErrorResponseMessageKey: message,
		ErrorResponseStatusKey:  code,
	}
}

func signUpMiddleWare(authParams map[string]interface{}) gin.HandlerFunc {
	authenticator := configAuthenticator(authParams)
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		if authorization == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, CreateError(JWTError, "JWT is missing"))
			return
		}
		claims, ve := authenticator.Authenticate(authorization)
		if ve == nil {
			c.Set(JwtKey, claims)
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, CreateError(JWTError, ve.Error()))
			return
		}
	}
}

func headerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := strings.Split(c.GetHeader(TracingPath), ".")
		if len(path) < 2 {
			c.Header(TracingPath, c.GetHeader(TracingPath)+".0")
		} else {
			inc, _ := strconv.ParseInt(path[1], 10, 32)
			inc++
			c.Header(TracingPath, c.GetHeader(TracingPath)+strconv.FormatInt(inc, 10))
		}
		c.Header(TracingRoot, c.GetHeader(TracingRoot))
		c.Next()
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(heartbeatResponse{Status: "OK"})
}
