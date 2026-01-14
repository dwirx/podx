# Advanced Encryption Features Design

**Date:** 2026-01-14
**Status:** Implementation In Progress

## Overview

This document describes the design for advanced encryption features in PODX v2, including:
- Dual encryption modes (Normal and Paranoid)
- XChaCha20-Poly1305 with Serpent cascade
- HKDF-SHA3-256 subkey derivation
- Keyed-BLAKE2b and HMAC-SHA3 MACs
- Configurable Argon2id parameters
- Streaming encryption for large files
- Any-file encryption support

## Encryption Modes

### Normal Mode
| Component | Implementation |
|-----------|----------------|
| Cipher | AES-256-GCM or XChaCha20-Poly1305 (user choice) |
| MAC | Keyed-BLAKE2b (256-bit key, 256-bit output) |
| KDF | Argon2id (4 passes, 256MB default, 4 threads) |

### Paranoid Mode
| Component | Implementation |
|-----------|----------------|
| Cipher | XChaCha20-Poly1305 → Serpent (cascade) |
| MAC | HMAC-SHA3-512 (256-bit key, 512-bit output) |
| KDF | Argon2id (8 passes, 512MB default, 8 threads) |

## Subkey Derivation (HKDF-SHA3-256)

```
Master Key (from Argon2id)
    ├── Header HMAC key (64 bytes) - for v2 header authentication
    ├── Payload MAC key (32 bytes) - BLAKE2b or HMAC-SHA3
    ├── Cipher key (32 bytes) - AES/XChaCha20
    └── Serpent key (32 bytes) - paranoid mode only
```

## File Format v2

```
┌─────────────────────────────────────────────────────────────┐
│ Magic (4 bytes): "PDX2"                                     │
├─────────────────────────────────────────────────────────────┤
│ Version (1 byte): 0x02                                      │
├─────────────────────────────────────────────────────────────┤
│ Flags (1 byte):                                             │
│   bit 0: mode (0=normal, 1=paranoid)                        │
│   bit 1-2: cipher (00=aes-gcm, 01=xchacha20, 10=cascade)    │
│   bit 3: streaming (0=binary, 1=streaming)                  │
├─────────────────────────────────────────────────────────────┤
│ Salt (16 bytes): Argon2id salt                              │
├─────────────────────────────────────────────────────────────┤
│ Header HMAC (64 bytes): HMAC of header fields               │
├─────────────────────────────────────────────────────────────┤
│ Payload: encrypted data with nonce prepended                │
├─────────────────────────────────────────────────────────────┤
│ Payload MAC (32 or 64 bytes): BLAKE2b or HMAC-SHA3          │
└─────────────────────────────────────────────────────────────┘
```

## Backward Compatibility

- v1 format detection: files without "PDX2" magic are treated as v1
- v1 files can be read but new files are always written as v2
- No automatic migration; user can manually re-encrypt

## Configuration

### CLI Flags
```bash
podx encrypt --mode paranoid --cipher xchacha20 file.txt
podx encrypt-all --mode normal
```

### .podx.yaml
```yaml
encryption:
  mode: normal        # normal | paranoid
  cipher: aes-gcm     # aes-gcm | xchacha20
  memory_mb: 256      # Argon2id memory (adaptive)
```

### Memory Adaptation
- Default: 256MB (normal), 512MB (paranoid)
- Adaptive: detect available RAM, reduce if < 2x required
- Minimum: 64MB with warning

## Streaming Encryption

For files > 100MB:
- Encrypt in 1MB chunks
- Each chunk has its own nonce (derived from chunk index)
- Progress bar in CLI and TUI
- Rekeying every 1GB (paranoid mode)

## Implementation Order

1. HKDF-SHA3-256 subkey derivation
2. XChaCha20-Poly1305 cipher
3. Serpent cipher
4. Keyed-BLAKE2b MAC
5. HMAC-SHA3-512 MAC
6. Argon2id adaptive parameters
7. Encryption modes integration
8. Streaming encryption
9. File format v2
10. CLI updates
11. TUI updates
12. Tests and documentation
