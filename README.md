# safe-write

safe-write provides methods to manage critical files where it is important that they are either completely
written to the disk or not at all – even when the process is unexpectedly interrupted or there are concurrent writes.

[![GoDoc](https://godoc.org/github.com/robojones/safe-write?status.svg)](https://godoc.org/github.com/robojones/safe-write)

```go
package main

import (
    "fmt"
    "github.com/robojones/safe-write"
)

func main() {
    data := []byte("{ \"data\": \"some important data\" }")
    
    safe.WriteFile("config.json", data)

    got, _ := safe.ReadFile("config.json")
    fmt.Printf("The data is: %s", string(got))
    
    safe.RemoveFile("config.json")
}
```
