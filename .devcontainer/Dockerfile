FROM composer:2.6 as composer
FROM golang:1.21 as go
FROM php:8.2-cli as base

ENV GOROOT /usr/local/go

COPY --from=composer /usr/bin/composer /usr/bin/composer
COPY --from=go /usr/local/go /usr/local/go
COPY --from=mlocati/php-extension-installer /usr/bin/install-php-extensions /usr/local/bin/
RUN \
    apt-get update && apt-get install -y procps git && \
    install-php-extensions opcache sockets protobuf msgpack curl zip

