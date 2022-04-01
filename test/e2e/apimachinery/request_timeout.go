/*
Copyright 2020 The Kubernetes Authors.

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

package apimachinery

import (
	"io"
	"net/http"
	"strings"

	e2eutils "k8s.io/kubernetes/test/e2e/framework/utils"

	"github.com/onsi/ginkgo"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/test/e2e/framework"
)

const (
	invalidTimeoutMessageExpected = "invalid timeout specified in the request URL"
)

var _ = SIGDescribe("Server request timeout", func() {
	f := framework.NewDefaultFramework("request-timeout")

	ginkgo.It("should return HTTP status code 400 if the user specifies an invalid timeout in the request URL", func() {
		rt := getRoundTripper(f)
		req := newRequest(f, "invalid")

		response, err := rt.RoundTrip(req)
		e2eutils.ExpectNoError(err)
		defer response.Body.Close()

		if response.StatusCode != http.StatusBadRequest {
			e2eutils.Failf("expected HTTP status code: %d, but got: %d", http.StatusBadRequest, response.StatusCode)
		}

		messageGot := readBody(response)
		if !strings.Contains(messageGot, invalidTimeoutMessageExpected) {
			e2eutils.Failf("expected HTTP status message to contain: %s, but got: %s", invalidTimeoutMessageExpected, messageGot)
		}
	})

	ginkgo.It("the request should be served with a default timeout if the specified timeout in the request URL exceeds maximum allowed", func() {
		rt := getRoundTripper(f)
		// Choose a timeout that exceeds the default timeout (60s) enforced by the apiserver
		req := newRequest(f, "3m")

		response, err := rt.RoundTrip(req)
		e2eutils.ExpectNoError(err)
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			e2eutils.Failf("expected HTTP status code: %d, but got: %d", http.StatusOK, response.StatusCode)
		}
	})

	ginkgo.It("default timeout should be used if the specified timeout in the request URL is 0s", func() {
		rt := getRoundTripper(f)
		req := newRequest(f, "0s")

		response, err := rt.RoundTrip(req)
		e2eutils.ExpectNoError(err)
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			e2eutils.Failf("expected HTTP status code: %d, but got: %d", http.StatusOK, response.StatusCode)
		}
	})
})

func getRoundTripper(f *framework.Framework) http.RoundTripper {
	config := rest.CopyConfig(f.ClientConfig())

	// ensure we don't enforce any transport timeout from the client side.
	config.Timeout = 0

	roundTripper, err := rest.TransportFor(config)
	e2eutils.ExpectNoError(err)

	return roundTripper
}

func newRequest(f *framework.Framework, timeout string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, f.ClientSet.CoreV1().RESTClient().Get().
		Param("timeout", timeout).AbsPath("version").URL().String(), nil)
	e2eutils.ExpectNoError(err)

	return req
}

func readBody(response *http.Response) string {
	raw, err := io.ReadAll(response.Body)
	e2eutils.ExpectNoError(err)

	return string(raw)
}
