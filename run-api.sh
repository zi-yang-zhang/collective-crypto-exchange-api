#!/usr/bin/env bash
docker run --rm --name cryptopia-api --net="host" -p 8080:8080 -v $PWD:/etc/krakend/ devopsfaith/krakend run -d -c /etc/krakend/krakend.json

