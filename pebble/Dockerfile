FROM golang:1.11.1-alpine3.8

RUN apk add --update openssl git && \
    rm -rf /var/cache/apk/*

RUN go get -u github.com/zimosworld/pebble/...
RUN cd $GOPATH/src/github.com/zimosworld/pebble && \
        go install ./...

ENV PEBBLE_WFE_NONCEREJECT=0
ENV PEBBLE_VA_ALWAYS_VALID=1
ENV PEBBLE_VA_NOSLEEP=1

RUN mkdir -p /var/pebble/certs/ca
RUN mkdir /var/pebble/certs/localhost
RUN mkdir /var/pebble/config

ADD pebble-config.json /var/pebble/config/pebble-config.json
ADD entrypoint.sh /entrypoint

RUN chmod a+x /entrypoint

EXPOSE 443

ENTRYPOINT ["/entrypoint"]
CMD ["pebble", "-config", "/var/pebble/config/pebble-config.json"]
