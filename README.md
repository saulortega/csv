# csv
Formats slices of structs to CSV

# Install
`go get https://github.com/saulortega/csv`

# Usage
```go
import (
	"fmt"
	"time"
	"github.com/saulortega/csv"
)

type SexyType struct {
	One   int
	Two   string
	Three time.Time
	Four  bool
}

var sexyTypes = []SexyType{
	SexyType{One: 1, Two: "s"},
	SexyType{One: 2, Two: "st"},
	SexyType{One: 3, Two: "str", Three: time.Now()},
	SexyType{One: 4, Two: "stri", Three: time.Now()},
	SexyType{One: 5, Two: "strin", Three: time.Now(), Four: true},
	SexyType{One: 6, Two: "string", Three: time.Now(), Four: true},
}

f, err := csv.Format(sexyTypes) //[][]string, error
if err != nil {
	panic(err)
}

//First row:
fmt.Println(f[0])

//Second Cell from first row:
fmt.Println(f[0][1])
```

Or you can write to Writer Interface:
```go
// w = http.ResponseWriter
err := csv.WriteTo(w, sexyTypes)
```

### Header

By default the first row is the header, which will take the names of the fields:
```go
type SexyType struct {
	One   int
	Two   string
	Three time.Time
	Four  bool
}
//First row:
//"One", "Two", "Three", "Four"
```

The names of header can be changed with the csv tag:
```go
type SexyType struct {
	One   int `csv:"uno"`
	Two   string `csv:"dos"`
	Three time.Time `csv:"tres"`
	Four  bool
}
//First row:
//"uno", "dos", "tres", "Four"
```

Or passing the parameter csv.Header:
```go
csv.Format(sexyTypes, csv.Header{"ONE", "dos", "THREE", "cuatro"})
//First row:
//"ONE", "dos", "THREE", "cuatro"
```

You can ommit the Header:
```go
csv.Format(sexyTypes, csv.Header{"-"})
//First row:
//"1", "s", "", "false"
```

### Ommiting fields

Ommit a single field with the `csv:"-"` tag:
```go
type SexyType struct {
	One   int `csv:"uno"`
	Two   string `csv:"dos"`
	Three time.Time `csv:"-"` //I will be omitted
	Four  bool
}
//First row:
//"uno", "dos", "Four"
```

Whitelist:
```go
csv.Format(sexyTypes, csv.Whitelist{"uno", "dos"}) //If you uses `csv:"uno"` and `csv:"dos"` tags
//First row:
//"uno", "dos"

csv.Format(sexyTypes, csv.Whitelist{"One", "Two"}) //If you don't uses csv tags
//First row:
//"One", "Two"
```

Blacklist:
```go
csv.Format(sexyTypes, csv.Blacklist{"One", "Two"}) //If you don't uses csv tags
//First row:
//"Three", "Four"
```

### More examples
```go
type SexyType struct {
	One   int `csv:"uno"`
	Two   string `csv:"dos"`
	Three time.Time `csv:"-"` //I will be omitted
	Four  bool
}

csv.Format(sexyTypes)
//First row:
//"uno", "dos", "Four"

csv.Format(sexyTypes, csv.Header{"1", "22", "333"})
//First row:
//"1", "22", "333"

csv.Format(sexyTypes, csv.Header{"-"})
//First row:
//"1", "s", "false"

csv.Format(sexyTypes, csv.Header{"-"}, csv.Whitelist{"dos", "Four"})
//First row:
//"s", "false"

csv.Format(sexyTypes, csv.Header{"-"}, csv.Blacklist{"dos", "Four"}) //Three is also ignored by `csv:"-"` tag
//First row:
//"1"

//w is a Writer Interface,
csv.WriteTo(w, sexyTypes, csv.Header{"-"}, csv.Whitelist{"One", "dos"}) //"uno" overwrites "One", so "One" has no effect
//First row:
//"s"
```

# Definitions

```go
func Format(obj interface{}, options ...interface{}) ([][]string, error)

func WriteTo(w io.Writer, obj interface{}, options ...interface{}) error

//options:
type Whitelist []string
//csv.Whitelist{"field1", "field2"}

type Blacklist []string
//csv.Blacklist{"field1", "field2"}

type Header []string
//csv.Header{"field1", "field2"}
```
