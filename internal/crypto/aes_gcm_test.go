package crypto

import (
	"crypto/rand"
	"testing"
)

func TestAESGCM_EncryptDecrypt(t *testing.T) {
	aes := NewAESGCM()
	key := make([]byte, 32) // 256-bit key
	rand.Read(key)

	tests := []struct {
		name      string
		plaintext string
	}{
		{"empty string", ""},
		{"simple text", "hello world"},
		{"special chars", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"unicode", "🔐 password with émojis and ñ"},
		{"long text", "This is a very long password that contains multiple words and should test the encryption with larger data"},
		{"json", `{"username":"test","password":"secret123"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := aes.Encrypt(key, tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			if ciphertext == "" && tt.plaintext != "" {
				t.Error("Ciphertext should not be empty for non-empty plaintext")
			}

			decrypted, err := aes.Decrypt(key, ciphertext)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("Expected '%s', got '%s'", tt.plaintext, decrypted)
			}
		})
	}
}

func TestAESGCM_DifferentKeys(t *testing.T) {
	aes := NewAESGCM()
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	rand.Read(key1)
	rand.Read(key2)

	plaintext := "secret message"

	ciphertext, err := aes.Encrypt(key1, plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Try to decrypt with wrong key
	_, err = aes.Decrypt(key2, ciphertext)
	if err == nil {
		t.Error("Decrypt should fail with wrong key")
	}
}

func TestAESGCM_InvalidInput(t *testing.T) {
	aes := NewAESGCM()

	// Test with invalid key size
	invalidKey := make([]byte, 15) // Invalid key size
	_, err := aes.Encrypt(invalidKey, "test")
	if err == nil {
		t.Error("Encrypt should fail with invalid key size")
	}

	// Test decrypt with invalid base64
	validKey := make([]byte, 32)
	rand.Read(validKey)
	_, err = aes.Decrypt(validKey, "invalid-base64!")
	if err == nil {
		t.Error("Decrypt should fail with invalid base64")
	}

	// Test decrypt with malformed ciphertext
	_, err = aes.Decrypt(validKey, "dGVzdA") // "test" in base64, too short
	if err == nil {
		t.Error("Decrypt should fail with malformed ciphertext")
	}
}

func TestAESGCM_Randomness(t *testing.T) {
	aes := NewAESGCM()
	key := make([]byte, 32)
	rand.Read(key)

	plaintext := "same message"
	
	// Encrypt same message multiple times
	ciphertext1, err := aes.Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("First encrypt failed: %v", err)
	}

	ciphertext2, err := aes.Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Second encrypt failed: %v", err)
	}

	// Ciphertexts should be different due to random nonce
	if ciphertext1 == ciphertext2 {
		t.Error("Ciphertexts should be different for same plaintext (nonce randomness)")
	}

	// But both should decrypt to same plaintext
	decrypted1, _ := aes.Decrypt(key, ciphertext1)
	decrypted2, _ := aes.Decrypt(key, ciphertext2)

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Error("Both ciphertexts should decrypt to original plaintext")
	}
}