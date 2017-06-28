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

// Package kms transforms values for storage at rest using a KMS provider
package kms

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"fmt"
	"strings"
	"sync"

	"k8s.io/apiserver/pkg/storage/value"
	aestransformer "k8s.io/apiserver/pkg/storage/value/encrypt/aes"
)

type kmsTransformer struct {
	kmsService value.KMSService

	transformers []value.PrefixTransformer

	storage value.KMSStorage

	refreshLock sync.RWMutex
}

// NewKMSTransformer returns a transformer which implements a KEK-DEK based envelope encryption scheme.
// It uses kmsService to communicate with the KEK store, and storage to communicate with the DEK store.
func NewKMSTransformer(kmsService value.KMSService, storage value.KMSStorage) (value.Transformer, error) {
	err := storage.Setup()
	if err != nil {
		return nil, err
	}
	transformer := &kmsTransformer{
		kmsService: kmsService,
		storage:    storage,
	}

	deks, err := transformer.storage.GetAllDEKs()
	if err != nil {
		return nil, err
	}
	if len(deks) == 0 {
		// Create new DEK in that case
		// If there are no keys, rotate(false) will create one.
		if err = transformer.Rotate(); err != nil {
			return nil, err
		}
	} else {
		transformer.Refresh()
	}

	return transformer, nil
}

// Rotate creates a new key and makes it the default for writing to disk. It refreshes the transformer
// once the new key has been written to disk.
func (t *kmsTransformer) Rotate() error {
	// Creates, encrypts and stores a new DEK.
	dekBytes, err := generateKey(32)
	if err != nil {
		return err
	}

	keyvalue, err := t.kmsService.Encrypt(dekBytes)
	if err != nil {
		return err
	}

	err = t.storage.StoreNewDEK(keyvalue)
	if err != nil {
		return err
	}

	return t.Refresh()
}

// Refresh reads the DEKs from disk and recreates the transformer.
func (t *kmsTransformer) Refresh() error {
	t.refreshLock.Lock()
	defer t.refreshLock.Unlock()

	deks, err := t.storage.GetAllDEKs()
	if err != nil {
		return err
	}
	transformers := []value.PrefixTransformer{}

	primary := false
	for keyname, encDek := range deks {
		// Check if the current key is the primary key. Necessary because maps are unordered.
		if strings.HasPrefix(keyname, "-") {
			keyname = keyname[1:]
			primary = true
		} else {
			primary = false
		}

		dekBytes, err := t.kmsService.Decrypt(encDek)
		if err != nil {
			return err
		}
		block, err := aes.NewCipher(dekBytes)
		if err != nil {
			return err
		}
		prefixTransformer := value.PrefixTransformer{
			Prefix:      []byte(keyname + ":"),
			Transformer: aestransformer.NewCBCTransformer(block),
		}

		if primary {
			// The primary key has to be at the beginning of the list
			transformers = append([]value.PrefixTransformer{prefixTransformer}, transformers...)
		} else {
			transformers = append(transformers, prefixTransformer)
		}
	}

	if len(transformers) == 0 {
		return fmt.Errorf("did not find any DEKs for KMS transformer")
	}

	// TODO(sakshams): Confirm that this can be done safely without any race conditions.
	t.transformers = transformers

	return nil
}

// TransformFromStorage implements value.Transformer
func (t *kmsTransformer) TransformFromStorage(data []byte, context value.Context) ([]byte, bool, error) {
	// Retry once, if the DEK was not found.
	// Consider the scenario with 2 apiservers. One of them rotates the DEK, and stores some data
	// encrypted with the new key.
	// The second master shall notice the new data, and fail to decrypt it. It shall then re-read DEKs
	// from etcd (Refresh), and be able to decrypt the data in its second attempt.
	// We choose to fail fast in the (possibly unlikely) case where the encrypted data has propagated
	// to the second apiserver, but the new DEK has not.
	for attempt := 0; attempt < 2; attempt++ {
		for i, transformer := range t.transformers {
			if bytes.HasPrefix(data, transformer.Prefix) {
				result, stale, err := transformer.Transformer.TransformFromStorage(data[len(transformer.Prefix):], context)
				if len(transformer.Prefix) == 0 && err != nil {
					continue
				}
				return result, stale || i != 0, err
			}
		}
		// A new key may have been added.
		t.Refresh()
	}
	return nil, false, fmt.Errorf("did not find a transformer to read key")
}

// TransformToStorage implements value.Transformer
func (t *kmsTransformer) TransformToStorage(data []byte, context value.Context) ([]byte, error) {
	transformer := t.transformers[0]
	prefixedData := make([]byte, len(transformer.Prefix), len(data)+len(transformer.Prefix))
	copy(prefixedData, transformer.Prefix)
	result, err := transformer.Transformer.TransformToStorage(data, context)
	if err != nil {
		return nil, err
	}
	prefixedData = append(prefixedData, result...)
	return prefixedData, nil
}

// generateKey generates a random key using system randomness.
func generateKey(length int) ([]byte, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return []byte{}, err
	}

	return key, nil
}
