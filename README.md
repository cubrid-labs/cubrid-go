# cubrid-go

**Pure Go database driver for CUBRID** — `database/sql` interface + GORM dialector, no CGO required.

[![Go Reference](https://pkg.go.dev/badge/github.com/cubrid-labs/cubrid-go.svg)](https://pkg.go.dev/github.com/cubrid-labs/cubrid-go)
[![Go 1.21+](https://img.shields.io/badge/go-1.21%2B-blue.svg)](https://go.dev)
[![license](https://img.shields.io/github/license/cubrid-labs/cubrid-go)](https://github.com/cubrid-labs/cubrid-go/blob/main/LICENSE)
[![GitHub stars](https://img.shields.io/github/stars/cubrid-labs/cubrid-go)](https://github.com/cubrid-labs/cubrid-go)
<!-- BADGES:END -->

## Why cubrid-go?

| | cubrid-go | CCI (C interface) |
|:---|:---|:---|
| **CGO Required** | No — pure Go | Yes |
| **Cross-compilation** | `GOOS=linux GOARCH=arm64 go build` | Requires C toolchain |
| **database/sql** | Native interface | Wrapper needed |
| **GORM Support** | Built-in dialector | Not available |
| **Connection Pool** | Go standard (`database/sql`) | Manual management |
| **Deployment** | Single binary | Shared library dependency |

cubrid-go speaks the CUBRID CAS protocol directly over TCP. No shared libraries, no CGO, no external dependencies — just `go get` and start coding.

## Installation

```bash
go get github.com/cubrid-labs/cubrid-go
```

**Requirements**: Go 1.21+

## Quick Start

### database/sql

```go
package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/cubrid-labs/cubrid-go"
)

func main() {
    db, err := sql.Open("cubrid", "cubrid://dba:@localhost:33000/demodb")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Verify connection
    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Connected to CUBRID!")

    // Query
    rows, err := db.Query("SELECT name, nation_code FROM athlete WHERE nation_code = ?", "KOR")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    for rows.Next() {
        var name, nation string
        rows.Scan(&name, &nation)
        fmt.Printf("%s (%s)\n", name, nation)
    }
}
```

### GORM

```go
package main

import (
    "fmt"
    "log"

    "gorm.io/gorm"
    cubrid "github.com/cubrid-labs/cubrid-go/dialector"
)

type Athlete struct {
    Code       int    `gorm:"primaryKey;autoIncrement"`
    Name       string `gorm:"size:40;not null"`
    NationCode string `gorm:"size:3"`
    Gender     string `gorm:"size:1"`
    Event      string `gorm:"size:40"`
}

func main() {
    db, err := gorm.Open(cubrid.Open("cubrid://dba:@localhost:33000/demodb"), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    // Auto-migrate creates the table if it doesn't exist
    db.AutoMigrate(&Athlete{})

    // Create
    db.Create(&Athlete{Name: "Hong Gildong", NationCode: "KOR", Gender: "M", Event: "Marathon"})

    // Read
    var athletes []Athlete
    db.Where("nation_code = ?", "KOR").Find(&athletes)
    for _, a := range athletes {
        fmt.Printf("%s - %s\n", a.Name, a.Event)
    }

    // Update
    db.Model(&Athlete{}).Where("name = ?", "Hong Gildong").Update("event", "Sprint")

    // Delete
    db.Delete(&Athlete{}, "nation_code = ?", "ZZZ")
}
```

## DSN Format

```
cubrid://[user[:password]]@host[:port]/database[?autocommit=true&timeout=30s]
```

| Parameter    | Default     | Description                         |
|:-------------|:------------|:------------------------------------|
| `host`       | `localhost` | CUBRID broker host                  |
| `port`       | `33000`     | CUBRID broker port                  |
| `database`   | *(required)* | Target database name               |
| `user`       | `""`        | Database user                       |
| `password`   | `""`        | Database password                   |
| `autocommit` | `true`      | Enable/disable auto-commit          |
| `timeout`    | `30s`       | Connection timeout ([Go duration](https://pkg.go.dev/time#ParseDuration)) |

**Examples**:

```go
// Local development
"cubrid://dba:@localhost:33000/demodb"

// Remote with credentials
"cubrid://admin:secret@db.example.com:33000/production"

// Manual transaction mode
"cubrid://dba:@localhost:33000/demodb?autocommit=false"

// Custom timeout
"cubrid://dba:@localhost:33000/demodb?timeout=60s"
```

## Supported Features

| Feature | Status | Notes |
|:--------|:-------|:------|
| Pure TCP (no shared library) | ✅ | No CGO, no external deps |
| `database/sql` driver | ✅ | Full interface |
| Parameterized queries (`?`) | ✅ | Client-side interpolation |
| Transactions (`Begin/Commit/Rollback`) | ✅ | |
| GORM dialector | ✅ | Models, migrations, CRUD |
| GORM AutoMigrate | ✅ | Create tables, add columns |
| Server-side cursor / lazy fetch | ✅ | Batches of 100 rows |
| Result streaming (FETCH) | ✅ | Memory-efficient |
| Last insert ID | ✅ | Via `SELECT LAST_INSERT_ID()` |
| Connection pool | ✅ | Go standard `database/sql` |
| Connection health check (Ping) | ✅ | Via `GET_DB_VERSION` |
| LOB (BLOB/CLOB) | ⚠️ | Raw bytes only |
| Timezone-aware types | ⚠️ | UTC only |
| ON CONFLICT / UPSERT | ❌ | Use raw SQL workaround |

## Type Mapping

### Go → CUBRID (Parameters)

| Go Type | CUBRID Literal | Example |
|:--------|:---------------|:--------|
| `nil` | `NULL` | |
| `bool` | `0` / `1` | CUBRID has no BOOLEAN |
| `int64` | Integer | `42` |
| `float64` | Float | `3.14` |
| `string` | `'escaped'` | Quotes doubled |
| `[]byte` | `X'cafe'` | Hex-encoded |
| `time.Time` | `DATETIME'...'` | Millisecond precision |

### CUBRID → Go (Results)

| CUBRID Type | Go Type |
|:------------|:--------|
| `SMALLINT`, `INTEGER`, `BIGINT` | `int64` |
| `FLOAT`, `DOUBLE`, `MONETARY` | `float64` |
| `NUMERIC` | `string` |
| `CHAR`, `VARCHAR`, `STRING` | `string` |
| `DATE`, `TIME`, `DATETIME`, `TIMESTAMP` | `string` |
| `BIT`, `BIT VARYING`, `BLOB`, `CLOB` | `[]byte` |
| `SET`, `MULTISET`, `SEQUENCE` | `string` |

## GORM Type Mapping

| GORM Field Type | CUBRID SQL Type |
|:----------------|:----------------|
| `Bool` | `SMALLINT` |
| `Int` (≤16 bits) | `SMALLINT` |
| `Int` (≤32 bits) | `INTEGER` |
| `Int` (>32 bits) | `BIGINT` |
| `Float` (≤32 bits) | `FLOAT` |
| `Float` (>32 bits) | `DOUBLE` |
| `String` | `VARCHAR(n)` (default 256) |
| `String` (very large) | `STRING` |
| `Time` | `DATETIME` |
| `Bytes` | `BLOB` |
| Auto-increment (≤32) | `INTEGER AUTO_INCREMENT` |
| Auto-increment (>32) | `BIGINT AUTO_INCREMENT` |

## Error Types

```go
import "github.com/cubrid-labs/cubrid-go"

// Base error
var cubridErr *cubrid.CubridError

// Constraint violation (unique, foreign key)
var integrityErr *cubrid.IntegrityError

// SQL syntax error, missing table/column
var progErr *cubrid.ProgrammingError

// Network/connection failure
var opErr *cubrid.OperationalError

// Use errors.As for type checking
if errors.As(err, &integrityErr) {
    fmt.Printf("Constraint violation: code=%d msg=%s\n", integrityErr.Code, integrityErr.Message)
}
```

## Architecture

```
cubrid-go/
├── constants.go    CAS protocol constants (function codes, data types)
├── packet.go       PacketWriter / PacketReader — big-endian binary codec
├── protocol.go     High-level packet builders and parsers
├── errors.go       CubridError, IntegrityError, ProgrammingError, OperationalError
├── types.go        FormatValue — Go → CUBRID type conversion
├── conn.go         TCP connection + two-step broker handshake
├── stmt.go         database/sql Stmt — PrepareAndExecute (FC=41)
├── rows.go         database/sql Rows — lazy server-side cursor (FC=8)
├── tx.go           database/sql Tx — COMMIT / ROLLBACK (FC=1)
├── driver.go       Driver registration + DSN parser
└── dialector/
    └── cubrid.go   GORM Dialector + Migrator
```

## Protocol Notes

cubrid-go speaks the CUBRID CAS (Client Application Server) protocol directly over TCP. The two-step connection sequence is:

1. **Broker handshake** — connect to `host:port`, send a 10-byte magic string, receive a redirected CAS port.
2. **Open database** — connect to the CAS port, send credentials (628 bytes), receive session info.

All subsequent requests use the `PREPARE_AND_EXECUTE` (function code 41) combined packet, and large result sets are streamed back via `FETCH` (function code 8).

## Documentation

| Document | Description |
|:---------|:------------|
| [API Reference](docs/API_REFERENCE.md) | Complete `database/sql` driver API, DSN format, type mapping, error types, protocol details |
| [GORM Guide](docs/GORM.md) | GORM dialector setup, models, migrations, CRUD, transactions, querying, schema inspection |
| [Troubleshooting](docs/TROUBLESHOOTING.md) | Connection, query, transaction, GORM, type, and performance issues with solutions |

## FAQ

### How do I connect to CUBRID with Go?

```go
import (
    "database/sql"
    _ "github.com/cubrid-labs/cubrid-go"
)
db, err := sql.Open("cubrid", "cubrid://dba:@localhost:33000/demodb")
```

### How do I use CUBRID with GORM?

```go
import (
    "gorm.io/gorm"
    cubrid "github.com/cubrid-labs/cubrid-go/dialector"
)
db, err := gorm.Open(cubrid.Open("cubrid://dba:@localhost:33000/demodb"), &gorm.Config{})
```

### Does cubrid-go require CGO?

No. cubrid-go is a pure Go implementation that speaks the CUBRID CAS protocol directly over TCP. No shared libraries or CGO needed.

### What Go version is required?

Go 1.21 or later.

### Does cubrid-go support connection pooling?

Yes, via Go's standard `database/sql` package. Configure with `db.SetMaxOpenConns()`, `db.SetMaxIdleConns()`, and `db.SetConnMaxLifetime()`.

### Does cubrid-go support transactions?

Yes. Use `db.Begin()` to start a transaction, then `tx.Commit()` or `tx.Rollback()`.

### Does GORM AutoMigrate work with CUBRID?

Yes. The GORM dialector supports `AutoMigrate` for creating and updating table schemas.

## Ecosystem

| Package | Description | Language |
|:--------|:------------|:---------|
| [cubrid-go](https://github.com/cubrid-labs/cubrid-go) | database/sql driver + GORM dialector | Go |
| [pycubrid](https://github.com/cubrid-labs/pycubrid) | PEP 249 DB-API 2.0 driver | Python |
| [sqlalchemy-cubrid](https://github.com/cubrid-labs/sqlalchemy-cubrid) | SQLAlchemy 2.0 dialect | Python |
| [cubrid-client](https://github.com/cubrid-labs/cubrid-client) | Modern TypeScript client | TypeScript |
| [drizzle-cubrid](https://github.com/cubrid-labs/drizzle-cubrid) | Drizzle ORM dialect | TypeScript |
| [cubrid-cookbook](https://github.com/cubrid-labs/cubrid-cookbook) | Working examples for all platforms | Multi |

## License

MIT
