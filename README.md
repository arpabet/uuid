# uuid

![build workflow](https://go.arpabet.com/uuid/actions/workflows/build.yaml/badge.svg)

A fast, allocation-conscious Golang UUID implementation with first-class support
for time-based (TimeUUID) versions.

A `UUID` is stored as two 64-bit words (`MostSigBits` / `LeastSigBits`), the same
layout used by `java.util.UUID`, which makes interop with JVM systems trivial.

## Features

- All standard versions: **v1** (time-based), **v2** (DCE security), **v3 / v5**
  (name-based MD5 / SHA-1), **v4** (random), and **v6 / v7** (RFC 9562
  time-ordered).
- Built-in generators: `NewV1` (monotonic), `RandomUUID` (v4), `NewV7`
  (Unix-time-ordered, sortable).
- Lexicographically **sortable** encodings for time UUIDs.
- Implements `encoding.TextMarshaler` / `TextUnmarshaler`,
  `encoding.BinaryMarshaler` / `BinaryUnmarshaler`, `json.Marshaler` /
  `Unmarshaler`, and `database/sql.Scanner` / `driver.Valuer`.
- Zero-allocation `Parse` and minimal-allocation `String` (see [Performance](#performance)).

## Requirements

- Go **1.25** or newer.

## Install

```sh
go get go.arpabet.com/uuid
```

```go
import "go.arpabet.com/uuid"
```

## Quick start

```go
// Generate a sortable, time-ordered v7 UUID (recommended for new systems).
id, err := uuid.NewV7()
if err != nil {
    log.Fatal(err)
}

fmt.Println(id.String())          // e.g. 0191d4e2-7f3a-7c10-8b2e-1a2b3c4d5e6f
fmt.Println(id.Version())         // UnixTimeVer7
fmt.Println(id.Variant())         // IETF

// Round-trip through text.
parsed, err := uuid.Parse(id.String())
fmt.Println(id.Equal(parsed))     // true
```

## Generators

```go
v1, _ := uuid.NewV1()      // time-based, monotonic, random node (multicast bit set)
v4, _ := uuid.RandomUUID() // cryptographically random
v7, _ := uuid.NewV7()      // RFC 9562 Unix-time-ordered, sortable as binary and text

// Name-based (deterministic) UUIDs.
v3, _ := uuid.NameUUIDFromBytes([]byte("example.com"), uuid.NamebasedVer3) // MD5
v5, _ := uuid.NameUUIDFromBytes([]byte("example.com"), uuid.NamebasedVer5) // SHA-1
```

`NewV1` is safe for concurrent use and guarantees that values produced within a
single process are strictly increasing, even if the wall clock does not advance.

## Building a TimeUUID manually

`New` creates an empty UUID of a given version; the setters let you compose the
timestamp, clock sequence, node and counter fields by hand:

```go
id := uuid.New(uuid.TimebasedVer1)
id.SetUnixTimeMillis(1718200000000)
id.SetCounter(555)

data, _ := id.MarshalBinary()
fmt.Printf("%x\n", data)

back, _ := uuid.Parse(id.String())
fmt.Println(id.Equal(back)) // true
```

## Parsing

`Parse` accepts every common textual form and rejects malformed input with an
error (it never silently returns the zero UUID):

```go
uuid.Parse("534b44a1-9bf1-3d20-b71e-cc4eb77c572f")            // canonical
uuid.Parse("534b44a19bf13d20b71ecc4eb77c572f")                // no dashes
uuid.Parse("{534b44a1-9bf1-3d20-b71e-cc4eb77c572f}")          // braces
uuid.Parse("\"534b44a1-9bf1-3d20-b71e-cc4eb77c572f\"")        // quotes
uuid.Parse("urn:uuid:534b44a1-9bf1-3d20-b71e-cc4eb77c572f")   // URN

if _, err := uuid.Parse("not-a-uuid"); err != nil {
    fmt.Println("rejected:", err)
}
```

`ParseBytes` is the same parser over a `[]byte`.

## Marshaling

```go
id, _ := uuid.NewV7()

// text
text, _ := id.MarshalText()        // 36-byte canonical form
fmt.Println(string(text))

// json (UUID encodes/decodes as a quoted string)
type Record struct {
    ID uuid.UUID `json:"id"`
}
b, _ := json.Marshal(Record{ID: id})
fmt.Println(string(b))             // {"id":"0191d4e2-7f3a-7c10-8b2e-1a2b3c4d5e6f"}

// binary (16 bytes, big-endian)
bin, _ := id.MarshalBinary()
var decoded uuid.UUID
_ = decoded.UnmarshalBinary(bin)
```

To avoid allocations, marshal into a caller-owned buffer:

```go
var buf [36]byte
_ = id.MarshalTextTo(buf[:])       // no allocation

var raw [16]byte
_ = id.MarshalBinaryTo(raw[:])     // no allocation
```

## Sortable encoding for TimeUUIDs

`MarshalSortableBinary` reorders the timestamp blocks and converts the counter to
unsigned bytes so that the resulting 16 bytes sort in chronological order — useful
for keys in ordered stores (RocksDB, LMDB, etc.). It is valid only for v1 UUIDs.

```go
id := uuid.New(uuid.TimebasedVer1)
id.SetTime(time.Now())
id.SetCounter(42)

key, err := id.MarshalSortableBinary()
if err == uuid.ErrorRequiredTimebasedUUID {
    // not a time-based UUID
}

var back uuid.UUID
_ = back.UnmarshalSortableBinary(key)
fmt.Println(id.Equal(back)) // true
```

For v7, the standard `MarshalBinary` output is already time-sortable, so no special
encoding is needed.

## Reading TimeUUID fields

```go
// version 1
id := uuid.New(uuid.TimebasedVer1)
id.SetTime(time.Now())
fmt.Println(id.Time())               // time.Time
fmt.Println(id.UnixTimeMillis())     // unix millis
fmt.Println(id.Node())               // 48-bit node
fmt.Println(id.ClockSequence())      // 14-bit clock sequence

// version 7
v7, _ := uuid.NewV7()
fmt.Println(v7.TimeV7())             // time.Time
fmt.Println(v7.UnixTimeMillisV7())   // unix millis
```

## Database usage

`uuid.UUID` implements `sql.Scanner` and `driver.Valuer`, storing the canonical
36-character string. `Scan` accepts a string, a textual `[]byte`, or a raw 16-byte
binary value, and treats SQL `NULL` as the zero UUID.

```go
var id uuid.UUID
err := db.QueryRow("SELECT id FROM users WHERE name = $1", name).Scan(&id)

_, err = db.Exec("INSERT INTO users(id, name) VALUES ($1, $2)", id, name)
```

## Interop with Java

```go
// from java.util.UUID.getMostSignificantBits()/getLeastSignificantBits()
id := uuid.Create(mostSigBits, leastSigBits)

msb := id.MostSignificantBits()
lsb := id.LeastSignificantBits()
```

## Versions and variants

```go
uuid.TimebasedVer1          // 1  time-based
uuid.DCESecurityVer2        // 2  DCE security
uuid.NamebasedVer3          // 3  name-based, MD5
uuid.RandomlyGeneratedVer4  // 4  random
uuid.NamebasedVer5          // 5  name-based, SHA-1
uuid.ReorderedTimeVer6      // 6  RFC 9562 reordered Gregorian time
uuid.UnixTimeVer7           // 7  RFC 9562 Unix epoch time

id.Version()   // e.g. uuid.UnixTimeVer7
id.Variant()   // e.g. uuid.IETF
```

## Performance

Benchmarks on Apple M4 (`go test -bench=.`):

| Operation        | ns/op | allocs/op |
|------------------|------:|----------:|
| `Parse` (dashed) |  ~17  |     0     |
| `Parse` (hex32)  |  ~13  |     0     |
| `String`         |  ~18  |     1     |
| `MarshalText`    |  ~10  |     0     |
| `MarshalBinary`  |  ~0.2 |     0     |
| `NewV7`          |  ~91  |     0     |
| `NewV1`          |  ~32  |     0     |

The single allocation in `String` is the returned string itself; use
`MarshalTextTo` with a stack buffer to avoid it.

## License

See [LICENSE](LICENSE).
