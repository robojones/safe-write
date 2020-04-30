# safe-write


```go
import "github.com/robojones/safe-write"

func main() {
    data := []byte("{ \"data\": \"some important data\" }")
    // Write the data to /data/file/asdf-12312124123
    // When the write is complete, create a hard link from /data/file/asdf to the written file.
    safe.WriteFile("/data/file/asdf", data)
}
```
