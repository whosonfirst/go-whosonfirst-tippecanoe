# go-whosonfirst-iterate-git

Go package implementing `whosonfirst/go-whosonfirst-iterate/v3.Iterator` functionality for Git repositories.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-whosonfirst-iterate-git.svg)](https://pkg.go.dev/github.com/whosonfirst/go-whosonfirst-iterate-git/v3)

## Example

Version 3.x of this package introduce major, backward-incompatible changes from earlier releases. That said, migragting from version 2.x to 3.x should be relatively straightforward as a the _basic_ concepts are still the same but (hopefully) simplified. Where version 2.x relied on defining a custom callback for looping over records version 3.x use Go's [iter.Seq2](https://pkg.go.dev/iter) iterator construct to yield records as they are encountered.


```
import (
	"context"
	"flag"
	"log"

	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v3"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
)

func main() {

     	var iterator_uri string

	flag.StringVar(&iterator_uri, "iterator-uri", "git:///tmp". "A registered whosonfirst/go-whosonfirst-iterate/v3.Iterator URI.")
	ctx := context.Background()
	
	iter, _:= iterate.NewIterator(ctx, iterator_uri)

	paths := flag.Args()
	
	for rec, _ := range iter.Iterate(ctx, paths...) {
		defer rec.Body.Close()
		log.Printf("Indexing %s\n", rec.Path)
	}
}
```

_Error handling removed for the sake of brevity._

### Version 2.x (the old way)

This is how you would do the same thing using the older version 2.x code:

```
import (
       "context"
       "flag"
       "io"
       "log"

       _ "github.com/whosonfirst/go-whosonfirst-iterate-github/v2"
       
       "github.com/whosonfirst/go-whosonfirst-iterate/emitter/v2"       
       "github.com/whosonfirst/go-whosonfirst-iterate/indexer/v2"
)

func main() {

	emitter_uri := flag.String("emitter-uri", "githubapi://", "A valid whosonfirst/go-whosonfirst-iterate/emitter URI")
	
     	flag.Parse()

	ctx := context.Background()

	emitter_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {
		log.Printf("Indexing %s\n", path)
		return nil
	}

	iter, _ := iterator.NewIterator(ctx, *emitter_uri, cb)

	uris := flag.Args()
	iter.IterateURIs(ctx, uris...)
}
```

_Error handling omitted for the sake of brevity._

## Iterators

This package exports the following iterators:

### git://

```
git://{PATH}?{PARAMETERS}
```

Where `{PATH}` is an optional path on disk where a repository will be clone to (default is to clone repository in memory) and `{PARAMETERS}` may be:

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| `include` | string | no | Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing. |
| `exclude` | string | no | Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing. |
| `include_mode` | string | no | A valid `aaronland/go-json-query` query mode string for testing inclusion rules. Default is "AND". |
| `exclude_mode` | string | no | A valid `aaronland/go-json-query` query mode string for testing exclusion rules. Default is "AND". |
| `preserve` | boolean | no | A boolean value indicating whether a Git repository (cloned to disk) should not be removed after processing. Default is false. |
| `depth` | int | no | An integer value indicating the number of commits to fetch. Default is 1. |

### githubapi://

```
githubapi://{GITHUB_ORGANIZATION}/{GITHUB_REPO}?{PARAMETERS}
```

Iterate through all the files in a specific GitHub respository using the GitHub API.

## Query parameters

In addition to the [default go-whosonfirst-iterate query parameters](https://github.com/whosonfirst/go-whosonfirst-iterate#query-parameters) the following query parameters are supported:

| Name | Value | Required | Notes
| --- | --- | --- | --- |
| access_token | String | Yes | A valid [GitHub API access token](https://docs.github.com/en/rest/overview/other-authentication-methods) |
| branch | String | No | The branch to use when iterating the repository contents |
| concurrent | Bool | No | If true iterate through documents concurrently. There is still a throttle on the number of API requests per second but this can speed things up significantly with the risk that you will still trigger GitHub API limits. |

_This iterator requires importing `github.com/whosonfirst/go-whosonfirst-iterate/v3/github`._

### githuborganization://

```
githuborganization://?{PARAMETERS}
```

Iterate through all the files in one or more GitHub respositories for an organization. The list of respositories is derived using the GitHub API but, as of this writing, fetching the respository files to iterate over is done using the [whosonfirst/go-whosonfirst-iterate-git/v3](https://github.com/whosonfirst/go-whosonfirst-iterate-git) package and plain-vanilla `git clone` operations rather than (or in addition to) the GitHub API.

In addition to the [default go-whosonfirst-iterate query parameters](https://github.com/whosonfirst/go-whosonfirst-iterate#query-parameters) the following query parameters are supported:

| Name | Value | Required | Notes
| --- | --- | --- | --- |
| dedupe | Boolean | No | Skip Who's On First records (IDs) which may occur in multiple repositories. |

The URIs passed to this iterator's `Iterate` method should be recognized `whosonfirst/go-whosonfirst-iterate-git URIs.

_This iterator requires importing `github.com/whosonfirst/go-whosonfirst-iterate/v3/github`._

## Filters

### QueryFilters

You can also specify inline queries by appending one or more `include` or `exclude` parameters to a `emitter.Emitter` URI, where the value is a string in the format of:

```
{PATH}={REGULAR EXPRESSION}
```

Paths follow the dot notation syntax used by the [tidwall/gjson](https://github.com/tidwall/gjson) package and regular expressions are any valid [Go language regular expression](https://golang.org/pkg/regexp/). Successful path lookups will be treated as a list of candidates and each candidate's string value will be tested against the regular expression's [MatchString](https://golang.org/pkg/regexp/#Regexp.MatchString) method.

For example:

```
repo://?include=properties.wof:placetype=region
```

You can pass multiple query parameters. For example:

```
repo://?include=properties.wof:placetype=region&include=properties.wof:name=(?i)new.*
```

The default query mode is to ensure that all queries match but you can also specify that only one or more queries need to match by appending a `include_mode` or `exclude_mode` parameter where the value is either "ANY" or "ALL".

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
Count files in one or more whosonfirst/go-whosonfirst-iterate/v3.Iterator sources.
Usage:
	 ./bin/count [options] uri(N) uri(N)
Valid options are:

  -iterator-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/v3.Iterator URI. Supported iterator URI schemes are: cwd://,directory://,featurecollection://,file://,filelist://,geojsonl://,git://,githubapi://,githuborg://,null://,repo:// (default "repo://")
  -verbose
    	Enable verbose (debug) logging.
```

For example:

```
$> ./bin/count -iterator-uri git:///tmp https://github.com/sfomuseum-data/sfomuseum-data-architecture.git
2025/06/24 05:27:43 INFO Counted records count=2019 time=7.248527917s
```

By default `go-whosonfirst-iterate-git` clones Git repositories in to memory. If your iterator URI contains a path then repositories will be cloned in that path:

By default repositories cloned in to a path are removed. If you want to preserve the cloned repository include a `?preserve=1` query parameter in your URI string:

```
$> /bin/count -iterator-uri 'git:///tmp?preserve=1' https://github.com/sfomuseum-data/sfomuseum-data-architecture.git
2025/06/24 05:29:22 INFO Counted records count=2019 time=4.728772811s

$> ls -al /tmp/sfomuseum-data-architecture.git 
total 48
drwxr-xr-x   9 asc   wheel    288 Jun 24 05:29 .
drwxrwxrwt   5 root  wheel    160 Jun 24 05:29 ..
drwxr-xr-x   8 asc   wheel    256 Jun 24 05:29 .git
-rw-r--r--   1 asc   wheel     14 Jun 24 05:29 .gitignore
drwxr-xr-x  18 asc   wheel    576 Jun 24 05:29 data
-rw-r--r--   1 asc   wheel  10462 Jun 24 05:29 LICENSE
-rw-r--r--   1 asc   wheel   1141 Jun 24 05:29 Makefile
drwxr-xr-x   3 asc   wheel     96 Jun 24 05:29 qgis
-rw-r--r--   1 asc   wheel    771 Jun 24 05:29 README.md
```

### emit

Emit records in one or more whosonfirst/go-whosonfirst-iterate/v3.Iterator sources as structured data.

```
$> ./bin/emit -h
Emit records in one or more whosonfirst/go-whosonfirst-iterate/v3.Iterator sources as structured data.
Usage:
	 ./bin/emit [options] uri(N) uri(N)
Valid options are:

  -geojson
    	Emit features as a well-formed GeoJSON FeatureCollection record.
  -iterator-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/v3.Iterator URI. Supported iterator URI schemes are: cwd://,directory://,featurecollection://,file://,filelist://,geojsonl://,git://,githubapi://,githuborg://,null://,repo:// (default "repo://")
  -json
    	Emit features as a well-formed JSON array.
  -null
    	Publish features to /dev/null
  -stdout
    	Publish features to STDOUT. (default true)
  -verbose
    	Enable verbose (debug) logging.
```

For example:

```
$> ./bin/emit \
	-geojson \
	-iterator-uri 'git://?include=properties.mz:is_current=1&include=properties.sfomuseum:placetype=gate' \
	https://github.com/sfomuseum-data/sfomuseum-data-architecture.git \

| jq '.features[]["properties"]["wof:label"]'

"B1 (2021-05-25)"
"C10R (2021-05-25)"
"A13R (2021-05-25)"
"G12S (2021-05-25)"
"G12V (2021-05-25)"
"F15M (2021-05-25)"
"C4U (2021-05-25)"
...and so on
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-iterate
* https://github.com/go-git/go-git