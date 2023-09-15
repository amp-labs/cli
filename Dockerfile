FROM alpine:latest

RUN apk --no-cache add curl

RUN apk --no-cache add ca-certificates \
  && update-ca-certificates

COPY amp /usr/bin

ENTRYPOINT ["/usr/bin/amp"]
