package apiserver

import (
	extensionsapiserver "k8s.io/apiextensions-apiserver/pkg/apiserver"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apiserveroptions "k8s.io/apiserver/pkg/server/options"
	aggregatorscheme "k8s.io/kube-aggregator/pkg/apiserver/scheme"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/controlplane/apiserver/options"
	generatedopenapi "k8s.io/kubernetes/pkg/generated/openapi"
	netutils "k8s.io/utils/net"
	"net"
	"testing"
)

func TestBuildGenericConfig(t *testing.T) {
	opts := options.NewOptions()
	s := (&apiserveroptions.SecureServingOptions{
		BindAddress: netutils.ParseIPSloppy("127.0.0.1"),
	}).WithLoopback()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on 127.0.0.1:0")
	}
	defer ln.Close()
	s.Listener = ln
	s.BindPort = ln.Addr().(*net.TCPAddr).Port
	opts.SecureServing = s

	completedOptions, err := opts.Complete(nil, nil)
	if err != nil {
		t.Fatalf("Failed to complete apiserver options: %v", err)
	}

	genericConfig, _, storageFactory, err := BuildGenericConfig(
		completedOptions,
		[]*runtime.Scheme{legacyscheme.Scheme, extensionsapiserver.Scheme, aggregatorscheme.Scheme},
		generatedopenapi.GetOpenAPIDefinitions,
	)
	if err != nil {
		t.Fatalf("Failed to build generic config: %v", err)
	}
	if genericConfig.StorageObjectCountTracker == nil {
		t.Errorf("genericConfig StorageObjectCountTracker is absent")
	}
	if genericConfig.StorageObjectCountTracker != storageFactory.StorageConfig.StorageObjectCountTracker {
		t.Errorf("There are different StorageObjectCountTracker in genericConfig and storageFactory")
	}

	restOptions, err := genericConfig.RESTOptionsGetter.GetRESTOptions(schema.GroupResource{Group: "", Resource: ""})
	if err != nil {
		t.Fatal(err)
	}
	if restOptions.StorageConfig.StorageObjectCountTracker != genericConfig.StorageObjectCountTracker {
		t.Errorf("There are different StorageObjectCountTracker in restOptions and serverConfig")
	}
}
