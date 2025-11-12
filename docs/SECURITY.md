# Security Architecture

This document details Coconut's security implementation, threat model, and cryptographic choices.

## Zero Knowledge Architecture

Coconut implements true Zero Knowledge Architecture where your master password never leaves your device and is never stored anywhere.

### How It Works

1. **Initialization**
   - User creates master password
   - Random 16-byte salt generated
   - Key derived using Argon2id
   - Verification token encrypted and stored
   - Only salt stored in database

2. **Unlocking**
   - User enters master password
   - Key derived from password + stored salt
   - Verification token decrypted to validate password
   - Key held in memory for session

3. **Locking**
   - Encryption key zeroed from memory
   - Session terminated
   - All secrets inaccessible until next unlock

## Cryptographic Details

### Encryption Algorithm

**AES-256-GCM (Galois/Counter Mode)**
- 256-bit key size
- Authenticated encryption (prevents tampering)
- 12-byte random nonce per operation
- Industry-standard, NIST-approved

### Key Derivation Function

**Argon2id** - Winner of the Password Hashing Competition
- **Time cost:** 3 iterations
- **Memory cost:** 64 MB
- **Parallelism:** 4 threads
- **Output:** 32 bytes (256-bit key)

**Why Argon2id?**
- Memory-hard algorithm (resistant to GPU/ASIC attacks)
- Hybrid approach (combines data-dependent and data-independent memory access)
- Superior to PBKDF2 used by most password managers
- Configurable parameters for future-proofing

### Random Number Generation

- Uses `crypto/rand` (cryptographically secure)
- 16-byte salt per vault
- 12-byte nonce per encryption operation

## Brute Force Resistance

### Attack Scenario Analysis

**Assumptions:**
- Attacker has full database access
- 100-core modern CPU
- Argon2id parameters: 64 MB memory, 3 iterations, 4 threads
- Attack speed: ~100 attempts/second per core (memory-hard constraint)

**Total attack speed:** 10,000 passwords/second

### Time to Crack

Using Coconut's character set (89 characters: a-z, A-Z, 0-9, special):

| Password Length | Entropy | Combinations | Time to Crack |
|----------------|---------|--------------|---------------|
| 8 chars | 52 bits | 3.9×10¹⁵ | 12 years |
| 12 chars | 78 bits | 7.5×10²³ | 2.4 billion years |
| 16 chars (default) | 105 bits | 3.9×10³¹ | 1.2×10¹⁸ years |
| 20 chars | 131 bits | 2.0×10³⁹ | 6.3×10²⁵ years |

**Recommendation:** Use 12+ character passwords with mixed case, numbers, and symbols.

### Why Argon2id Matters

Compared to PBKDF2:
- **10x slower** for attackers (memory-hard vs compute-hard)
- **GPU resistance** - Requires 64 MB per attempt (expensive to parallelize)
- **ASIC resistance** - Memory requirements make custom hardware impractical

## Industry Comparison

| Feature | Coconut | 1Password | Bitwarden | LastPass |
|---------|---------|-----------|-----------|----------|
| Key Derivation | Argon2id | PBKDF2 (650K) | PBKDF2 (600K) | PBKDF2 (100K) |
| Encryption | AES-256-GCM | AES-256-GCM | AES-256-CBC | AES-256-CBC |
| Zero Knowledge | Yes | Yes | Yes | Yes |
| Local-first | Yes | Cloud | Optional | Cloud |
| Open Source | Yes | No | Yes | No |

**Coconut's advantage:** Argon2id provides superior protection against brute force attacks compared to PBKDF2.

## Session Management

### autoLockSecs Setting

Controls session timeout behavior:

- **Default:** 300 seconds (5 minutes)
- Session stays active for specified duration
- Encrypted key cached in database during session
- Auto-locks after inactivity timeout
- **Trade-off:** Active sessions increase attack surface

**Note:** Lower values provide better security at the cost of more frequent password prompts.

## Threat Model

### What Coconut Protects Against

- **Full database theft** - Encrypted data useless without master password
- **Brute force attacks** - Argon2id makes attacks computationally infeasible
- **Memory dumps (when locked)** - Keys zeroed from memory
- **Tampering** - AES-GCM authenticated encryption detects modifications

### What Coconut Does NOT Protect Against

- **Keyloggers** - Can capture master password when entered
- **Memory dumps (when unlocked)** - Key present in memory during session
- **Malicious code execution** - Attacker with code execution can extract keys
- **Weak master passwords** - User responsibility to choose strong passwords
- **Physical access attacks** - Cold boot attacks, hardware keyloggers, etc.  

## Best Practices

1. **Strong Master Password**
   - Minimum 12 characters
   - Mix uppercase, lowercase, numbers, symbols
   - Avoid dictionary words
   - Don't reuse from other services

2. **Secure Your Device**
   - Full disk encryption
   - Screen lock when away
   - Keep OS and software updated

3. **Lock When Not in Use**
   - Run `coconut lock` when done
   - Keys removed from memory
   - Prevents memory dump attacks

4. **Backup Your Database**
   - Copy `~/.coconut/coconut.db` to secure location
   - Encrypted backup still requires master password
   - Test restore procedure

5. **Monitor Logs**
   - Check `~/.coconut/logs/coconut.log` for suspicious activity
   - Look for unexpected unlock attempts

## Security Audits

Coconut is open source and welcomes security reviews. If you discover a vulnerability:

1. **Do not** open a public issue
2. Email: patilom001@gmail.com with details
3. Allow reasonable time for fix before disclosure




