#!/bin/bash

#Java Setup
setup_java(){
  cd ${BUILD_DIR}
  # Make sure we can see uname
  export PATH=$PATH:/bin:/usr/bin

  #unpack Java - we support Mac OS 64bit and Linux 64bit otherwise we require JAVA_HOME to point to JDK
  mkdir java
  case "$OSTYPE" in
    darwin*)
      tar zxvf openjdk/openjdk-10.0.1_osx-x64_bin.tar.gz -C java --strip-components=4
      ;;
    linux*)
      tar zxvf openjdk/openjdk-10.0.1_linux-x64_bin.tar.gz -C java --strip-components=1
      ;;
    *)
      if [ ! -d $JAVA_HOME ]; then
        echo "Set JAVA_HOME properly for non Linux/Darwin builds."
        exit 1
      fi
      ;;
  esac
  export JAVA_HOME=${BUILD_DIR}/java

  #setup Java path
  export PATH=$JAVA_HOME/bin:$PATH
}

#Maven Setup
setup_maven(){
  cd ${BUILD_DIR}
  tar zxvf maven/apache-maven-3.5.4-bin.tar.gz -C maven --strip-components=1
  export M2_HOME=${BUILD_DIR}/maven
  export PATH=$M2_HOME/bin:$PATH
}

#Cleanup Java files from BUILD_DIR
cleanup_java(){
  cd ${BUILD_DIR}
  rm -rf java
  rm -rf openjdk
}

#Cleanup Maven files from BUILD_DIR
cleanup_maven(){
  cd ${BUILD_DIR}
  rm -rf maven
}

