package db_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ompatil-15/coconut/internal/crypto"
	"github.com/ompatil-15/coconut/internal/db"
	"github.com/ompatil-15/coconut/internal/db/boltdb"
	"github.com/ompatil-15/coconut/internal/db/model"
	"github.com/ompatil-15/coconut/internal/vault"
)

func BenchmarkEncryptedRepository_Add(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "coconut-bench")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "bench.db")
	boltStore, err := boltdb.NewBoltStore(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer boltStore.Close()

	if err := boltStore.CreateBucket("secrets"); err != nil {
		b.Fatal(err)
	}

	baseRepo := db.NewBaseRepository(boltStore, "secrets")

	// Setup Vault
	v := vault.NewVault(crypto.NewAESGCM(), []byte("salt"))
	// 32 byte key for AES-256
	key := make([]byte, 32)
	v.Unlock(key)

	repo := db.NewEncryptedRepository(baseRepo, v, "secrets")

	secret := model.Secret{
		Username:    "benchuser",
		Password:    "benchpassword123!",
		URL:         "https://benchmark.com",
		Description: "A secret for benchmarking",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		secret.ID = fmt.Sprintf("bench-%d", i)
		_, err := repo.Add(secret)
		if err != nil {
			b.Fatalf("Add failed: %v", err)
		}
	}
}

func BenchmarkEncryptedRepository_Get(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "coconut-bench-get")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "bench_get.db")
	boltStore, err := boltdb.NewBoltStore(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer boltStore.Close()

	if err := boltStore.CreateBucket("secrets"); err != nil {
		b.Fatal(err)
	}

	baseRepo := db.NewBaseRepository(boltStore, "secrets")
	v := vault.NewVault(crypto.NewAESGCM(), []byte("salt"))
	key := make([]byte, 32)
	v.Unlock(key)
	repo := db.NewEncryptedRepository(baseRepo, v, "secrets")

	// Pre-populate
	numSecrets := 1000
	ids := make([]string, numSecrets)
	for i := 0; i < numSecrets; i++ {
		secret := model.Secret{
			ID:        fmt.Sprintf("bench-%d", i),
			Username:  "user",
			Password:  "pass",
			CreatedAt: time.Now(),
		}
		ids[i], _ = repo.Add(secret)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.Get(ids[i%numSecrets])
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}
