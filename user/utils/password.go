package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"
)

var (
	ErrPasswordMismatch = errors.New("password verification failed")
	ErrDecryptionFailed = errors.New("failed to decrypt stored password")
)

// TODO: encryptionKey should set in .env
var encryptionKey = []byte("LIFE_IS_FUCKER_MOVIE!=HE_TOLD_ME")

func HashForStorage(h1Hex string) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	h1Bytes, err := hex.DecodeString(h1Hex)
	if err != nil {
		return "", fmt.Errorf("decode H1: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, h1Bytes, nil)

	return hex.EncodeToString(ciphertext), nil
}

func VerifyPassword(storedEncrypted, clientHash, requestID string) error {
	h1Bytes, err := decryptH1(storedEncrypted)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}
	h1Hex := hex.EncodeToString(h1Bytes)

	// use UTC
	nowSec := time.Now().UTC().Unix()
	timestamps := make([]int64, 0, 11)
	for i := int64(-5); i <= 5; i++ {
		timestamps = append(timestamps, nowSec+i)
	}

	for _, ts := range timestamps {
		plain := fmt.Sprintf("%s:%d:%s", h1Hex, ts, requestID)
		h := sha256.Sum256([]byte(plain))
		expectedHash := hex.EncodeToString(h[:])

		if subtle.ConstantTimeCompare([]byte(expectedHash), []byte(clientHash)) == 1 {
			return nil
		}
	}

	return ErrPasswordMismatch
}

func decryptH1(encryptedHex string) ([]byte, error) {
	ciphertext, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}
