package profile

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hailocab/gocassa"
	"github.com/zi-yang-zhang/cryptopia-api/core"
	"github.com/zi-yang-zhang/go-oauth-authenticator"
	"log"
)

const (
	SignUpUserExists        = "SUE1"
	SignUpUserCreateError   = "SUE2"
	SignInUserNotFound      = "AUTH1"
	DatabaseConnectionError = "DBE1"
)

type user struct {
	UserID      string    `cql:"uid" json:"uid"`
	Email       string    `cql:"email" json:"email"`
	DisplayName string    `cql:"displayName" json:"displayName"`
	Picture     string    `cql:"pictureUrl" json:"pictureUrl"`
	GivenName   string    `cql:"givenName" json:"givenName"`
	LastName    string    `cql:"lastName" json:"lastName"`
	Created     time.Time `cql:"createdAt" json:"createdAt"`
	userConfig  `cql:",squash" json:"config"`
}

func userSignInEndpoint(provider UserTableProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userTable, err := provider()
		if err != nil {
			log.Print("Cannot get user table", err.Error())
			c.JSON(http.StatusServiceUnavailable, core.CreateError(DatabaseConnectionError, err.Error()))
			return
		}

		token, ok := c.Get(core.JwtKey)
		if !ok {
			c.JSON(http.StatusBadRequest, core.CreateError(core.JWTError, "JWT not found"))
		}
		tokenProvider, ok := token.(auth.AuthenticationInfo)
		existingUser := user{}
		if err := userTable.Where(gocassa.Eq("uid", tokenProvider.GetId())).ReadOne(&existingUser).Run(); err != nil {
			c.JSON(http.StatusNotFound, core.CreateError(SignInUserNotFound, "User not found"))
			return
		}
		c.JSON(http.StatusOK, existingUser)
	}
}

func userSignUpEndpoint(provider UserTableProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userTable, err := provider()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, core.CreateError(DatabaseConnectionError, err.Error()))
			return
		}
		token, ok := c.Get(core.JwtKey)
		if !ok {
			c.JSON(http.StatusBadRequest, core.CreateError(core.JWTError, "JWT not found"))
		}
		tokenProvider, ok := token.(auth.AuthenticationInfo)
		switch issuer := tokenProvider.GetIssuer(); issuer {

		case auth.IssuerGoogle:
			googleSignUp(userTable, c, tokenProvider)
			return
		}
	}
}

func googleSignUp(userTable gocassa.Table, c *gin.Context, token auth.AuthenticationInfo) {
	googleClaims := token.(*auth.GoogleJWTClaims)

	existingUser := user{}
	if err := userTable.Where(gocassa.Eq("uid", googleClaims.Subject)).ReadOne(&existingUser).Run(); err == nil {
		c.JSON(http.StatusOK, core.CreateError(SignUpUserExists, "User exists"))
		return
	}

	newUser := user{
		UserID:      googleClaims.Subject,
		Email:       googleClaims.Email,
		DisplayName: googleClaims.DisplayName,
		Picture:     googleClaims.Picture,
		GivenName:   googleClaims.GivenName,
		LastName:    googleClaims.FamilyName,
		Created:     time.Now(),
		userConfig:  userConfig{},
	}
	err := userTable.Set(newUser).Run()
	if err != nil {
		c.JSON(http.StatusOK, core.CreateError(SignUpUserCreateError, "Cannot create user"))
		return
	}
	c.JSON(http.StatusCreated, newUser)
	return
}
