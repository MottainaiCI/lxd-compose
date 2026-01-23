/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package helpers_security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

// Derived Key Argon2 options
type DKA_Opts struct {
	TimeIterations uint32
	MemoryUsage    uint32
	KeyLength      uint32
	Parallelism    uint8
}

func NewDKAOpts(timeIterations, memoryUsage, keyLength uint32, parallelism uint8) *DKA_Opts {
	return &DKA_Opts{
		TimeIterations: timeIterations,
		MemoryUsage:    memoryUsage,
		Parallelism:    parallelism,
		KeyLength:      keyLength,
	}
}

func NewDKAOptsDefault() *DKA_Opts {
	return NewDKAOpts(
		3,       // Number of cycles
		64*1024, // 64 MB of RAM
		32,      // Length of the key
		4,       // Number of threads/core to use
	)
}

func Encrypt(plaintext []byte, key []byte, opts *DKA_Opts) ([]byte, error) {
	// Generate a randow salt
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	// Elaborate key with derived key argon2 algorithm
	dKey := argon2.IDKey(key, salt,
		opts.TimeIterations,
		opts.MemoryUsage,
		opts.Parallelism,
		opts.KeyLength,
	)

	block, err := aes.NewCipher(dKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate a random nonce for the encryption
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Add salt at the begin
	result := append(salt, nonce...)

	return gcm.Seal(result, nonce, plaintext, nil), nil
}

func Decrypt(cipherText []byte, key []byte, opts *DKA_Opts) ([]byte, error) {

	if len(cipherText) < 16 {
		return nil, fmt.Errorf("invalid cipher text too short")
	}

	// Retrieve sal and nonce
	salt := cipherText[:16]

	// Elaborate key with derived key argon2 algorithm
	dKey := argon2.IDKey(key, salt,
		opts.TimeIterations,
		opts.MemoryUsage,
		opts.Parallelism,
		opts.KeyLength,
	)

	block, err := aes.NewCipher(dKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < 16+nonceSize {
		return nil, fmt.Errorf("invalid cipher text and nonce too short")
	}

	nonce, cipherText := cipherText[16:16+nonceSize], cipherText[16+nonceSize:]
	return gcm.Open(nil, nonce, cipherText, nil)
}
