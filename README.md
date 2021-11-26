# go-way
Golang geolocation library.

## Installation

go-way may be installed using the go get command:

```
go get github.com/pghq/go-way
```
## Usage

```
import "github.com/pghq/go-way"
```

To create a new client:

```
client, err := way.NewClient()
if err != nil{
    panic(err)
}

loc, err := client.Locations.IP("1.2.3.4")
if err != nil{
    panic(err)
}

loc, err = client.Locations.City("US", "NY", "Brooklyn")
if err != nil{
    panic(err)
}
```

## Powered by
* GeoNames - http://www.geonames.org
* MaxMind - https://www.maxmind.com 
