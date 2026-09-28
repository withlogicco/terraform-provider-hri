#!/usr/bin/env bash
set -euo pipefail
: "\${TF_STATE_JSON:?Set TF_STATE_JSON to a terraform show -json file}"
: "\${HRI_INVENTORY_JSON:?Set HRI_INVENTORY_JSON to terraform output -json containing hri_servers}"
python3 - "$TF_STATE_JSON" "$HRI_INVENTORY_JSON" <<'PY'
import json, sys
state = json.load(open(sys.argv[1]))
inventory = json.load(open(sys.argv[2]))
tracked = set()
def walk(module):
    for resource in module.get("resources", []):
        if resource.get("type") == "hri_server":
            tracked.add(str(resource.get("values", {}).get("server_number")))
    for child in module.get("child_modules", []): walk(child)
walk(state.get("values", {}).get("root_module", {}))
servers = {str(n) for n in inventory.get("server_numbers", {}).get("value", [])}
missing, untracked = sorted(tracked - servers), sorted(servers - tracked)
if missing or untracked:
    if missing: print("Tracked servers absent from Robot: " + ", ".join(missing))
    if untracked: print("Untracked Robot servers: " + ", ".join(untracked))
    sys.exit(1)
print("Terraform state matches the Robot inventory.")
PY
