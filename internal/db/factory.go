package db

import (
	"fmt"

	"github.com/ompatil-15/coconut/internal/vault"
)

type RepositoryFactory struct {
	db    DB
	vault *vault.Vault
}

func NewRepositoryFactory(db DB, v *vault.Vault, buckets ...string) *RepositoryFactory {
	for _, bucket := range buckets {
		if err := db.CreateBucket(bucket); err != nil {
			panic(fmt.Sprintf("failed to create bucket %q: %v", bucket, err))
		}
	}

	return &RepositoryFactory{
		db:    db,
		vault: v,
	}
}

func (f *RepositoryFactory) NewBaseRepository(bucket string) *BaseRepository {
	return &BaseRepository{
		db:     f.db,
		bucket: bucket,
	}
}

func (f *RepositoryFactory) NewEncryptedRepository(bucket string) SecretRepository {
	base := &BaseRepository{
		db:     f.db,
		bucket: bucket,
	}

	return &EncryptedRepository{
		repo:  base,
		vault: f.vault,
	}
}
