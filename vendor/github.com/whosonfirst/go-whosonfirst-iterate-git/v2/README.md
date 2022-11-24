# go-whosonfirst-iterate-git

Go package implementing go-whosonfirst-iterate/emitter functionality for Git repositories.

## Important

Documentation for this package is incomplete and will be updated shortly.

## Example

```
package main

import (
	"context"
	"flag"
	"fmt"
	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v2"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"io"
	"log"
	"os"
	"strings"
	"sync/atomic"
)

func main() {

	valid_schemes := strings.Join(emitter.Schemes(), ",")
	emitter_desc := fmt.Sprintf("A valid whosonfirst/go-whosonfirst-iterate/v2 URI. Supported emitter URI schemes are: %s", valid_schemes)

	var emitter_uri = flag.String("emitter-uri", "git://", emitter_desc)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Count files in one or more whosonfirst/go-whosonfirst-iterate/v2 sources.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t %s [options] uri(N) uri(N)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	var count int64
	count = 0

	emitter_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {
		atomic.AddInt64(&count, 1)
		return nil
	}

	ctx := context.Background()

	iter, _ := iterator.NewIterator(ctx, *emitter_uri, emitter_cb)

	uris := flag.Args()
	iter.IterateURIs(ctx, uris...)

	log.Printf("Counted %d records (saw %d records)\n", count, iter.Seen)
}
```

_Error handling omitted for the sake of brevity._

## Tools

```
$> make cli
go build -mod vendor -o bin/count cmd/count/main.go
go build -mod vendor -o bin/emit cmd/emit/main.go
```

### count

Count files in one or more whosonfirst/go-whosonfirst-iterate/emitter sources.

```
$> ./bin/count -h
Count files in one or more whosonfirst/go-whosonfirst-iterate/emitter sources.
Usage:
	 ./bin/count [options] uri(N) uri(N)
Valid options are:

  -emitter-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/emitter URI. Supported emitter URI schemes are: directory://,featurecollection://,file://,filelist://,geojsonl://,git://,repo:// (default "git://")
```

```
$> ./bin/count \
	https://github.com/sfomuseum-data/sfomuseum-data-architecture.git

2021/02/17 15:54:32 time to index paths (1) 26.076332877s
2021/02/17 15:54:32 Counted 857 records (indexed 857 records)
```


```
$> ./bin/count \
	-emitter-uri 'git://?include=properties.mz:is_current=1&include=properties.sfomuseum:placetype=gate' \
	https://github.com/sfomuseum-data/sfomuseum-data-architecture.git

2021/02/17 16:00:17 time to index paths (1) 24.470490474s
2021/02/17 16:00:17 Counted 120 records (indexed 120 records)
```

By default `go-whosonfirst-iterate-git` clones Git repositories in to memory. If your emitter URI contains a path then repositories will be cloned in that path:

```
$> bin/count \
	-emitter-uri 'git:///tmp/data' \
	git@github.com:whosonfirst-data/whosonfirst-data-admin-is.git

2021/02/17 15:56:54 time to index paths (1) 3.742559429s
2021/02/17 15:56:54 Counted 436 records (indexed 436 records)
```

By default repositories cloned in to a path are removed. If you want to preserve the cloned repository include a `?preserve=1` query parameter in your URI string:

```
$> bin/count \
	-emitter-uri 'git:///tmp/data?preserve=1' \
	git@github.com:whosonfirst-data/whosonfirst-data-admin-is.git

2021/02/17 15:57:49 time to index paths (1) 3.465746865s
2021/02/17 15:57:49 Counted 436 records (indexed 436 records)
```

In this example the clone repository will be store in `/tmp/data/whosonfirst-data-admin-is.git`.

### emit

Publish features from one or more whosonfirst/go-whosonfirst-iterate sources.

```
> ./bin/emit -h
Publish features from one or more whosonfirst/go-whosonfirst-iterate/emitter sources.
Usage:
	 ./bin/emit [options] uri(N) uri(N)
Valid options are:

  -emitter-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/emitter URI. Supported emitter URI schemes are: directory://,featurecollection://,file://,filelist://,geojsonl://,git://,repo:// (default "git://")
  -geojson
    	Emit features as a well-formed GeoJSON FeatureCollection record.
  -json
    	Emit features as a well-formed JSON array.
  -null
    	Publish features to /dev/null
  -stdout
    	Publish features to STDOUT. (default true)
```

For example:

```
$> ./bin/emit \
	-geojson \
	-emitter-uri 'git://?include=properties.mz:is_current=1&include=properties.sfomuseum:placetype=gate' \
	https://github.com/sfomuseum-data/sfomuseum-data-architecture.git \

| jq '.features[]["properties"]["wof:label"]'

"C45 (2019)"
"C42A (2019)"
"C48A (2019)"
"F77 (2019)"
"F84D (2019)"
"F84C (2019)"
"F84B (2019)"
"F70A (2019)"
"F84A (2019)
...and so on
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-iterate
* https://godoc.org/gopkg.in/src-d/go-git.v4