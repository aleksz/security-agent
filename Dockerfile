FROM resin/rpi-raspbian:jessie
#FROM debian:jessie

ENV GO_VERSION=1.7.4
ENV GO_OS=linux
#ENV GO_ARCH=amd64
ENV GO_ARCH=armv6l
ENV CGO_ENABLED=0
ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/home/security-agent/go

RUN useradd -ms /bin/bash security-agent && adduser security-agent dialout
WORKDIR /home/security-agent

RUN apt-get update && \
	apt-get install -y curl git gcc && \
	curl -O https://storage.googleapis.com/golang/go$GO_VERSION.$GO_OS-$GO_ARCH.tar.gz && \
	tar -C /usr/local -xzf go$GO_VERSION.$GO_OS-$GO_ARCH.tar.gz && \
	rm go$GO_VERSION.$GO_OS-$GO_ARCH.tar.gz && \
	rm -rf /var/lib/apt/lists/*

ADD . go/src/github.com/aleksz/security-agent/

RUN cd go/src/github.com/aleksz/security-agent/ && \
	go get ./... && \
	go build . && \
	ls -lh && cp security-agent /home/security-agent/security-agent && \
	cd ~ && rm -rf go && rm -rf /usr/local/go && \
	chown security-agent /home/security-agent

USER security-agent
ENTRYPOINT ["/home/security-agent/security-agent", "/home/security-agent/config.yml"]