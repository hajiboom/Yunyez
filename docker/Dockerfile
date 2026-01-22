# basic golang image
FROM golang:1.24.8-alpine3.21

RUN apk update && \
    apk add --no-cache \
        curl \
        tcpdump \
        ca-certificates \
        tzdata && \
    rm -rf /var/cache/apk/*

ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

