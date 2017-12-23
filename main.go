package main

import (
	"log"

	"github.com/zi-yang-zhang/cryptopia-api/profile"
	"golang.org/x/sync/errgroup"
)

var (
	g errgroup.Group
)

func main() {

	g.Go(func() error {
		return profile.ProfileService()
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
