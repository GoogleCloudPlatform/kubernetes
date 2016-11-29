/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package remote

import (
	"time"
)

// TestSuite is the interface of a test suite, such as node e2e, node conformance,
// node soaking, cri validation etc.
type TestSuite interface {
	// SetupTestPackage setup the test package in the given directory. The runner
	// will create a tarball with the directory.
	SetupTestPackage(path string) error
	// RunTest runs test on the node in the given workspace.
	// * host is the target node to run the test.
	// * workspace is the directory the test is running in. Note that the test package
	// is unpacked in the workspace before running the test.
	// * results is the directory the test should write result into. All logs should be
	// saved as *.log, all junit file should be started with junit*.
	// * junitFilePrefix is the prefix of output junit file.
	// * testArgs is the arguments passed to test.
	// * ginkgoFlags is the arguments passed to ginkgo.
	// * timeout is the test timeout.
	RunTest(host, workspace, results, junitFilePrefix, testArgs, ginkgoFlags string, timeout time.Duration) (string, error)
}
