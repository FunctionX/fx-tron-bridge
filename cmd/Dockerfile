FROM alpine:latest

WORKDIR root

COPY ./build/bin/fxtronbridge /usr/bin/fxtronbridge

ENV FX_ADDRESS_PREFIX="fx"

EXPOSE 9811/tcp

VOLUME ["/root"]

ENTRYPOINT ["fxtronbridge"]
