# uuid

![build workflow](https://go.arpabet.com/uuid/actions/workflows/build.yaml/badge.svg)

Golang UUID implementation that supports TimeUUID version

### Quick start example:
```
	id := uuid.New(uuid.TimebasedVer1)
	id.SetUnixTimeMillis(123)
	id.SetCounter(555)
	fmt.Print(id.MarshalBinary())
	uuid.Parse(id.String())
```
