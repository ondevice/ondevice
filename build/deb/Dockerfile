# multi-stage build for the ondevice .deb packages

#
# stage0: build
#
ARG SOURCE_IMAGE
FROM ${SOURCE_IMAGE}

#ARG ARCH
#RUN test "${ARCH}" = "$(dpkg --print-architecture)"

RUN apt-get update && apt-get -y install devscripts git curl wget

RUN adduser --gecos builduser,,, --disabled-password user
RUN mkdir -p /target/

# install go
#WORKDIR /usr/local/
#ARG GOARCH
#RUN wget https://storage.googleapis.com/golang/go1.9.2.linux-${GOARCH}.tar.gz
#RUN tar xfz go1.9.2*.tar.gz
#RUN ln -s /usr/local/go/bin/* /usr/local/bin/

#ENV GOPATH=/go/
#ENV PATH=/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
#RUN mkdir -p /go/bin /go/src

# install glide
RUN curl https://glide.sh/get | sh

COPY / /go/src/github.com/ondevice/ondevice/

COPY /build/deb/debian/ /go/src/github.com/ondevice/ondevice/debian

RUN chown -R user:user /go/ /target/
USER user
WORKDIR /go/src/github.com/ondevice/ondevice/

RUN glide install
ARG BUILD_ARGS
RUN dpkg-buildpackage -us -uc ${BUILD_ARGS}
RUN cp -a ../ondevice*.* /target/

#
# test stage (doesn't support cross-compiling)
#
#FROM ${SOURCE_IMAGE}
#RUN apt-get update && apt-get -y install ssh rsync ca-certificates

#RUN mkdir -p /target/
#COPY --from=0 /target/ /target/
#COPY /build/test.sh /
#RUN dpkg -i /target/ondevice_*.deb
#RUN apt-get -f install
#RUN /test.sh

#WORKDIR /target/
