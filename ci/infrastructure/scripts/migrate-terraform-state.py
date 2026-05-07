#!/usr/bin/env python3
"""Migrate Terraform state from v3 (TF 0.11) to v4 (TF 1.x).

BBL v8 used Terraform 0.11 which produces state format version 3.
BBL v9 uses Terraform 1.4+ which requires state format version 4.
Terraform's official upgrade path requires going through 0.12 and 0.13,
but we can convert the JSON directly since the schema is well-defined.
"""
import json
import sys


def convert_output_type(tf11_type, value):
    """Convert TF 0.11 output type to TF 1.x type expression."""
    if tf11_type == "list":
        count = len(value) if isinstance(value, list) else 0
        return ["tuple", ["string"] * count]
    if tf11_type == "map":
        if isinstance(value, dict):
            return ["object", {k: "string" for k in value}]
        return ["object", {}]
    return "string"


def migrate_v3_to_v4(state_v3):
    provider_map = {
        "provider.google": 'provider["registry.terraform.io/hashicorp/google"]',
    }

    state_v4 = {
        "version": 4,
        "terraform_version": "1.4.6",
        "serial": state_v3["serial"] + 1,
        "lineage": state_v3["lineage"],
        "outputs": {},
        "resources": [],
    }

    module = state_v3["modules"][0]

    for out_name, out_val in module.get("outputs", {}).items():
        state_v4["outputs"][out_name] = {
            "value": out_val["value"],
            "type": convert_output_type(out_val.get("type", "string"), out_val["value"]),
        }
        if out_val.get("sensitive"):
            state_v4["outputs"][out_name]["sensitive"] = True

    for res_key, res_data in module.get("resources", {}).items():
        parts = res_key.split(".")
        if parts[0] == "data":
            mode = "data"
            res_type = parts[1]
            res_name = ".".join(parts[2:])
        else:
            mode = "managed"
            res_type = parts[0]
            res_name = ".".join(parts[1:])

        provider_v3 = res_data.get("provider", "provider.google")
        provider_v4 = provider_map.get(provider_v3, provider_v3)

        primary = res_data.get("primary", {})
        attributes = primary.get("attributes", {})
        meta = primary.get("meta", {})
        schema_version = int(meta.get("schema_version", 0)) if isinstance(meta, dict) else 0

        instance = {
            "schema_version": schema_version,
            "attributes_flat": attributes,
        }

        deps = res_data.get("depends_on", [])
        if deps:
            instance["dependencies"] = deps

        resource = {
            "mode": mode,
            "type": res_type,
            "name": res_name,
            "provider": provider_v4,
            "instances": [instance],
        }

        state_v4["resources"].append(resource)

    return state_v4


def main():
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} <terraform.tfstate>", file=sys.stderr)
        sys.exit(1)

    state_file = sys.argv[1]

    with open(state_file) as f:
        state = json.load(f)

    if state.get("version") != 3:
        print(f"State is already version {state.get('version')}, skipping migration.")
        sys.exit(0)

    print(f"Migrating state from TF {state['terraform_version']} (v3) to v4 format...")
    state_v4 = migrate_v3_to_v4(state)
    print(f"Converted {len(state_v4['resources'])} resources, {len(state_v4['outputs'])} outputs.")

    with open(state_file, "w") as f:
        json.dump(state_v4, f, indent=2)

    print(f"State file {state_file} migrated successfully.")


if __name__ == "__main__":
    main()
