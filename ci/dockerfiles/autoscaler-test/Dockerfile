FROM ubuntu:18.04

MAINTAINER qiye
RUN \
      apt-get update && \
      apt-get -qqy install --fix-missing \
            inetutils-ping \
            vim \
            curl \
            wget \
            unzip \
            zip \
            gcc \
      && \
      apt-get clean
# get maven 3.5.4
RUN wget --no-verbose -O /tmp/apache-maven-3.5.4.tar.gz http://archive.apache.org/dist/maven/maven-3/3.5.4/binaries/apache-maven-3.5.4-bin.tar.gz

# verify checksum
#RUN echo "87e5cc81bc4ab9b83986b3e77e6b3095 /tmp/apache-maven-3.5.4.tar.gz" | md5sum -c

# install maven
RUN tar xzf /tmp/apache-maven-3.5.4.tar.gz -C /opt/
RUN ln -s /opt/apache-maven-3.5.4 /opt/maven
RUN ln -s /opt/maven/bin/mvn /usr/local/bin
RUN rm -f /tmp/apache-maven-3.5.4.tar.gz
ENV MAVEN_HOME /opt/maven

# install openjdk10
#RUN \
#  wget -q -O openjdk-10.0.1_linux-x64_bin.tar.gz https://download.java.net/java/GA/jdk10/10.0.1/fb4372174a714e6b8c52526dc134031e/10/openjdk-10.0.1_linux-x64_bin.tar.gz \
#  && tar -xvf openjdk-10.0.1_linux-x64_bin.tar.gz \
#  && export PATH=$PATH:/jdk-10.0.1/bin \
#  && rm -f openjdk-10.0.1_linux-x64_bin.tar.gz \
#  && java -version
#ENV PATH /jdk-10.0.1/bin:$PATH 
#ENV JAVA_HOME /jdk-10.0.1
RUN apt-get install -y openjdk-8-jdk

# install nodejs
ENV NPM_CONFIG_LOGLEVEL info
ENV NODE_VERSION 8.11.2
RUN wget -q -O node-v8.11.2-linux-x64.tar.gz https://nodejs.org/dist/v8.11.2/node-v8.11.2-linux-x64.tar.gz
RUN tar -xvf node-v8.11.2-linux-x64.tar.gz
RUN export PATH=$PATH:/node-v8.11.2-linux-x64/bin
ENV PATH /node-v8.11.2-linux-x64/bin:$PATH 
RUN rm -f node-v8.11.2-linux-x64.tar.gz

# install golang
ENV GOPATH $HOME/go
ENV PATH $HOME/go/bin:/usr/local/go/bin:$PATH
RUN \
  wget -q https://dl.google.com/go/go1.11.linux-amd64.tar.gz -P /tmp && \
  tar xzvf /tmp/go1.11.linux-amd64.tar.gz -C /usr/local && \
  mkdir $GOPATH && \
  rm -rf /tmp/*

# install postgres10
RUN apt-get install -y gnupg gnupg2
ENV PGDATA /var/lib/postgresql/10/main
ENV PGCONFIG /etc/postgresql/10/main
RUN wget -q https://www.postgresql.org/media/keys/ACCC4CF8.asc -O- | apt-key add - 
RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ bionic-pgdg main" | tee /etc/apt/sources.list.d/postgresql.list 
RUN apt-get update \
  && apt-get install -y postgresql-10 \
  && sed -i 's/peer/trust/' ${PGCONFIG}/pg_hba.conf \
  && sed -i 's/md5/trust/' ${PGCONFIG}/pg_hba.conf 