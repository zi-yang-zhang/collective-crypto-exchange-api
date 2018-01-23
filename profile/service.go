package profile

import (
	"net/http"
	"time"

	"github.com/hailocab/gocassa"
	"github.com/zi-yang-zhang/cryptopia-api/core"
)

func profileHandler(authParams map[string]interface{}) http.Handler {
	e := core.SignUpEnabled(authParams)
	userTable := getUserTable()
	e.GET("/signUp", userSignUpEndpoint(userTable))
	e.GET("/signIn", userSignInEndpoint(userTable))
	return e
}

func ProfileService(authParams map[string]interface{}) error {
	server := &http.Server{
		Addr:         ":9100",
		Handler:      profileHandler(authParams),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return server.ListenAndServe()
}

func getKeySpace() gocassa.KeySpace {
	keySpace, err := gocassa.ConnectToKeySpace("user_space", []string{"127.0.0.1"}, "", "")
	if err != nil {
		panic(err)
	}

	return keySpace
}

func getUserTable() gocassa.Table {
	keySpace := getKeySpace()
	userTable := keySpace.Table("user", &user{}, gocassa.Keys{
		PartitionKeys: []string{"uid"},
	})
	err := userTable.CreateIfNotExist()
	if err != nil {
		panic(err)
	}
	return userTable
}
