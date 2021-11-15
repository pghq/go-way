# go-poco
Go client library for efficient postal code lookup (powered by GeoNames)

## Installation

go-poco may be installed using the go get command:

```
go get github.com/pghq/go-poco
```
## Usage

```
import "github.com/pghq/go-poco/poco"
```

To create a new client:

```
client, err := poco.New()
if err != nil{
    panic(err)
}

// TODO: See tests for specific use cases...
```

