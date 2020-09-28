#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
goarch="$(go env GOARCH)"
goos="$(go env GOOS)" 
version="$1"
base_url="https://github.com/kubernetes-sigs/kubebuilder/releases/download"
url="${base_url}/v${version}/kubebuilder_${version}_${goos}_${goarch}.tar.gz"
tmp_dir="$(mktemp -d)"
(cd ${tmp_dir} && curl -sL ${url} | tar -xz)
cp ${tmp_dir}/kubebuilder_${version}_${goos}_${goarch}/bin/* $(dirname ${script_dir})/bin
rm -fr ${tmp_dir}
