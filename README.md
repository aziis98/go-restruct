# Restuct

This is a small library that uses generics and reflection to provide a nice way to convert structs and types in a highly configurable way.

## Example

```go
// First format

type Foo struct {
    First  int
    Second string
}

// Second format

type SlugString string

func NewSlugString(s string) SlugString {
    return SlugString(strings.ReplaceAll(strings.TrimSpace(s), " ", "-"))
}

type Bar struct {
    FirstField  int
    SecondField SlugString
}
```

Then we can convert from `Foo` to `Bar` using the following

```go
func main() {
    foo1 := Foo{
        First:  1,
        Second: "  this is my foo    ",
    }

    bar, _ := restruct.Convert[Bar](foo1,
        restruct.MustFunc[SlugString, string](NewSlugString),
        restruct.StructFromStruct[Bar, Foo]{
            "FirstField":  "First",
            "SecondField": "Second",
        },
    )

    fmt.Printf("%#v", bar)
    // Bar{
    //     FirstField:  1,
    //     SecondField: SlugString("this-is-my-foo"),
    // }
}
```

## Usage

```bash shell
go get github.com/aziis98/go-restruct
```

```go
import "github.com/aziis98/go-restruct"
```

## Reference

See the docs for now

TODO
