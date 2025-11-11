package crypto

import (
	"crypto/rand"
	"log"

	"golang.org/x/crypto/argon2"
)

func GenerateRandomSalt(size int) []byte {
	s := make([]byte, size)
	if _, err := rand.Read(s); err != nil {
		log.Fatalf("failed to generate random salt: %v", err)
	}
	return s
}

func DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
}
