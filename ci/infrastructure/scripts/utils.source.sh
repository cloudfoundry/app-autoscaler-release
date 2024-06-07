
bosh_upload_stemcell_opts="${BOSH_UPLOAD_STEMCELL_OPTS:-""}"
function find_or_upload_stemcell_from(){
  deployment_manifest=$1
  # Determine if we need to upload a stemcell at this point.
  stemcell_os=$(yq eval '.stemcells[] | select(.alias == "default").os' ${deployment_manifest})
  stemcell_version=$(yq eval '.stemcells[] | select(.alias == "default").version' ${deployment_manifest})
  stemcell_name="bosh-google-kvm-${stemcell_os}-go_agent"

  if ! bosh stemcells | grep "${stemcell_name}" >/dev/null; then
    URL="https://bosh.io/d/stemcells/${stemcell_name}"
    if [ "${stemcell_version}" != "latest" ]; then
	    URL="${URL}?v=${stemcell_version}"
    fi
    wget "$URL" -O stemcell.tgz
    bosh -n upload-stemcell $bosh_upload_stemcell_opts stemcell.tgz
  fi
}

function load_bbl_vars() {
  if [ -z "${bbl_state_path}" ]; then
    echo "ERROR: bbl_state_path is not set"
    exit 1
  fi

  director_store="${bbl_state_path}/vars/director-vars-store.yml"
  log "director_store = '${director_store}'"

  pushd "${bbl_state_path}" > /dev/null
    eval "$(bbl print-en  v)"
  popd > /dev/null
}

function validate_ops_files() {
  for ops_file in ${ops_files}; do
      echo "ERROR: could not find ops file ${OPS_FILE} in ${PWD}"
      exit 1
  done
}
