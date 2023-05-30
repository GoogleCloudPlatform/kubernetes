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

// Package aes transforms values for storage at rest using AES-GCM.
package aes

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/hkdf"

	"k8s.io/apiserver/pkg/storage/value"
	"k8s.io/klog/v2"
)

// TODO comment
const commonSize = 32

type gcm struct {
	aead      cipher.AEAD
	nonceFunc func([]byte) error
	infoLen   int
}

// NewGCMTransformer takes the given block cipher and performs encryption and decryption on the given data.
// It implements AEAD encryption of the provided values given a cipher.Block algorithm.
// The authenticated data provided as part of the value.Context method must match when the same
// value is set to and loaded from storage. In order to ensure that values cannot be copied by
// an attacker from a location under their control, use characteristics of the storage location
// (such as the etcd key) as part of the authenticated data.
//
// Because this mode requires a generated IV and IV reuse is a known weakness of AES-GCM, keys
// must be rotated before a birthday attack becomes feasible. NIST SP 800-38D
// (http://csrc.nist.gov/publications/nistpubs/800-38D/SP-800-38D.pdf) recommends using the same
// key with random 96-bit nonces (the default nonce length) no more than 2^32 times, and
// therefore transformers using this implementation *must* ensure they allow for frequent key
// rotation. Future work should include investigation of AES-GCM-SIV as an alternative to
// random nonces.
func NewGCMTransformer(block cipher.Block) (value.Transformer, error) {
	aead, err := newGCM(block)
	if err != nil {
		return nil, err
	}

	return &gcm{aead: aead, nonceFunc: randomNonce}, nil
}

// NewGCMTransformerWithUniqueKeyUnsafe is the same as NewGCMTransformer but is unsafe for general
// use because it makes assumptions about the key underlying the block cipher.  Specifically,
// it uses a 96-bit nonce where the first 32 bits are random data and the remaining 64 bits are
// a monotonically incrementing atomic counter.  This means that the key must be randomly generated
// on process startup and must never be used for encryption outside the lifetime of the process.
// Unlike NewGCMTransformer, this function is immune to the birthday attack and thus the key can
// be used for 2^64-1 writes without rotation.  Furthermore, cryptographic wear out of AES-GCM with
// a sequential nonce occurs after 2^64 encryptions, which is not a concern for our use cases.
// Even if that occurs, the nonce counter would overflow and crash the process.  We have no concerns
// around plaintext length because all stored items are small (less than 2 MB).  To prevent the
// chance of the block cipher being accidentally re-used, it is not taken in as input.  Instead,
// a new random key is generated and returned on every invocation of this function.  This key is
// used as the input to the block cipher.  If the key is stored and retrieved at a later point,
// it can be passed to NewGCMTransformer(aes.NewCipher(key)) to construct a transformer capable
// of decrypting values encrypted by this transformer (that transformer must not be used for encryption).
func NewGCMTransformerWithUniqueKeyUnsafe() (value.Transformer, []byte, error) {
	key, err := generateKey(commonSize)
	if err != nil {
		return nil, nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	transformer, err := newGCMTransformerWithUniqueKeyUnsafe(block, newNonceGenerator())
	if err != nil {
		return nil, nil, err
	}
	return transformer, key, nil
}

func newNonceGenerator() *nonceGenerator {
	nonceGen := &nonceGenerator{
		// we start the nonce counter at one billion so that we are
		// guaranteed to detect rollover across different go routines
		zero:  1_000_000_000,
		fatal: die,
	}
	nonceGen.nonce.Add(nonceGen.zero)
	return nonceGen
}

func newGCMTransformerWithUniqueKeyUnsafe(block cipher.Block, nonceGen *nonceGenerator) (*gcm, error) {
	aead, err := newGCM(block)
	if err != nil {
		return nil, err
	}

	nonceFunc := func(b []byte) error {
		// we only need 8 bytes to store our 64 bit incrementing nonce
		// instead of leaving the unused bytes as zeros, set those to random bits
		// this mostly protects us from weird edge cases like a VM restore that rewinds our atomic counter
		randNonceSize := len(b) - 8

		if err := randomNonce(b[:randNonceSize]); err != nil {
			return err
		}

		nonceGen.next(b[randNonceSize:])

		return nil
	}

	return &gcm{aead: aead, nonceFunc: nonceFunc}, nil
}

func newGCM(block cipher.Block) (cipher.AEAD, error) {
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if nonceSize := aead.NonceSize(); nonceSize != 12 { // all data in etcd will be broken if this ever changes
		return nil, fmt.Errorf("crypto/cipher.NewGCM returned unexpected nonce size: %d", nonceSize)
	}
	return aead, nil
}

func randomNonce(b []byte) error {
	_, err := rand.Read(b)
	return err
}

type nonceGenerator struct {
	// even at one million encryptions per second, this counter is enough for half a million years
	// using this struct avoids alignment bugs: https://pkg.go.dev/sync/atomic#pkg-note-BUG
	nonce atomic.Uint64
	zero  uint64
	fatal func(msg string)
}

func (n *nonceGenerator) next(b []byte) {
	incrementingNonce := n.nonce.Add(1)
	if incrementingNonce <= n.zero {
		// this should never happen, and is unrecoverable if it does
		n.fatal("aes-gcm detected nonce overflow - cryptographic wear out has occurred")
	}
	binary.LittleEndian.PutUint64(b, incrementingNonce)
}

func die(msg string) {
	// nolint:logcheck // we want the stack traces, log flushing, and process exiting logic from FatalDepth
	klog.FatalDepth(1, msg)
}

// generateKey generates a random key using system randomness.
func generateKey(length int) (key []byte, err error) {
	defer func(start time.Time) {
		value.RecordDataKeyGeneration(start, err)
	}(time.Now())
	key = make([]byte, length)
	if _, err = rand.Read(key); err != nil {
		return nil, err
	}

	return key, nil
}

// TODO comment
func NewReadOnlyKDFExtendedNonceGCMTransformerFromUniqueKeyUnsafe(key []byte) value.Read {
	return newReadOnlyExtendedNonceGCMTransformerFromUniqueKeyUnsafe(key, nil) // TODO set cache
}

func newReadOnlyExtendedNonceGCMTransformerFromUniqueKeyUnsafe(key []byte, cache *simpleCache) value.Read {
	return &readOnlyExtendedNonceGCM{
		e: &extendedNonceGCM{
			key:   key,
			cache: cache,
		},
	}
}

// readOnlyExtendedNonceGCM is just a wrapper that prevents callers from accessing extendedNonceGCM's write method
type readOnlyExtendedNonceGCM struct {
	e *extendedNonceGCM
}

func (r *readOnlyExtendedNonceGCM) TransformFromStorage(ctx context.Context, data []byte, dataCtx value.Context) ([]byte, bool, error) {
	return r.e.TransformFromStorage(ctx, data, dataCtx)
}

// TODO comment
func NewKDFExtendedNonceGCMTransformerWithUniqueKeyUnsafe() (value.Transformer, []byte, error) {
	key, err := generateKey(commonSize)
	if err != nil {
		return nil, nil, err
	}
	return newExtendedNonceGCMTransformerWithUniqueKeyUnsafe(key, nil), key, nil // TODO set cache
}

func newExtendedNonceGCMTransformerWithUniqueKeyUnsafe(key []byte, cache *simpleCache) value.Transformer {
	return &extendedNonceGCM{
		key:   key,
		cache: cache,
	}
}

type extendedNonceGCM struct {
	key   []byte
	cache *simpleCache
}

func (e *extendedNonceGCM) TransformFromStorage(ctx context.Context, data []byte, dataCtx value.Context) ([]byte, bool, error) {
	if len(data) < commonSize {
		return nil, false, errors.New("the stored data was shorter than the required size")
	}

	info := data[:commonSize]

	transformer, err := e.derivedKeyTransformer(info, dataCtx)
	if err != nil {
		return nil, false, err // TODO fmt.Err
	}

	return transformer.TransformFromStorage(ctx, data, dataCtx)
}

func (e *extendedNonceGCM) TransformToStorage(ctx context.Context, data []byte, dataCtx value.Context) ([]byte, error) {
	info := make([]byte, commonSize)

	if err := randomNonce(info); err != nil {
		return nil, err // TODO fmt.Err
	}

	transformer, err := e.derivedKeyTransformer(info, dataCtx)
	if err != nil {
		return nil, err // TODO fmt.Err
	}

	transformedData, err := transformer.TransformToStorage(ctx, data, dataCtx)
	if err != nil {
		return nil, err // TODO fmt.Err
	}

	copy(transformedData, info)

	return transformedData, nil
}

func (e *extendedNonceGCM) derivedKeyTransformer(info []byte, dataCtx value.Context) (value.Transformer, error) {
	if e.cache != nil {
		if transformer := e.cache.get(info, dataCtx); transformer != nil {
			return transformer, nil
		}
	}

	key, err := e.sha256KDFExpandOnly(info)
	if err != nil {
		return nil, err // TODO fmt.Err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err // TODO fmt.Err
	}

	transformer, err := NewGCMTransformer(block)
	if err != nil {
		return nil, err // TODO fmt.Err
	}
	transformer.(*gcm).infoLen = commonSize // TODO probably better to set as part of the constructor

	if e.cache != nil {
		e.cache.set(info, dataCtx, transformer)
	}

	return transformer, nil
}

func (e *extendedNonceGCM) sha256KDFExpandOnly(info []byte) ([]byte, error) {
	// TODO come up with a way to pre-compute this
	kdf := hkdf.Expand(sha256.New, e.key, info)

	derivedKey := make([]byte, commonSize)
	if _, err := io.ReadFull(kdf, derivedKey); err != nil {
		return nil, err // TODO fmt.Err
	}

	return derivedKey, nil
}

func (t *gcm) TransformFromStorage(ctx context.Context, data []byte, dataCtx value.Context) ([]byte, bool, error) {
	nonceSize := t.aead.NonceSize()
	if len(data) < t.infoLen+nonceSize {
		return nil, false, errors.New("the stored data was shorter than the required size")
	}
	result, err := t.aead.Open(nil, data[t.infoLen:t.infoLen+nonceSize], data[t.infoLen+nonceSize:], dataCtx.AuthenticatedData())
	return result, false, err
}

func (t *gcm) TransformToStorage(ctx context.Context, data []byte, dataCtx value.Context) ([]byte, error) {
	nonceSize := t.aead.NonceSize()
	result := make([]byte, t.infoLen+nonceSize+t.aead.Overhead()+len(data))

	if err := t.nonceFunc(result[t.infoLen : t.infoLen+nonceSize]); err != nil {
		return nil, fmt.Errorf("failed to write nonce for AES-GCM: %w", err)
	}

	cipherText := t.aead.Seal(result[t.infoLen+nonceSize:t.infoLen+nonceSize], result[t.infoLen:t.infoLen+nonceSize], data, dataCtx.AuthenticatedData())
	return result[:t.infoLen+nonceSize+len(cipherText)], nil
}

// cbc implements encryption at rest of the provided values given a cipher.Block algorithm.
type cbc struct {
	block cipher.Block
}

// NewCBCTransformer takes the given block cipher and performs encryption and decryption on the given
// data.
func NewCBCTransformer(block cipher.Block) value.Transformer {
	return &cbc{block: block}
}

var (
	errInvalidBlockSize    = errors.New("the stored data is not a multiple of the block size")
	errInvalidPKCS7Data    = errors.New("invalid PKCS7 data (empty or not padded)")
	errInvalidPKCS7Padding = errors.New("invalid padding on input")
)

func (t *cbc) TransformFromStorage(ctx context.Context, data []byte, dataCtx value.Context) ([]byte, bool, error) {
	blockSize := aes.BlockSize
	if len(data) < blockSize {
		return nil, false, errors.New("the stored data was shorter than the required size")
	}
	iv := data[:blockSize]
	data = data[blockSize:]

	if len(data)%blockSize != 0 {
		return nil, false, errInvalidBlockSize
	}

	result := make([]byte, len(data))
	copy(result, data)
	mode := cipher.NewCBCDecrypter(t.block, iv)
	mode.CryptBlocks(result, result)

	// remove and verify PKCS#7 padding for CBC
	c := result[len(result)-1]
	paddingSize := int(c)
	size := len(result) - paddingSize
	if paddingSize == 0 || paddingSize > len(result) {
		return nil, false, errInvalidPKCS7Data
	}
	for i := 0; i < paddingSize; i++ {
		if result[size+i] != c {
			return nil, false, errInvalidPKCS7Padding
		}
	}

	return result[:size], false, nil
}

func (t *cbc) TransformToStorage(ctx context.Context, data []byte, dataCtx value.Context) ([]byte, error) {
	blockSize := aes.BlockSize
	paddingSize := blockSize - (len(data) % blockSize)
	result := make([]byte, blockSize+len(data)+paddingSize)
	iv := result[:blockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, errors.New("unable to read sufficient random bytes")
	}
	copy(result[blockSize:], data)

	// add PKCS#7 padding for CBC
	copy(result[blockSize+len(data):], bytes.Repeat([]byte{byte(paddingSize)}, paddingSize))

	mode := cipher.NewCBCEncrypter(t.block, iv)
	mode.CryptBlocks(result[blockSize:], result[blockSize:])
	return result, nil
}
