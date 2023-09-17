# pgcapture-benchmark

```
goos: darwin
goarch: arm64
pkg: github.com/KennyChenFight/pgcapture-benchmark
BenchmarkPostgresSink/single_insert/default_txSize_100-12         	       1	1081910375 ns/op	  194688 B/op	    3110 allocs/op
BenchmarkPostgresSink/single_insert/pipeline_txSize_100-12        	       1	1027360375 ns/op	  239280 B/op	    1963 allocs/op
BenchmarkPostgresSink/single_insert/default_txSize_500-12         	       1	1067740376 ns/op	  533384 B/op	   13908 allocs/op
BenchmarkPostgresSink/single_insert/pipeline_txSize_500-12        	       1	1099963833 ns/op	  771360 B/op	    7686 allocs/op
BenchmarkPostgresSink/single_insert/default_txSize_1000-12        	       1	2055781292 ns/op	 1016216 B/op	   27909 allocs/op
BenchmarkPostgresSink/single_insert/pipeline_txSize_1000-12       	       1	1077802500 ns/op	 1497360 B/op	   15310 allocs/op
BenchmarkPostgresSink/single_insert/default_txSize_2000-12        	       1	3077476457 ns/op	 1983704 B/op	   55915 allocs/op
BenchmarkPostgresSink/single_insert/pipeline_txSize_2000-12       	       1	1034498833 ns/op	 2939096 B/op	   30540 allocs/op
BenchmarkPostgresSink/single_insert/default_txSize_3000-12        	       1	5099181167 ns/op	 2947584 B/op	   83912 allocs/op
BenchmarkPostgresSink/single_insert/pipeline_txSize_3000-12       	       1	1080704125 ns/op	 4381872 B/op	   45762 allocs/op
BenchmarkPostgresSink/single_insert/default_txSize_4000-12        	       1	6099933084 ns/op	 3916144 B/op	  111913 allocs/op
BenchmarkPostgresSink/single_insert/pipeline_txSize_4000-12       	       1	1043409458 ns/op	 5825760 B/op	   60996 allocs/op
BenchmarkPostgresSink/single_insert/default_txSize_5000-12        	       1	7018961667 ns/op	 4881472 B/op	  139914 allocs/op
BenchmarkPostgresSink/single_insert/pipeline_txSize_5000-12       	       1	1003574459 ns/op	 7269216 B/op	   76222 allocs/op
BenchmarkPostgresSink/single_insert/default_txSize_6000-12        	       1	9000671542 ns/op	 5848824 B/op	  167919 allocs/op
BenchmarkPostgresSink/single_insert/pipeline_txSize_6000-12       	       1	2033802542 ns/op	 8714008 B/op	   91462 allocs/op
BenchmarkPostgresSink/single_insert/default_txSize_7000-12        	       1	10081351667 ns/op	 6811568 B/op	  195911 allocs/op
BenchmarkPostgresSink/single_insert/pipeline_txSize_7000-12       	       1	2011987334 ns/op	10157560 B/op	  106691 allocs/op
BenchmarkPostgresSink/single_insert/default_txSize_8000-12        	       1	11099647708 ns/op	 7781024 B/op	  223919 allocs/op
BenchmarkPostgresSink/single_insert/pipeline_txSize_8000-12       	       1	2001378625 ns/op	11602136 B/op	  121929 allocs/op
BenchmarkPostgresSink/single_insert/default_txSize_9000-12        	       1	13008063791 ns/op	 8747600 B/op	  251919 allocs/op
BenchmarkPostgresSink/single_insert/pipeline_txSize_9000-12       	       1	2071056124 ns/op	13044216 B/op	  137148 allocs/op
BenchmarkPostgresSink/single_insert/default_txSize_10000-12       	       1	14025051000 ns/op	 9712184 B/op	  279920 allocs/op
BenchmarkPostgresSink/single_insert/pipeline_txSize_10000-12      	       1	2069321083 ns/op	14487848 B/op	  152374 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/default_txSize_100-12         	       1	1045263333 ns/op	  202272 B/op	    4779 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/pipeline_txSize_100-12        	       1	1068958333 ns/op	  301632 B/op	    3393 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/default_txSize_500-12         	       1	2100533791 ns/op	  804584 B/op	   23969 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/pipeline_txSize_500-12        	       1	1016294499 ns/op	 1311888 B/op	   16696 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/default_txSize_1000-12        	       1	3064628500 ns/op	 1560568 B/op	   47972 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/pipeline_txSize_1000-12       	       1	1098673167 ns/op	 2574840 B/op	   33327 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/default_txSize_2000-12        	       1	5027738708 ns/op	 3071688 B/op	   95977 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/pipeline_txSize_2000-12       	       1	2096576333 ns/op	 5098184 B/op	   66573 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/default_txSize_3000-12        	       1	7083048625 ns/op	 4582008 B/op	  143978 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/pipeline_txSize_3000-12       	       1	2015351417 ns/op	 7621832 B/op	   99821 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/default_txSize_4000-12        	       1	10065022334 ns/op	 6092616 B/op	  191976 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/pipeline_txSize_4000-12       	       1	3067695042 ns/op	10145688 B/op	  133071 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/default_txSize_5000-12        	       1	12092050583 ns/op	 7602920 B/op	  239972 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/pipeline_txSize_5000-12       	       1	3088680417 ns/op	12669688 B/op	  166321 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/default_txSize_6000-12        	       1	14010165041 ns/op	 9112312 B/op	  287972 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/pipeline_txSize_6000-12       	       1	4007572500 ns/op	15196992 B/op	  199576 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/default_txSize_7000-12        	       1	16045347334 ns/op	10620528 B/op	  335979 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/pipeline_txSize_7000-12       	       1	4080933917 ns/op	17719216 B/op	  232823 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/default_txSize_8000-12        	       1	19071556792 ns/op	12135328 B/op	  383983 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/pipeline_txSize_8000-12       	       1	5035610667 ns/op	20244128 B/op	  266079 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/default_txSize_9000-12        	       1	21075506709 ns/op	13643024 B/op	  431983 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/pipeline_txSize_9000-12       	       1	5073299375 ns/op	22767440 B/op	  299325 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/default_txSize_10000-12       	       1	24002158291 ns/op	15153376 B/op	  479976 allocs/op
BenchmarkPostgresSink/single_insert_update_delete/pipeline_txSize_10000-12      	       1	6093729166 ns/op	25291760 B/op	  332578 allocs/op
PASS
ok  	github.com/KennyChenFight/pgcapture-benchmark	297.281s
```