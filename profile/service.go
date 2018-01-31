package profile

import (
	"net/http"
	"time"

	"github.com/hailocab/gocassa"
	"github.com/zi-yang-zhang/cryptopia-api/core"
)

type UserTableProvider func() (userTable gocassa.Table, err error)

func profileHandler(authParams map[string]interface{}) http.Handler {
	e := core.SignUpEnabled(authParams)
	e.GET("/signUp", userSignUpEndpoint(getUserTableProvider()))
	e.GET("/signIn", userSignInEndpoint(getUserTableProvider()))
	return e
}

func Start(authParams map[string]interface{}) error {
	server := &http.Server{
		Addr:         ":9100",
		Handler:      profileHandler(authParams),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return server.ListenAndServe()
}

func getKeySpace() (keySpace gocassa.KeySpace, err error) {
	keySpace, err = gocassa.ConnectToKeySpace("user_space", []string{"127.0.0.1"}, "", "")
	return
}

func getUserTableProvider() UserTableProvider {
	var keySpace gocassa.KeySpace
	var err error
	return func() (gocassa.Table, error) {
		if keySpace == nil {
			keySpace, err = getKeySpace()
		}
		if err != nil {
			keySpace = nil
			return nil, err
		}
		userTable := keySpace.Table("user", &user{}, gocassa.Keys{
			PartitionKeys: []string{"uid"},
		})
		err = userTable.CreateIfNotExist()
		return userTable, err
	}
}
