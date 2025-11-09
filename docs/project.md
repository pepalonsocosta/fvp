# Fast VPN Project Structure

## Overview

A minimal, high-performance VPN server built with Go using onion-layer architecture. The project implements a complete VPN server with ChaCha20-Poly1305 encryption, TUN interface management, and UDP-based client communication.

## Project Structure

```
vpnTest/
├── cmd/
│   └── server/              # Main server executable
├── internal/
│   ├── protocol/            # Packet definitions and parsing
│   ├── crypto/              # Encryption/decryption logic
│   ├── network/             # TUN interface and routing
│   └── server/              # Server core logic
├── docs/
│   ├── protocol.md          # Protocol specification
│   └── project.md           # This file
├── server.example.yaml      # Server configuration example
├── go.mod
└── README.md
```

## Internal Packages

### `internal/protocol/` - Layer 1: Core Packet Handling

- **Packet definitions**: FVP packet structure with 12-byte header
- **Packet parsing**: Encode/decode functions for binary packet format
- **Packet validation**: Magic number, type, and length validation
- **Protocol constants**: Packet types (Data, Auth, Ping, Pong)

### `internal/crypto/` - Layer 2: Encryption

- **ChaCha20-Poly1305 encryption**: Authenticated encryption for payloads
- **Pre-shared key management**: YAML-based client key configuration
- **Key validation**: 32-byte key length enforcement
- **Nonce generation**: Sequence-based nonce creation for anti-replay

### `internal/network/` - Layer 3: TUN Interface

- **TUN device creation**: System command-based TUN interface setup
- **IP packet capture**: Raw IP packet reading/writing
- **Interface management**: TUN interface lifecycle and configuration
- **Mock implementation**: Testing support without real TUN devices

### `internal/server/` - Layer 4: Server Logic

- **Client management**: Dynamic IP assignment, timeout handling, sequence validation
- **Packet processing**: Incoming/outgoing packet routing and encryption
- **Connection handling**: UDP server with client authentication
- **Response generation**: Auth and Pong response packet creation
- **Modular architecture**: Separated into config, handlers, responses, and routing

## Versioning Strategy

### Build-Time Version Injection

The project uses industry-standard build-time version injection with `-ldflags`:

```bash
# Build with version
go build -ldflags "-X main.version=1.1.1" -o ultrafast-vpn ./cmd/server

# Or using Makefile
make build VERSION=1.2.3
```

### Version Components

- **Major Version**: Protocol compatibility (e.g., 1, 2, 3)
- **Minor Version**: Feature additions (e.g., 0, 1, 2)
- **Patch Version**: Bug fixes (e.g., 0, 1, 2)

### Version Flow

1. **Build Time**: Version injected via `-ldflags` as "major.minor.patch" string
2. **Runtime**: Version parsed and set in protocol constants
3. **Protocol**: Major version used in packet headers for compatibility
4. **Compatibility**: Clients must match major version

### Version Constants

```go
// Default fallback versions
const (
    ProtocolVersionMajor = 1
    ProtocolVersionMinor = 0
    ProtocolVersionPatch = 0
)

// Build-time injected version
var version = "dev" // Set via -ldflags as "1.1.1" format
```

### Version Parsing

The `SetVersion()` function parses the version string and updates all three components:

- Input: "1.1.1" → ProtocolVersionMajor=1, ProtocolVersionMinor=1, ProtocolVersionPatch=1
- Input: "2.0.5" → ProtocolVersionMajor=2, ProtocolVersionMinor=0, ProtocolVersionPatch=5
- Input: "dev" or empty → Uses default values (1.0.0)

## Implementation Status

### ✅ Completed Features

- **Protocol Layer**: Complete packet parsing and validation
- **Crypto Layer**: ChaCha20-Poly1305 encryption with pre-shared keys
- **Network Layer**: TUN interface management with system commands
- **Server Layer**: Complete VPN server with client management
- **Authentication**: Client key-based authentication with dynamic IP assignment
- **Packet Processing**: Bidirectional packet routing between clients and TUN
- **Keep-Alive**: Ping/Pong mechanism for connection monitoring
- **Error Handling**: Comprehensive error handling with custom error types
- **Testing**: Unit tests for all major components
- **Modular Design**: Clean separation of concerns across multiple files

### 🔧 Technical Specifications

- **Encryption**: ChaCha20-Poly1305 with 32-byte keys
- **Transport**: UDP on port 1194 (configurable)
- **Interface**: TUN interface named "fvp0"
- **Client Limit**: 256 concurrent clients (ClientID 1-255)
- **IP Range**: 10.0.0.2 to 10.0.0.255 for client assignments
- **Timeout**: 30-minute client inactivity timeout
- **Sequence**: 32-bit sequence numbers for anti-replay protection
- **Configuration**: YAML-based client key management

### 📁 Server Architecture

The server is organized into focused modules:

- **`server.go`**: Core lifecycle (Start/Stop) and main orchestration
- **`server_config.go`**: Configuration loading and component creation
- **`server_handlers.go`**: High-level packet handling (Auth, Data, Ping, Pong)
- **`server_responses.go`**: Response packet creation and sending
- **`server_routing.go`**: Outgoing packet routing from TUN to clients
- **`client_manager.go`**: Client state management and IP assignment
- **`packet_processor.go`**: Low-level packet processing and encryption

## Development Approach

- **Onion layers**: Each layer builds on the previous
- **Test-driven**: Unit tests for each layer with mock interfaces
- **Server-first**: Complete server implementation with client simulation
- **Performance-optimized**: Minimal overhead design with efficient packet processing
- **Modular design**: Clean separation of concerns for maintainability
