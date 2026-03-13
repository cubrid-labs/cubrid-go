# cubrid-go

A pure-Go `database/sql` driver and GORM dialector for [CUBRID](https://www.cubrid.org) databases.

## Features

- Pure Go — no CGO, no Java, no CUBRID client library required
- Implements the `database/sql/driver` standard interface
- GORM dialector included (`dialector/` sub-package)
- Typed bind parameters sent over the wire (no string interpolation)
- Prepared statement reuse via server-side handles
- Transaction support (commit / rollback)
- Server-side cursor with configurable fetch size

## Requirements

- Go 1.21+
- CUBRID 10.x or 11.x server

## Installation

```bash
go get github.com/cubrid-labs/cubrid-go
```

## DSN Format

```
cubrid://[user[:password]]@host[:port]/database[?param=value&...]
```

| Parameter    | Default | Description                      |
|--------------|---------|----------------------------------|
| `autocommit` | `true`  | Enable/disable auto-commit mode  |
| `timeout`    | `30s`   | Connection and I/O deadline      |

**Examples:**

```
cubrid://dba:@localhost:33000/demodb
cubrid://admin:secret@db.example.com:33000/mydb?autocommit=false&timeout=10s
```

## Usage with `database/sql`

```go
import (
    "database/sql"
    _ "github.com/cubrid-labs/cubrid-go"
)

db, err := sql.Open("cubrid", "cubrid://dba:@localhost:33000/demodb")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Query
rows, err := db.Query("SELECT id, name FROM users WHERE active = ?", 1)
defer rows.Close()
for rows.Next() {
    var id int
    var name string
    rows.Scan(&id, &name)
}

// Exec
res, err := db.Exec("INSERT INTO users (name) VALUES (?)", "Alice")
id, _       := res.LastInsertId()
affected, _ := res.RowsAffected()

// Transaction
tx, _ := db.Begin()
tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", 100, 1)
tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", 100, 2)
tx.Commit()
```

## Usage with GORM

```go
import (
    "gorm.io/gorm"
    cubrid "github.com/cubrid-labs/cubrid-go/dialector"
)

db, err := gorm.Open(cubrid.Open("cubrid://dba:@localhost:33000/demodb"), &gorm.Config{})
if err != nil {
    log.Fatal(err)
}

// AutoMigrate
type User struct {
    ID   uint   `gorm:"primaryKey;autoIncrement"`
    Name string `gorm:"size:100"`
}
db.AutoMigrate(&User{})

// CRUD
db.Create(&User{Name: "Alice"})

var user User
db.First(&user, 1)

db.Model(&user).Update("Name", "Bob")
db.Delete(&user)
```

## Supported Data Types

| CUBRID Type           | Go Type     |
|-----------------------|-------------|
| `SMALLINT`            | `int64`     |
| `INTEGER`             | `int64`     |
| `BIGINT`              | `int64`     |
| `FLOAT`               | `float64`   |
| `DOUBLE`, `MONETARY`  | `float64`   |
| `NUMERIC`             | `string`    |
| `CHAR`, `VARCHAR`     | `string`    |
| `NCHAR`, `NVARCHAR`   | `string`    |
| `ENUM`                | `string`    |
| `DATE`                | `time.Time` |
| `TIME`                | `time.Time` |
| `TIMESTAMP`           | `time.Time` |
| `DATETIME`            | `time.Time` |
| `BIT`, `BIT VARYING`  | `[]byte`    |
| `BLOB`, `CLOB`        | `[]byte`    |
| `NULL`                | `nil`       |

## Error Handling

Three exported error types allow callers to handle different failure categories:

```go
import "errors"
import cubrid "github.com/cubrid-labs/cubrid-go"

if err != nil {
    var intErr  *cubrid.IntegrityError    // unique key, foreign key violation
    var progErr *cubrid.ProgrammingError  // syntax error, table not found
    var opErr   *cubrid.OperationalError  // network or server-side failure

    switch {
    case errors.As(err, &intErr):
        fmt.Println("constraint violation:", intErr.Code, intErr.Message)
    case errors.As(err, &progErr):
        fmt.Println("programming error:", progErr.Message)
    case errors.As(err, &opErr):
        fmt.Println("operational error:", opErr.Message)
    }
}
```

## Running Tests

**Unit tests** (no CUBRID server required):

```bash
go test ./...
```

**Integration tests** (requires a running CUBRID server):

```bash
CUBRID_DSN=cubrid://dba:@localhost:33000/demodb go test -tags integration ./...
```

## Project Structure

```
cubrid-go/
├── driver.go       # Driver registration, DSN parsing
├── conn.go         # TCP connection, handshake, database/sql Conn interface
├── stmt.go         # Prepared statement (Stmt interface)
├── rows.go         # Result rows with server-side cursor fetch
├── tx.go           # Transaction (commit / rollback)
├── packet.go       # Binary serializer / deserializer (big-endian)
├── protocol.go     # CAS broker wire protocol (FC=2,3,6,8,15,31,40,41)
├── errors.go       # Exported error types
├── constants.go    # Protocol constants (function codes, data types, sizes)
├── types.go        # SQL value formatting utilities
└── dialector/
    └── cubrid.go   # GORM dialector + migrator
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Write tests for new functionality
4. Ensure `go test ./...` passes
5. Open a pull request

## License

MIT
