package main

import (
	"encoding/json"
	"github.com/zi-yang-zhang/cryptopia-api/core"
	"github.com/zi-yang-zhang/cryptopia-api/gateway"
	"github.com/zi-yang-zhang/cryptopia-api/profile"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"log"
)

var (
	g errgroup.Group
)

func main() {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	config := core.Config{}
	json.Unmarshal(data, &config)

	g.Go(func() error {
		return profile.Start(config.AuthParams)
	})

	gateway.Start()

}
