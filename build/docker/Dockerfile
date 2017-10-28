FROM alpine

RUN apk add --no-cache ca-certificates tini su-exec && update-ca-certificates
RUN adduser -D ondevice

COPY /build/docker/docker-entrypoint.sh /
COPY /target/alpine/ondevice /usr/local/bin/

# making the config directory a volume allows us to use --volumes-from
# in other containers to talk to this ondevice dameon
VOLUME /home/ondevice/.config/ondevice/

# default credentials for the 'demo' user:
ENV ONDEVICE_USER=demo
ENV ONDEVICE_AUTH=5h42l5xylznw

ENTRYPOINT [ "tini", "--", "/docker-entrypoint.sh" ]
#ENTRYPOINT [ "/usr/local/bin/ondevice" ]
