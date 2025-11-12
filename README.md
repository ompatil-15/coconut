# Coconut

A secure, local-first CLI password manager with Zero Knowledge Architecture.

## Why Coconut?

Coconut is a zero-knowledge, local-first CLI password manager built for software engineers who prefer the terminal. Secure, fast, and fully under your control. No cloud, no tracking, no compromises.

**Key Features:**
- **Zero Knowledge Architecture** - Your master password never leaves your device
- **Local-first** - All data stays on your machine, no cloud sync
- **Strong encryption** - AES-256-GCM with Argon2id key derivation
- **Simple CLI** - Quick access without context switching

## Installation

```bash
git clone https://github.com/ompatil-15/coconut.git
cd coconut
make build
```

**Requirements:** Go 1.25.3 or higher

## Quick Start

```bash
# Initialize your vault
coconut init

# Add a password
coconut add -u myusername -p mypassword

# Generate a strong password
coconut generate

# List all passwords
coconut list

# Get a password
coconut get 1
```

## Usage

### Vault Management
```bash
coconut init      # Create a new vault
coconut unlock    # Start a session
coconut lock      # End session
```

### Password Management
```bash
coconut add -u <username> -p <password>     # Add password
coconut list                                # List all
coconut get <index>                         # Get password
coconut update <index> -u <user> -p <pass>  # Update
coconut delete <index>                      # Delete
```

### Utilities
```bash
coconut generate    # Generate strong password
coconut config      # View/modify settings
```

## Security

Coconut implements true Zero Knowledge Architecture:

- **Master password never stored** - Only a random salt is kept
- **Argon2id key derivation** - Memory-hard algorithm resistant to GPU attacks
- **AES-256-GCM encryption** - Industry-standard authenticated encryption
- **Memory safety** - Keys are zeroed when vault locks

**Security vs Usability:** Configure `autoLockSecs` setting for session timeout (default: 300 seconds)

**Learn more:** [Security Architecture](docs/SECURITY.md) | [Design Decisions](docs/ARCHITECTURE.md)

## Documentation

- [Security Details](docs/SECURITY.md) - Encryption, key derivation, and threat model
- [Architecture](docs/ARCHITECTURE.md) - Design patterns and extensibility
- [Development](docs/DEVELOPMENT.md) - Contributing and building

## Configuration

- **autoLockSecs = 0**: Maximum security - no session caching, password required for every operation
- **autoLockSecs > 0**: Session timeout in seconds (default: 300)
- Lower timeout values provide better security with more frequent password prompts

## Data Storage

- **Database:** `~/.coconut/coconut.db`
- **Logs:** `~/.coconut/logs/coconut.log`

## Contributing

Contributions welcome! See [DEVELOPMENT.md](docs/DEVELOPMENT.md) for setup instructions.

```bash
# Development workflow
make build      # Build and install
make logs       # View logs
make clear_db   # Clear database
make db_dump    # Dump database
```

## License

Licensed under the Apache License, Version 2.0 - see [LICENSE](LICENSE) file for details.

## Author

**Om Patil** - [patilom001@gmail.com](mailto:patilom001@gmail.com)

---

**Remember:** Your master password is the key to all your secrets. Choose wisely, and never share it!
