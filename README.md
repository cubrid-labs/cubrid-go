# cubrid-go

<!-- BADGES:START -->
[![Go Reference](https://pkg.go.dev/badge/github.com/cubrid-labs/cubrid-go.svg)](https://pkg.go.dev/github.com/cubrid-labs/cubrid-go)
[![Go 1.21+](https://img.shields.io/badge/go-1.21%2B-blue.svg)](https://go.dev)
[![license](https://img.shields.io/github/license/cubrid-labs/cubrid-go)](https://github.com/cubrid-labs/cubrid-go/blob/main/LICENSE)
[![GitHub stars](https://img.shields.io/github/stars/cubrid-labs/cubrid-go)](https://github.com/cubrid-labs/cubrid-go)
<!-- BADGES:END -->


Pure-Go CUBRID database driver for Go's `database/sql` package and [GORM](https://gorm.io).

Ported from [pycubrid](https://github.com/cubrid-labs/pycubrid) — no CGO, no native libraries required.

## Installation

```bash
go get github.com/cubrid-labs/cubrid-go
```

## DSN Format

```
cubrid://[user[:password]]@host[:port]/database[?autocommit=true&timeout=30s]
```

| Parameter    | Default     | Description                         |
|--------------|-------------|-------------------------------------|
| `host`       | `localhost` | CUBRID broker host                  |
| `port`       | `33000`     | CUBRID broker port                  |
| `database`   | (required)  | Target database name                |
| `user`       | (empty)     | Database user                       |
| `password`   | (empty)     | Database password                   |
| `autocommit` | `true`      | Enable/disable auto-commit          |
| `timeout`    | `30s`       | Connection timeout (Go duration)    |

## Usage with `database/sql`

```go
import (
    "database/sql"
    _ "github.com/cubrid-labs/cubrid-go"
)

db, err := sql.Open("cubrid", "cubrid://dba:@localhost:33000/demodb")
if err != nil {
    panic(err)
}
defer db.Close()

rows, err := db.Query("SELECT * FROM athlete WHERE nation_code = ?", "KOR")
```

## Usage with GORM

```go
import (
    "gorm.io/gorm"
    cubrid "github.com/cubrid-labs/cubrid-go/dialector"
)

db, err := gorm.Open(cubrid.Open("cubrid://dba:@localhost:33000/demodb"), &gorm.Config{})
if err != nil {
    panic(err)
}

// Define a model.
type Athlete struct {
    Code       int    `gorm:"primaryKey;autoIncrement"`
    Name       string `gorm:"size:40;not null"`
    NationCode string `gorm:"size:3"`
    Gender     string `gorm:"size:1"`
    Event      string `gorm:"size:40"`
}

// Auto-migrate creates the table if it doesn't exist.
db.AutoMigrate(&Athlete{})

// Create
db.Create(&Athlete{Name: "Hong Gildong", NationCode: "KOR", Gender: "M", Event: "Marathon"})

// Find
var athletes []Athlete
db.Where("nation_code = ?", "KOR").Find(&athletes)

// Update
db.Model(&Athlete{}).Where("name = ?", "Hong Gildong").Update("event", "Sprint")

// Delete
db.Delete(&Athlete{}, "nation_code = ?", "ZZZ")
```

## Supported Features

| Feature                         | Status |
|---------------------------------|--------|
| Pure TCP (no shared library)    | ✅     |
| `database/sql` driver           | ✅     |
| Parameterised queries (`?`)     | ✅     |
| Transactions (`Begin/Commit/Rollback`) | ✅ |
| GORM dialector                  | ✅     |
| GORM AutoMigrate                | ✅     |
| Server-side cursor / lazy fetch | ✅     |
| Result streaming (FETCH)        | ✅     |
| Last insert ID                  | ✅     |
| Connection pool (via `database/sql`) | ✅ |
| LOB (BLOB/CLOB)                 | ⚠️ raw bytes only |
| Timezone-aware types            | ⚠️ UTC only       |

## Architecture

```
cubrid-go/
├── constants.go   CAS protocol constants (function codes, data types, …)
├── packet.go      PacketWriter / PacketReader — big-endian binary codec
├── protocol.go    High-level packet builders and parsers
├── errors.go      CubridError, IntegrityError, ProgrammingError
├── types.go       Client-side parameter interpolation
├── conn.go        TCP connection + broker handshake
├── stmt.go        database/sql Stmt — PrepareAndExecute
├── rows.go        database/sql Rows — lazy server-side cursor
├── tx.go          database/sql Tx — COMMIT / ROLLBACK
├── driver.go      Driver registration + DSN parser
└── dialector/
    └── cubrid.go  GORM Dialector + Migrator
```

## Protocol Notes

cubrid-go speaks the CUBRID CAS (Client Application Server) protocol directly over TCP.
The two-step connection sequence is:

1. **Broker handshake** — connect to `host:port`, send a 10-byte magic string,
   receive a redirected CAS port.
2. **Open database** — connect to the CAS port, send credentials (628 bytes),
   receive session info.

All subsequent requests use the `PREPARE_AND_EXECUTE` (function code 41) combined
packet, and large result sets are streamed back via `FETCH` (function code 8).

## License

MIT
