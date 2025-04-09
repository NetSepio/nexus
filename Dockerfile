#LABEL Maintainer Punarv Name punarv@netsepio.com

FROM golang:alpine AS build-app
RUN apk update && apk add --no-cache git
WORKDIR /app
COPY . .
RUN go build -ldflags "-X main.version=1.1.1-alpha -X main.codeHash=$(git rev-parse HEAD)" -o erebrus .

FROM alpine:latest
RUN apk update && apk add --no-cache git
WORKDIR /app
COPY --from=build-app /app/erebrus .
COPY --from=build-app /app/webapp ./webapp
COPY wg-watcher.sh .
RUN chmod +x ./erebrus ./wg-watcher.sh
RUN apk update && apk add --no-cache bash openresolv bind-tools wireguard-tools gettext inotify-tools iptables
ENV LOAD_CONFIG_FILE=$LOAD_CONFIG_FILE RUNTYPE=$RUNTYPE SERVER=$SERVER HTTP_PORT=$HTTP_PORT GRPC_PORT=$GRPC_PORT GATEWAY_DOMAIN=$GATEWAY_DOMAIN
ENV NODE_NAME=$NODE_NAME REGION=$REGION DOMAIN=$DOMAIN REGION_NAME=$REGION_NAME REGION_CODE=$REGION_CODE
ENV WG_CONF_DIR=$WG_CONF_DIR WG_CLIENTS_DIR=$WG_CLIENTS_DIR WG_KEYS_DIR=$WG_KEYS_DIR WG_INTERFACE_NAME=$WG_INTERFACE_NAME
ENV WG_ENDPOINT_HOST=$WG_ENDPOINT_HOST WG_ENDPOINT_PORT=$WG_ENDPOINT_PORT WG_IPv4_SUBNET=$WG_IPv4_SUBNET WG_IPv6_SUBNET=$WG_IPv6_SUBNET
ENV WG_DNS=$WG_DNS WG_ALLOWED_IP_1=$WG_ALLOWED_IP_1 WG_ALLOWED_IP_2=$WG_ALLOWED_IP_2
ENV WG_PRE_UP=$WG_PRE_UP WG_POST_UP=$WG_POST_UP WG_PRE_DOWN=$WG_PRE_DOWN WG_POST_DOWN=$WG_POST_DOWN
ENV NODE_CONFIG=$NODE_CONFIG NODE_ACCESS=$NODE_ACCESS
RUN echo $'#!/usr/bin/env bash\n\
    set -eo pipefail\n\
    /app/erebrus &\n\
    ./wg-watcher.sh\n\
    sleep infinity' > /app/start.sh && chmod +x /app/start.sh
ENTRYPOINT ["/app/start.sh"]