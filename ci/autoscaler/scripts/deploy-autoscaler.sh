#!/usr/bin/env bash
# shellcheck disable=SC2086,SC2034,SC2155
set -euo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
source "${script_dir}/vars.source.sh"
source "${script_dir}/common.sh"

deployment_manifest="${autoscaler_dir}/templates/app-autoscaler.yml"
bosh_deploy_opts="${BOSH_DEPLOY_OPTS:-""}"
BOSH_DEPLOY_VARS="${BOSH_DEPLOY_VARS:-""}"
bosh_upload_release_opts="${BOSH_UPLOAD_RELEASE_OPTS:-""}"
bosh_upload_stemcell_opts="${BOSH_UPLOAD_STEMCELL_OPTS:-""}"

ops_files=${OPS_FILES:-$(cat <<EOF
${autoscaler_dir}/operations/add-releases.yml
${autoscaler_dir}/operations/instance-identity-cert-from-cf.yml
${autoscaler_dir}/operations/add-postgres-variables.yml
${autoscaler_dir}/operations/connect_to_postgres_with_certs.yml
${autoscaler_dir}/operations/enable-nats-tls.yml
${autoscaler_dir}/operations/add-extra-plan.yml
${autoscaler_dir}/operations/set-release-version.yml
${autoscaler_dir}/operations/enable-metricsforwarder-via-syslog-agent.yml
${autoscaler_dir}/operations/enable-scheduler-logging.yml
EOF
)}


CURRENT_COMMIT_HASH=$(cd "${autoscaler_dir}"; git log -1 --pretty=format:"%H")
bosh_release_version=${RELEASE_VERSION:-${CURRENT_COMMIT_HASH}-${deployment_name}}

bosh_login "${BBL_STATE_PATH}"

function setup_autoscaler_uaac() {
	local uaac_authorities="cloud_controller.read,cloud_controller.admin,uaa.resource,routing.routes.write,routing.routes.read,routing.router_groups.read"
	local autoscaler_secret="autoscaler_client_secret"
	local uaa_client_secret=$(credhub get -n /bosh-autoscaler/cf/uaa_admin_client_secret --quiet)

	uaac target "https://uaa.${system_domain}" --skip-ssl-validation > /dev/null
	uaac token client get admin -s "${uaa_client_secret}" > /dev/null

	if uaac client get autoscaler_client_id >/dev/null; then
		step "updating autoscaler uaac client"
		uaac client update "autoscaler_client_id" --authorities "$uaac_authorities" > /dev/null
	else
		step "creating autoscaler uaac client"
		uaac client add "autoscaler_client_id" \
			--authorized_grant_types "client_credentials" \
			--authorities "$uaac_authorities" \
			--secret "$autoscaler_secret" > /dev/null
	fi
}

function get_postgres_external_port() {
	[[ -z "${PR_NUMBER}" ]] && echo "5432" || echo "${PR_NUMBER}"
}

function create_manifest() {
	local tmp_dir
	local perform_as_gh_action="${GITHUB_ACTIONS:-false}"

	if "${perform_as_gh_action}" != 'false'; then
		tmp_dir="${RUNNER_TEMP}"
	else
		tmp_dir="$(pwd)/dev_releases"
		mkdir -p "${tmp_dir}"
	fi

	tmp_manifest_file="$(mktemp "${tmp_dir}/${deployment_name}.bosh-manifest.yaml.XXX")"
	credhub interpolate -f "${autoscaler_dir}/ci/autoscaler/scripts/autoscaler-secrets.yml.tpl" > /tmp/autoscaler-secrets.yml

	add_variable "deployment_name" "${deployment_name}"
	add_variable "system_domain" "${system_domain}"
	add_variable "app_autoscaler_version" "${bosh_release_version}"
	add_variable "cf_client_id" "autoscaler_client_id"
	add_variable "cf_client_secret" "autoscaler_client_secret"
	add_variable "postgres_external_port" "$(get_postgres_external_port)"


	bosh_deploy_vars=""

	# always use the ops file that tags all deployment related VMs and disks
	OPS_FILES_TO_USE="${OPS_FILES_TO_USE} -o ${ci_dir}/operations/tag-vms-and-disks.yml"

	# add deployment name
	bosh -n -d "${deployment_name}" interpolate "${deployment_manifest}" ${OPS_FILES_TO_USE} \
		${bosh_deploy_opts} ${BOSH_DEPLOY_VARS} \
		--vars-file=/tmp/autoscaler-secrets.yml -v skip_ssl_validation=true > "${tmp_manifest_file}"

	if [[ -z "${debug}" || "${debug}" = "false" ]]; then
		# shellcheck disable=SC2064
		trap "rm ${tmp_manifest_file}" EXIT
	fi
}

add_variable() {
	local variable_name=$1
	local variable_value=$2
	BOSH_DEPLOY_VARS="${BOSH_DEPLOY_VARS} -v ${variable_name}=${variable_value}"
}

function check_ops_files() {
	step "Using Ops files: '${ops_files}'"

	OPS_FILES_TO_USE=""
	for OPS_FILE in ${ops_files}; do
		if [[ -f "${OPS_FILE}" ]]; then
			OPS_FILES_TO_USE="${OPS_FILES_TO_USE} -o ${OPS_FILE}"
		else
			echo "ERROR: could not find ops file ${OPS_FILE} in ${PWD}"
			exit 1
		fi
	done
}

function deploy() {
	create_manifest
	log "creating Bosh deployment '${deployment_name}' with version '${bosh_release_version}' in system domain '${system_domain}'"
	debug "tmp_manifest_file=${tmp_manifest_file}"
	step "Using Ops files: '${OPS_FILES_TO_USE}'"
	step "Deploy options: '${bosh_deploy_opts}'"
	bosh -n -d "${deployment_name}" deploy "${tmp_manifest_file}"
	postgres_ip="$(bosh curl "/deployments/${deployment_name}/vms" | jq '. | .[] | select(.job == "postgres") | .ips[0]' -r)"
	credhub set -n "/bosh-autoscaler/${deployment_name}/postgres_ip" -t value -v "${postgres_ip}"
}

function find_or_upload_stemcell() {
	local stemcell_os stemcell_version stemcell_name
	stemcell_os=$(yq eval '.stemcells[] | select(.alias == "default").os' "${deployment_manifest}")
	stemcell_version=$(yq eval '.stemcells[] | select(.alias == "default").version' "${deployment_manifest}")
	stemcell_name="bosh-google-kvm-${stemcell_os}-go_agent"

	if ! bosh stemcells | grep "${stemcell_name}" >/dev/null; then
		local URL="https://bosh.io/d/stemcells/${stemcell_name}"
		[[ "${stemcell_version}" != "latest" ]] && URL="${URL}?v=${stemcell_version}"
		wget "${URL}" -O stemcell.tgz
		bosh -n upload-stemcell $bosh_upload_stemcell_opts stemcell.tgz
	fi
}

function find_or_upload_release() {
	if ! bosh releases | grep -E "${bosh_release_version}[*]*\s" > /dev/null; then
		local release_desc_file="dev_releases/app-autoscaler/app-autoscaler-${bosh_release_version}.yml"
		if [[ ! -f "${release_desc_file}" ]]; then
			echo "Creating Release with bosh version ${bosh_release_version}"
			bosh create-release --force --version="${bosh_release_version}"
		else
			# shellcheck disable=SC2006
			echo -e "Release with bosh-version ${bosh_release_version} already locally present. Reusing it."\
				"\n\tIf this does not work, please consider executing `bosh reset-release`."
		fi

		echo "Uploading Release"
		bosh upload-release ${bosh_upload_release_opts} "${release_desc_file}"
	else
		echo "the app-autoscaler release is already uploaded with the commit ${bosh_release_version}"
		echo "Attempting redeploy..."
	fi
}

function pre_deploy() {
	case "${cpu_upper_threshold}" in
		"100") ;;
		"200") ops_files+=" ${autoscaler_dir}/operations/cpu_upper_threshold_200.yml" ;;
		"400") ops_files+=" ${autoscaler_dir}/operations/cpu_upper_threshold_400.yml" ;;
		*) echo "No Ops file for cpu_upper_threshold of ${cpu_upper_threshold}"; exit 1 ;;
	esac
}

log "Deploying autoscaler '${bosh_release_version}' with name '${deployment_name}'"
setup_autoscaler_uaac
pushd "${autoscaler_dir}" > /dev/null
pre_deploy
check_ops_files
find_or_upload_stemcell
find_or_upload_release
deploy
popd > /dev/null
