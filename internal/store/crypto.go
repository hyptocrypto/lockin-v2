package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// Salt for key derivation - must match your existing data
var keySalt = []byte{0x59, 0xa8, 0x42, 0x85, 0x8d, 0x95, 0xe1, 0xb9, 0x0e, 0x19, 0x11, 0x17, 0x03, 0x2e, 0x0a, 0x9d}

// XOR mask for obfuscating the master key in memory
var xorMask = []byte{
	0x3c, 0x7f, 0x1a, 0x9e, 0x5b, 0xd2, 0x48, 0xe3,
	0x71, 0x0c, 0x8a, 0xf5, 0x29, 0x64, 0xb7, 0x03,
	0x3c, 0x7f, 0x1a, 0x9e, 0x5b, 0xd2, 0x48, 0xe3,
	0x71, 0x0c, 0x8a, 0xf5, 0x29, 0x64, 0xb7, 0x03,
}

// deriveKey derives a 32-byte key from the master password using PBKDF2
func deriveKey(password string) []byte {
	return pbkdf2.Key([]byte(password), keySalt, 100000, 32, sha256.New)
}

// xorBytes applies XOR operation to obfuscate/deobfuscate data
func xorBytes(data, mask []byte) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] ^ mask[i%len(mask)]
	}
	return result
}

// getMasterKey retrieves and deobfuscates the master key
func (v *FileVault) getMasterKey() []byte {
	if v.obfuscatedKey == nil {
		return nil
	}
	return xorBytes(v.obfuscatedKey, xorMask)
}

// setMasterKey obfuscates and stores the master key
func (v *FileVault) setMasterKey(key []byte) {
	v.obfuscatedKey = xorBytes(key, xorMask)
}

// clearMasterKey securely clears the master key from memory
func (v *FileVault) clearMasterKey() {
	if v.obfuscatedKey != nil {
		for i := range v.obfuscatedKey {
			v.obfuscatedKey[i] = 0
		}
		v.obfuscatedKey = nil
	}
}

// encrypt encrypts plaintext using AES-GCM
func (v *FileVault) encrypt(plaintext string) (string, error) {
	key := v.getMasterKey()
	if key == nil {
		return "", ErrVaultLocked
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts ciphertext using AES-GCM
func (v *FileVault) decrypt(ciphertext string) (string, error) {
	key := v.getMasterKey()
	if key == nil {
		return "", ErrVaultLocked
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
