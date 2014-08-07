go-themoviedb [![Build Status](https://travis-ci.org/amahi/go-themoviedb.png?branch=master)](https://travis-ci.org/amahi/go-themoviedb)
=============

Golang interface to The Movie DB (TMDb) APIs

See the godoc [go-themoviedb documentation](http://godoc.org/github.com/amahi/go-themoviedb)


Golang library for requesting metadata from themoviedb.org It's used by

1) Initializing the library via Init(), with the caller's API key from http://www.themoviedb.org like

2) Calling MovieData() to get the actual data, like

```go
package main

import "fmt"
import "github.com/amahi/go-themoviedb"

func main() {
        tmdb := tmdb.Init("your-api-key")
        metadata, err := tmdb.MovieData("Pulp Fiction")
        if err != nil {
                fmt.Printf("Error: %s\n", err)
        } else {
                fmt.Printf("TMDb Metadata: %s\n", metadata)
        }
}
```


the metadata is returned in XML format according to TMDB guidelines.
