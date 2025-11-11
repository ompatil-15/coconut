package db

type DB interface {
	Put(bucket string, key string, value []byte) error
	Get(bucket string, key string) ([]byte, error)
	Delete(bucket string, key string) error
	ListKeys(bucket string) ([]string, error)
	CreateBucket(bucket string) error
	Close() error
}
