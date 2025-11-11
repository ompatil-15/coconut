package db

import "github.com/ompatil-15/coconut/internal/db/model"

type Repository interface {
	Put(key string, value []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	ListKeys() ([]string, error)
}

type SecretRepository interface {
	Add(secret model.Secret) (string, error)
	Get(key string) (*model.Secret, error)
	Update(secret model.Secret) error
	Delete(key string) error
	List() ([]model.Secret, error)
}
