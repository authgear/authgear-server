#!/bin/sh -eu
MODULE=$1

mkdir -p tmp
make -e DOC_PATH="tmp/doc-$MODULE.yaml" "generate-doc-$MODULE"
docker run --rm -p 9001:8080 -e URL="/doc-$MODULE.yaml" -v "$PWD/tmp/doc-$MODULE.yaml:/usr/share/nginx/html/doc-$MODULE.yaml" swaggerapi/swagger-ui &
sleep 1
python -m webbrowser http://localhost:9001
wait
