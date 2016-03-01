#!/bin/sh

set -e

: ${DIST:=/home/ubuntu/myapp}

while getopts ":s:n:r:" opt; do
  case $opt in
    s)
      STACK_NAME=$OPTARG
      ;;
    n)
      RESOURCE_NAME=$OPTARG
      ;;
    r)
      AWS_REGION=$OPTARG
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      exit 1
      ;;
    :)
      echo "Option -$OPTARG requires an argument." >&2
      exit 1
      ;;
  esac
done

# Generate self-signed cert for SSL
if [ ! -f $DIST/nginx-privkey.pem ]; then
  openssl req -x509 -newkey rsa:4096 -nodes -days 365 \
    -subj "/C=AU/ST=Some-State/O=Internet Widgits Pty Ltd/CN=localhost" \
    -keyout $DIST/nginx-privkey.pem -out $DIST/nginx-cert.pem
  chown ubuntu:ubuntu $DIST/nginx-privkey.pem $DIST/nginx-cert.pem
fi

if [ ! -f $DIST/development.ini ]; then
  if [ ! -z "$STACK_NAME" ] && [ ! -z "$RESOURCE_NAME" ] && [ ! -z "$AWS_REGION" ]; then
    METADATA=`cfn-get-metadata -s "$STACK_NAME" -r "$RESOURCE_NAME" --region "$AWS_REGION"`
    if [ $? -eq 1 ]; then
      METADATA="{}"
    fi
  else
    METADATA="{}"
  fi
  echo $METADATA | jinja2 $DIST/development.ini.tmpl \
    -D api_key=changeme \
    -D app_name=myapp \
    > $DIST/development.ini
  chown ubuntu:ubuntu $DIST/development.ini
fi

docker-compose -f $DIST/docker-compose.yml pull
docker-compose -f $DIST/docker-compose.yml up -d db redis web server

if [ ! -z "$STACK_NAME" ] && [ ! -z "$RESOURCE_NAME" ] && [ ! -z "$AWS_REGION" ]; then
  # NOTE: Should change this to checking the server http port for a success
  # response instead of waiting.
  sleep 10
  cfn-signal --success true --stack "$STACK_NAME" --resource "$RESOURCE_NAME" --region "$AWS_REGION"
fi
