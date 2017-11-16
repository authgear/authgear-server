FROM skygeario/skygear-godev:go1.8

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

# Copy a minimal set of files to restore Go dependencies to get advantage
# of Docker build cache
COPY Gopkg.toml Gopkg.lock /go/src/app/
RUN dep ensure

COPY . /go/src/app

RUN go-wrapper download && \
    go-wrapper install --tags zmq

VOLUME /go/src/app/data

EXPOSE 3000

CMD ["go-wrapper", "run"]
