docker run --rm -p 8080:8080 -v $PWD:/etc/krakend/ --name cryptopia-api --net="host" devopsfaith/krakend run -d
