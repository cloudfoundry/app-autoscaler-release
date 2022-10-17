FROM mikefarah/yq:4.27.5 AS yq

FROM ubuntu:22.04
MAINTAINER autoscaler-team

ENV DEBIAN_FRONTEND="noninteractive" TZ="Europe/London"

RUN apt-get update && \
    apt-get -qqy install --fix-missing gnupg apt-transport-https wget && \
    apt-get clean

RUN wget -q -O - https://packages.cloudfoundry.org/debian/cli.cloudfoundry.org.key | apt-key add - && \
      echo "deb https://packages.cloudfoundry.org/debian stable main" | tee /etc/apt/sources.list.d/cloudfoundry-cli.list && \
      wget -q -O - https://cli.github.com/packages/githubcli-archive-keyring.gpg | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg && \
      echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null

RUN apt-get update && \
    apt-get -qqy install --fix-missing \
            lsb-release \
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
            gcc \
            openjdk-11-jdk \
            zip \
            ruby \
            ruby-dev \
            python3 \
            ca-certificates \
            cf8-cli \
            gnupg \
            gnupg2 \
            gh && \
    apt-get clean

RUN wget -q https://www.postgresql.org/media/keys/ACCC4CF8.asc -O- | apt-key add -
RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ $(lsb_release -cs)-pgdg main" | tee /etc/apt/sources.list.d/postgresql.list

# install golang
ARG GO_VERSION=1.19
ENV GOPATH $HOME/go
ENV PATH $HOME/go/bin:/usr/local/go/bin:$PATH
RUN wget -q https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz -P /tmp &&\
    tar xzvf /tmp/go${GO_VERSION}.linux-amd64.tar.gz -C /usr/local &&\
    mkdir $GOPATH &&\
    rm -rf /tmp/* &&\
    go version

ENV POSTGRES_VERSION 12
RUN apt-get update && \
    apt-get install -y postgresql-${POSTGRES_VERSION} && \
    apt-get install -y libjson-perl && \
    apt-get clean

# Install bosh_cli
ENV BOSH_VERSION 6.4.4
RUN wget -q https://github.com/cloudfoundry/bosh-cli/releases/download/v${BOSH_VERSION}/bosh-cli-${BOSH_VERSION}-linux-amd64 && \
  mv bosh-cli-* /usr/local/bin/bosh && \
  chmod +x /usr/local/bin/bosh

# Install bbl
ENV BBL_VERSION 8.4.44
RUN wget -q https://github.com/cloudfoundry/bosh-bootloader/releases/download/v${BBL_VERSION}/bbl-v${BBL_VERSION}_linux_x86-64 && \
  mv bbl-* /usr/local/bin/bbl &&\
  chmod +x /usr/local/bin/bbl &&\
  bbl --version

# Install credhub
ENV CREDHUB_VERSION 2.9.0
RUN wget -q https://github.com/cloudfoundry-incubator/credhub-cli/releases/download/${CREDHUB_VERSION}/credhub-linux-${CREDHUB_VERSION}.tgz && \
  tar xvfz credhub-linux-${CREDHUB_VERSION}.tgz && \
  mv credhub /usr/local/bin/credhub &&\
  credhub --version

# Install uaac
RUN gem install cf-uaac &&\
    uaac --version

# Install jq as a nice to have on container debugging
ENV JQ_VERSION 1.6
RUN wget -q https://github.com/stedolan/jq/releases/download/jq-${JQ_VERSION}/jq-linux64 && \
    mv jq-linux64 /usr/local/bin/jq && \
    chmod +x /usr/local/bin/jq &&\
    jq --version

# get maven
ENV MAVEN_VERSION 3.6.3
ENV MAVEN_HOME /opt/maven
RUN wget --no-verbose -O /tmp/apache-maven-${MAVEN_VERSION}.tar.gz http://archive.apache.org/dist/maven/maven-3/${MAVEN_VERSION}/binaries/apache-maven-${MAVEN_VERSION}-bin.tar.gz && \
	tar xzf /tmp/apache-maven-${MAVEN_VERSION}.tar.gz -C /opt/ && \
	ln -s /opt/apache-maven-${MAVEN_VERSION} /opt/maven && \
	ln -s /opt/maven/bin/mvn /usr/local/bin && \
	rm -f /tmp/apache-maven-${MAVEN_VERSION}.tar.gz &&\
    mvn --version


# install postgres
ENV PGDATA /var/lib/postgresql/${POSTGRES_VERSION}/main
ENV PGCONFIG /etc/postgresql/${POSTGRES_VERSION}/main
RUN sed -i 's/peer/trust/' ${PGCONFIG}/pg_hba.conf \
  	&& sed -i 's/md5/trust/' ${PGCONFIG}/pg_hba.conf

# Install gcloud
RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] http://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
  curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key --keyring /usr/share/keyrings/cloud.google.gpg  add - && \
  apt-get update -y &&\
  apt-get install google-cloud-sdk -y &&\
  apt-get clean &&\
  gcloud version


ARG GINKGO_VERSION=2.1.3
RUN go install github.com/onsi/ginkgo/v2/ginkgo@v${GINKGO_VERSION} && \
    ginkgo version

RUN curl -sL https://deb.nodesource.com/setup_16.x | bash - \
    && apt-get update -y \
    && apt install nodejs -y \
    && apt-get clean \
    && npm --version \
    && node -v

RUN  cf install-plugin -f -r CF-Community app-autoscaler-plugin\
     && cf plugins
