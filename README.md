# safe-write

safe-write provides methods to manage configuration files where it is important that they are either completely
written to the disk or not at all – even when the process is unexpectedly interrupted or there are concurrent writes.

[![GoDoc](https://godoc.org/github.com/robojones/safe-write?status.svg)](https://godoc.org/github.com/robojones/safe-write)

```go
package main

import "github.com/robojones/safe-write"

func main() {
    data := []byte("{ \"data\": \"some important data\" }")
    // Write the data to /data/file/asdf-12312124123
    // When the write is complete, create a hard link from /data/file/asdf to the written file.
    safe.WriteFile("/data/file/asdf", data)
}
```
