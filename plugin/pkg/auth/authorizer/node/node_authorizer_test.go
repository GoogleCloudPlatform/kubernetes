/*
Copyright 2017 The Kubernetes Authors.

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

package node

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"sync/atomic"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/featuregate"
	"k8s.io/kubernetes/pkg/auth/nodeidentifier"
	"k8s.io/kubernetes/plugin/pkg/auth/authorizer/rbac/bootstrappolicy"
)

func TestAuthorizer(t *testing.T) {
	g := NewGraph()

	opts := &sampleDataOpts{
		nodes:                              2,
		namespaces:                         2,
		podsPerNode:                        2,
		attachmentsPerNode:                 1,
		sharedConfigMapsPerPod:             0,
		uniqueConfigMapsPerPod:             1,
		sharedSecretsPerPod:                1,
		uniqueSecretsPerPod:                1,
		sharedPVCsPerPod:                   0,
		uniquePVCsPerPod:                   1,
		uniqueResourceClaimsPerPod:         1,
		uniqueResourceClaimTemplatesPerPod: 1,
		uniqueResourceClaimTemplatesWithClaimPerPod: 1,
	}
	nodes, pods, pvs, attachments := generate(opts)
	populate(g, nodes, pods, pvs, attachments)

	identifier := nodeidentifier.NewDefaultNodeIdentifier()
	authz := NewAuthorizer(g, identifier, bootstrappolicy.NodeRules())

	node0 := &user.DefaultInfo{Name: "system:node:node0", Groups: []string{"system:nodes"}}

	tests := []struct {
		name     string
		attrs    authorizer.AttributesRecord
		expect   authorizer.Decision
		features featuregate.FeatureGate
	}{
		{
			name:   "allowed configmap",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "configmaps", Name: "configmap0-pod0-node0", Namespace: "ns0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "allowed secret via pod",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "secrets", Name: "secret0-pod0-node0", Namespace: "ns0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "list allowed secret via pod",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "list", Resource: "secrets", Name: "secret0-pod0-node0", Namespace: "ns0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "watch allowed secret via pod",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "watch", Resource: "secrets", Name: "secret0-pod0-node0", Namespace: "ns0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "disallowed list many secrets",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "list", Resource: "secrets", Name: "", Namespace: "ns0"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed watch many secrets",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "watch", Resource: "secrets", Name: "", Namespace: "ns0"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed list secrets from all namespaces with name",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "list", Resource: "secrets", Name: "secret0-pod0-node0", Namespace: ""},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "allowed shared secret via pod",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "secrets", Name: "secret0-shared", Namespace: "ns0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "allowed shared secret via pvc",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "secrets", Name: "secret-pv0-pod0-node0-ns0", Namespace: "ns0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "allowed pvc",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "persistentvolumeclaims", Name: "pvc0-pod0-node0", Namespace: "ns0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "allowed resource claim",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "resourceclaims", APIGroup: "resource.k8s.io", Name: "claim0-pod0-node0-ns0", Namespace: "ns0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "allowed resource claim with template",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "resourceclaims", APIGroup: "resource.k8s.io", Name: "generated-claim-pod0-node0-ns0-0", Namespace: "ns0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "allowed pv",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "persistentvolumes", Name: "pv0-pod0-node0-ns0", Namespace: ""},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "disallowed configmap",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "configmaps", Name: "configmap0-pod0-node1", Namespace: "ns0"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed secret via pod",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "secrets", Name: "secret0-pod0-node1", Namespace: "ns0"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed shared secret via pvc",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "secrets", Name: "secret-pv0-pod0-node1-ns0", Namespace: "ns0"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed pvc",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "persistentvolumeclaims", Name: "pvc0-pod0-node1", Namespace: "ns0"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed resource claim, other node",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "resourceclaims", APIGroup: "resource.k8s.io", Name: "claim0-pod0-node1-ns0", Namespace: "ns0"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed resource claim with template",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "resourceclaims", APIGroup: "resource.k8s.io", Name: "pod0-node1-claimtemplate0", Namespace: "ns0"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed pv",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "persistentvolumes", Name: "pv0-pod0-node1-ns0", Namespace: ""},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "allowed attachment",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "volumeattachments", APIGroup: "storage.k8s.io", Name: "attachment0-node0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "allowed svcacct token create",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "create", Resource: "serviceaccounts", Subresource: "token", Name: "svcacct0-node0", Namespace: "ns0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "disallowed svcacct token create - serviceaccount not attached to node",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "create", Resource: "serviceaccounts", Subresource: "token", Name: "svcacct0-node1", Namespace: "ns0"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed svcacct token create - no subresource",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "create", Resource: "serviceaccounts", Name: "svcacct0-node0", Namespace: "ns0"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed svcacct token create - non create",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "update", Resource: "serviceaccounts", Subresource: "token", Name: "svcacct0-node0", Namespace: "ns0"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed get lease in namespace other than kube-node-lease - feature enabled",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "leases", APIGroup: "coordination.k8s.io", Name: "node0", Namespace: "foo"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed create lease in namespace other than kube-node-lease - feature enabled",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "create", Resource: "leases", APIGroup: "coordination.k8s.io", Name: "node0", Namespace: "foo"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed update lease in namespace other than kube-node-lease - feature enabled",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "update", Resource: "leases", APIGroup: "coordination.k8s.io", Name: "node0", Namespace: "foo"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed patch lease in namespace other than kube-node-lease - feature enabled",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "patch", Resource: "leases", APIGroup: "coordination.k8s.io", Name: "node0", Namespace: "foo"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed delete lease in namespace other than kube-node-lease - feature enabled",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "delete", Resource: "leases", APIGroup: "coordination.k8s.io", Name: "node0", Namespace: "foo"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed get another node's lease - feature enabled",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "leases", APIGroup: "coordination.k8s.io", Name: "node1", Namespace: corev1.NamespaceNodeLease},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed update another node's lease - feature enabled",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "update", Resource: "leases", APIGroup: "coordination.k8s.io", Name: "node1", Namespace: corev1.NamespaceNodeLease},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed patch another node's lease - feature enabled",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "patch", Resource: "leases", APIGroup: "coordination.k8s.io", Name: "node1", Namespace: corev1.NamespaceNodeLease},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed delete another node's lease - feature enabled",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "delete", Resource: "leases", APIGroup: "coordination.k8s.io", Name: "node1", Namespace: corev1.NamespaceNodeLease},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed list node leases - feature enabled",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "list", Resource: "leases", APIGroup: "coordination.k8s.io", Namespace: corev1.NamespaceNodeLease},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed watch node leases - feature enabled",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "watch", Resource: "leases", APIGroup: "coordination.k8s.io", Namespace: corev1.NamespaceNodeLease},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "allowed get node lease - feature enabled",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "leases", APIGroup: "coordination.k8s.io", Name: "node0", Namespace: corev1.NamespaceNodeLease},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "allowed create node lease - feature enabled",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "create", Resource: "leases", APIGroup: "coordination.k8s.io", Name: "node0", Namespace: corev1.NamespaceNodeLease},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "allowed update node lease - feature enabled",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "update", Resource: "leases", APIGroup: "coordination.k8s.io", Name: "node0", Namespace: corev1.NamespaceNodeLease},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "allowed patch node lease - feature enabled",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "patch", Resource: "leases", APIGroup: "coordination.k8s.io", Name: "node0", Namespace: corev1.NamespaceNodeLease},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "allowed delete node lease - feature enabled",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "delete", Resource: "leases", APIGroup: "coordination.k8s.io", Name: "node0", Namespace: corev1.NamespaceNodeLease},
			expect: authorizer.DecisionAllow,
		},
		// CSINode
		{
			name:   "disallowed CSINode with subresource - feature enabled",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "csinodes", Subresource: "csiDrivers", APIGroup: "storage.k8s.io", Name: "node0"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed get another node's CSINode",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "csinodes", APIGroup: "storage.k8s.io", Name: "node1"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed update another node's CSINode",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "update", Resource: "csinodes", APIGroup: "storage.k8s.io", Name: "node1"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed patch another node's CSINode",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "patch", Resource: "csinodes", APIGroup: "storage.k8s.io", Name: "node1"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed delete another node's CSINode",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "delete", Resource: "csinodes", APIGroup: "storage.k8s.io", Name: "node1"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed list CSINodes",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "list", Resource: "csinodes", APIGroup: "storage.k8s.io"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed watch CSINodes",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "watch", Resource: "csinodes", APIGroup: "storage.k8s.io"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "allowed get CSINode",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "csinodes", APIGroup: "storage.k8s.io", Name: "node0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "allowed create CSINode",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "create", Resource: "csinodes", APIGroup: "storage.k8s.io", Name: "node0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "allowed update CSINode",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "update", Resource: "csinodes", APIGroup: "storage.k8s.io", Name: "node0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "allowed patch CSINode",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "patch", Resource: "csinodes", APIGroup: "storage.k8s.io", Name: "node0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "allowed delete CSINode",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "delete", Resource: "csinodes", APIGroup: "storage.k8s.io", Name: "node0"},
			expect: authorizer.DecisionAllow,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.features == nil {
				authz.features = utilfeature.DefaultFeatureGate
			} else {
				authz.features = tc.features
			}
			decision, _, _ := authz.Authorize(context.Background(), tc.attrs)
			if decision != tc.expect {
				t.Errorf("expected %v, got %v", tc.expect, decision)
			}
		})
	}
}

func TestAuthorizerSharedResources(t *testing.T) {
	g := NewGraph()
	g.destinationEdgeThreshold = 1
	identifier := nodeidentifier.NewDefaultNodeIdentifier()
	authz := NewAuthorizer(g, identifier, bootstrappolicy.NodeRules())

	node1 := &user.DefaultInfo{Name: "system:node:node1", Groups: []string{"system:nodes"}}
	node2 := &user.DefaultInfo{Name: "system:node:node2", Groups: []string{"system:nodes"}}
	node3 := &user.DefaultInfo{Name: "system:node:node3", Groups: []string{"system:nodes"}}

	g.AddPod(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod1-node1", Namespace: "ns1"},
		Spec: corev1.PodSpec{
			NodeName: "node1",
			Volumes: []corev1.Volume{
				{VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "node1-only"}}},
				{VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "node1-node2-only"}}},
				{VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "shared-all"}}},
			},
		},
	})
	g.AddPod(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod2-node2", Namespace: "ns1"},
		Spec: corev1.PodSpec{
			NodeName: "node2",
			Volumes: []corev1.Volume{
				{VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "node1-node2-only"}}},
				{VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "shared-all"}}},
			},
		},
	})

	pod3 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod3-node3", Namespace: "ns1"},
		Spec: corev1.PodSpec{
			NodeName: "node3",
			Volumes: []corev1.Volume{
				{VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "shared-all"}}},
			},
		},
	}
	g.AddPod(pod3)

	testcases := []struct {
		User      user.Info
		Secret    string
		ConfigMap string
		Decision  authorizer.Decision
	}{
		{User: node1, Decision: authorizer.DecisionAllow, Secret: "node1-only"},
		{User: node1, Decision: authorizer.DecisionAllow, Secret: "node1-node2-only"},
		{User: node1, Decision: authorizer.DecisionAllow, Secret: "shared-all"},

		{User: node2, Decision: authorizer.DecisionNoOpinion, Secret: "node1-only"},
		{User: node2, Decision: authorizer.DecisionAllow, Secret: "node1-node2-only"},
		{User: node2, Decision: authorizer.DecisionAllow, Secret: "shared-all"},

		{User: node3, Decision: authorizer.DecisionNoOpinion, Secret: "node1-only"},
		{User: node3, Decision: authorizer.DecisionNoOpinion, Secret: "node1-node2-only"},
		{User: node3, Decision: authorizer.DecisionAllow, Secret: "shared-all"},
	}

	for i, tc := range testcases {
		var (
			decision authorizer.Decision
			err      error
		)

		if len(tc.Secret) > 0 {
			decision, _, err = authz.Authorize(context.Background(), authorizer.AttributesRecord{User: tc.User, ResourceRequest: true, Verb: "get", Resource: "secrets", Namespace: "ns1", Name: tc.Secret})
			if err != nil {
				t.Errorf("%d: unexpected error: %v", i, err)
				continue
			}
		} else if len(tc.ConfigMap) > 0 {
			decision, _, err = authz.Authorize(context.Background(), authorizer.AttributesRecord{User: tc.User, ResourceRequest: true, Verb: "get", Resource: "configmaps", Namespace: "ns1", Name: tc.ConfigMap})
			if err != nil {
				t.Errorf("%d: unexpected error: %v", i, err)
				continue
			}
		} else {
			t.Fatalf("test case must include a request for a Secret or ConfigMap")
		}

		if decision != tc.Decision {
			t.Errorf("%d: expected %v, got %v", i, tc.Decision, decision)
		}
	}

	{
		node3SharedSecretGet := authorizer.AttributesRecord{User: node3, ResourceRequest: true, Verb: "get", Resource: "secrets", Namespace: "ns1", Name: "shared-all"}

		decision, _, err := authz.Authorize(context.Background(), node3SharedSecretGet)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if decision != authorizer.DecisionAllow {
			t.Error("expected allowed")
		}

		// should trigger recalculation of the shared secret index
		pod3.Spec.Volumes = nil
		g.AddPod(pod3)

		decision, _, err = authz.Authorize(context.Background(), node3SharedSecretGet)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if decision == authorizer.DecisionAllow {
			t.Errorf("unexpectedly allowed")
		}
	}
}

type sampleDataOpts struct {
	nodes       int
	namespaces  int
	podsPerNode int

	attachmentsPerNode int

	// sharedConfigMapsPerNamespaces defines number of shared configmaps in a given
	// namespace. Each pod then mounts a random set of size `sharedConfigMapsPerPod`
	// from that set. sharedConfigMapsPerPod is used if greater.
	sharedConfigMapsPerNamespace int
	// sharedSecretsPerNamespaces defines number of shared secrets in a given
	// namespace. Each pod then mounts a random set of size `sharedSecretsPerPod`
	// from that set. sharedSecretsPerPod is used if greater.
	sharedSecretsPerNamespace int
	// sharedPVCsPerNamespaces defines number of shared pvcs in a given
	// namespace. Each pod then mounts a random set of size `sharedPVCsPerPod`
	// from that set. sharedPVCsPerPod is used if greater.
	sharedPVCsPerNamespace int

	sharedConfigMapsPerPod int
	sharedSecretsPerPod    int
	sharedPVCsPerPod       int

	uniqueSecretsPerPod                         int
	uniqueConfigMapsPerPod                      int
	uniquePVCsPerPod                            int
	uniqueResourceClaimsPerPod                  int
	uniqueResourceClaimTemplatesPerPod          int
	uniqueResourceClaimTemplatesWithClaimPerPod int
}

func BenchmarkPopulationAllocation(b *testing.B) {
	opts := &sampleDataOpts{
		nodes:                  500,
		namespaces:             200,
		podsPerNode:            200,
		attachmentsPerNode:     20,
		sharedConfigMapsPerPod: 0,
		uniqueConfigMapsPerPod: 1,
		sharedSecretsPerPod:    1,
		uniqueSecretsPerPod:    1,
		sharedPVCsPerPod:       0,
		uniquePVCsPerPod:       1,
	}

	nodes, pods, pvs, attachments := generate(opts)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		g := NewGraph()
		populate(g, nodes, pods, pvs, attachments)
	}
}

func BenchmarkPopulationRetention(b *testing.B) {

	// Run with:
	// go test ./plugin/pkg/auth/authorizer/node -benchmem -bench . -run None -v -o node.test -timeout 300m

	// Evaluate retained memory with:
	// go tool pprof --inuse_space node.test plugin/pkg/auth/authorizer/node/BenchmarkPopulationRetention.profile
	// list populate

	opts := &sampleDataOpts{
		nodes:                  500,
		namespaces:             200,
		podsPerNode:            200,
		attachmentsPerNode:     20,
		sharedConfigMapsPerPod: 0,
		uniqueConfigMapsPerPod: 1,
		sharedSecretsPerPod:    1,
		uniqueSecretsPerPod:    1,
		sharedPVCsPerPod:       0,
		uniquePVCsPerPod:       1,
	}

	nodes, pods, pvs, attachments := generate(opts)
	// Garbage collect before the first iteration
	runtime.GC()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		g := NewGraph()
		populate(g, nodes, pods, pvs, attachments)

		if i == 0 {
			f, _ := os.Create("BenchmarkPopulationRetention.profile")
			runtime.GC()
			pprof.WriteHeapProfile(f)
			f.Close()
			// reference the graph to keep it from getting garbage collected
			_ = fmt.Sprintf("%T\n", g)
		}
	}
}

func BenchmarkWriteIndexMaintenance(b *testing.B) {

	// Run with:
	// go test ./plugin/pkg/auth/authorizer/node -benchmem -bench BenchmarkWriteIndexMaintenance -run None

	opts := &sampleDataOpts{
		// simulate high replication in a small number of namespaces:
		nodes:                  5000,
		namespaces:             1,
		podsPerNode:            1,
		attachmentsPerNode:     20,
		sharedConfigMapsPerPod: 0,
		uniqueConfigMapsPerPod: 1,
		sharedSecretsPerPod:    1,
		uniqueSecretsPerPod:    1,
		sharedPVCsPerPod:       0,
		uniquePVCsPerPod:       1,
	}
	nodes, pods, pvs, attachments := generate(opts)
	g := NewGraph()
	populate(g, nodes, pods, pvs, attachments)
	// Garbage collect before the first iteration
	runtime.GC()
	b.ResetTimer()

	b.SetParallelism(100)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			g.AddPod(pods[0])
		}
	})
}

func BenchmarkUnauthorizedRequests(b *testing.B) {

	// Run with:
	// go test ./plugin/pkg/auth/authorizer/node -benchmem -bench BenchmarkUnauthorizedRequests -run None

	opts := &sampleDataOpts{
		nodes:                  5000,
		namespaces:             1,
		podsPerNode:            1,
		sharedConfigMapsPerPod: 1,
	}
	nodes, pods, pvs, attachments := generate(opts)

	// Create an additional Node that doesn't have access to a shared ConfigMap
	// that all the other Nodes are authorized to read.
	additionalNodeName := "nodeX"
	nsName := "ns0"
	additionalNode := &user.DefaultInfo{Name: fmt.Sprintf("system:node:%s", additionalNodeName), Groups: []string{"system:nodes"}}
	pod := &corev1.Pod{}
	pod.Name = "pod-X"
	pod.Namespace = nsName
	pod.Spec.NodeName = additionalNodeName
	pod.Spec.ServiceAccountName = "svcacct-X"
	pods = append(pods, pod)

	g := NewGraph()
	populate(g, nodes, pods, pvs, attachments)

	identifier := nodeidentifier.NewDefaultNodeIdentifier()
	authz := NewAuthorizer(g, identifier, bootstrappolicy.NodeRules())

	attrs := authorizer.AttributesRecord{User: additionalNode, ResourceRequest: true, Verb: "get", Resource: "configmaps", Name: "configmap0-shared", Namespace: nsName}

	b.SetParallelism(500)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			decision, _, _ := authz.Authorize(context.Background(), attrs)
			if decision != authorizer.DecisionNoOpinion {
				b.Errorf("expected %v, got %v", authorizer.DecisionNoOpinion, decision)
			}
		}
	})
}

func BenchmarkAuthorization(b *testing.B) {
	g := NewGraph()

	opts := &sampleDataOpts{
		// To simulate high replication in a small number of namespaces:
		// nodes:       5000,
		// namespaces:  10,
		// podsPerNode: 10,
		nodes:                  500,
		namespaces:             200,
		podsPerNode:            200,
		attachmentsPerNode:     20,
		sharedConfigMapsPerPod: 0,
		uniqueConfigMapsPerPod: 1,
		sharedSecretsPerPod:    1,
		uniqueSecretsPerPod:    1,
		sharedPVCsPerPod:       0,
		uniquePVCsPerPod:       1,
	}
	nodes, pods, pvs, attachments := generate(opts)
	populate(g, nodes, pods, pvs, attachments)

	identifier := nodeidentifier.NewDefaultNodeIdentifier()
	authz := NewAuthorizer(g, identifier, bootstrappolicy.NodeRules())

	node0 := &user.DefaultInfo{Name: "system:node:node0", Groups: []string{"system:nodes"}}

	tests := []struct {
		name     string
		attrs    authorizer.AttributesRecord
		expect   authorizer.Decision
		features featuregate.FeatureGate
	}{
		{
			name:   "allowed configmap",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "configmaps", Name: "configmap0-pod0-node0", Namespace: "ns0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "allowed secret via pod",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "secrets", Name: "secret0-pod0-node0", Namespace: "ns0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "allowed shared secret via pod",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "secrets", Name: "secret0-shared", Namespace: "ns0"},
			expect: authorizer.DecisionAllow,
		},
		{
			name:   "disallowed configmap",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "configmaps", Name: "configmap0-pod0-node1", Namespace: "ns0"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed secret via pod",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "secrets", Name: "secret0-pod0-node1", Namespace: "ns0"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed shared secret via pvc",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "secrets", Name: "secret-pv0-pod0-node1-ns0", Namespace: "ns0"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed pvc",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "persistentvolumeclaims", Name: "pvc0-pod0-node1", Namespace: "ns0"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed pv",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "persistentvolumes", Name: "pv0-pod0-node1-ns0", Namespace: ""},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "disallowed attachment - no relationship",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "volumeattachments", APIGroup: "storage.k8s.io", Name: "attachment0-node1"},
			expect: authorizer.DecisionNoOpinion,
		},
		{
			name:   "allowed attachment",
			attrs:  authorizer.AttributesRecord{User: node0, ResourceRequest: true, Verb: "get", Resource: "volumeattachments", APIGroup: "storage.k8s.io", Name: "attachment0-node0"},
			expect: authorizer.DecisionAllow,
		},
	}

	podToAdd, _ := generatePod("testwrite", "ns0", "node0", "default", opts)

	b.ResetTimer()
	for _, testWriteContention := range []bool{false, true} {

		shouldWrite := int32(1)
		writes := int64(0)
		_1ms := int64(0)
		_10ms := int64(0)
		_25ms := int64(0)
		_50ms := int64(0)
		_100ms := int64(0)
		_250ms := int64(0)
		_500ms := int64(0)
		_1000ms := int64(0)
		_1s := int64(0)

		contentionPrefix := ""
		if testWriteContention {
			contentionPrefix = "contentious "
			// Start a writer pushing graph modifications 100x a second
			go func() {
				for shouldWrite == 1 {
					go func() {
						start := time.Now()
						authz.graph.AddPod(podToAdd)
						diff := time.Since(start)
						atomic.AddInt64(&writes, 1)
						switch {
						case diff < time.Millisecond:
							atomic.AddInt64(&_1ms, 1)
						case diff < 10*time.Millisecond:
							atomic.AddInt64(&_10ms, 1)
						case diff < 25*time.Millisecond:
							atomic.AddInt64(&_25ms, 1)
						case diff < 50*time.Millisecond:
							atomic.AddInt64(&_50ms, 1)
						case diff < 100*time.Millisecond:
							atomic.AddInt64(&_100ms, 1)
						case diff < 250*time.Millisecond:
							atomic.AddInt64(&_250ms, 1)
						case diff < 500*time.Millisecond:
							atomic.AddInt64(&_500ms, 1)
						case diff < 1000*time.Millisecond:
							atomic.AddInt64(&_1000ms, 1)
						default:
							atomic.AddInt64(&_1s, 1)
						}
					}()
					time.Sleep(10 * time.Millisecond)
				}
			}()
		}

		for _, tc := range tests {
			if tc.features == nil {
				authz.features = utilfeature.DefaultFeatureGate
			} else {
				authz.features = tc.features
			}
			b.Run(contentionPrefix+tc.name, func(b *testing.B) {
				// Run authorization checks in parallel
				b.SetParallelism(5000)
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						decision, _, _ := authz.Authorize(context.Background(), tc.attrs)
						if decision != tc.expect {
							b.Errorf("expected %v, got %v", tc.expect, decision)
						}
					}
				})
			})
		}

		atomic.StoreInt32(&shouldWrite, 0)
		if testWriteContention {
			b.Logf("graph modifications during contention test: %d", writes)
			b.Logf("<1ms=%d, <10ms=%d, <25ms=%d, <50ms=%d, <100ms=%d, <250ms=%d, <500ms=%d, <1000ms=%d, >1000ms=%d", _1ms, _10ms, _25ms, _50ms, _100ms, _250ms, _500ms, _1000ms, _1s)
		} else {
			b.Logf("graph modifications during non-contention test: %d", writes)
		}
	}
}

func populate(graph *Graph, nodes []*corev1.Node, pods []*corev1.Pod, pvs []*corev1.PersistentVolume, attachments []*storagev1.VolumeAttachment) {
	p := &graphPopulator{}
	p.graph = graph
	for _, pod := range pods {
		p.addPod(pod)
	}
	for _, pv := range pvs {
		p.addPV(pv)
	}
	for _, attachment := range attachments {
		p.addVolumeAttachment(attachment)
	}
}

func randomSubset(a, b int) []int {
	if b < a {
		b = a
	}
	return rand.Perm(b)[:a]
}

// generate creates sample pods and persistent volumes based on the provided options.
// the secret/configmap/pvc/node references in the pod and pv objects are named to indicate the connections between the objects.
// for example, secret0-pod0-node0 is a secret referenced by pod0 which is bound to node0.
// when populated into the graph, the node authorizer should allow node0 to access that secret, but not node1.
func generate(opts *sampleDataOpts) ([]*corev1.Node, []*corev1.Pod, []*corev1.PersistentVolume, []*storagev1.VolumeAttachment) {
	nodes := make([]*corev1.Node, 0, opts.nodes)
	pods := make([]*corev1.Pod, 0, opts.nodes*opts.podsPerNode)
	pvs := make([]*corev1.PersistentVolume, 0, (opts.nodes*opts.podsPerNode*opts.uniquePVCsPerPod)+(opts.sharedPVCsPerPod*opts.namespaces))
	attachments := make([]*storagev1.VolumeAttachment, 0, opts.nodes*opts.attachmentsPerNode)

	rand.Seed(12345)

	for n := 0; n < opts.nodes; n++ {
		nodeName := fmt.Sprintf("node%d", n)
		for p := 0; p < opts.podsPerNode; p++ {
			name := fmt.Sprintf("pod%d-%s", p, nodeName)
			namespace := fmt.Sprintf("ns%d", p%opts.namespaces)
			svcAccountName := fmt.Sprintf("svcacct%d-%s", p, nodeName)

			pod, podPVs := generatePod(name, namespace, nodeName, svcAccountName, opts)
			pods = append(pods, pod)
			pvs = append(pvs, podPVs...)
		}
		for a := 0; a < opts.attachmentsPerNode; a++ {
			attachment := &storagev1.VolumeAttachment{}
			attachment.Name = fmt.Sprintf("attachment%d-%s", a, nodeName)
			attachment.Spec.NodeName = nodeName
			attachments = append(attachments, attachment)
		}

		nodes = append(nodes, &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: nodeName},
			Spec:       corev1.NodeSpec{},
		})
	}
	return nodes, pods, pvs, attachments
}

func generatePod(name, namespace, nodeName, svcAccountName string, opts *sampleDataOpts) (*corev1.Pod, []*corev1.PersistentVolume) {
	pvs := make([]*corev1.PersistentVolume, 0, opts.uniquePVCsPerPod+opts.sharedPVCsPerPod)

	pod := &corev1.Pod{}
	pod.Name = name
	pod.Namespace = namespace
	pod.Spec.NodeName = nodeName
	pod.Spec.ServiceAccountName = svcAccountName

	for i := 0; i < opts.uniqueSecretsPerPod; i++ {
		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{SecretName: fmt.Sprintf("secret%d-%s", i, pod.Name)},
		}})
	}
	// Choose shared secrets randomly from shared secrets in a namespace.
	subset := randomSubset(opts.sharedSecretsPerPod, opts.sharedSecretsPerNamespace)
	for _, i := range subset {
		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{SecretName: fmt.Sprintf("secret%d-shared", i)},
		}})
	}

	for i := 0; i < opts.uniqueConfigMapsPerPod; i++ {
		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("configmap%d-%s", i, pod.Name)}},
		}})
	}
	// Choose shared configmaps randomly from shared configmaps in a namespace.
	subset = randomSubset(opts.sharedConfigMapsPerPod, opts.sharedConfigMapsPerNamespace)
	for _, i := range subset {
		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("configmap%d-shared", i)}},
		}})
	}

	for i := 0; i < opts.uniquePVCsPerPod; i++ {
		pv := &corev1.PersistentVolume{}
		pv.Name = fmt.Sprintf("pv%d-%s-%s", i, pod.Name, pod.Namespace)
		pv.Spec.FlexVolume = &corev1.FlexPersistentVolumeSource{SecretRef: &corev1.SecretReference{Name: fmt.Sprintf("secret-%s", pv.Name)}}
		pv.Spec.ClaimRef = &corev1.ObjectReference{Name: fmt.Sprintf("pvc%d-%s", i, pod.Name), Namespace: pod.Namespace}
		pvs = append(pvs, pv)

		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: pv.Spec.ClaimRef.Name},
		}})
	}
	for i := 0; i < opts.uniqueResourceClaimsPerPod; i++ {
		claimName := fmt.Sprintf("claim%d-%s-%s", i, pod.Name, pod.Namespace)
		pod.Spec.ResourceClaims = append(pod.Spec.ResourceClaims, corev1.PodResourceClaim{
			Name: fmt.Sprintf("claim%d", i),
			Source: corev1.ClaimSource{
				ResourceClaimName: &claimName,
			},
		})
	}
	for i := 0; i < opts.uniqueResourceClaimTemplatesPerPod; i++ {
		claimTemplateName := fmt.Sprintf("claimtemplate%d-%s-%s", i, pod.Name, pod.Namespace)
		podClaimName := fmt.Sprintf("claimtemplate%d", i)
		pod.Spec.ResourceClaims = append(pod.Spec.ResourceClaims, corev1.PodResourceClaim{
			Name: podClaimName,
			Source: corev1.ClaimSource{
				ResourceClaimTemplateName: &claimTemplateName,
			},
		})
	}
	for i := 0; i < opts.uniqueResourceClaimTemplatesWithClaimPerPod; i++ {
		claimTemplateName := fmt.Sprintf("claimtemplate%d-%s-%s", i, pod.Name, pod.Namespace)
		podClaimName := fmt.Sprintf("claimtemplate-with-claim%d", i)
		claimName := fmt.Sprintf("generated-claim-%s-%s-%d", pod.Name, pod.Namespace, i)
		pod.Spec.ResourceClaims = append(pod.Spec.ResourceClaims, corev1.PodResourceClaim{
			Name: podClaimName,
			Source: corev1.ClaimSource{
				ResourceClaimTemplateName: &claimTemplateName,
			},
		})
		pod.Status.ResourceClaimStatuses = append(pod.Status.ResourceClaimStatuses, corev1.PodResourceClaimStatus{
			Name:              podClaimName,
			ResourceClaimName: &claimName,
		})
	}
	// Choose shared pvcs randomly from shared pvcs in a namespace.
	subset = randomSubset(opts.sharedPVCsPerPod, opts.sharedPVCsPerNamespace)
	for _, i := range subset {
		pv := &corev1.PersistentVolume{}
		pv.Name = fmt.Sprintf("pv%d-shared-%s", i, pod.Namespace)
		pv.Spec.FlexVolume = &corev1.FlexPersistentVolumeSource{SecretRef: &corev1.SecretReference{Name: fmt.Sprintf("secret-%s", pv.Name)}}
		pv.Spec.ClaimRef = &corev1.ObjectReference{Name: fmt.Sprintf("pvc%d-shared", i), Namespace: pod.Namespace}
		pvs = append(pvs, pv)

		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: pv.Spec.ClaimRef.Name},
		}})
	}

	return pod, pvs
}
