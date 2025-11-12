# Architecture

Coconut is built with extensibility and maintainability in mind. The architecture allows swapping concrete implementations without changing core logic.

## Design Principles

1. **Separation of Concerns** - Each layer has a single responsibility
2. **Dependency Injection** - Components receive dependencies, not create them
3. **Interface-based Design** - Program to interfaces, not implementations
4. **Strategy Pattern** - Swap algorithms at runtime

## System Architecture

![Architecture Diagram](architecture.svg)

*See [architecture.puml](architecture.puml) for PlantUML source*

## Core Components

### 1. CLI Layer

**Responsibility:** User interaction and command handling

- Built with [Cobra](https://github.com/spf13/cobra)
- Each command in separate file
- Minimal business logic (delegates to lower layers)
- Input validation and error formatting

### 2. Factory Pattern

**Responsibility:** Dependency injection and component lifecycle

```go
type Factory struct {
    DB      *db.DB
    System  db.Repository
    Secrets db.Repository
    Session *session.Manager
    Logger  *logger.Logger
    IO      *iostreams.IOStreams
}
```

**Benefits:**
- Centralized initialization
- Easy testing (inject mocks)
- Single source of truth for dependencies
- Proper cleanup on shutdown

### 3. Vault Layer

**Responsibility:** Core password management logic

```go
type Vault struct {
    strategy crypto.CryptoStrategy
    key      []byte
    salt     []byte
    unlocked bool
}
```

**Key Operations:**
- `Unlock(key)` - Start session with derived key
- `Lock()` - End session, zero memory
- `Encrypt(plaintext)` - Encrypt data
- `Decrypt(ciphertext)` - Decrypt data
- `VerifyPassword(token)` - Validate master password

**Design Decision:** Vault doesn't know about password derivation or storage - it only handles encryption/decryption with a provided key.

### 4. Crypto Layer

**Responsibility:** Encryption algorithms and key derivation

**Strategy Pattern:**
```go
type CryptoStrategy interface {
    Encrypt(key []byte, plaintext string) (string, error)
    Decrypt(key []byte, ciphertext string) (string, error)
}
```

**Current Implementation:**
- `AESGCM` - AES-256-GCM encryption

**Key Derivation:**
```go
func DeriveKey(password string, salt []byte) []byte {
    return argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
}
```

**Why Strategy Pattern?**
- Easy to add new encryption algorithms
- Swap implementations without changing vault logic
- Test with mock crypto for unit tests

### 5. Storage Layer

**Responsibility:** Data persistence

**Repository Pattern:**
```go
type Repository interface {
    Get(key string) ([]byte, error)
    Put(key string, value []byte) error
    Delete(key string) error
    ListKeys() ([]string, error)
}
```

**Implementations:**
- `BoltRepository` - BoltDB key-value store
- `EncryptedRepository` - Transparent encryption wrapper

**Database Structure:**
```
coconut.db (BoltDB)
├── system/
│   ├── salt                    # Random salt for key derivation
│   ├── vault_verification      # Encrypted verification token
│   └── config                  # JSON configuration
└── secrets/
    ├── 1                       # Encrypted secret (username:password)
    ├── 2
    └── ...
```

**Why Repository Pattern?**
- Abstract storage implementation
- Easy to swap BoltDB for SQLite, PostgreSQL, etc.
- Testable with in-memory implementation

### 6. Session Management

**Responsibility:** Track vault unlock state

```go
type Manager struct {
    repo Repository
    cfg  *Config
}
```

**Features:**
- Manages session state and cached keys
- Auto-lock after inactivity timeout
- Session validation and activity tracking

### 7. Configuration

**Responsibility:** User settings management

```go
type Config struct {
    AutoLockSecs int  // Session timeout in seconds
}
```

**Storage:** JSON in database `system` bucket

## Design Patterns Used

### 1. Strategy Pattern (Crypto)

**Problem:** Need to support multiple encryption algorithms

**Solution:** Define `CryptoStrategy` interface, implement concrete strategies

**Benefit:** Add new algorithms without changing vault code

### 2. Repository Pattern (Storage)

**Problem:** Need to abstract database operations

**Solution:** Define `Repository` interface, implement for different stores

**Benefit:** Swap storage backends, easy testing

### 3. Factory Pattern (DI)

**Problem:** Complex object creation and wiring

**Solution:** Centralized factory creates and injects dependencies

**Benefit:** Single initialization point, proper lifecycle management

### 4. Decorator Pattern (Encrypted Repository)

**Problem:** Need transparent encryption for secrets

**Solution:** Wrap base repository with encryption layer

**Benefit:** Encryption logic separate from storage logic

## Data Flow

### Adding a Secret

```
User Input
    ↓
CLI Layer
    ↓
Vault.Encrypt(plaintext)
    ↓
CryptoStrategy.Encrypt(key, plaintext)
    ↓
Repository.Put(key, ciphertext)
    ↓
Database
```

### Getting a Secret

```
User Input
    ↓
CLI Layer
    ↓
Repository.Get(key)
    ↓
Database
    ↓
Vault.Decrypt(ciphertext)
    ↓
CryptoStrategy.Decrypt(key, ciphertext)
    ↓
Clipboard
```

## Extensibility Points

### Adding a New Encryption Algorithm

1. Implement `CryptoStrategy` interface
2. Add to factory initialization
3. Update configuration to select strategy

```go
type MyNewCrypto struct{}

func (m *MyNewCrypto) Encrypt(key []byte, plaintext string) (string, error) {
    // Implementation
}

func (m *MyNewCrypto) Decrypt(key []byte, ciphertext string) (string, error) {
    // Implementation
}
```

### Adding a New Storage Backend

1. Implement `Repository` interface
2. Update factory to create new backend
3. Migration tool for existing data

```go
type PostgresRepository struct {
    conn *sql.DB
}

func (p *PostgresRepository) Get(key string) ([]byte, error) {
    // Implementation
}
// ... other methods
```

### Adding a New Key Derivation Function

1. Add function to crypto utilities
2. Update configuration to select KDF
3. Provide migration path for existing vaults

```go
func DeriveKeyScrypt(password string, salt []byte) []byte {
    key, _ := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
    return key
}
```

## Testing Strategy

### Unit Tests
- Mock `CryptoStrategy` for vault tests
- Mock `Repository` for command tests
- Test each layer independently

### Integration Tests
- Test full flow with real implementations
- Verify encryption/decryption round-trip
- Test database operations

### Security Tests
- Verify key zeroing on lock
- Test brute force resistance (timing)
- Validate encryption parameters
