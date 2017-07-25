FROM alpine:3.4

RUN apk --update --no-cache add libc6-compat libstdc++ zlib ca-certificates \
        libsodium libzmq

RUN apk --update --no-cache add --virtual build-dependencies \
        coreutils gcc tar libtool zlib make musl-dev openssl-dev g++ \
        zlib-dev libsodium-dev zeromq-dev git automake autoconf && \
    git clone --branch v4.0.2 --depth 1 git://github.com/zeromq/czmq.git && \
    ( \
        cd czmq; \
        ./autogen.sh; \
        ./configure --prefix=/usr \
                    --sysconfdir=/etc \
                    --mandir=/usr/share/man \
                    --infodir=/usr/share/info; \
        make; make install; \
    ) && rm -rf czmq && \
    apk del build-dependencies && \
    ln -s /lib /lib64
