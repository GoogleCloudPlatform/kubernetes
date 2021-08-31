#!/usr/bin/env bash

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

# Tear down a Kubernetes cluster.

set -o errexit
set -o nounset
set -o pipefail

# TODO(b/197113765): Remove this script and use binary directly.
if [[ -e "$(dirname "${BASH_SOURCE[0]}")/../../hack/lib/util.sh" ]]; then
  # When kubectl.sh is used directly from the repo, it's under gke/cluster.
  KUBE_ROOT=$(dirname "${BASH_SOURCE[0]}")/../..
else
  # When kubectl.sh is used from unpacked tarball, it's under cluster.
  KUBE_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
fi

if [ -f "$(dirname "${BASH_SOURCE[0]}")/env.sh" ]; then
    source "$(dirname "${BASH_SOURCE[0]}")/env.sh"
fi

source "$(dirname "${BASH_SOURCE[0]}")/kube-util.sh"

echo "Bringing down cluster using provider: $KUBERNETES_PROVIDER"

echo "... calling verify-prereqs" >&2
verify-prereqs
echo "... calling verify-kube-binaries" >&2
verify-kube-binaries
echo "... calling kube-down" >&2
kube-down

echo "Done"
