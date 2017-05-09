#!/bin/bash

# Copyright 2014 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# GoFmt apparently is changing @ head...

set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE}")/..
source "${KUBE_ROOT}/hack/lib/init.sh"

cd "${KUBE_ROOT}"

# Prefer bazel's gofmt.
gofmt="external/io_bazel_rules_go_toolchain/bin/gofmt"
if [[ ! -x "${gofmt}" ]]; then
  gofmt=$(which gofmt)
  kube::golang::verify_go_version
fi

find_files() {
  find . -not \( \
      \( \
        -wholename './output' \
        -o -wholename './_output' \
        -o -wholename './_gopath' \
        -o -wholename './release' \
        -o -wholename './target' \
        -o -wholename '*/third_party/*' \
        -o -wholename '*/vendor/*' \
        -o -wholename './staging/src/k8s.io/client-go/*vendor/*' \
        -o -wholename '*/bindata.go' \
      \) -prune \
    \) -name '*.go'
}

diff=$(find_files | xargs ${gofmt} -d -s 2>&1)
if [[ -n "${diff}" ]]; then
  echo "${diff}"
  exit 1
fi

# Find comments not prefaced with a space.  gofmt doesn't do this.
# Multiple grep statements is easier to extend debug, although not elegant.
bad_files=$(find_files | xargs grep '//[^ ]' | grep -v "://" | grep -v swagger | grep -v "\"")
bad_files_count="`echo \"${bad_files}\" | wc -l`"
if [[ ${bad_files_count} -ne "0" ]]; then
  echo "Comments in the above files need to be in idiomatic format i.e. '// hello kube'"
  echo "We found $bad_files_count files with unidiomatic comments" 
  echo "EXAMPLES: ${bad_files}" | head -5
  echo "Currently, this will not cause the build to fail, but over time we want to decrease this number"
  ### Don't exit 1 yet, because it would break the build. 
fi
