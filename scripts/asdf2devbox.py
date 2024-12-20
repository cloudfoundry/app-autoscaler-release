#!/usr/bin/env python3

import json
import os
import subprocess

# Determine the script directory
script_dir = os.path.dirname(os.path.realpath(__file__))

# Function to determine the currently installed version of the package and return it
def get_installed_version(package):
    with open(os.path.join(script_dir, '..', 'devbox.json'), 'r') as f:
        data = json.load(f)
        try:
            return data['packages'][package]
        except KeyError:
            return None

# Read the .tool-versions file and process each line
if __name__ == "__main__":
    with open(os.path.join(script_dir, '..', '.tool-versions'), 'r') as f:
        for line in f:
            program, version = line.strip().split(' ')

            # Mapping of ASDF to Nix program names
            program_mapping = {
                "bosh": "bosh-cli",
                "cf": "cloudfoundry-cli",
                "concourse": "fly",
                "credhub": "credhub-cli",
                "gcloud": "google-cloud-sdk",
                "golang": "go",
                "java": "temurin-bin-21",
                "make": "gnumake",
                "yq": "yq-go"
            }
            program = program_mapping.get(program, program)

            # Check if the package is already installed in the desired version
            installed_version = get_installed_version(program)
            if installed_version != version:
                # Try to add the package with the specified version
                try:
                    subprocess.run(['devbox', 'add', f"{program}@{version}"], check=True)
                except subprocess.CalledProcessError:
                    # Fallback to latest in case the exact version is not available and there's no previous version installed
                    if installed_version is None:
                        print(f"Could not find {program}@{version}, using latest instead")
                        subprocess.run(['devbox', 'add', f"{program}@latest"], check=True)
                    else:
                        # Readd the previously installed version
                        print(f"Could not find {program}@{version}, readding {program}@{installed_version}")
                        subprocess.run(['devbox', 'add', f"{program}@{installed_version}"], check=True)

                    subprocess.run(['devbox', 'info', program], check=True)
            else:
                print(f"{program}@{version} is already installed")
