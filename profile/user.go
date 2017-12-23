package profile

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hailocab/gocassa"
	"github.com/zi-yang-zhang/cryptopia-api/core"
	auth "github.com/zi-yang-zhang/go-oauth-authenticator"
)

type user struct {
	UserID      string    `cql:"uid"`
	Email       string    `cql:"email"`
	DisplayName string    `cql:"displayName"`
	Picture     string    `cql:"pictureUrl"`
	GivenName   string    `cql:"givenName"`
	LastName    string    `cql:"lastName"`
	Created     time.Time `cql:"createdAt"`
	userConfig  `cql:",squash"`
}

type userResponse struct {
	UserID      string     `json:"uid"`
	Email       string     `json:"email"`
	DisplayName string     `json:"displayName"`
	Picture     string     `json:"pictureUrl"`
	GivenName   string     `json:"givenName"`
	LastName    string     `json:"lastName"`
	Created     time.Time  `json:"createdAt"`
	Config      userConfig `json:"config"`
}

func userGoogleSignUpEndpoint(keyspace gocassa.KeySpace) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwt, ok := auth.GetGoogleClaims(c)
		if !ok {
			panic("Jwt not found")
		}

		userTable := keyspace.Table("user", &user{}, gocassa.Keys{
			PartitionKeys: []string{"uid"},
		})
		err := userTable.CreateIfNotExist()
		if err != nil {
			panic(err)
		}
		existingUser := user{}
		if err = userTable.Where(gocassa.Eq("uid", jwt.Subject)).ReadOne(&existingUser).Run(); err == nil {
			c.JSON(http.StatusOK, gin.H{
				core.ErrorResponseKey: "User exists",
			})
			return
		}

		newUser := user{
			UserID:      jwt.Subject,
			Email:       jwt.Email,
			DisplayName: jwt.DisplayName,
			Picture:     jwt.Picture,
			GivenName:   jwt.GivenName,
			LastName:    jwt.FamilyName,
			Created:     time.Now(),
			userConfig:  userConfig{},
		}
		err = userTable.Set(newUser).Run()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				core.ErrorResponseKey: "Cannot create user",
			})
			return
		}
	}
}
