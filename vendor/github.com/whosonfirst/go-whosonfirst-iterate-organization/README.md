# go-whosonfirst-iterate-organization

Go package to implement the `whosonfirst/go-whosonfirst-iterate/v2/emitter.Emitter` interface for iterating multiple repositories in a GitHub organization.

## Motivation

This package is designed for situations where you need to iterate over multiple Who's On First -style repositories in a GitHub organization.

Under the hood it is thin wrapper around the [whosonfirst/go-whosonfirst-github](https://github.com/whosonfirst/go-whosonfirst-github) package, used to fetch a list of repositories for an organization, and the [whosonfirst/go-whosonfirst-iterate-git](https://github.com/whosonfirst/go-whosonfirst-iterate-git) to fetch each repository and iterate over its files.

It implements the [whosonfirst/go-whosonfirst-iterate/v2/emitter.Emitter](https://github.com/whosonfirst/go-whosonfirst-iterate) interface so it will work with any existing code and callback functions designed for use with the `go-whosonfirst-iterate` package.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-whosonfirst-iterate-organization.svg)](https://pkg.go.dev/github.com/whosonfirst/go-whosonfirst-iterate-organization)

## Example

Iterate over all the [sfomuseum-data/sfomuseum-data-flights-*](https://github.com/sfomuseum-data/?q=sfomuseum-data-flights&type=all&language=&sort=) repositories, excluding `sfomuseum-data-flights-YYYY-MM`, cloning each repository to `/tmp` before processing. (Each repository will be deleted after processing).

```
package main

import (
	_ "github.com/whosonfirst/go-whosonfirst-iterate-organization"
)

import (
	"context"
	"io"
	"fmt"
	"testing"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
)

func main()

	ctx := context.Background()

	iter_uri := "org:///tmp"
	
	iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {
		fmt.Println(path)
		return nil
	}

	iter, _ := iterator.NewIterator(ctx, iter_uri, iter_cb)

	iter.IterateURIs(ctx, "sfomuseum-data://?prefix=sfomuseum-data-flights-&exclude=sfomuseum-data-flights-YYYY-MM")
}
```

_Error handling omitted for the sake of brevity._

## See also

* https://github.com/whosonfirst/go-whosonfirst-iterate
* https://github.com/whosonfirst/go-whosonfirst-iterate-git
* https://github.com/whosonfirst/go-whosonfirst-github
