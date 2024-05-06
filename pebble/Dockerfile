FROM golang:1.21-alpine

RUN apk add --no-cache openssl git ca-certificates bash curl vim

RUN mkdir -p /var/pebble/src && \
  git clone \
  https://github.com/podcloud/pebble-static-CA.git \
  /var/pebble/src

WORKDIR /var/pebble/src

RUN go install ./cmd/pebble

ENV PEBBLE_WFE_NONCEREJECT=0
ENV PEBBLE_VA_ALWAYS_VALID=1
ENV PEBBLE_VA_NOSLEEP=1

RUN mkdir -p /var/pebble/certs/ca && \
  mkdir -p /var/pebble/certs/server && \
  mkdir -p /var/pebble/config

COPY pebble-config.json /var/pebble/config/pebble-config.json
COPY entrypoint.sh /entrypoint

RUN chmod a+x /entrypoint

EXPOSE 443

ENTRYPOINT ["/entrypoint"]
CMD ["pebble", "-config", "/var/pebble/config/pebble-config.json"]
