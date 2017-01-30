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

package rest

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	gruntime "runtime"
	"strings"
	"time"

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/version"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/flowcontrol"
)

const (
	DefaultQPS   float32 = 5.0
	DefaultBurst int     = 10
)

// Config holds the common attributes that can be passed to a Kubernetes client on
// initialization.
type Config struct {
	// Host must be a host string, a host:port pair, or a URL to the base of the apiserver.
	// If a URL is given then the (optional) Path of that URL represents a prefix that must
	// be appended to all request URIs used to access the apiserver. This allows a frontend
	// proxy to easily relocate all of the apiserver endpoints.
	Host string

	// APIPath is a sub-path that points to an API root.
	APIPath string
	// Prefix is the sub path of the server. If not specified, the client will set
	// a default value.  Use "/" to indicate the server root should be used
	Prefix string

	// ContentConfig contains settings that affect how objects are transformed when
	// sent to the server.
	ContentConfig

	// Server requires Basic authentication
	Username string
	Password string

	// Server requires Bearer authentication. This client will not attempt to use
	// refresh tokens for an OAuth2 flow.
	// TODO: demonstrate an OAuth2 compatible client.
	BearerToken string

	// Impersonate is the configuration that RESTClient will use for impersonation.
	Impersonate ImpersonationConfig

	// Server requires plugin-specified authentication.
	AuthProvider *clientcmdapi.AuthProviderConfig

	// Callback to persist config for AuthProvider.
	AuthConfigPersister AuthProviderConfigPersister

	// TLSClientConfig contains settings to enable transport layer security
	TLSClientConfig

	// Server should be accessed without verifying the TLS
	// certificate. For testing only.
	Insecure bool

	// UserAgent is an optional field that specifies the caller of this request.
	UserAgent string

	// Transport may be used for custom HTTP behavior. This attribute may not
	// be specified with the TLS client certificate options. Use WrapTransport
	// for most client level operations.
	Transport http.RoundTripper
	// WrapTransport will be invoked for custom HTTP behavior after the underlying
	// transport is initialized (either the transport created from TLSClientConfig,
	// Transport, or http.DefaultTransport). The config may layer other RoundTrippers
	// on top of the returned RoundTripper.
	WrapTransport func(rt http.RoundTripper) http.RoundTripper

	// QPS indicates the maximum QPS to the master from this client.
	// If it's zero, the created RESTClient will use DefaultQPS: 5
	QPS float32

	// Maximum burst for throttle.
	// If it's zero, the created RESTClient will use DefaultBurst: 10.
	Burst int

	// Rate limiter for limiting connections to the master from this client. If present overwrites QPS/Burst
	RateLimiter flowcontrol.RateLimiter

	// The maximum length of time to wait before giving up on a server request. A value of zero means no timeout.
	Timeout time.Duration

	// AlternateHosts is additional hosts that can be used for purpose of fault tolerance
	AlternateHosts []string

	// Version forces a specific version to be used (if registered)
	// Do we need this?
	// Version string
}

// AllHosts combines Host and AlternateHosts into single slice
func (c *Config) Hosts() []string {
	var hosts []string
	if len(c.Host) > 0 {
		hosts = make([]string, 0, 1+len(c.AlternateHosts))
		hosts = append(hosts, c.Host)
	} else {
		hosts = make([]string, 0, len(c.AlternateHosts))
	}
	for _, host := range c.AlternateHosts {
		if host != c.Host {
			hosts = append(hosts, host)
		}
	}
	return hosts
}

// ImpersonationConfig has all the available impersonation options
type ImpersonationConfig struct {
	// UserName is the username to impersonate on each request.
	UserName string
	// Groups are the groups to impersonate on each request.
	Groups []string
	// Extra is a free-form field which can be used to link some authentication information
	// to authorization information.  This field allows you to impersonate it.
	Extra map[string][]string
}

// TLSClientConfig contains settings to enable transport layer security
type TLSClientConfig struct {
	// Server requires TLS client certificate authentication
	CertFile string
	// Server requires TLS client certificate authentication
	KeyFile string
	// Trusted root certificates for server
	CAFile string

	// CertData holds PEM-encoded bytes (typically read from a client certificate file).
	// CertData takes precedence over CertFile
	CertData []byte
	// KeyData holds PEM-encoded bytes (typically read from a client certificate key file).
	// KeyData takes precedence over KeyFile
	KeyData []byte
	// CAData holds PEM-encoded bytes (typically read from a root certificates bundle).
	// CAData takes precedence over CAFile
	CAData []byte
}

type ContentConfig struct {
	// AcceptContentTypes specifies the types the client will accept and is optional.
	// If not set, ContentType will be used to define the Accept header
	AcceptContentTypes string
	// ContentType specifies the wire format used to communicate with the server.
	// This value will be set as the Accept header on requests made to the server, and
	// as the default content type on any object sent to the server. If not set,
	// "application/json" is used.
	ContentType string
	// GroupVersion is the API version to talk to. Must be provided when initializing
	// a RESTClient directly. When initializing a Client, will be set with the default
	// code version.
	GroupVersion *schema.GroupVersion
	// NegotiatedSerializer is used for obtaining encoders and decoders for multiple
	// supported media types.
	NegotiatedSerializer runtime.NegotiatedSerializer
}

// RESTClientFor returns a RESTClient that satisfies the requested attributes on a client Config
// object. Note that a RESTClient may require fields that are optional when initializing a Client.
// A RESTClient created by this method is generic - it expects to operate on an API that follows
// the Kubernetes conventions, but may not be the Kubernetes API.
func RESTClientFor(config *Config) (*RESTClient, error) {
	if config.GroupVersion == nil {
		return nil, fmt.Errorf("GroupVersion is required when initializing a RESTClient")
	}
	if config.NegotiatedSerializer == nil {
		return nil, fmt.Errorf("NegotiatedSerializer is required when initializing a RESTClient")
	}
	qps := config.QPS
	if config.QPS == 0.0 {
		qps = DefaultQPS
	}
	burst := config.Burst
	if config.Burst == 0 {
		burst = DefaultBurst
	}

	hosts, versionedAPIPath, err := defaultServerUrlsFor(config)
	if err != nil {
		return nil, err
	}

	transport, err := TransportFor(config)
	if err != nil {
		return nil, err
	}

	var httpClient *http.Client
	if transport != http.DefaultTransport {
		httpClient = &http.Client{Transport: transport}
		if config.Timeout > 0 {
			httpClient.Timeout = config.Timeout
		}
	}
	return NewRESTClient(hosts, versionedAPIPath, config.ContentConfig, qps, burst, config.RateLimiter, httpClient)
}

// UnversionedRESTClientFor is the same as RESTClientFor, except that it allows
// the config.Version to be empty.
func UnversionedRESTClientFor(config *Config) (*RESTClient, error) {
	if config.NegotiatedSerializer == nil {
		return nil, fmt.Errorf("NeogitatedSerializer is required when initializing a RESTClient")
	}

	hosts, versionedAPIPath, err := defaultServerUrlsFor(config)
	if err != nil {
		return nil, err
	}

	transport, err := TransportFor(config)
	if err != nil {
		return nil, err
	}

	var httpClient *http.Client
	if transport != http.DefaultTransport {
		httpClient = &http.Client{Transport: transport}
		if config.Timeout > 0 {
			httpClient.Timeout = config.Timeout
		}
	}

	versionConfig := config.ContentConfig
	if versionConfig.GroupVersion == nil {
		v := metav1.SchemeGroupVersion
		versionConfig.GroupVersion = &v
	}

	return NewRESTClient(hosts, versionedAPIPath, versionConfig, config.QPS, config.Burst, config.RateLimiter, httpClient)
}

// SetKubernetesDefaults sets default values on the provided client config for accessing the
// Kubernetes API or returns an error if any of the defaults are impossible or invalid.
func SetKubernetesDefaults(config *Config) error {
	if len(config.UserAgent) == 0 {
		config.UserAgent = DefaultKubernetesUserAgent()
	}
	return nil
}

// DefaultKubernetesUserAgent returns the default user agent that clients can use.
func DefaultKubernetesUserAgent() string {
	commit := version.Get().GitCommit
	if len(commit) > 7 {
		commit = commit[:7]
	}
	if len(commit) == 0 {
		commit = "unknown"
	}
	version := version.Get().GitVersion
	seg := strings.SplitN(version, "-", 2)
	version = seg[0]
	return fmt.Sprintf("%s/%s (%s/%s) kubernetes/%s", path.Base(os.Args[0]), version, gruntime.GOOS, gruntime.GOARCH, commit)
}

// InClusterConfig returns a config object which uses the service account
// kubernetes gives to pods. It's intended for clients that expect to be
// running inside a pod running on kubernetes. It will return an error if
// called from a process not running in a kubernetes environment.
func InClusterConfig() (*Config, error) {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		return nil, fmt.Errorf("unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")
	}

	token, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/" + api.ServiceAccountTokenKey)
	if err != nil {
		return nil, err
	}
	tlsClientConfig := TLSClientConfig{}
	rootCAFile := "/var/run/secrets/kubernetes.io/serviceaccount/" + api.ServiceAccountRootCAKey
	if _, err := certutil.NewPool(rootCAFile); err != nil {
		glog.Errorf("Expected to load root CA config from %s, but got err: %v", rootCAFile, err)
	} else {
		tlsClientConfig.CAFile = rootCAFile
	}

	return &Config{
		// TODO: switch to using cluster DNS.
		Host:            "https://" + net.JoinHostPort(host, port),
		BearerToken:     string(token),
		TLSClientConfig: tlsClientConfig,
	}, nil
}

// IsConfigTransportTLS returns true if and only if the provided
// config will result in a protected connection to the server when it
// is passed to restclient.RESTClientFor().  Use to determine when to
// send credentials over the wire.
//
// Note: the Insecure flag is ignored when testing for this value, so MITM attacks are
// still possible.
func IsConfigTransportTLS(config Config) (bool, error) {
	hosts, _, err := defaultServerUrlsFor(&config)
	allHttps := 0
	// TODO Propagate errors from defaultServerUrlsFor
	if err != nil {
		return false, nil
	}
	for _, host := range hosts {
		if host.Scheme == "https" {
			allHttps++
		}
	}
	if allHttps == 0 {
		return false, nil
	} else if allHttps == len(hosts) {
		return true, nil
	}
	return false, fmt.Errorf("https and http can't be mixed, hosts: %v", hosts)
}

// LoadTLSFiles copies the data from the CertFile, KeyFile, and CAFile fields into the CertData,
// KeyData, and CAFile fields, or returns an error. If no error is returned, all three fields are
// either populated or were empty to start.
func LoadTLSFiles(c *Config) error {
	var err error
	c.CAData, err = dataFromSliceOrFile(c.CAData, c.CAFile)
	if err != nil {
		return err
	}

	c.CertData, err = dataFromSliceOrFile(c.CertData, c.CertFile)
	if err != nil {
		return err
	}

	c.KeyData, err = dataFromSliceOrFile(c.KeyData, c.KeyFile)
	if err != nil {
		return err
	}
	return nil
}

// dataFromSliceOrFile returns data from the slice (if non-empty), or from the file,
// or an error if an error occurred reading the file
func dataFromSliceOrFile(data []byte, file string) ([]byte, error) {
	if len(data) > 0 {
		return data, nil
	}
	if len(file) > 0 {
		fileData, err := ioutil.ReadFile(file)
		if err != nil {
			return []byte{}, err
		}
		return fileData, nil
	}
	return nil, nil
}

func AddUserAgent(config *Config, userAgent string) *Config {
	fullUserAgent := DefaultKubernetesUserAgent() + "/" + userAgent
	config.UserAgent = fullUserAgent
	return config
}

// AnonymousClientConfig returns a copy of the given config with all user credentials (cert/key, bearer token, and username/password) removed
func AnonymousClientConfig(config *Config) *Config {
	// copy only known safe fields
	return &Config{
		Host:           config.Host,
		AlternateHosts: config.AlternateHosts,
		APIPath:        config.APIPath,
		Prefix:         config.Prefix,
		ContentConfig:  config.ContentConfig,
		TLSClientConfig: TLSClientConfig{
			CAFile: config.TLSClientConfig.CAFile,
			CAData: config.TLSClientConfig.CAData,
		},
		RateLimiter:   config.RateLimiter,
		Insecure:      config.Insecure,
		UserAgent:     config.UserAgent,
		Transport:     config.Transport,
		WrapTransport: config.WrapTransport,
		QPS:           config.QPS,
		Burst:         config.Burst,
		Timeout:       config.Timeout,
	}
}
