FROM resin/rpi-raspbian:jessie
#FROM debian:jessie

ENV GO_VERSION=1.7.4
ENV GO_OS=linux
#ENV GO_ARCH=amd64
ENV GO_ARCH=armv6l
ENV CGO_ENABLED=0 #could not run go build witout it
ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/home/security-agent/go

RUN useradd -ms /bin/bash security-agent && adduser security-agent dialout
WORKDIR /home/security-agent
ADD . go/src/github.com/aleksz/security-agent/

RUN apt-get update && \
	apt-get install -y curl git gcc && \
	curl -O https://storage.googleapis.com/golang/go$GO_VERSION.$GO_OS-$GO_ARCH.tar.gz && \
	tar -C /usr/local -xzf go$GO_VERSION.$GO_OS-$GO_ARCH.tar.gz && \
	rm go$GO_VERSION.$GO_OS-$GO_ARCH.tar.gz && \
	cd go/src/github.com/aleksz/security-agent/ && \
	go get ./... && \
	go build . && \
	ls -lh && cp security-agent /home/security-agent/security-agent && \
	cd ~ && rm -rf go && rm -rf /usr/local/go && \
	apt-get remove -y curl git gcc && \
	apt-get -y autoremove && \
	apt-get clean && \
	chown security-agent /home/security-agent

USER security-agent
ENTRYPOINT ["/home/security-agent/security-agent", "/home/security-agent/config.yml"]