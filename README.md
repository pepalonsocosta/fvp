# Fast VPN Protocol (FVP)

A minimal VPN solution built with Go. Fast and easy to use.

## Why FVP?

FVP enables secure, transparent access to your local network services. Run the server on your home network and connect remotely to access services like Home Assistant, local web apps, or file servers—without exposing HTTP ports or relying on SSH. All traffic is encrypted with ChaCha20-Poly1305, providing a secure tunnel that makes remote services appear as if they're on your local network.

## Download

Download pre-built binaries from the [Releases](https://github.com/pepalonsocosta/fvp/releases) page:

- **Server**: `fvp-server-linux-amd64.tar.gz` or `fvp-server-linux-arm64.tar.gz`
- **Client**: `fvp-client-linux-amd64.tar.gz` or `fvp-client-linux-arm64.tar.gz`

Extract and place the binary in your `PATH`, or run directly:

```bash
# Linux
tar -xzf fvp-server-linux-amd64.tar.gz
sudo mv fvps /usr/local/bin/

tar -xzf fvp-client-linux-amd64.tar.gz
sudo mv fvpc /usr/local/bin/
```

## Quick Start

### Server Setup

1. **Initialize configuration:**

```bash
fvps setup --port 1194 --timeout 30
```

2. **Add a client:**

```bash
fvps add-client
```

This generates a client ID and key. Save the key - you'll need it on the client side.

3. **Start the server:**

```bash
sudo fvps up
```

### Client Connection

1. **Connect to server:**

```bash
fvpc connect --server <server-ip>:1194
```

2. **Disconnect:**

```bash
fvpc disconnect
```

Or press `Ctrl+C` while connected.

## Server Commands

| Command                                        | Description                             |
| ---------------------------------------------- | --------------------------------------- |
| `fvps setup --port <port> --timeout <minutes>` | Create initial server configuration     |
| `fvps up`                                      | Start the VPN server                    |
| `fvps status`                                  | Show server status and statistics       |
| `fvps add-client`                              | Add a new client and generate a key     |
| `fvps list-clients`                            | List all clients with connection status |
| `fvps remove-client --id <id>`                 | Remove a client from configuration      |

## Client Commands

| Command                             | Description                           |
| ----------------------------------- | ------------------------------------- |
| `fvpc connect --server <ip>:<port>` | Connect to a VPN server               |
| `fvpc disconnect`                   | Disconnect from the VPN server        |
| `fvpc status`                       | Show connection status and statistics |
| `fvpc version`                      | Show version information              |
| `fvpc help`                         | Show help message                     |

## Requirements

- **Server**: Linux with root privileges (for TUN interface)
- **Client**: Linux (macOS and Windows support coming soon)
- Pre-shared keys for authentication

## License

MIT License - see [LICENSE](LICENSE) file for details.
