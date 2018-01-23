package main

import (
	"log"

	"encoding/json"
	"github.com/zi-yang-zhang/cryptopia-api/core"
	"github.com/zi-yang-zhang/cryptopia-api/profile"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
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
		return profile.ProfileService(config.AuthParams)
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
