# go-whosonfirst-tippecanoe

Go package to produce tippecanoe-compatible input using the `whosonfirst/go-whosonfirst-iterwriter` and `whosonfirst/go-writer-featurecollection` packages.

## Documentation

Documentation is incomplete at this time

## Tools

```
$> make cli
go build -mod vendor -o bin/features cmd/features/main.go
```

### features

Emit features derived from a `whosonfirst/go-whosonfirst-iterator` instance as JSON-L or a GeoJSON FeatureCollection.

```
$> ./bin/features -h
  -as-spr
    	Replace Feature properties with Who's On First Standard Places Result (SPR) derived from that feature. (default true)
  -forgiving
    	Be "forgiving" of failed writes, logging the issue(s) but not triggering errors	
  -include-alt-files
    	Include alternate geometry files in output.
  -iterator-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/v3 URI. (default "repo://")
  -monitor-uri string
    	A valid sfomuseum/go-timings URI. (default "counter://PT60S")
  -require-polygons
    	Require that geometry type be 'Polygon' or 'MultiPolygon' to be included in output.
  -spr-append-property value
    	Zero or more properties in a given feature to append to SPR output
  -writer-uri value
    	One or more valid whosonfirst/go-writer/v2 URIs, each encoded as a gocloud.dev/runtimevar URI.
```

#### Local data

Generate a PMTiles database of all the records in a local checkout of the [sfomuseum-data-whosonfirst](https://github.com/sfomuseum-data/sfomuseum-data-whosonfirst) repository:

```
$> bin/features \
	-writer-uri 'constant://?val=featurecollection://?writer=stdout://' \
	/usr/local/data/sfomuseum-data-whosonfirst/ \
	
	| tippecanoe -zg -o wof.pmtiles
	
For layer 0, using name "wof"
2022/11/23 18:39:31 time to index paths (1) 22.630520632s
1586 features, 68185835 bytes of geometry, 1663363 bytes of separate metadata, 3052364 bytes of string pool
Choosing a maxzoom of -z2 for features typically 345007 feet (105158 meters) apart, and at least 29034 feet (8850 meters) apart
Choosing a maxzoom of -z9 for resolution of about 537 feet (163 meters) within features
tile 0/0/0 size is 1442119 with detail 12, >500000    
tile 0/0/0 size is 1195585 with detail 11, >500000    
tile 0/0/0 size is 980818 with detail 10, >500000    
tile 0/0/0 size is 813356 with detail 9, >500000    
tile 0/0/0 size is 673514 with detail 8, >500000    
tile 1/1/0 size is 910427 with detail 12, >500000    
tile 1/0/0 size is 763555 with detail 12, >500000    
tile 1/1/0 size is 799283 with detail 11, >500000    
tile 1/0/0 size is 619950 with detail 11, >500000    
tile 1/1/0 size is 696214 with detail 10, >500000    
tile 1/1/0 size is 589930 with detail 9, >500000    
tile 2/1/1 size is 570412 with detail 12, >500000    
tile 2/2/1 size is 724332 with detail 12, >500000    
tile 2/2/1 size is 648960 with detail 11, >500000    
tile 2/2/1 size is 584143 with detail 10, >500000    
  99.9%  9/409/233  
```

To take advantage of parallel reads in `tippecanoe` be sure to use the [whosonfirst/go-writer-jsonl](https://github.com/whosonfirst/go-writer-jsonl) writer (instead of `featurecollection`). For example:

```
$> ./bin/features \
	-writer-uri 'constant://?val=jsonl://?writer=stdout://' \
	/usr/local/data/whosonfirst-data-admin-us \

	| tippecanoe -P -zg -o us.pmtiles --drop-densest-as-needed
```

#### Remote data

Generate a PMTiles database of all the records from repositories in the [sfomuseum-data](https://github.com/sfomuseum-data) organization with a prefix of `sfomuseum-data-maps`:

```
$> ./bin/features \
	-writer-uri 'constant://?val=jsonl://?writer=stdout://' \
	-iterator-uri 'githuborg:///tmp' \
	'sfomuseum-data://?prefix=sfomuseum-data-maps' \
	
	| tippecanoe -zg -o sfomuseum.pmtiles
	
For layer 0, using name "sfomuseum"
2022/11/23 18:57:25 time to index paths (1) 2.758176132s
2022/11/23 18:57:25 time to index paths (1) 3.040810943s
38 features, 2793 bytes of geometry, 5975 bytes of separate metadata, 13994 bytes of string pool
Choosing a maxzoom of -z7 for features typically 4537 feet (1383 meters) apart, and at least 735 feet (224 meters) apart
  99.9%  7/20/49
```

#### Data for use with `whosonfirst/go-whosonfirst-spatial-pmtiles`

To generate a MBTiles database suitable for converting to a Protomaps PMTiles database for use with the [whosonfirst/go-whosonfirst-spatial-pmtiles](https://github.com/whosonfirst/go-whosonfirst-spatial-pmtiles) package it is helpful to specify the `-require-polygons` and `-as-spr` flags. These will exclude features that don't have a geometry of type "Polygon" or "MultiPolygon" and replace they keys and values in a feature's `properties` element with a [Standard Place Result](https://github.com/whosonfirst/go-whosonfirst-spr) instance, respectively. This will greatly reduce the size of the final database.

Likewise, pass the `-pf` and `-pk` flags to `tippecanoe` to ensure no features are excluded from the MBTiles database. I also like to specify a fixed zoom level; 12 is usually a good balance between the number of features in a given tile and the total number of tiles.

```
$> ./bin/features \
	-as-spr \
	-require-polygons \
	-writer-uri 'constant://?val=jsonl://?writer=stdout://' \
	/usr/local/data/whosonfirst-data-admin-us/ \

	| tippecanoe -P -z 12 -pf -pk -o us.pmtiles
```

Note: As of this writing the `go-whosonfirst-spatial-pmtiles` doesn't play well with alternate geometry features (which are included by specifying the `-include-alt-files` flag).

#### Appending addition properties to SPR output

Pass zero or more `-spr-append-property` flags to append addition properties from a feature to SPR output. Paths may be relative or absolute. For example:

```
$> bin/features \
	-as-spr \
	-require-polygons \
	-spr-append-property 'properties.src:geom' \
	-spr-append-property wof:hierarchy \
	-writer-uri 'constant://?val=jsonl://?writer=stdout://' \
	/usr/local/data/sfomuseum-data-architecture/
```

For a more complete example take a look at the [docker/build.sh](https://github.com/whosonfirst/go-whosonfirst-spatial-pmtiles/blob/main/docker/build.sh) script in the `go-whosonfirst-spatial-pmtiles` package.

#### Filtering data

You can limit features to be included in the final output by appending filtering parameters to the `-iterator-uri` paramater. For details consult the filtering documentation in the [whosonfirst/go-whosonfirst-iterator](https://github.com/whosonfirst/go-whosonfirst-iterate) package here:

* https://github.com/whosonfirst/go-whosonfirst-iterate#filters

#### Considerations when reading data from an `githuborg://` iterator

If you are fetching data from multiple GitHub repositories using the `githuborg://` iterator you may want to assign the `?_max_procs=2` parameter (as in `org:///tmp?_max_procs=2`) especially if you are running the code from a machine with lots of CPUs. By default the `go-whosonfirst-iterator` code will process as many sources are there are CPUs concurrently.

Normally this is not a problem but if you are fetching very large repositories (like `whosonfirst-data-admin-us`) and, say, seven other small repositories (on an 8-CPU machine) that will cause enough network traffic that Go will eventually start emitting "STREAM" errors and the iterator will fail. 

It seems as this behaviour can be adjusted by configuring the `http.Client` instance but I haven't figured out where, or even if it's possible, to do this in the underlying Git package.

As with all things the details will vary depending on the environent you are working in. You might be able to get away with `?_max_procs=3` or higher but the point is there is an upper limit to the number of (large) repositories that can be read over the wire in parallel.

Anecedotaly, generating a PMTiles database for all the `whosonfirst-data-admin-` repositories, at zoom level 13, on a machine with 8 cores (see notes above) takes between 12-24 hours and produces a final database that is a little over 9GB in size. At least half that time appears to be `tippecanoe` doing it's thing. Any tips or pointers on how to speed things up would be welcome but those numbers may just be "the cost of doing business" for the time being.

## See also

* https://github.com/whosonfirst/go-whosonfirst-iterwriter
* https://github.com/whosonfirst/go-whosonfirst-iterate
* https://github.com/whosonfirst/go-whosonfirst-iterate-git
* https://github.com/whosonfirst/go-writer-featurecollection
* https://github.com/whosonfirst/go-writer-jsonl
* https://github.com/felt/tippecanoe
