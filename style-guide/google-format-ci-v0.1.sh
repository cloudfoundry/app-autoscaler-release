#! /usr/bin/env bash
set -eu -o pipefail
export PATH="${HOME}/go/bin:${PATH}"

###################################################################################################
 # This script downloads google formatter and displays formatting issues on Github actions
###################################################################################################
GOOGLE_JAR_VERSION=${GOOGLE_JAR_VERSION:-"1.22.0"}
GOOGLE_JAR_NAME=${GOOGLE_JAR_NAME:-"google-java-format-${GOOGLE_JAR_VERSION}-all-deps.jar"}
readonly PATH GOOGLE_JAR_VERSION GOOGLE_JAR_NAME

! [ -e "${GOOGLE_JAR_NAME}" ] && \
	curl --fail --location --remote-name --remote-header-name \
			 "https://github.com/google/google-java-format/releases/download/v${GOOGLE_JAR_VERSION}/${GOOGLE_JAR_NAME}"
# shellcheck disable=SC2046
files_to_be_formatted="$(java \
							-jar "${GOOGLE_JAR_NAME}" --dry-run --skip-javadoc-formatting $(find src/scheduler -name '*.java'))"

if [ -n "${files_to_be_formatted}" ]
then
	# # This output of proposed changes seems to be unusable. Please overwrite instead.
	# # shellcheck disable=SC2046
	# proprosed_changes="$(java -jar "${GOOGLE_JAR_NAME}" --skip-javadoc-formatting $(find src/scheduler -name '*.java'))"
	cat <<-EOF
			Formatter results …
			Files, that require reformatting:
			${files_to_be_formatted}
	EOF
	exit 1
fi
