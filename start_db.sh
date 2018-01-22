#!/usr/bin/env bash
docker run -it --rm --name user-db -v $PWD/db:/var/lib/cassandra -p 9042:9042 cassandra:3.0.15
