FROM alpine:3.4

RUN apk --update --no-cache add libc6-compat libstdc++ zlib ca-certificates \
        libsodium libzmq && \
    ln -s /lib /lib64
