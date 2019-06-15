# uuid
Golang UUID implementation that supports TimeUUID version

### Checkout
```
go get "github.com/consensusdb/uuid"
```

### Import
```
import "github.com/consensusdb/uuid"
```

### Quick start example:
```
	uuid := uuid.NewUUID(uuid.TimebasedUUID)
	uuid.SetUnixTimeMillis(123)
	uuid.SetCounter(555)
	fmt.Print(uuid.MarshalBinary())
	uuid.Parse(uuid.String())
```