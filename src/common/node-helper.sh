#!/bin/bash

#NPM setup
setup_npm(){
  cd ${BUILD_DIR}
  # Make sure we can see uname
  export PATH=$PATH:/bin:/usr/bin
  #unpack NodeJs - we support Mac OS 64bit and Linux 64bit
  mkdir npm
  case "$OSTYPE" in
    darwin*)
      tar xvf nodejs/node-v8.11.4-darwin-x64.tar.gz -C npm --strip-components=1
      ;;
    linux*)
      tar xvf nodejs/node-v8.11.4-linux-x64.tar.xz -C npm --strip-components=1
      ;;
  esac
  #setup npm path
  export PATH=${BUILD_DIR}/npm/bin:$PATH
}

#NPM Cleanup
cleanup_npm(){
  cd ${BUILD_DIR}
  rm -rf npm
  rm -rf nodejs
}

