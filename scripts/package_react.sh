#!/bin/bash

# Make bash more strict
set -euo pipefail

function die {
    echo "$@" 1>&2
    exit 1
}

# It's good practice to not assume the running CWD, this helps with that
script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd -P)

# cd to where this script lives
cd "${script_dir}"

# Find the root of the project and cd there
cd "$(git rev-parse --show-toplevel)"

cd www
rm -rf dist
yarn package
