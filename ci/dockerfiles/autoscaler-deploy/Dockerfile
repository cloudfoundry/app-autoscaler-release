FROM ubuntu:16.04
MAINTAINER qiye

RUN \
      apt-get update && \
      apt-get -qqy install --fix-missing \
            build-essential \
            curl \
            libcurl4-openssl-dev \
            git \
            libreadline6 \
            libreadline6-dev \
            wget \
            runit \
            openssl \
            unzip \
            zip \
            iptables \
      && \
      apt-get clean

# Install go
RUN wget https://storage.googleapis.com/golang/go1.9.2.linux-amd64.tar.gz && \
  tar -C /usr/local -xzf go1.9.2.linux-amd64.tar.gz && \
  rm -rf go1.9.2.linux-amd64.tar.gz

# Create directory for GOPATH
RUN mkdir -p /go/bin

# set GOPATH
ENV GOPATH /go


# add go and GOPATH/bin to PATH
ENV PATH $PATH:$GOPATH/bin:/usr/local/go/bin

# Install bosh_cli
RUN wget https://github.com/cloudfoundry/bosh-cli/releases/download/v5.3.1/bosh-cli-5.3.1-linux-amd64 && \
  mv bosh-cli-* /usr/local/bin/bosh && \
  chmod +x /usr/local/bin/bosh


# Install jq as a nice to have on container debugging
RUN wget https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64 && \
  mv jq-linux64 /usr/local/bin/jq && \
  chmod +x /usr/local/bin/jq

# Install yq
RUN wget https://github.com/mikefarah/yq/releases/download/2.1.2/yq_linux_amd64 && \
  mv yq_linux_amd64 /usr/local/bin/yq && \
  chmod +x /usr/local/bin/yq
