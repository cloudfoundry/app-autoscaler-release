#! /usr/bin/env bash

set -eu -o pipefail
# set -x

# For following definitions, see:
# <https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_(Select_Graphic_Rendition)_parameters>
declare -r TERM_STYLE_BLUE='\e[38;2;0;0;255m'
declare -r TERM_STYLE_YELLOW='\e[38;2;255;255;0m' # TODO: Check value!
declare -r TERM_STYLE_RESET='\e[0m'

# This function gets a list of names of shell-parameters and prints
# them as well as its values
function print_parameters () {
  local -a -r param_list=("$@")
  for param_name in "${param_list[@]}"
  do
    echo -ne "${TERM_STYLE_BLUE}"
    echo "${param_name} = ${!param_name}"
    echo -ne "${TERM_STYLE_RESET}"
  done

  return 0
}

# Prints out a nice description for a parameter provided as bash-tuple
# ('name' 'description').
function describe_parameter () {
  local -a -r param=${1}
  local -r name=${param[0]}
  local -r description=${param[1]}

  echo -ne "${TERM_STYLE_YELLOW}"
  echo "${name}: ${description}"
  echo -ne "${TERM_STYLE_RESET}"

  return 0
}

function describe_all_parameters () {
  local -a -r param_desc_list=("$@")

  for param_desc in "${param_desc_list[@]}"
  do
    describe_parameter "${param_desc}"
  done

  return 0
}

function print_help_for_parameters () {
  local -a -r param_desc_list=("$@")
  describe_all_parameters "${param_desc_list[@]}"

  local -a param_list
  for param_desc in "${param_desc_list[@]}"
  do
    local -a pd=("${param_desc[@]}")
    local param=${pd[0]}
    echo "!!!pd=${pd}"
    echo "!!!param = ${param}"
    param_list+="${param}"
  done

  print_parameters "${param_list[@]}"

  return 0
}

# When called as a script
if [ "${0}" = "${BASH_SOURCE[0]}" ]
then
  # TODO: Try associative array: <https://stackoverflow.com/a/3113285>
  declare -a -r input=("$@")
  # https://stackoverflow.com/a/3816747
  print_help_for_parameters "${input[@]}"
fi