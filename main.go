package main

import (
	"github.com/gin-gonic/gin"
	"github.com/zi-yang-zhang/cryptopia-api/signin"
)

func main() {
	router := gin.Default()
	api := router.Group("/api")
	v1 := api.Group("/v1")
	setupV1Endpoint(v1)
	router.Run(":9000")
}

func setupV1Endpoint(v1 *gin.RouterGroup) {
	signIn := v1.Group("/signin")
	tokenSignIn := signIn.Group("/token")
	tokenSignIn.GET("/google", signin.GoogleSignIn)

}
