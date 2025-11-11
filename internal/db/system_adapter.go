package db

type SystemRepoAdapter struct {
	repo Repository
}

func NewSystemRepoAdapter(r Repository) *SystemRepoAdapter {
	return &SystemRepoAdapter{repo: r}
}

func (a *SystemRepoAdapter) Get(key string) ([]byte, error) {
	return a.repo.Get(key)
}

func (a *SystemRepoAdapter) Put(key string, value []byte) error {
	return a.repo.Put(key, value)
}
