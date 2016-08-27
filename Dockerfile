FROM skygeario/skygear-godev:latest

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

# Copy a minimal set of files to restore Go dependencies to get advantage
# of Docker build cache
COPY glide.yaml glide.lock /go/src/app/
RUN glide install --skip-test

COPY . /go/src/app

RUN go-wrapper download && \
    go-wrapper install --tags zmq

VOLUME /go/src/app/data

EXPOSE 3000

CMD ["go-wrapper", "run"]
