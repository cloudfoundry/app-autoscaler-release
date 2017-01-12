#Java Setup
setup_java(){
  cd ${BUILD_DIR}
  # Make sure we can see uname
  export PATH=$PATH:/bin:/usr/bin

  #unpack Java - we support Mac OS 64bit and Linux 64bit otherwise we require JAVA_HOME to point to JDK
  mkdir java
  case "$OSTYPE" in
    darwin*)
      tar zxvf openjdk/openjdk-1.8.0_101-x86_64-mountainlion.tar.gz -C java
      ;;
    linux*)
      tar zxvf openjdk/openjdk-1.8.0_101-x86_64-trusty.tar.gz -C java
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
  tar zxvf maven/apache-maven-3.3.9-bin.tar.gz -C maven --strip-components=1
  export M2_HOME=${BUILD_DIR}/maven
  export PATH=$M2_HOME/bin:$PATH
}


#NPM setup
setup_npm(){
  cd ${BUILD_DIR}
  # Make sure we can see uname
  export PATH=$PATH:/bin:/usr/bin
  #unpack NodeJs - we support Mac OS 64bit and Linux 64bit
  mkdir npm
  case "$OSTYPE" in
    darwin*)
      tar xvf nodejs/node-v6.9.2-darwin-x64.tar.gz -C npm --strip-components=1
      ;;
    linux*)
      tar xvf nodejs/node-v6.9.2-linux-x64.tar.xz -C npm --strip-components=1
      ;;
  esac
  #setup npm path
  export PATH=${BUILD_DIR}/npm/bin:$PATH
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

#NPM Cleanup
cleanup_npm(){
  cd ${BUILD_DIR}
  rm -rf npm
  rm -rf nodejs
}
