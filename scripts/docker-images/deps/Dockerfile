FROM alpine:3.4

ENV LIBSODIUM_VERSION=1.0.5 \
    ZMQ_VERSION=4.1.3 \
    CZMQ_VERSION=3.0.2

RUN apk --update --no-cache add libc6-compat libstdc++ zlib

RUN apk --update --no-cache add --virtual build-dependencies \
        coreutils gcc tar libtool zlib make musl-dev openssl-dev g++ \
        zlib-dev git automake autoconf && \
    git clone --branch $LIBSODIUM_VERSION --depth 1 git://github.com/jedisct1/libsodium.git && \
    ( \
        cd libsodium; \
        ./autogen.sh; \
        ./configure --prefix=/usr \
                    --sysconfdir=/etc \
                    --mandir=/usr/share/man \
                    --infodir=/usr/share/info; \
        make; make install; \
    ) && rm -rf libsodium && \
    git clone --branch v$ZMQ_VERSION --depth 1 git://github.com/zeromq/zeromq4-1.git && \
    ( \
        cd zeromq4-1; \
        ./autogen.sh; \
        ./configure --prefix=/usr \
                    --sysconfdir=/etc \
                    --mandir=/usr/share/man \
                    --infodir=/usr/share/info; \
        make; make install; \
    ) && rm -rf zeromq4-1 && \
    git clone --branch v$CZMQ_VERSION --depth 1 git://github.com/zeromq/czmq.git && \
    ( \
        cd czmq; \
        ./autogen.sh; \
        ./configure --prefix=/usr \
                    --sysconfdir=/etc \
                    --mandir=/usr/share/man \
                    --infodir=/usr/share/info; \
        make; make install; \
    ) && rm -rf czmq && \
    apk del build-dependencies

RUN ln -s /lib /lib64
