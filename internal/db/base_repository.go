package db

type BaseRepository struct {
	db     DB
	bucket string
}

func NewBaseRepository(db DB, bucket string) *BaseRepository {
	return &BaseRepository{
		db:     db,
		bucket: bucket,
	}
}

func (r *BaseRepository) Put(key string, value []byte) error {
	return r.db.Put(r.bucket, key, value)
}

func (r *BaseRepository) Get(key string) ([]byte, error) {
	return r.db.Get(r.bucket, key)
}

func (r *BaseRepository) Delete(key string) error {
	return r.db.Delete(r.bucket, key)
}

func (r *BaseRepository) ListKeys() ([]string, error) {
	return r.db.ListKeys(r.bucket)
}
