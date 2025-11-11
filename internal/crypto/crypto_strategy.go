package crypto

type CryptoStrategy interface {
	Encrypt(key []byte, plaintext string) (string, error)
	Decrypt(key []byte, ciphertext string) (string, error)
}
