# go-way
Golang postal level geo-lookup library.

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
radar := way.New()

// optionally wait for background refresh to occur
radar.Wait()
if err := radar.Error(); err != nil{
    panic(err)
}

loc, err := radar.IP("1.2.3.4")
if err != nil{
    panic(err)
}

loc, err = radar.City(country.UnitedStatesAmerica, "NY", "Brooklyn")
if err != nil{
    panic(err)
}

loc, err = radar.Postal(country.UnitedStatesAmerica, "NY", "10027")
if err != nil{
    panic(err)
}
```

## Powered by
* GeoNames - http://www.geonames.org
* MaxMind - https://www.maxmind.com 
