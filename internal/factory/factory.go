package factory

import (
	"fmt"

	"github.com/ompatil-15/coconut/internal/config"
	"github.com/ompatil-15/coconut/internal/crypto"
	"github.com/ompatil-15/coconut/internal/db"
	"github.com/ompatil-15/coconut/internal/db/boltdb"
	"github.com/ompatil-15/coconut/internal/iostreams"
	"github.com/ompatil-15/coconut/internal/logger"
	"github.com/ompatil-15/coconut/internal/session"
	"github.com/ompatil-15/coconut/internal/vault"
)

type Factory struct {
	IO      *iostreams.IOStreams
	Logger  *logger.Logger
	Config  *config.Config
	DB      db.DB
	Vault   *vault.Vault
	Crypto  crypto.CryptoStrategy
	Repo    *db.RepositoryFactory
	System  db.Repository
	Secrets db.SecretRepository
	Session *session.Manager
}

func New() (*Factory, error) {
	io := iostreams.System()
	log, err := logger.New()
	if err != nil {
		return nil, fmt.Errorf("logger init: %w", err)
	}

	cfg := config.Default()

	bdb, err := boltdb.NewBoltStore(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("db open: %w", err)
	}

	repoFactory := db.NewRepositoryFactory(bdb, nil, cfg.SystemBucket, cfg.SecretsBucket)

	systemRepo := repoFactory.NewBaseRepository(cfg.SystemBucket)

	cfg, err = config.Load(systemRepo)
	if err != nil {
		return nil, fmt.Errorf("config load: %w", err)
	}

	strategy := crypto.NewAESGCM()
	v := vault.NewVault(strategy, nil)

	repoFactory.SetVault(v)

	secretRepo := repoFactory.NewEncryptedRepository(cfg.SecretsBucket)

	sessionRepo := systemRepo
	sessionMgr := session.NewManager(sessionRepo, cfg)

	return &Factory{
		IO:      io,
		Logger:  log,
		Config:  cfg,
		DB:      bdb,
		Vault:   v,
		Crypto:  strategy,
		Repo:    repoFactory,
		System:  systemRepo,
		Secrets: secretRepo,
		Session: sessionMgr,
	}, nil
}

func (f *Factory) Close() {
	if f.Logger != nil {
		f.Logger.Close()
	}
	if f.DB != nil {
		_ = f.DB.Close()
	}
}
