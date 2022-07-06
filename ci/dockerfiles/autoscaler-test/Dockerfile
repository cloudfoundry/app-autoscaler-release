FROM ubuntu:22.04

MAINTAINER autoscaler-team

ENV DEBIAN_FRONTEND="noninteractive" TZ="Europe/London"

RUN \
      apt-get update && \
      apt-get -qqy install --fix-missing \
            lsb-release \
            build-essential \
            inetutils-ping \
            vim \
            curl \
            wget \
            unzip \
            zip \
            gcc \
            git \
            openjdk-11-jdk \
            gnupg \
            gnupg2 \
            ruby \
            ruby-dev &&\
    apt-get clean

RUN wget -q https://www.postgresql.org/media/keys/ACCC4CF8.asc -O- | apt-key add - 
RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ $(lsb_release -cs)-pgdg main" | tee /etc/apt/sources.list.d/postgresql.list 

ENV POSTGRES_VERSION 12
RUN \
      apt-get update && \
      apt-get install -y postgresql-${POSTGRES_VERSION} && \
      apt-get install -y libjson-perl && \
      apt-get clean

# get maven
ENV MAVEN_VERSION 3.6.3
RUN wget --no-verbose -O /tmp/apache-maven-${MAVEN_VERSION}.tar.gz http://archive.apache.org/dist/maven/maven-3/${MAVEN_VERSION}/binaries/apache-maven-${MAVEN_VERSION}-bin.tar.gz && \
	tar xzf /tmp/apache-maven-${MAVEN_VERSION}.tar.gz -C /opt/ && \
	ln -s /opt/apache-maven-${MAVEN_VERSION} /opt/maven && \
	ln -s /opt/maven/bin/mvn /usr/local/bin && \
	rm -f /tmp/apache-maven-${MAVEN_VERSION}.tar.gz
ENV MAVEN_HOME /opt/maven

# install golang
ENV GO_VERSION 1.18
ENV GOPATH $HOME/go
ENV PATH $HOME/go/bin:/usr/local/go/bin:$PATH
RUN \
  wget -q https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz -P /tmp && \
  tar xzvf /tmp/go${GO_VERSION}.linux-amd64.tar.gz -C /usr/local && \
  mkdir $GOPATH && \
  rm -rf /tmp/*

# Install bosh_cli
ENV BOSH_VERSION 6.4.4
RUN wget -q https://github.com/cloudfoundry/bosh-cli/releases/download/v${BOSH_VERSION}/bosh-cli-${BOSH_VERSION}-linux-amd64 && \
  mv bosh-cli-* /usr/local/bin/bosh && \
  chmod +x /usr/local/bin/bosh
# Install uaac
RUN gem install cf-uaac

# install postgres
ENV PGDATA /var/lib/postgresql/${POSTGRES_VERSION}/main
ENV PGCONFIG /etc/postgresql/${POSTGRES_VERSION}/main
RUN sed -i 's/peer/trust/' ${PGCONFIG}/pg_hba.conf \
  	&& sed -i 's/md5/trust/' ${PGCONFIG}/pg_hba.conf

# Install jq
ENV JQ_VERSION 1.6
RUN wget -q https://github.com/stedolan/jq/releases/download/jq-${JQ_VERSION}/jq-linux64 && \
  mv jq-linux64 /usr/local/bin/jq && \
  chmod +x /usr/local/bin/jq
