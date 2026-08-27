package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
	"time"
)

func init() {
	// Test key must be 32 bytes (AES-256)
	encryptionKey = []byte("TEST_KEY_32_BYTES_LONG!!!!!!....")
}

func TestHashForStorage(t *testing.T) {
	data := "123456"
	hash := sha256.Sum256([]byte(data))
	expectedHash := "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92"

	if fmt.Sprintf("%x", hash) != expectedHash {
		t.Errorf("SHA256 hash mismatch.\nExpected: %s\nActual: %x", expectedHash, hash)
	}
	t.Log("✅ SHA256 hash verification passed")
}

// TestFullAuthFlow dynamically simulates the complete registration->login flow
func TestFullAuthFlow(t *testing.T) {
	// === 1. Simulate registration phase ===
	// H1 (SHA256 of password) sent during frontend registration
	registerH1Hex := "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92"

	storedEncrypted, err := HashForStorage(registerH1Hex)
	if err != nil {
		t.Fatalf("HashForStorage failed: %v", err)
	}
	t.Logf("✅ Registration successful, stored ciphertext: %s", storedEncrypted)

	// === 2. Simulate login phase (using current time) ===
	requestID := "1999-aa71-6be8"
	nowSec := time.Now().Unix()

	// Frontend computation: clientHash = SHA256(H1 + ":" + timestamp + ":" + requestId)
	plain := fmt.Sprintf("%s:%d:%s", registerH1Hex, nowSec, requestID)
	h := sha256.Sum256([]byte(plain))
	clientHash := hex.EncodeToString(h[:])

	t.Logf("Login parameters: timestamp=%d, requestId=%s", nowSec, requestID)
	t.Logf("Login clientHash: %s", clientHash)

	// === 3. Backend verification ===
	err = VerifyPassword(storedEncrypted, clientHash, requestID)
	if err != nil {
		t.Fatalf("❌ VerifyPassword failed: %v", err)
	}
	t.Log("✅ Login verification passed!")
}

// TestLoginWithExpiredTimestamp uses the original log data you provided
// Expected result: should return ErrPasswordMismatch (timestamp expired beyond ±5s)
func TestLoginWithExpiredTimestamp(t *testing.T) {
	// Generate stored ciphertext using the registration H1
	registerH1Hex := "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92"
	storedEncrypted, err := HashForStorage(registerH1Hex)
	if err != nil {
		t.Fatalf("HashForStorage failed: %v", err)
	}

	// Use the fixed clientHash from your original login log
	fixedClientHash := "148d51da422cd1f6598bd0f5252428ea7bab7f1396fc432ffd3729cec65d257a"
	requestID := "1999-aa71-6be8"

	err = VerifyPassword(storedEncrypted, fixedClientHash, requestID)
	if err == nil {
		t.Fatal("⚠️ Expected verification to fail due to expired timestamp, but it passed! Please check if server time is abnormal.")
	}

	// Verify it's the expected error type
	if errors.Is(err, ErrPasswordMismatch) {
		t.Logf("✅ Expected: expired clientHash correctly rejected: %v", err)
	} else {
		t.Fatalf("❌ Unexpected error type: %v", err)
	}
}

// TestDecryptIntegrity independently verifies encryption/decryption integrity
func TestDecryptIntegrity(t *testing.T) {
	originalH1 := "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92"

	encrypted, err := HashForStorage(originalH1)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	decryptedBytes, err := decryptH1(encrypted)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	decryptedHex := hex.EncodeToString(decryptedBytes)
	if decryptedHex != originalH1 {
		t.Fatalf("❌ Decryption result mismatch!\nExpected: %s\nActual: %s", originalH1, decryptedHex)
	}
	t.Log("✅ AES-GCM encryption/decryption integrity verification passed")
}

// TestWrongRequestId verifies that mismatched requestId is rejected
func TestWrongRequestId(t *testing.T) {
	registerH1Hex := "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92"
	storedEncrypted, _ := HashForStorage(registerH1Hex)

	nowSec := time.Now().Unix()
	// Generate hash with correct requestId
	plain := fmt.Sprintf("%s:%d:%s", registerH1Hex, nowSec, "correct-id")
	h := sha256.Sum256([]byte(plain))
	clientHash := hex.EncodeToString(h[:])

	// But verify with wrong requestId
	err := VerifyPassword(storedEncrypted, clientHash, "wrong-id")
	if !errors.Is(err, ErrPasswordMismatch) {
		t.Fatalf("❌ Expected ErrPasswordMismatch, got: %v", err)
	}
	t.Log("✅ requestId tampering detection passed")
}
