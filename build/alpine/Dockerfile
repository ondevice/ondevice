# Alpine Linux build image
#
# Use `make build-alpine` in the repo's root directory to build this image
#
# This serves as base for the ondevice docker image in build/docker/

FROM golang:alpine

RUN apk add --no-cache git glide

#RUN git clone -b stable --single-branch https://github.com/ondevice/ondevice /go/src/github.com/ondevice/ondevice/
COPY ./ /go/src/github.com/ondevice/ondevice/

WORKDIR /go/src/github.com/ondevice/ondevice/

RUN glide update
RUN go build -o /build/ondevice
