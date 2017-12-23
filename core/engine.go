package core

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	auth "github.com/zi-yang-zhang/go-oauth-authenticator"
)

const (
	ErrorResponseKey       = "error"
	errorResponseReturnKey = "errorResponse"
	responseReturnKey      = "data"
)

type heartbeatResponse struct {
	Status string `json:"status"`
}

type interceptResponstWritter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w interceptResponstWritter) Write(b []byte) (int, error) {
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
	authenticator := &auth.GoogleAuthenticator{}
	engine.Use(authenticator.AuthenticateMiddleware("963603210536-qau1i4c8l9lj5hvkl09o7fvu8d045b1r.apps.googleusercontent.com"))
	return engine
}

func postProcessResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		intercepted := c.Writer.(*interceptResponstWritter)
		var response map[string]interface{}
		json.Unmarshal(intercepted.body.Bytes(), &response)
		c.Writer = intercepted.ResponseWriter
		_, hasError := response[ErrorResponseKey]
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
		c.Writer = &interceptResponstWritter{body: new(bytes.Buffer), ResponseWriter: c.Writer}
		c.Next()
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(heartbeatResponse{Status: "OK"})
}
