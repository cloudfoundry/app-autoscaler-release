#! /usr/bin/env bash

set -eu -o pipefail

# This function gets a list of names of shell-parameters and prints
# them as well as its values
function print_parameters () {
  local -a -r param_list=("$@")
  for param_name in "${param_list[@]}"
  do
    echo "${param_name} = ${!param_name}"
  done
}

# When called as a script
if [ "${0}" = "${BASH_SOURCE[0]}" ]
then
  # https://stackoverflow.com/a/3816747
  print_parameters "$@"
fi