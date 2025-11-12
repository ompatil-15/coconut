package crypto

import (
	"crypto/rand"
	"testing"
)

func BenchmarkAESGCM_Encrypt(b *testing.B) {
	aes := NewAESGCM()
	key := make([]byte, 32)
	rand.Read(key)
	plaintext := "This is a test message for encryption benchmarking"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := aes.Encrypt(key, plaintext)
		if err != nil {
			b.Fatalf("Encrypt failed: %v", err)
		}
	}
}

func BenchmarkAESGCM_Decrypt(b *testing.B) {
	aes := NewAESGCM()
	key := make([]byte, 32)
	rand.Read(key)
	plaintext := "This is a test message for decryption benchmarking"

	ciphertext, err := aes.Encrypt(key, plaintext)
	if err != nil {
		b.Fatalf("Setup encrypt failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := aes.Decrypt(key, ciphertext)
		if err != nil {
			b.Fatalf("Decrypt failed: %v", err)
		}
	}
}

func BenchmarkAESGCM_EncryptDecrypt(b *testing.B) {
	aes := NewAESGCM()
	key := make([]byte, 32)
	rand.Read(key)
	plaintext := "This is a test message for full round-trip benchmarking"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ciphertext, err := aes.Encrypt(key, plaintext)
		if err != nil {
			b.Fatalf("Encrypt failed: %v", err)
		}
		_, err = aes.Decrypt(key, ciphertext)
		if err != nil {
			b.Fatalf("Decrypt failed: %v", err)
		}
	}
}

func BenchmarkDeriveKey(b *testing.B) {
	password := "test-password-for-benchmarking"
	salt := GenerateRandomSalt(16)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DeriveKey(password, salt)
	}
}

func BenchmarkGenerateRandomSalt(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateRandomSalt(16)
	}
}

// Benchmark different message sizes
func BenchmarkAESGCM_EncryptSizes(b *testing.B) {
	aes := NewAESGCM()
	key := make([]byte, 32)
	rand.Read(key)

	sizes := []struct {
		name string
		size int
	}{
		{"16B", 16},
		{"64B", 64},
		{"256B", 256},
		{"1KB", 1024},
		{"4KB", 4096},
		{"16KB", 16384},
	}

	for _, size := range sizes {
		plaintext := make([]byte, size.size)
		rand.Read(plaintext)
		plaintextStr := string(plaintext)

		b.Run(size.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := aes.Encrypt(key, plaintextStr)
				if err != nil {
					b.Fatalf("Encrypt failed: %v", err)
				}
			}
		})
	}
}

// Benchmark key derivation with different parameters
func BenchmarkDeriveKey_Passwords(b *testing.B) {
	salt := GenerateRandomSalt(16)

	passwords := []struct {
		name     string
		password string
	}{
		{"Short", "pass"},
		{"Medium", "medium-length-password"},
		{"Long", "this-is-a-very-long-password-with-many-characters-for-testing"},
		{"Unicode", "pássw🔐rd-with-ñ-and-émojis"},
	}

	for _, pwd := range passwords {
		b.Run(pwd.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = DeriveKey(pwd.password, salt)
			}
		})
	}
}

// Memory allocation benchmarks
func BenchmarkAESGCM_EncryptAllocs(b *testing.B) {
	aes := NewAESGCM()
	key := make([]byte, 32)
	rand.Read(key)
	plaintext := "test message for allocation benchmarking"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := aes.Encrypt(key, plaintext)
		if err != nil {
			b.Fatalf("Encrypt failed: %v", err)
		}
	}
}

func BenchmarkDeriveKey_Allocs(b *testing.B) {
	password := "test-password"
	salt := GenerateRandomSalt(16)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DeriveKey(password, salt)
	}
}