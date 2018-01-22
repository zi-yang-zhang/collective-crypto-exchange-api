package core

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zi-yang-zhang/go-oauth-authenticator"
)

const (
	ErrorResponseMessageKey = "errorMessage"
	ErrorResponseStatusKey  = "errorCode"
	errorResponseReturnKey  = "errorResponse"
	responseReturnKey       = "data"
	JwtKey                  = "jwt_key"
	ClientID                = "X-Request-Client-ID"
	ClientEmail             = "X-Request-Client-Email"
	JWTError                = "JWTE1"
)

type heartbeatResponse struct {
	Status string `json:"status"`
}

type interceptResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w interceptResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

//Default gin engine
func Default() *gin.Engine {
	engine := gin.New()
	engine.Use(preProcessRequest(), gin.Logger(), gin.Recovery(), postProcessResponse())
	return engine
}

//AuthEnabled gin engine
func AuthEnabled() *gin.Engine {
	engine := Default()
	engine.Use(authenticationMiddleWare())
	return engine
}

//AuthEnabled gin engine
func SignUpEnabled() *gin.Engine {
	engine := Default()
	engine.Use(signUpMiddleWare())
	return engine
}

func configAuthenticator() *auth.AuthenticationProvider {
	authParam := map[string]interface{}{"google": "963603210536-qau1i4c8l9lj5hvkl09o7fvu8d045b1r.apps.googleusercontent.com"}
	authenticator := auth.New(authParam)
	return authenticator
}

func CreateError(code string, message string) gin.H {
	return gin.H{
		ErrorResponseMessageKey: message,
		ErrorResponseStatusKey:  code,
	}
}

func signUpMiddleWare() gin.HandlerFunc {
	authenticator := configAuthenticator()
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

func authenticationMiddleWare() gin.HandlerFunc {
	authenticator := configAuthenticator()
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		claims, ve := authenticator.Authenticate(authorization)
		if ve != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, CreateError(JWTError, ve.Error()))
			return
		}
		c.Writer.Header().Set(ClientID, claims.GetId())
		c.Writer.Header().Set(ClientEmail, claims.GetEmail())
		c.Next()
	}

}
func postProcessResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		intercepted := c.Writer.(*interceptResponseWriter)
		var response map[string]interface{}
		json.Unmarshal(intercepted.body.Bytes(), &response)
		c.Writer = intercepted.ResponseWriter
		_, hasError := response[ErrorResponseMessageKey]
		var responseKey string
		var status int
		if c.Writer.Status() >= 400 {
			responseKey = errorResponseReturnKey
			status = c.Writer.Status()
		} else if hasError {
			responseKey = errorResponseReturnKey
			status = http.StatusOK
		} else {
			responseKey = responseReturnKey
			status = c.Writer.Status()
		}
		c.JSON(status, gin.H{
			responseKey: response,
		})
		c.Abort()

	}
}
func preProcessRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Request-Root-Id", uuid.New().String())
		c.Writer.Header().Set("X-Request-Path-Id", uuid.New().String())
		c.Writer = &interceptResponseWriter{body: new(bytes.Buffer), ResponseWriter: c.Writer}
		c.Next()
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(heartbeatResponse{Status: "OK"})
}
