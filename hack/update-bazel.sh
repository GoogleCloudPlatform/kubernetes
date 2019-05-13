#!/usr/bin/env bash
# Copyright 2016 The Kubernetes Authors.
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

set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
# shellcheck source=../hack/lib/init.sh
source "${KUBE_ROOT}/hack/lib/init.sh"

# Ensure that we find the binaries we build before anything else.
export GOBIN="${KUBE_OUTPUT_BINPATH}"
PATH="${GOBIN}:${PATH}"

# Install tools we need, but only from vendor/...
pushd "${KUBE_ROOT}/vendor"
  go install ./github.com/bazelbuild/bazel-gazelle/cmd/gazelle
  go install ./github.com/bazelbuild/buildtools/buildozer
  go install ./k8s.io/repo-infra/kazel
popd

# Find all of the staging repos.
while IFS='' read -r repo; do staging_repos+=("${repo}"); done <\
  <(cd "${KUBE_ROOT}/staging/src" && find k8s.io -mindepth 1 -maxdepth 1 -type d | LANG=C sort)

# Save the staging repos into a Starlark list that can be used by Bazel rules.
(
  cat "${KUBE_ROOT}/hack/boilerplate/boilerplate.generatebzl.txt"
  echo "# This file is autogenerated by hack/update-bazel.sh."
  # avoid getting caught by the boilerplate checker
  rev <<<".TIDE TON OD #"
  echo
  echo "staging_repos = ["
  for repo in "${staging_repos[@]}"; do
    echo "    \"${repo}\","
  done
  echo "]"
) >"${KUBE_ROOT}/staging/repos_generated.bzl"

# Ensure there's a BUILD file at vendor/ so the all-srcs rule in the root
# BUILD.bazel file is isolated from changes under vendor/.
touch "${KUBE_ROOT}/vendor/BUILD"
# Ensure there's a BUILD file at the root of each staging repo. This prevents
# the package-srcs glob in vendor/BUILD from following the symlinks
# from vendor/k8s.io into staging. Following these symlinks through
# vendor results in files being excluded from the kubernetes-src tarball
# generated by Bazel.
for repo in "${staging_repos[@]}"; do
  touch "${KUBE_ROOT}/staging/src/${repo}/BUILD"
done

# Run gazelle to update Go rules in BUILD files.
gazelle fix \
    -external=vendored \
    -mode=fix \
    -repo_root "${KUBE_ROOT}" \
    "${KUBE_ROOT}"

# Run kazel to update the pkg-srcs and all-srcs rules as well as handle
# Kubernetes code generators.
kazel

# make targets in vendor manual
# buildozer exits 3 when no changes are made ¯\_(ツ)_/¯
# https://github.com/bazelbuild/buildtools/tree/master/buildozer#error-code
buildozer -quiet 'add tags manual' '//vendor/...:%go_binary' '//vendor/...:%go_test' && ret=$? || ret=$?
if [[ $ret != 0 && $ret != 3 ]]; then
  exit 1
fi
