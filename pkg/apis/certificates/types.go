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

package certificates

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	api "k8s.io/kubernetes/pkg/apis/core"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Describes a certificate signing request
type CertificateSigningRequest struct {
	metav1.TypeMeta
	// +optional
	metav1.ObjectMeta

	// The certificate request itself and any additional information.
	// +optional
	Spec CertificateSigningRequestSpec

	// Derived information about the request.
	// +optional
	Status CertificateSigningRequestStatus
}

// This information is immutable after the request is created. Only the Request
// and Usages fields can be set on creation, other fields are derived by
// Kubernetes and cannot be modified by users.
type CertificateSigningRequestSpec struct {
	// Base64-encoded PKCS#10 CSR data
	Request []byte

	// signerName indicates the requested signer, and is a qualified name.
	//
	// List/watch requests for CertificateSigningRequests can filter on this field using a "spec.signerName=NAME" fieldSelector.
	//
	// Well-known Kubernetes signers are:
	//  1. "kubernetes.io/kube-apiserver-client": issues client certificates that can be used to authenticate to kube-apiserver.
	//   Requests for this signer are never auto-approved by kube-controller-manager, can be issued by the "csrsigning" controller in kube-controller-manager.
	//  2. "kubernetes.io/kube-apiserver-client-kubelet": issues client certificates that kubelets use to authenticate to kube-apiserver.
	//   Requests for this signer can be auto-approved by the "csrapproving" controller in kube-controller-manager, and can be issued by the "csrsigning" controller in kube-controller-manager.
	//  3. "kubernetes.io/kubelet-serving" issues serving certificates that kubelets use to serve TLS endpoints, which kube-apiserver can connect to securely.
	//   Requests for this signer are never auto-approved by kube-controller-manager, and can be issued by the "csrsigning" controller in kube-controller-manager.
	//
	// More details are available at https://k8s.io/docs/reference/access-authn-authz/certificate-signing-requests/#kubernetes-signers
	//
	// Custom signerNames can also be specified. The signer defines:
	//  1. Trust distribution: how trust (CA bundles) are distributed.
	//  2. Permitted subjects: and behavior when a disallowed subject is requested.
	//  3. Required, permitted, or forbidden x509 extensions in the request (including whether subjectAltNames are allowed, which types, restrictions on allowed values) and behavior when a disallowed extension is requested.
	//  4. Required, permitted, or forbidden key usages / extended key usages.
	//  5. Expiration/certificate lifetime: whether it is fixed by the signer, configurable by the admin.
	//  6. Whether or not requests for CA certificates are allowed.
	SignerName string

	// expirationSeconds is the requested duration of validity of the issued
	// certificate. The certificate signer may issue a certificate with a different
	// validity duration so a client must check the delta between the notBefore and
	// and notAfter fields in the issued certificate to determine the actual duration.
	//
	// The v1.22+ in-tree implementations of the well-known Kubernetes signers will
	// honor this field as long as the requested duration is not greater than the
	// maximum duration they will honor per the --cluster-signing-duration CLI
	// flag to the Kubernetes controller manager.
	//
	// Certificate signers may not honor this field for various reasons:
	//
	//   1. Old signer that is unaware of the field (such as the in-tree
	//      implementations prior to v1.22)
	//   2. Signer whose configured maximum is shorter than the requested duration
	//   3. Signer whose configured minimum is longer than the requested duration
	//
	// The minimum valid value for expirationSeconds is 600, i.e. 10 minutes.
	//
	// +optional
	ExpirationSeconds *int32

	// usages specifies a set of usage contexts the key will be
	// valid for.
	// See:
	//	https://tools.ietf.org/html/rfc5280#section-4.2.1.3
	//	https://tools.ietf.org/html/rfc5280#section-4.2.1.12
	Usages []KeyUsage

	// Information about the requesting user.
	// See user.Info interface for details.
	// +optional
	Username string
	// UID information about the requesting user.
	// See user.Info interface for details.
	// +optional
	UID string
	// Group information about the requesting user.
	// See user.Info interface for details.
	// +optional
	Groups []string
	// Extra information about the requesting user.
	// See user.Info interface for details.
	// +optional
	Extra map[string]ExtraValue
}

// Built in signerName values that are honoured by kube-controller-manager.
// None of these usages are related to ServiceAccount token secrets
// `.data[ca.crt]` in any way.
const (
	// Signs certificates that will be honored as client-certs by the
	// kube-apiserver. Never auto-approved by kube-controller-manager.
	KubeAPIServerClientSignerName = "kubernetes.io/kube-apiserver-client"

	// Signs client certificates that will be honored as client-certs by the
	// kube-apiserver for a kubelet.
	// May be auto-approved by kube-controller-manager.
	KubeAPIServerClientKubeletSignerName = "kubernetes.io/kube-apiserver-client-kubelet"

	// "kubernetes.io/kube-apiserver-client-pod" issues client certificates that pods can use to authenticate to kube-apiserver.
	// Pods can only obtain these certificates by using PodCertificate projected volumes.
	// Can be auto-approved by the "csrapproving" controller in kube-controller-manager.
	// Can be issued by the "csrsigning" controller in kube-controller-manager.
	KubeAPIServerClientPodSignerName = "kubernetes.io/kube-apiserver-client-pod"

	// Signs serving certificates that are honored as a valid kubelet serving
	// certificate by the kube-apiserver, but has no other guarantees.
	KubeletServingSignerName = "kubernetes.io/kubelet-serving"

	// Has no guarantees for trust at all. Some distributions may honor these
	// as client certs, but that behavior is not standard kubernetes behavior.
	LegacyUnknownSignerName = "kubernetes.io/legacy-unknown"
)

// ExtraValue masks the value so protobuf can generate
type ExtraValue []string

type CertificateSigningRequestStatus struct {
	// Conditions applied to the request, such as approval or denial.
	// +optional
	Conditions []CertificateSigningRequestCondition

	// If request was approved, the controller will place the issued certificate here.
	// +optional
	Certificate []byte
}

type RequestConditionType string

// These are the possible conditions for a certificate request.
const (
	CertificateApproved RequestConditionType = "Approved"
	CertificateDenied   RequestConditionType = "Denied"
	CertificateFailed   RequestConditionType = "Failed"
)

type CertificateSigningRequestCondition struct {
	// type of the condition. Known conditions include "Approved", "Denied", and "Failed".
	Type RequestConditionType
	// Status of the condition, one of True, False, Unknown.
	// Approved, Denied, and Failed conditions may not be "False" or "Unknown".
	// If unset, should be treated as "True".
	// +optional
	Status api.ConditionStatus
	// brief reason for the request state
	// +optional
	Reason string
	// human readable message with details about the request state
	// +optional
	Message string
	// timestamp for the last update to this condition
	// +optional
	LastUpdateTime metav1.Time
	// lastTransitionTime is the time the condition last transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CertificateSigningRequestList struct {
	metav1.TypeMeta
	// +optional
	metav1.ListMeta

	// +optional
	Items []CertificateSigningRequest
}

// KeyUsages specifies valid usage contexts for keys.
// See:
//
//	https://tools.ietf.org/html/rfc5280#section-4.2.1.3
//	https://tools.ietf.org/html/rfc5280#section-4.2.1.12
type KeyUsage string

const (
	UsageSigning           KeyUsage = "signing"
	UsageDigitalSignature  KeyUsage = "digital signature"
	UsageContentCommitment KeyUsage = "content commitment"
	UsageKeyEncipherment   KeyUsage = "key encipherment"
	UsageKeyAgreement      KeyUsage = "key agreement"
	UsageDataEncipherment  KeyUsage = "data encipherment"
	UsageCertSign          KeyUsage = "cert sign"
	UsageCRLSign           KeyUsage = "crl sign"
	UsageEncipherOnly      KeyUsage = "encipher only"
	UsageDecipherOnly      KeyUsage = "decipher only"
	UsageAny               KeyUsage = "any"
	UsageServerAuth        KeyUsage = "server auth"
	UsageClientAuth        KeyUsage = "client auth"
	UsageCodeSigning       KeyUsage = "code signing"
	UsageEmailProtection   KeyUsage = "email protection"
	UsageSMIME             KeyUsage = "s/mime"
	UsageIPsecEndSystem    KeyUsage = "ipsec end system"
	UsageIPsecTunnel       KeyUsage = "ipsec tunnel"
	UsageIPsecUser         KeyUsage = "ipsec user"
	UsageTimestamping      KeyUsage = "timestamping"
	UsageOCSPSigning       KeyUsage = "ocsp signing"
	UsageMicrosoftSGC      KeyUsage = "microsoft sgc"
	UsageNetscapeSGC       KeyUsage = "netscape sgc"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterTrustBundle is a cluster-scoped container for X.509 trust anchors
// (root certificates).
//
// ClusterTrustBundle objects are considered to be readable by any authenticated
// user in the cluster.
//
// It can be optionally associated with a particular assigner, in which case it
// contains one valid set of trust anchors for that signer. Signers may have
// multiple associated ClusterTrustBundles; each is an independent set of trust
// anchors for that signer.
type ClusterTrustBundle struct {
	metav1.TypeMeta
	// +optional
	metav1.ObjectMeta

	// Spec contains the signer (if any) and trust anchors.
	// +optional
	Spec ClusterTrustBundleSpec
}

// ClusterTrustBundleSpec contains the signer and trust anchors.
type ClusterTrustBundleSpec struct {
	// SignerName indicates the associated signer, if any.
	SignerName string

	// TrustBundle contains the individual X.509 trust anchors for this
	// bundle, as PEM bundle of PEM-wrapped, DER-formatted X.509 certificates.
	//
	// The data must consist only of PEM certificate blocks that parse as valid
	// X.509 certificates.  Each certificate must include a basic constraints
	// extension with the CA bit set.  The API server will reject objects that
	// contain duplicate certificates, or that use PEM block headers.
	//
	// Users of ClusterTrustBundles, including Kubelet, are free to reorder and
	// deduplicate certificate blocks in this file according to their own logic,
	// as well as to drop PEM block headers and inter-block data.
	TrustBundle string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterTrustBundleList is a collection of ClusterTrustBundle objects
type ClusterTrustBundleList struct {
	metav1.TypeMeta
	// +optional
	metav1.ListMeta

	// Items is a collection of ClusterTrustBundle objects
	Items []ClusterTrustBundle
}

// MaxTrustBundleSize is the maximimum size of a single trust bundle field.
const MaxTrustBundleSize = 1 * 1024 * 1024

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodCertificateRequest encodes a pod requesting a certificate from a given
// signer.
//
// Kubelets use this API to implement podCertificate projected volumes
type PodCertificateRequest struct {
	metav1.TypeMeta
	// +optional
	metav1.ObjectMeta

	// Spec contains the details about the certificate being requested.
	Spec PodCertificateRequestSpec

	// Status contains the issued certificate, and a standard set of conditions.
	// +optional
	Status PodCertificateRequestStatus
}

// PodCertificateRequestSpec describes the certificate request.  All fields are
// immutable after creation.
type PodCertificateRequestSpec struct {
	// SignerName indicates the request signer.
	SignerName string

	// PodName is the name of the pod into which the certificate will be mounted.
	PodName string
	// PodUID is the UID of the pod into which the certificate will be mounted.
	PodUID types.UID

	// ServiceAccountname is the name of the service account the pod is running as.
	ServiceAccountName string

	// ServiceAccountUID is the UID of the service account the pod is running as.
	ServiceAccountUID types.UID

	// NodeName is the name of the node the pod is assigned to.
	NodeName types.NodeName

	// NodeUID is the UID of the node the pod is assigned to.
	NodeUID types.UID

	// PKIXPublicKey is the PKIX-serialized public key the signer should issue
	// the certificate to.
	//
	// The key must be one of RSA-3072, RSA-4096, ECDSA-P256 or ECDSA-P384.
	PKIXPublicKey []byte

	// ProofOfPossession proves that the requesting Kubelet holds the private
	// key corresponding to pkixPublicKey.
	//
	// If the key is an RSA key, then the signature is over the ASCII bytes of
	// the pod UID, using  RSASSA-PKCS1-V1_5-SIGN from RSA PKCS #1 v1.5 (as
	// implemented by the golang function crypto/rsa.SignPKCS1v15).
	//
	// If the key is an ECDSA key, then the signature is as described by [SEC 1,
	// Version 2.0](https://www.secg.org/sec1-v2.pdf) (as implemented by the
	// golang library function crypto/ecdsa.SignASN1)
	ProofOfPossession []byte
}

type PodCertificateRequestStatus struct {
	// Conditions applied to the request. Known conditions are "Denied",
	// "Failed", and "SuggestedKeyType".
	//
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition

	// CertificateChain is populated with an issued certificate by the signer.
	// This field is set via the /status subresource. Once populated, this field
	// is immutable.
	//
	// If the certificate signing request is denied, a condition of type
	// "Denied" is added and this field remains empty. If the signer cannot
	// issue the certificate, a condition of type "Failed" is added and this
	// field remains empty.
	//
	// Validation requirements:
	//  1. certificateChain must contain one or more entries.
	//  2. Each entry must contain a single DER-encoded ASN.1 Certificate structure as described in section 4 of RFC5280.
	//
	// If more than one entry is present, and the definition of the requested
	// spec.signerName does not indicate otherwise, the first entry is the
	// issued certificate, and subsequent entries should be treated as
	// intermediate certificates and presented in TLS handshakes.  When
	// projecting the chain into a pod volume, kubelet will preserve the order
	// of certificateChain.
	//
	// As a reminder, when serialized as JSON or YAML, each entry is
	// additionally base64-encoded.
	//
	// +listType=atomic
	// +optional
	CertificateChain [][]byte

	// IssuedAt is the time at which the signer issued the certificate.  This
	// field is set via the /status subresource.  Once populated, it is
	// immutable.
	//
	// +optional
	IssuedAt metav1.Time

	// NotBefore is the time at which the certificate becomes valid.  This field
	// is set via the /status subresource.  Once populated, it is immutable.
	//
	// +optional
	NotBefore metav1.Time

	// BeginRefreshAt is the time at which the kubelet should begin trying to
	// refresh the certificate.  This field is set via the /status subresource,
	// and must be set at the same time as certificateChain.  Once populated,
	// this field is immutable.
	//
	// +optional
	BeginRefreshAt metav1.Time

	// NotAfter is the time at which the certificate expires.  This field is set
	// via the /status subresource.  Once populated, it is immutable.
	//
	// +optional
	NotAfter metav1.Time
}

// Well-known condition types for PodCertificateRequests
const (
	// Denied indicates the request was denied by the signer.
	PodCertificateRequestConditionTypeDenied string = "Denied"
	// Failed indicates the signer failed to issue the certificate.
	PodCertificateRequestConditionTypeFailed string = "Failed"
	// SuggestedKeyType is an auxiliary condition that a signer can attach if it
	// denied the request due to an unsupported key type.
	PodCertificateRequestConditionTypeSuggestedKeyType string = "SuggestedKeyType"
)

// Well-known condition reasons for PodCertificateRequests
const (
	// UnsupportedKeyType should be set on "Denied" conditions when the signer
	// doesn't support the key type of publicKey.
	PodCertificateRequestConditionUnsupportedKeyType string = "UnsupportedKeyType"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodCertificateRequestList is a collection of PodCertificateRequest objects.
type PodCertificateRequestList struct {
	metav1.TypeMeta
	// +optional
	metav1.ListMeta

	// Items is a collection of PodCertificateRequest objects
	Items []PodCertificateRequest
}
