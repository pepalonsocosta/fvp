# Fast VPN Protocol (FVP)

## Overview

A minimal, high-performance VPN protocol designed for maximum speed with essential security. This protocol implements a complete VPN solution with ChaCha20-Poly1305 encryption, UDP transport, and TUN interface integration.

## Packet Format

### Header (12 bytes)

```
Byte 0-2:   Magic "FVP"           - Protocol identifier
Byte 3:     Type                  - Packet type (1-4)
Byte 4:     ClientID              - Client identifier (0-255)
Byte 5-8:   Sequence              - Sequence number (0-4,294,967,295)
Byte 9-10:  Length                - Payload size in bytes
Byte 11:    Version               - Protocol version (currently 1)
Byte 12+:   Payload               - Encrypted data
```

### Packet Types

- `1` - Data: Encrypted IP packet
- `2` - Auth: Authentication request
- `3` - Ping: Keep-alive request
- `4` - Pong: Keep-alive response

## Security

- **Encryption**: ChaCha20-Poly1305 with authenticated encryption
- **Keys**: Pre-shared 32-byte keys per client (hex-encoded in config)
- **Anti-replay**: Strict sequential sequence numbers with validation
- **Authentication**: Client key-based authentication with dynamic IP assignment
- **Nonce**: Sequence number + 8 zero bytes for 12-byte nonce

## Client Limits

- Maximum 256 concurrent clients (ClientID 1-255)
- 32-bit sequence numbers per client (0-4,294,967,295)
- Pre-shared key authentication via YAML configuration
- Dynamic IP assignment (10.0.0.2 to 10.0.0.255)
- 30-minute inactivity timeout

## Protocol Flow

### Authentication

```
Client → Server: Auth packet (ClientID, empty payload)
Server: Validates client key from configuration
Server: Assigns dynamic IP (10.0.0.x)
Server → Client: Auth packet (success confirmation)
```

### Data Transfer

```
Client → Server: Data packet (encrypted IP payload, sequential sequence)
Server: Validates sequence (must be lastSeq + 1)
Server: Decrypts payload and writes to TUN interface
Server: Reads from TUN, encrypts, sends to client
```

### Keep-Alive

```
Client → Server: Ping packet (every 30 seconds)
Server → Client: Pong packet (immediate response)
Server: Updates client LastSeen timestamp
Timeout: 30 minutes without activity = disconnect
```

### Packet Processing Pipeline

```
1. Client sends encrypted FVP packet via UDP
2. Server decodes packet header and validates
3. Server decrypts payload using client key
4. Server writes decrypted IP packet to TUN interface
5. TUN interface routes packet to internet
6. Response comes back through TUN interface
7. Server reads IP packet from TUN
8. Server determines target client by destination IP
9. Server encrypts IP packet with client key
10. Server sends encrypted FVP packet to client
```

## Example Packet

```
FVP 1 5 1000 1500 1 [encrypted_payload]
│   │ │ │    │    │
│   │ │ │    │    └─ Version: 1
│   │ │ │    └────── Length: 1500 bytes
│   │ │ └─────────── Sequence: 1000
│   │ └───────────── ClientID: 5
│   └─────────────── Type: Data
└─────────────────── Magic: "FVP"
```

## Implementation Details

### Server Architecture

The FVP server is implemented as a modular Go application with the following components:

- **Protocol Layer**: Packet parsing and validation
- **Crypto Layer**: ChaCha20-Poly1305 encryption with pre-shared keys
- **Network Layer**: TUN interface management and IP packet handling
- **Server Layer**: Client management, packet processing, and UDP communication

### Configuration

Client keys are managed via YAML configuration:

```yaml
clients:
  - id: 1
    key: "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"
  - id: 2
    key: "fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"
```

### Error Handling

The protocol includes comprehensive error handling:

- Invalid packet format
- Authentication failures
- Sequence number validation
- Client timeout management
- Encryption/decryption errors

### Testing

Complete unit test coverage for all components using mock interfaces for TUN device simulation.
