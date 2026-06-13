# uuid

![build workflow](https://go.arpabet.com/uuid/actions/workflows/build.yaml/badge.svg)

Golang UUID implementation that supports TimeUUID version

Supported versions: v1 (time-based), v2 (DCE), v3/v5 (name-based), v4 (random),
and v6/v7 (RFC 9562 time-ordered, sortable). Implements `encoding.TextMarshaler`,
`encoding.BinaryMarshaler`, `json.Marshaler`, `database/sql.Scanner` and
`database/sql/driver.Valuer`.

### Quick start example:
```go
	id := uuid.New(uuid.TimebasedVer1)
	id.SetUnixTimeMillis(123)
	id.SetCounter(555)
	fmt.Print(id.MarshalBinary())
	uuid.Parse(id.String())
```

### Generators:
```go
	v1, _ := uuid.NewV1()      // time-based, monotonic, random node
	v4, _ := uuid.RandomUUID() // random
	v7, _ := uuid.NewV7()      // RFC 9562 Unix-time-ordered, sortable
```

### Database usage:
```go
	// uuid.UUID satisfies sql.Scanner and driver.Valuer
	row.Scan(&id)
	db.Exec("INSERT INTO t(id) VALUES ($1)", id)
```
