# safe-write

safe-write provides methods to manage critical files where it is important that they are either completely
written to the disk or not at all. If the process is unexpectedly interrupted while updating the contents of a file, the original version of the file remains untouched.

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

## How it works

**TL;DR:** When overwriting files, the `WriteFile` method uses hard links and temporary files to ensure that there is always a consistent version of your file on the disk.

Lets assume we are using the `WriteFile` method to create a file called `config.json`. The write procedure is as follows.

1. Write the contents of the file to a temporary file (e.g. `config.json.2020-01-02T15-04-05.000000`)
2. Create a hard link from `config.json.1` to the temporary file
3. Create a hard link from `config.json` to the temporary file
4. Remove the temporary file. Because we are using hard links, the contents of the file are still available using the file names `config.json` and `config.json.1`

The reason why the intermediate link `config.json.1` is created becomes clear when we overwrite the contents of the file. Again, we are using the `WriteFile` method.

1. Write the updated contents of the file to a new temporary file (e.g. `config.json.2020-02-01T10-03-08.004001`)
2. Remove the previous hard link `config.json.1`
3. Create a new hard link from `config.json.1` to our new temporary file
4. Remove the previous hard link `config.json`
5. Create a new hard link from `config.json` to the new temporary file
6. Remove the temporary file

The intermediate link `config.json.1` is created so there is always one of our links `config.json` or `config.json.1`
pointing to a complete version of our config file.
Even if our process is interrupted by a server crash, at any point of the overwrite process,
there is always either `config.json` or `config.json.1` safely written on the disk.

The `ReadFile` method of this module always checks for both links, so even if the `config.json` link is missing,
there is a valid file available via `config.json.1`.
