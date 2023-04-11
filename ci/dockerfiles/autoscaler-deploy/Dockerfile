FROM ubuntu:jammy
MAINTAINER autoscaler-team

RUN apt-get update && \
    apt-get -qqy install --fix-missing gnupg apt-transport-https wget && \
    apt-get clean
 
RUN wget -q -O - https://packages.cloudfoundry.org/debian/cli.cloudfoundry.org.key | apt-key add - && \
      echo "deb https://packages.cloudfoundry.org/debian stable main" | tee /etc/apt/sources.list.d/cloudfoundry-cli.list && \
      wget -q -O - https://cli.github.com/packages/githubcli-archive-keyring.gpg | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg && \
      echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null

RUN apt-get update && \
    apt-get -qqy install --fix-missing \
            build-essential \
            inetutils-ping \
            vim \
            curl \
            libcurl4-openssl-dev \
            git \
            wget \
            runit \
            openssl \
            unzip \
            zip \
            ruby \
            ruby-dev \
            python3 \
            ca-certificates \
            cf8-cli \
            gh && \
    apt-get clean

# install golang
# renovate: datasource=golang-version depName=golang
ARG GO_VERSION=1.18
RUN wget -q https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz && \
  tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
  rm -rf go${GO_VERSION}.linux-amd64.tar.gz && \
  mkdir -p /go/bin

# set GOPATH
ENV GOPATH /go
# add go and GOPATH/bin to PATH
ENV PATH $PATH:$GOPATH/bin:/usr/local/go/bin

# Install bosh_cli
# renovate: datasource=github-releases depName=bosh-cli lookupName=cloudfoundry/bosh-cli
ENV BOSH_VERSION 7.2.2
RUN wget -q https://github.com/cloudfoundry/bosh-cli/releases/download/v${BOSH_VERSION}/bosh-cli-${BOSH_VERSION}-linux-amd64 && \
  mv bosh-cli-* /usr/local/bin/bosh && \
  chmod +x /usr/local/bin/bosh

# Install bbl
# renovate: datasource=github-releases depName=bosh-bootloader lookupName=cloudfoundry/bosh-bootloader
ENV BBL_VERSION v8.4.111
RUN wget -q https://github.com/cloudfoundry/bosh-bootloader/releases/download/${BBL_VERSION}/bbl-${BBL_VERSION}_linux_x86-64 && \
  mv bbl-* /usr/local/bin/bbl && \
  chmod +x /usr/local/bin/bbl

# Install credhub
# renovate: datasource=github-releases depName=credhub-cli lookupName=cloudfoundry/credhub-cli
ENV CREDHUB_VERSION 2.9.13
RUN wget -q https://github.com/cloudfoundry-incubator/credhub-cli/releases/download/${CREDHUB_VERSION}/credhub-linux-${CREDHUB_VERSION}.tgz && \
  tar xvfz credhub-linux-${CREDHUB_VERSION}.tgz && \
  mv credhub /usr/local/bin/credhub

# Install uaac
RUN gem install cf-uaac

# Install jq as a nice to have on container debugging
# renovate: datasource=github-releases depName=jq lookupName=stedolan/jq
ENV JQ_VERSION jq-1.6
RUN wget -q https://github.com/stedolan/jq/releases/download/${JQ_VERSION}/jq-linux64 && \
    mv jq-linux64 /usr/local/bin/jq && \
    chmod +x /usr/local/bin/jq

# install yq
# renovate: datasource=github-releases depName=yq lookupName=mikefarah/yq
ARG YQ_VERSION=v4.33.3
RUN wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64 && \
    chmod a+x /usr/local/bin/yq && \
    yq --version

# Install gcloud
RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] http://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
  curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key --keyring /usr/share/keyrings/cloud.google.gpg  add - && \
  apt-get update -y &&\
  apt-get install google-cloud-sdk -y &&\
  apt-get clean &&\
  gcloud version

# renovate: datasource=github-releases depName=ginkgo lookupName=onsi/ginkgo
ARG GINKGO_VERSION=v2.9.2
RUN go install github.com/onsi/ginkgo/v2/ginkgo@${GINKGO_VERSION} &&\
    ginkgo version

ARG NODE_VERSION=18
RUN curl -sL "https://deb.nodesource.com/setup_${NODE_VERSION}.x" | bash - \
    && apt-get update -y \
    && apt install nodejs -y \
    && apt-get clean \
    && npm --version \
    && node -v
