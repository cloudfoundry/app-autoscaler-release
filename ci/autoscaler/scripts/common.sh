#! /usr/bin/env bash

# This file is intended to be loaded via the `source`-command.

function step(){
	echo "# $1"
}

function retry(){
	max_retries=$1
	shift
	retries=0
	command="$*"
	until [ "${retries}" -eq "${max_retries}" ] || $command; do
		((retries=retries+1))
		echo " - retrying command '${command}' attempt: ${retries}"
	done
	[ "${retries}" -lt "${max_retries}" ] || { echo "ERROR: Command '$*' failed after ${max_retries} attempts"; return 1; }
}

function concourse_login() {
	local -r concourse_aas_release_target="${1}"
	if ! fly targets | rg --only-matching --regexp="^${concourse_aas_release_target}[[:space:]]" > /dev/null
	then
		echo "There is no concourse-target for precicely \"${concourse_aas_release_target}\"." \
				 'Login required!'
		fly --target="${concourse_aas_release_target}" login \
				--team-name='app-autoscaler'\
				--concourse-url='https://concourse.app-runtime-interfaces.ci.cloudfoundry.org'
	fi
}

function bosh_login() {
	step "bosh login"
	local -r bbl_state_path="${1}"
	if [[ ! -d "${bbl_state_path}" ]]
	then
		echo "⛔ FAILED: Did not find bbl-state folder at ${bbl_state_path}"
		echo 'Make sure you have checked out the app-autoscaler-env-bbl-state repository next to the app-autoscaler-release repository to run this target or indicate its location via BBL_STATE_PATH'
		exit 1;
	fi

	pushd "${bbl_state_path}" > /dev/null
		eval "$(bbl print-env)"
	popd > /dev/null
}

function cf_login(){
	step 'login to cf'
	cf api "https://api.${system_domain}" --skip-ssl-validation
	cf_admin_password="$(credhub get --quiet --name='/bosh-autoscaler/cf/cf_admin_password')"
	cf auth admin "$cf_admin_password"
}

function cleanup_acceptance_run(){
	step "cleaning up from acceptance tests"
	pushd "${ci_dir}/../src/acceptance" > /dev/null
		retry 5 ./cleanup.sh
	popd > /dev/null
}

function cleanup_service_broker(){
	step "deleting service broker for deployment '${deployment_name}'"
	SERVICE_BROKER_EXISTS=$(cf service-brokers | grep -c "${service_broker_name}.${system_domain}" || true)
	if [[ $SERVICE_BROKER_EXISTS == 1 ]]; then
		echo "- Service Broker exists, deleting broker '${deployment_name}'"
		retry 3 cf delete-service-broker "${deployment_name}" -f
	fi
}

function cleanup_bosh_deployment(){
	step "deleting bosh deployment '${deployment_name}'"
	retry 3 bosh delete-deployment -d "${deployment_name}" -n
}

function delete_releases(){
	step "deleting releases"
	if [ -n "${deployment_name}" ]
	then
		for release in $(bosh releases | grep -E "${deployment_name}\s+"  | awk '{print $2}')
		do
			 echo "- Deleting bosh release '${release}'"
			 bosh delete-release -n "app-autoscaler/${release}" &
		done
		wait
	fi
}

function cleanup_bosh(){
	step "cleaning up bosh"
	retry 3 bosh clean-up --all -n
}

function cleanup_credhub(){
	step "cleaning up credhub: '/bosh-autoscaler/${deployment_name}/*'"
	retry 3 credhub delete --path="/bosh-autoscaler/${deployment_name}"
}

function cleanup_apps(){
	step "cleaning up apps"
	local mtar_app
	local space_guid

	cf_target "${autoscaler_org}" "${autoscaler_space}"

	space_guid="$(cf space --guid "${autoscaler_space}")"
	mtar_app="$(curl --header "Authorization: $(cf oauth-token)" "deploy-service.${system_domain}/api/v2/spaces/${space_guid}/mtas"  | jq ". | .[] | .metadata | .id" -r)"

	if [ -n "${mtar_app}" ]; then
		set +e
		cf undeploy "${mtar_app}" -f --delete-service-brokers --delete-service-keys --delete-services --do-not-fail-on-missing-permissions
		set -e
	else
		 echo "No app to undeploy"
	fi

	if cf spaces | grep --quiet --regexp="^${AUTOSCALER_SPACE}$"; then
		cf delete-space -f "${AUTOSCALER_SPACE}"
	fi

	if cf orgs | grep --quiet --regexp="^${AUTOSCALER_ORG}$"
	then
		cf delete-org -f "${AUTOSCALER_ORG}"
	fi
}


function unset_vars() {
	unset PR_NUMBER
	unset DEPLOYMENT_NAME
	unset SYSTEM_DOMAIN
	unset BBL_STATE_PATH
	unset AUTOSCALER_DIR
	unset CI_DIR
	unset SERVICE_NAME
	unset SERVICE_BROKER_NAME
	unset NAME_PREFIX
	unset GINKGO_OPTS
}

function find_or_create_org(){
	step "finding or creating org"
	local org_name="$1"
	if ! cf orgs | grep --quiet --regexp="^${org_name}$"
	then
		cf create-org "${org_name}"
	fi
	echo "targeting org ${org_name}"
	cf target -o "${org_name}"
}

function find_or_create_space(){
	step "finding or creating space"
	local space_name="$1"
	if ! cf spaces | grep --quiet --regexp="^${space_name}$"
	then
		cf create-space "${space_name}"
	fi
	echo "targeting space ${space_name}"
	cf target -s "${space_name}"
}

function cf_target(){
	local org_name="$1"
	local space_name="$2"

	find_or_create_org "${org_name}"
	find_or_create_space "${space_name}"
}

# 🚀 Initialise and start the PostgreSQL DBMS-server, create a user 'test-pg' and a database
# 'test-pg'.
#
# ⚠️ This subprogram is meant to be only run in container derived from the image
# 'app-autoscaler-release-tools'. It assumes a working devbox-installation and devbox being
# responsible for PostgreSQL.
#
# 🚸 Add `trap 'devbox services stop postgresql' EXIT` to your script to shut it down when the
# process finishes.
function ci_prepare_postgres_db() {
	# devbox makes sure that the environment-variables PGHOST and PGDATA are set appropriately.
	set -x # 🚧 To-do: Debug-code
	echo "pwd: $(pwd)" # 🚧 To-do: Debug-code
	echo "ls -lah .: $(ls -lah .)" # 🚧 To-do: Debug-code
	initdb
	devbox services --config='/code' up postgresql --background # 🚧 To-do: Can we avoid the `--config`-parameter?
	#devbox services  up postgresql --background	# pg_ctl will not work as it is not aware of where to
																							# create the socket.
	createuser tests-pg
	createdb tests-pg # Needed to be done like this, because 'tests-pg' does not have the required#
										# priviledges.
	set +x # 🚧 To-do: Debug-code
}
