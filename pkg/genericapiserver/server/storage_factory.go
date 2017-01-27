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

package server

import (
	"fmt"
	"mime"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/recognizer"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/storage/storagebackend"

	"github.com/golang/glog"
)

// StorageFactory is the interface to locate the storage for a given GroupResource
type StorageFactory interface {
	// New finds the storage destination for the given group and resource. It will
	// return an error if the group has no storage destination configured.
	NewConfig(groupResource schema.GroupResource) (*storagebackend.Config, error)

	// ResourcePrefix returns the overridden resource prefix for the GroupResource
	// This allows for cohabitation of resources with different native types and provides
	// centralized control over the shape of etcd directories
	ResourcePrefix(groupResource schema.GroupResource) string

	// Backends gets all backends for all registered storage destinations.
	// Used for getting all instances for health validations.
	Backends() []string
}

// DefaultStorageFactory takes a GroupResource and returns back its storage interface.  This result includes:
// 1. Merged etcd config, including: auth, server locations, prefixes
// 2. Resource encodings for storage: group,version,kind to store as
// 3. Cohabitating default: some resources like hpa are exposed through multiple APIs.  They must agree on 1 and 2
type DefaultStorageFactory struct {
	// StorageConfig describes how to create a storage backend in general.
	// Its authentication information will be used for every storage.Interface returned.
	StorageConfig storagebackend.Config

	Overrides map[schema.GroupResource]groupResourceOverrides

	DefaultResourcePrefixes map[schema.GroupResource]string

	// DefaultMediaType is the media type used to store resources. If it is not set, "application/json" is used.
	DefaultMediaType string

	// DefaultSerializer is used to create encoders and decoders for the storage.Interface.
	DefaultSerializer runtime.StorageSerializer

	// ResourceEncodingConfig describes how to encode a particular GroupVersionResource
	ResourceEncodingConfig ResourceEncodingConfig

	// APIResourceConfigSource indicates whether the *storage* is enabled, NOT the API
	// This is discrete from resource enablement because those are separate concerns.  How this source is configured
	// is left to the caller.
	APIResourceConfigSource APIResourceConfigSource

	// newStorageCodecFn exists to be overwritten for unit testing.
	newStorageCodecFn func(opts StorageCodecOptions) (codec runtime.Codec, err error)
}

// StorageCodecOptions are the arguments passed to newStorageCodecFn
type StorageCodecOptions struct {
	StorageMediaType  string
	StorageSerializer runtime.StorageSerializer
	StorageVersion    schema.GroupVersion
	MemoryVersion     schema.GroupVersion
	Config            storagebackend.Config

	EncoderDecoratorFn func(runtime.Encoder) runtime.Encoder
	DecoderDecoratorFn func([]runtime.Decoder) []runtime.Decoder
}

type groupResourceOverrides struct {
	// etcdLocation contains the list of "special" locations that are used for particular GroupResources
	// These are merged on top of the StorageConfig when requesting the storage.Interface for a given GroupResource
	etcdLocation []string
	// etcdPrefix is the base location for a GroupResource.
	etcdPrefix string
	// etcdResourcePrefix is the location to use to store a particular type under the `etcdPrefix` location
	// If empty, the default mapping is used.  If the default mapping doesn't contain an entry, it will use
	// the ToLowered name of the resource, not including the group.
	etcdResourcePrefix string
	// mediaType is the desired serializer to choose. If empty, the default is chosen.
	mediaType string
	// serializer contains the list of "special" serializers for a GroupResource.  Resource=* means for the entire group
	serializer runtime.StorageSerializer
	// cohabitatingResources keeps track of which resources must be stored together.  This happens when we have multiple ways
	// of exposing one set of concepts.  autoscaling.HPA and extensions.HPA as a for instance
	// The order of the slice matters!  It is the priority order of lookup for finding a storage location
	cohabitatingResources []schema.GroupResource
	// encoderDecoratorFn is optional and may wrap the provided encoder prior to being serialized.
	encoderDecoratorFn func(runtime.Encoder) runtime.Encoder
	// decoderDecoratorFn is optional and may wrap the provided decoders (can add new decoders). The order of
	// returned decoders will be priority for attempt to decode.
	decoderDecoratorFn func([]runtime.Decoder) []runtime.Decoder
}

// Apply overrides the provided config and options if the override has a value in that position
func (o groupResourceOverrides) Apply(config *storagebackend.Config, options *StorageCodecOptions) {
	if len(o.etcdLocation) > 0 {
		config.ServerList = o.etcdLocation
	}
	if len(o.etcdPrefix) > 0 {
		config.Prefix = o.etcdPrefix
	}

	if len(o.mediaType) > 0 {
		options.StorageMediaType = o.mediaType
	}
	if o.serializer != nil {
		options.StorageSerializer = o.serializer
	}
	if o.encoderDecoratorFn != nil {
		options.EncoderDecoratorFn = o.encoderDecoratorFn
	}
	if o.decoderDecoratorFn != nil {
		options.DecoderDecoratorFn = o.decoderDecoratorFn
	}
}

var _ StorageFactory = &DefaultStorageFactory{}

const AllResources = "*"

// specialDefaultResourcePrefixes are prefixes compiled into Kubernetes.
// TODO: move out of this package, it is not generic
var specialDefaultResourcePrefixes = map[schema.GroupResource]string{
	schema.GroupResource{Group: "", Resource: "replicationControllers"}:        "controllers",
	schema.GroupResource{Group: "", Resource: "replicationcontrollers"}:        "controllers",
	schema.GroupResource{Group: "", Resource: "endpoints"}:                     "services/endpoints",
	schema.GroupResource{Group: "", Resource: "nodes"}:                         "minions",
	schema.GroupResource{Group: "", Resource: "services"}:                      "services/specs",
	schema.GroupResource{Group: "extensions", Resource: "ingresses"}:           "ingress",
	schema.GroupResource{Group: "extensions", Resource: "podsecuritypolicies"}: "podsecuritypolicy",
}

func NewDefaultStorageFactory(config storagebackend.Config, defaultMediaType string, defaultSerializer runtime.StorageSerializer, resourceEncodingConfig ResourceEncodingConfig, resourceConfig APIResourceConfigSource) *DefaultStorageFactory {
	if len(defaultMediaType) == 0 {
		defaultMediaType = runtime.ContentTypeJSON
	}
	return &DefaultStorageFactory{
		StorageConfig:           config,
		Overrides:               map[schema.GroupResource]groupResourceOverrides{},
		DefaultMediaType:        defaultMediaType,
		DefaultSerializer:       defaultSerializer,
		ResourceEncodingConfig:  resourceEncodingConfig,
		APIResourceConfigSource: resourceConfig,
		DefaultResourcePrefixes: specialDefaultResourcePrefixes,

		newStorageCodecFn: NewStorageCodec,
	}
}

func (s *DefaultStorageFactory) SetEtcdLocation(groupResource schema.GroupResource, location []string) {
	overrides := s.Overrides[groupResource]
	overrides.etcdLocation = location
	s.Overrides[groupResource] = overrides
}

func (s *DefaultStorageFactory) SetEtcdPrefix(groupResource schema.GroupResource, prefix string) {
	overrides := s.Overrides[groupResource]
	overrides.etcdPrefix = prefix
	s.Overrides[groupResource] = overrides
}

// SetResourceEtcdPrefix sets the prefix for a resource, but not the base-dir.  You'll end up in `etcdPrefix/resourceEtcdPrefix`.
func (s *DefaultStorageFactory) SetResourceEtcdPrefix(groupResource schema.GroupResource, prefix string) {
	overrides := s.Overrides[groupResource]
	overrides.etcdResourcePrefix = prefix
	s.Overrides[groupResource] = overrides
}

func (s *DefaultStorageFactory) SetSerializer(groupResource schema.GroupResource, mediaType string, serializer runtime.StorageSerializer) {
	overrides := s.Overrides[groupResource]
	overrides.mediaType = mediaType
	overrides.serializer = serializer
	s.Overrides[groupResource] = overrides
}

// AddCohabitatingResources links resources together the order of the slice matters!  its the priority order of lookup for finding a storage location
func (s *DefaultStorageFactory) AddCohabitatingResources(groupResources ...schema.GroupResource) {
	for _, groupResource := range groupResources {
		overrides := s.Overrides[groupResource]
		overrides.cohabitatingResources = groupResources
		s.Overrides[groupResource] = overrides
	}
}

func (s *DefaultStorageFactory) AddSerializationChains(encoderDecoratorFn func(runtime.Encoder) runtime.Encoder, decoderDecoratorFn func([]runtime.Decoder) []runtime.Decoder, groupResources ...schema.GroupResource) {
	for _, groupResource := range groupResources {
		overrides := s.Overrides[groupResource]
		overrides.encoderDecoratorFn = encoderDecoratorFn
		overrides.decoderDecoratorFn = decoderDecoratorFn
		s.Overrides[groupResource] = overrides
	}
}

func getAllResourcesAlias(resource schema.GroupResource) schema.GroupResource {
	return schema.GroupResource{Group: resource.Group, Resource: AllResources}
}

func (s *DefaultStorageFactory) getStorageGroupResource(groupResource schema.GroupResource) schema.GroupResource {
	for _, potentialStorageResource := range s.Overrides[groupResource].cohabitatingResources {
		if s.APIResourceConfigSource.AnyVersionOfResourceEnabled(potentialStorageResource) {
			return potentialStorageResource
		}
	}

	return groupResource
}

// New finds the storage destination for the given group and resource. It will
// return an error if the group has no storage destination configured.
func (s *DefaultStorageFactory) NewConfig(groupResource schema.GroupResource) (*storagebackend.Config, error) {
	chosenStorageResource := s.getStorageGroupResource(groupResource)

	// operate on copy
	config := s.StorageConfig
	options := StorageCodecOptions{
		StorageMediaType:  s.DefaultMediaType,
		StorageSerializer: s.DefaultSerializer,
	}

	if override, ok := s.Overrides[getAllResourcesAlias(chosenStorageResource)]; ok {
		override.Apply(&config, &options)
	}
	if override, ok := s.Overrides[chosenStorageResource]; ok {
		override.Apply(&config, &options)
	}

	var err error
	options.StorageVersion, err = s.ResourceEncodingConfig.StorageEncodingFor(chosenStorageResource)
	if err != nil {
		return nil, err
	}
	options.MemoryVersion, err = s.ResourceEncodingConfig.InMemoryEncodingFor(groupResource)
	if err != nil {
		return nil, err
	}
	options.Config = config

	config.Codec, err = s.newStorageCodecFn(options)
	if err != nil {
		return nil, err
	}
	glog.V(3).Infof("storing %v in %v, reading as %v from %v", groupResource, options.StorageVersion, options.MemoryVersion, options.Config)

	return &config, nil
}

// Get all backends for all registered storage destinations.
// Used for getting all instances for health validations.
func (s *DefaultStorageFactory) Backends() []string {
	backends := sets.NewString(s.StorageConfig.ServerList...)

	for _, overrides := range s.Overrides {
		backends.Insert(overrides.etcdLocation...)
	}
	return backends.List()
}

// NewStorageCodec assembles a storage codec for the provided storage media type, the provided serializer, and the requested
// storage and memory versions.
func NewStorageCodec(opts StorageCodecOptions) (runtime.Codec, error) {
	mediaType, _, err := mime.ParseMediaType(opts.StorageMediaType)
	if err != nil {
		return nil, fmt.Errorf("%q is not a valid mime-type", opts.StorageMediaType)
	}
	serializer, ok := runtime.SerializerInfoForMediaType(opts.StorageSerializer.SupportedMediaTypes(), mediaType)
	if !ok {
		return nil, fmt.Errorf("unable to find serializer for %q", opts.StorageMediaType)
	}

	s := serializer.Serializer

	// etcd2 only supports string data - we must wrap any result before returning
	// TODO: storagebackend should return a boolean indicating whether it supports binary data
	if !serializer.EncodesAsText && (opts.Config.Type == storagebackend.StorageTypeUnset || opts.Config.Type == storagebackend.StorageTypeETCD2) {
		glog.V(4).Infof("Wrapping the underlying binary storage serializer with a base64 encoding for etcd2")
		s = runtime.NewBase64Serializer(s)
	}

	// Give callers the opportunity to wrap encoders and decoders.  For decoders, each returned decoder will
	// be passed to the recognizer so that multiple decoders are available.
	var encoder runtime.Encoder = s
	if opts.EncoderDecoratorFn != nil {
		encoder = opts.EncoderDecoratorFn(encoder)
	}
	decoders := []runtime.Decoder{s, opts.StorageSerializer.UniversalDeserializer()}
	if opts.DecoderDecoratorFn != nil {
		decoders = opts.DecoderDecoratorFn(decoders)
	}

	// Ensure the storage receives the correct version.
	encoder = opts.StorageSerializer.EncoderForVersion(
		encoder,
		runtime.NewMultiGroupVersioner(
			opts.StorageVersion,
			schema.GroupKind{Group: opts.StorageVersion.Group},
			schema.GroupKind{Group: opts.MemoryVersion.Group},
		),
	)
	decoder := opts.StorageSerializer.DecoderToVersion(
		recognizer.NewDecoder(decoders...),
		runtime.NewMultiGroupVersioner(
			opts.MemoryVersion,
			schema.GroupKind{Group: opts.MemoryVersion.Group},
			schema.GroupKind{Group: opts.StorageVersion.Group},
		),
	)

	return runtime.NewCodec(encoder, decoder), nil
}

func (s *DefaultStorageFactory) ResourcePrefix(groupResource schema.GroupResource) string {
	chosenStorageResource := s.getStorageGroupResource(groupResource)
	groupOverride := s.Overrides[getAllResourcesAlias(chosenStorageResource)]
	exactResourceOverride := s.Overrides[chosenStorageResource]

	etcdResourcePrefix := s.DefaultResourcePrefixes[chosenStorageResource]
	if len(groupOverride.etcdResourcePrefix) > 0 {
		etcdResourcePrefix = groupOverride.etcdResourcePrefix
	}
	if len(exactResourceOverride.etcdResourcePrefix) > 0 {
		etcdResourcePrefix = exactResourceOverride.etcdResourcePrefix
	}
	if len(etcdResourcePrefix) == 0 {
		etcdResourcePrefix = strings.ToLower(chosenStorageResource.Resource)
	}

	return etcdResourcePrefix
}
