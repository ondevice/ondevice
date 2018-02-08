# Builds the official ondevice docker image
#
# Use `make build-docker` in the repo's root directory to build this image
#
# This docker image uses a multi-stage build (introduced in Docker v17.05)
FROM golang:alpine

RUN apk add --no-cache git glide

#COPY ./ /go/src/github.com/ondevice/ondevice/
COPY /.git/ /tmp/src.git/
RUN git clone /tmp/src.git/ /go/src/github.com/ondevice/ondevice/

WORKDIR /go/src/github.com/ondevice/ondevice/

# trying to pinpoint a docker hub build issue:
RUN glide update
ARG VERSION
RUN go build -ldflags "-X github.com/ondevice/ondevice/config.version=${VERSION}" -o /build/ondevice

#
# stage 2
#
FROM alpine

RUN apk add --no-cache ca-certificates tini su-exec openssh rsync && update-ca-certificates
RUN adduser -D ondevice

COPY /build/docker/docker-entrypoint.sh /
COPY --from=0 /build/ondevice /usr/local/bin/

# making the config directory a volume allows us to use --volumes-from
# in other containers to talk to this ondevice dameon
VOLUME /home/ondevice/.config/ondevice/

# default credentials for the 'demo' user:
ENV ONDEVICE_USER=demo
ENV ONDEVICE_AUTH=5h42l5xylznw
ENV SSH_ADDR=ssh:22

ENTRYPOINT [ "tini", "--", "/docker-entrypoint.sh" ]
#ENTRYPOINT [ "/usr/local/bin/ondevice" ]
