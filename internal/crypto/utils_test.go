package crypto

import (
	"bytes"
	"testing"
)

func TestGenerateRandomSalt(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"16 bytes", 16},
		{"32 bytes", 32},
		{"8 bytes", 8},
		{"1 byte", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			salt := GenerateRandomSalt(tt.size)
			
			if len(salt) != tt.size {
				t.Errorf("Expected salt size %d, got %d", tt.size, len(salt))
			}

			// Generate another salt and ensure they're different
			salt2 := GenerateRandomSalt(tt.size)
			if bytes.Equal(salt, salt2) {
				t.Error("Two random salts should not be identical")
			}
		})
	}
}

func TestDeriveKey(t *testing.T) {
	password := "test-password-123"
	salt := GenerateRandomSalt(16)

	key := DeriveKey(password, salt)

	// Key should be 32 bytes (256 bits)
	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}

	// Same password and salt should produce same key
	key2 := DeriveKey(password, salt)
	if !bytes.Equal(key, key2) {
		t.Error("Same password and salt should produce identical keys")
	}

	// Different password should produce different key
	key3 := DeriveKey("different-password", salt)
	if bytes.Equal(key, key3) {
		t.Error("Different passwords should produce different keys")
	}

	// Different salt should produce different key
	salt2 := GenerateRandomSalt(16)
	key4 := DeriveKey(password, salt2)
	if bytes.Equal(key, key4) {
		t.Error("Different salts should produce different keys")
	}
}

func TestDeriveKey_EmptyInputs(t *testing.T) {
	salt := GenerateRandomSalt(16)
	
	// Empty password
	key1 := DeriveKey("", salt)
	if len(key1) != 32 {
		t.Error("Key derivation should work with empty password")
	}

	// Empty salt
	key2 := DeriveKey("password", []byte{})
	if len(key2) != 32 {
		t.Error("Key derivation should work with empty salt")
	}

	// Both empty
	key3 := DeriveKey("", []byte{})
	if len(key3) != 32 {
		t.Error("Key derivation should work with both empty")
	}
}

func TestDeriveKey_Consistency(t *testing.T) {
	// Test that key derivation is deterministic
	password := "consistent-password"
	salt := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	keys := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		keys[i] = DeriveKey(password, salt)
	}

	// All keys should be identical
	for i := 1; i < len(keys); i++ {
		if !bytes.Equal(keys[0], keys[i]) {
			t.Errorf("Key derivation not consistent: iteration %d differs", i)
		}
	}
}