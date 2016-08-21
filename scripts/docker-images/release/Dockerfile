FROM skygeario/skygear-server-deps:latest

ARG version
ARG sha
ARG build_date

ENV SKYGEAR_VERSION=$version

LABEL \
    io.skygear.role=server \
    io.skygear.repo=SkygearIO/skygear-server \
    io.skygear.commit=$sha \
    io.skygear.version=$version \
    io.skygear.build-date=$build_date

ADD skygear-server /usr/local/bin/
RUN chmod a+x /usr/local/bin/skygear-server \
    && mkdir -p /app/data \
    && chown nobody:nobody /app/data

WORKDIR /app
VOLUME /app/data
USER nobody

EXPOSE 3000

CMD ["/usr/local/bin/skygear-server"]
