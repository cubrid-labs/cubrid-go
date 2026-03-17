# AGENT.md — cubrid-go development guide

## Overview

cubrid-go is a pure-Go `database/sql` driver for CUBRID databases.
It communicates directly with the CUBRID CAS (Client Application Server) broker
over TCP using a custom binary protocol ported from [pycubrid](https://github.com/cubrid-labs/pycubrid).

## Build and Test

```bash
# Build
go build ./...

# Unit tests (no server needed)
go test ./...

# Unit tests with verbose output
go test -v ./...

# Integration tests (requires CUBRID server)
CUBRID_DSN=cubrid://dba:@localhost:33000/demodb go test -tags integration ./...
```

## Architecture

### Request/Response Flow

```
db.Query("SELECT ...")
  → conn.Prepare()    → writePrepare()    → FC=2  → parsePrepare()
  → stmt.Query()      → writeExecute()   → FC=3  → parseExecute()
  → rows.Next()       → writeFetch()     → FC=8  → parseFetch()   (if more rows)
  → stmt.Close()      → writeCloseReqHandle() → FC=6
```

### Connection Handshake

```
Client                          Broker (port 33000)
  │── writeClientInfoExchange (10 bytes) ──▶│
  │◀── new CAS port (4 bytes) ─────────────│

Client                          CAS (new port)
  │── writeOpenDatabase (628 bytes) ────────▶│
  │◀── openDatabaseResult (CASInfo + version)│
```

## Key Files

| File            | Responsibility |
|-----------------|----------------|
| `driver.go`     | `sql.Register`, DSN parsing (`parseDSN`) |
| `conn.go`       | TCP connection lifecycle, all I/O helpers (`send`/`recv`) |
| `stmt.go`       | Prepared statement (`Exec`, `Query`) |
| `rows.go`       | Cursor iteration with server-side fetch buffer |
| `tx.go`         | `Commit` / `Rollback` via `writeEndTran` |
| `packet.go`     | `packetWriter` / `packetReader` — big-endian binary codec |
| `protocol.go`   | All `write*` / `parse*` protocol functions |
| `constants.go`  | Function codes (`FuncPrepare`, etc.), type codes, sizes |
| `errors.go`     | `CubridError`, `IntegrityError`, `ProgrammingError`, `OperationalError` |
| `types.go`      | `interpolateArgs`, `formatValue` (used only for logging) |
| `dialector/`    | GORM dialector + migrator |

## Protocol Reference

CAS function codes (defined in `constants.go`):

| Constant                 | Value | Description                      |
|--------------------------|-------|----------------------------------|
| `FuncEndTran`            | 1     | Commit or rollback               |
| `FuncPrepare`            | 2     | Prepare SQL statement            |
| `FuncExecute`            | 3     | Execute prepared statement       |
| `FuncCloseReqHandle`     | 6     | Release server-side handle       |
| `FuncFetch`              | 8     | Fetch next batch of rows         |
| `FuncGetDbVersion`       | 15    | Get CUBRID server version        |
| `FuncConClose`           | 31    | Close connection                 |
| `FuncGetLastInsertId`    | 40    | Get last auto-increment ID       |
| `FuncPrepareAndExecute`  | 41    | Prepare + execute in one round-trip |

### Packet framing

Every request (except handshake packets) is framed with an 8-byte header:

```
[4 bytes: DATA_LENGTH (big-endian)] [4 bytes: CAS_INFO]
[DATA_LENGTH bytes: payload]
```

Responses have a 4-byte length prefix only (no CAS_INFO in the response header
after the initial handshake; CAS_INFO is embedded in the response body).

### Response body layout

```
[4 bytes: CAS_INFO] [4 bytes: response_code] [payload...]
```

`response_code < 0` means error. In that case the payload is:
```
[4 bytes: error_code] [N bytes: null-terminated error message]
```

## Adding a New Protocol Feature

1. Add the function code constant to `constants.go` if missing.
2. Add `write<Feature>(...)` and `parse<Feature>(...)` functions to `protocol.go`.
   - Keep all types/functions **unexported** (lowercase).
3. Call from `conn.go`, `stmt.go`, or `rows.go` as appropriate.
4. Add tests to `protocol_test.go`.

## Code Conventions

- All wire-protocol internals are **unexported** (lowercase). Only error types
  and `Driver` are part of the public API.
- Packet functions follow the naming pattern: `write<X>` / `parse<X>`.
- `packetWriter` builds payloads; `buildProtocolHeader` prepends the 8-byte header.
- `encodeBindParams` / `encodeOneParam` handle typed bind parameter encoding for FC=3.
- `parseRowData` handles result row decoding; `readValue` dispatches by type code.

## Exported Public API

```go
// Driver — registered automatically via init()
type Driver struct{}

// Error types
type CubridError    struct { Code int; Message string }
type IntegrityError  struct { CubridError }
type ProgrammingError struct { CubridError }
type OperationalError  struct { CubridError }

// Version
const Version = "0.0.1-dev"
```

## Running a Local CUBRID Server (Docker)

```bash
docker run -d --name cubrid \
  -p 33000:33000 \
  cubrid/cubrid:11.3

# Create demodb
docker exec -it cubrid csql -u dba demodb
```
