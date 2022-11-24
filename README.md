# go-whosonfirst-tippecanoe

Go package to produce tippecanoe-compatible input using the `whosonfirst/go-whosonfirst-iterwriter` and `whosonfirst/go-writer-featurecollection` packages.

## Documenation

Documentation is incomplete at this time

## Example

```
> go run -mod vendor cmd/features/main.go \
	-writer-uri 'constant://?val=featurecollection://?writer=stdout://' \
	/usr/local/data/sfomuseum-data-whosonfirst/ \
	
	| tippecanoe -zg -o wof.mbtiles
	
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

## See also

* https://github.com/whosonfirst/go-whosonfirst-iterwriter
* https://github.com/whosonfirst/go-writer-featurecollection
* https://github.com/felt/tippecanoe