# post2post

A simple Go library for string transformations and utilities.

## Installation

```bash
go get github.com/esa/post2post
```

## Usage

```go
package main

import (
    "fmt"
    "strings"
    "github.com/esa/post2post"
)

func main() {
    // Transform a string using a custom function
    result := post2post.Transform("hello", strings.ToUpper)
    fmt.Println(result) // Output: HELLO
    
    // Reverse a string
    reversed := post2post.Reverse("hello")
    fmt.Println(reversed) // Output: olleh
    
    // Convert to title case
    title := post2post.ToTitle("hello world")
    fmt.Println(title) // Output: Hello World
    
    // Greet someone
    greeting := post2post.Greet("Alice")
    fmt.Println(greeting) // Output: Hello, Alice!
}
```

## API

### `Transform(input string, fn func(string) string) string`
Applies a transformation function to a string.

### `Reverse(s string) string`
Returns the reversed string (Unicode-safe).

### `ToTitle(s string) string`
Converts string to title case.

### `Greet(name string) string`
Returns a greeting message.

## Testing

Run tests with:

```bash
go test
```

## License

MIT