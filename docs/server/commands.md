# FVP Server Commands

## `fvps setup`

Creates the initial server configuration.

```bash
fvps setup --port 1194 --timeout 30
```

## `fvps up`

Starts the VPN server.

```bash
fvps up
```

## `fvps status`

Shows server status and statistics.

```bash
fvps status
```

## `fvps add-client`

Adds a new client and generates a key.

```bash
fvps add-client
```

## `fvps list-clients`

Lists all clients with connection status.

```bash
fvps list-clients
```

## `fvps remove-client`

Removes a client from the configuration.

```bash
fvps remove-client --id 2
```
