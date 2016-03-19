FROM alpine:3.3

ENV LIBSODIUM_VERSION=1.0.5 \
    ZMQ_VERSION=4.1.3 \
    CZMQ_VERSION=3.0.2

RUN apk update \
    && apk add libc6-compat libstdc++ curl \
    && apk add \
        coreutils gcc tar libtool zlib make musl-dev openssl-dev g++ zlib-dev \
    && echo " ... adding libsodium" \
        && curl -L https://download.libsodium.org/libsodium/releases/libsodium-${LIBSODIUM_VERSION}.tar.gz -o /tmp/libsodium.tar.gz \
        && cd /tmp/ \
        && echo "bfcafc678c7dac87866c50f9b99aa821750762edcf8e56fc6d13ba0ffbef8bab libsodium.tar.gz" | sha256sum -c - \
        && tar -xf /tmp/libsodium.tar.gz \
        && cd /tmp/libsodium*/ \
        && ./configure --prefix=/usr \
                       --sysconfdir=/etc \
                       --mandir=/usr/share/man \
                       --infodir=/usr/share/info \
        && make && make install \
        && rm -rf /tmp/libsodium* \
        && cd .. \
    && echo " ... adding zeromq" \
        && curl -L http://download.zeromq.org/zeromq-${ZMQ_VERSION}.tar.gz -o /tmp/zeromq.tar.gz \
        && cd /tmp/ \
        && echo "61b31c830db377777e417235a24d3660a4bcc3f40d303ee58df082fcd68bf411 zeromq.tar.gz" | sha256sum -c - \
        && tar -xf /tmp/zeromq.tar.gz \
        && cd /tmp/zeromq*/ \
        && ./configure --prefix=/usr \
                       --sysconfdir=/etc \
                       --mandir=/usr/share/man \
                       --infodir=/usr/share/info \
        && make && make install \
        && rm -rf /tmp/zeromq* \
        && cd .. \
    && echo " ... adding czmq" \
        && curl -L http://download.zeromq.org/czmq-${CZMQ_VERSION}.tar.gz -o /tmp/czmq.tar.gz \
        && cd /tmp/ \
        && echo "8bca39ab69375fa4e981daf87b3feae85384d5b40cef6adbe9d5eb063357699a czmq.tar.gz" | sha256sum -c - \
        && tar -xf /tmp/czmq.tar.gz \
        && cd /tmp/czmq*/ \
        && ./configure --prefix=/usr \
                       --sysconfdir=/etc \
                       --mandir=/usr/share/man \
                       --infodir=/usr/share/info \
        && make && make install \
        && rm -rf /tmp/czmq* \
        && cd .. \
    && apk del coreutils gcc tar libtool make musl-dev openssl-dev g++ zlib-dev \
    && rm -rf /var/cache/apk/* \
    && ln -s /lib /lib64

ADD skygear-server /usr/local/bin/
RUN chmod a+x /usr/local/bin/skygear-server \
    && mkdir -p /app/data

WORKDIR /app
VOLUME /app/data
USER nobody
ADD config.ini /app/

EXPOSE 3000

CMD ["/usr/local/bin/skygear-server", "config.ini"]
